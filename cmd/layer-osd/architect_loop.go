package main

import (
	"context"
	"fmt"
	"hash/crc32"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"layer-os/internal/runtime"
)

const (
	defaultArchitectLoopInterval = 15 * time.Second
	defaultArchitectPromoteLimit = 3
	architectVerificationScope   = "architect_loop"
	architectRepoStampPrefix     = "repo_stamp:"
	architectEvidenceMaxChars    = 1600
)

type architectLoopConfig struct {
	Enabled           bool
	AutoDispatch      bool
	AutoVerify        bool
	Interval          time.Duration
	PromoteLimit      int
	AutoRecoverGemini bool
	CleanupGemini     bool
	AutoRecoverCorpus bool
	CleanupCorpus     bool
}

type architectTickResult struct {
	Recovery        runtime.GeminiAbsorbResult
	CorpusRecovery  runtime.CorpusMarkdownRecoverResult
	Verification    *runtime.VerificationRecord
	VerificationRan bool
	RepairJob       *runtime.AgentJob
	RepairDispatch  *runtime.AgentDispatchResult
	RepairCreated   bool
	RepairExisting  bool
	Promotion       runtime.AgentJobPromotionResult
}

type architectLoopStatus struct {
	mu     sync.RWMutex
	config architectLoopConfig
	item   runtime.DaemonArchitectStatus
}

func loadArchitectLoopConfig() architectLoopConfig {
	return architectLoopConfig{
		Enabled:           architectLoopEnabled(),
		AutoDispatch:      architectAutoDispatchEnabled(),
		AutoVerify:        architectAutoVerifyEnabled(),
		Interval:          architectLoopInterval(),
		PromoteLimit:      architectPromoteLimit(),
		AutoRecoverGemini: architectAutoRecoverGeminiEnabled(),
		CleanupGemini:     architectCleanupGeminiEnabled(),
		AutoRecoverCorpus: architectAutoRecoverCorpusEnabled(),
		CleanupCorpus:     architectCleanupCorpusEnabled(),
	}
}

func newArchitectLoopStatus(config architectLoopConfig) *architectLoopStatus {
	return &architectLoopStatus{
		config: config,
		item: runtime.DaemonArchitectStatus{
			Enabled:      config.Enabled,
			AutoDispatch: config.AutoDispatch,
			Interval:     config.Interval.String(),
			PromoteLimit: config.PromoteLimit,
		},
	}
}

func (s *architectLoopStatus) Snapshot() *runtime.DaemonArchitectStatus {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item := s.item
	if s.item.LastRunAt != nil {
		lastRunAt := *s.item.LastRunAt
		item.LastRunAt = &lastRunAt
	}
	if s.item.LastError != nil {
		lastError := *s.item.LastError
		item.LastError = &lastError
	}
	return &item
}

func (s *architectLoopStatus) record(result architectTickResult, err error) {
	if s == nil {
		return
	}
	now := time.Now().UTC()
	considered := result.Promotion.Considered + result.Recovery.Considered + result.CorpusRecovery.Considered
	created := result.Promotion.Created + result.Recovery.Created + result.CorpusRecovery.Created
	existing := result.Promotion.Existing + result.Recovery.Existing + result.CorpusRecovery.Existing
	dispatched := result.Promotion.Dispatched
	failed := result.Promotion.Failed + result.Recovery.Failed + result.CorpusRecovery.Failed
	if result.RepairCreated {
		created++
	}
	if result.RepairExisting {
		existing++
	}
	if result.RepairDispatch != nil {
		dispatched++
	}
	if result.VerificationRan && result.Verification != nil && result.Verification.Status == "failed" {
		failed++
	}
	idle := architectTickIsIdle(result)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.item.LastRunAt = &now
	s.item.LastIdle = idle
	s.item.LastConsidered = considered
	s.item.LastCreated = created
	s.item.LastExisting = existing
	s.item.LastDispatched = dispatched
	s.item.LastFailed = failed
	if err != nil {
		message := err.Error()
		s.item.LastError = &message
		return
	}
	s.item.LastError = nil
}

func startArchitectLoop(ctx context.Context, service *runtime.Service, status *architectLoopStatus) {
	if status == nil {
		status = newArchitectLoopStatus(loadArchitectLoopConfig())
	}
	config := status.config
	if !config.Enabled {
		log.Println("architect loop: disabled")
		return
	}
	log.Printf("architect loop: started interval=%s limit=%d dispatch=%t auto_verify=%t gemini_recover=%t gemini_cleanup=%t corpus_recover=%t corpus_cleanup=%t", config.Interval, config.PromoteLimit, config.AutoDispatch, config.AutoVerify, config.AutoRecoverGemini, config.CleanupGemini, config.AutoRecoverCorpus, config.CleanupCorpus)
	runArchitectTickWithLogging(service, status, config)
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("architect loop: stopped")
			return
		case <-ticker.C:
			runArchitectTickWithLogging(service, status, config)
		}
	}
}

func runArchitectTickWithLogging(service *runtime.Service, status *architectLoopStatus, config architectLoopConfig) {
	result, err := runArchitectTick(service, config)
	if status != nil {
		status.record(result, err)
	}
	if result.Recovery.Considered > 0 || result.Recovery.Cleaned > 0 || result.Recovery.Failed > 0 {
		log.Printf("architect loop: gemini considered=%d created=%d existing=%d cleaned=%d failed=%d", result.Recovery.Considered, result.Recovery.Created, result.Recovery.Existing, result.Recovery.Cleaned, result.Recovery.Failed)
	}
	if result.CorpusRecovery.Considered > 0 || result.CorpusRecovery.Cleaned > 0 || result.CorpusRecovery.Failed > 0 {
		log.Printf("architect loop: corpus considered=%d created=%d existing=%d cleaned=%d failed=%d", result.CorpusRecovery.Considered, result.CorpusRecovery.Created, result.CorpusRecovery.Existing, result.CorpusRecovery.Cleaned, result.CorpusRecovery.Failed)
	}
	if result.VerificationRan && result.Verification != nil {
		log.Printf("architect loop: verification id=%s status=%s scope=%s", result.Verification.RecordID, result.Verification.Status, result.Verification.Scope)
	}
	if result.RepairJob != nil {
		dispatched := result.RepairDispatch != nil
		log.Printf("architect loop: repair job=%s status=%s created=%t existing=%t dispatched=%t", result.RepairJob.JobID, result.RepairJob.Status, result.RepairCreated, result.RepairExisting, dispatched)
	}
	if err != nil {
		log.Printf("architect loop: tick warning: %v", err)
	}
	if architectTickIsIdle(result) {
		log.Printf("architect loop: idle considered=%d recoveries=gemini:%d corpus:%d verify=%t", result.Promotion.Considered, result.Recovery.Considered, result.CorpusRecovery.Considered, result.VerificationRan)
		return
	}
	if result.Promotion.Created == 0 && result.Promotion.Dispatched == 0 && result.Promotion.Failed == 0 {
		return
	}
	log.Printf("architect loop: considered=%d created=%d existing=%d dispatched=%d failed=%d", result.Promotion.Considered, result.Promotion.Created, result.Promotion.Existing, result.Promotion.Dispatched, result.Promotion.Failed)
}

func architectTickIsIdle(result architectTickResult) bool {
	if result.Recovery.Considered > 0 || result.CorpusRecovery.Considered > 0 {
		return false
	}
	if result.VerificationRan || result.RepairJob != nil || result.RepairDispatch != nil || result.RepairCreated || result.RepairExisting {
		return false
	}
	if result.Promotion.Considered > 0 || result.Promotion.Created > 0 || result.Promotion.Existing > 0 || result.Promotion.Dispatched > 0 || result.Promotion.Failed > 0 {
		return false
	}
	return true
}

func runArchitectTick(service *runtime.Service, config architectLoopConfig) (architectTickResult, error) {
	result := architectTickResult{}
	var warningParts []string
	if config.AutoRecoverGemini {
		recovery, err := service.AbsorbGeminiArtifacts(0, config.CleanupGemini)
		result.Recovery = recovery
		if err != nil {
			warningParts = append(warningParts, "gemini recovery: "+err.Error())
		}
	}
	if config.AutoRecoverCorpus {
		corpusRecovery, err := service.AbsorbCorpusMarkdown(0, config.CleanupCorpus, false)
		result.CorpusRecovery = corpusRecovery
		if err != nil {
			warningParts = append(warningParts, "corpus recovery: "+err.Error())
		}
	}
	if config.AutoVerify {
		verification, ran, err := maybeRunArchitectVerification(service)
		if ran {
			result.VerificationRan = true
			copy := verification
			result.Verification = &copy
		}
		if err != nil {
			warningParts = append(warningParts, "verification: "+err.Error())
		}
		if ran && verification.Status == "failed" {
			repairJob, repairDispatch, created, existing, repairErr := ensureArchitectVerificationRepairJob(service, verification, config.AutoDispatch)
			result.RepairJob = repairJob
			result.RepairDispatch = repairDispatch
			result.RepairCreated = created
			result.RepairExisting = existing
			if repairErr != nil {
				warningParts = append(warningParts, "verification repair: "+repairErr.Error())
			}
			if len(warningParts) > 0 {
				return result, logError(strings.Join(warningParts, "; "))
			}
			return result, nil
		}
	}
	limit := config.PromoteLimit
	if limit <= 0 {
		limit = defaultArchitectPromoteLimit
	}
	promotion, err := service.PromoteContextJobs(limit, config.AutoDispatch)
	result.Promotion = promotion
	if err != nil {
		warningParts = append(warningParts, "context promotion: "+err.Error())
	}
	if len(warningParts) > 0 {
		return result, logError(strings.Join(warningParts, "; "))
	}
	return result, nil
}

func architectLoopEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_LOOP")
}

func architectAutoDispatchEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("LAYER_OS_ARCHITECT_AUTODISPATCH")))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func architectAutoVerifyEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_AUTOVERIFY")
}

func architectAutoRecoverGeminiEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_GEMINI_RECOVERY")
}

func architectCleanupGeminiEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_GEMINI_CLEANUP")
}

func architectAutoRecoverCorpusEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_CORPUS_RECOVERY")
}

func architectCleanupCorpusEnabled() bool {
	return !envBoolFalse("LAYER_OS_ARCHITECT_CORPUS_CLEANUP")
}

func architectLoopInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("LAYER_OS_ARCHITECT_INTERVAL"))
	if raw == "" {
		return defaultArchitectLoopInterval
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return defaultArchitectLoopInterval
	}
	return value
}

func architectPromoteLimit() int {
	raw := strings.TrimSpace(os.Getenv("LAYER_OS_ARCHITECT_PROMOTE_LIMIT"))
	if raw == "" {
		return defaultArchitectPromoteLimit
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultArchitectPromoteLimit
	}
	return value
}

func envBoolFalse(key string) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "0", "false", "no", "off":
		return true
	default:
		return false
	}
}

func maybeRunArchitectVerification(service *runtime.Service) (runtime.VerificationRecord, bool, error) {
	repoRoot, err := architectRepoRoot()
	if err != nil {
		return runtime.VerificationRecord{}, false, err
	}
	stamp, err := repoVerificationStamp(repoRoot)
	if err != nil {
		return runtime.VerificationRecord{}, false, err
	}
	if stamp == "" || stamp == lastArchitectVerificationStamp(service.ListVerifications()) {
		return runtime.VerificationRecord{}, false, nil
	}
	recordID := fmt.Sprintf("verify_architect_%d", time.Now().UTC().UnixNano())
	notes := []string{"architect_loop", architectRepoStampPrefix + stamp}
	command, err := runtime.DefaultGoTestVerificationCommand()
	if err != nil {
		return runtime.VerificationRecord{}, false, err
	}
	item, err := service.RunVerification(recordID, architectVerificationScope, repoRoot, command, notes)
	return item, true, err
}

func architectRepoRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT")); root != "" {
		return root, nil
	}
	return os.Getwd()
}

func repoVerificationStamp(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", nil
	}
	hasher := crc32.NewIEEE()
	count := 0
	var totalSize int64
	var latest time.Time
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != root && architectShouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !architectShouldTrackFile(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		modified := info.ModTime().UTC()
		if modified.After(latest) {
			latest = modified
		}
		count++
		totalSize += info.Size()
		_, _ = hasher.Write([]byte(filepath.ToSlash(rel)))
		_, _ = hasher.Write([]byte{'|'})
		_, _ = hasher.Write([]byte(strconv.FormatInt(modified.UnixNano(), 10)))
		_, _ = hasher.Write([]byte{'|'})
		_, _ = hasher.Write([]byte(strconv.FormatInt(info.Size(), 10)))
		_, _ = hasher.Write([]byte{'\n'})
		return nil
	})
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", nil
	}
	return fmt.Sprintf("%d-%d-%d-%08x", count, totalSize, latest.UnixNano(), hasher.Sum32()), nil
}

func architectShouldSkipDir(name string) bool {
	switch strings.TrimSpace(name) {
	case ".git", ".layer-os", ".tmp", "output":
		return true
	default:
		return false
	}
}

func architectShouldTrackFile(path string) bool {
	base := filepath.Base(path)
	if base == "go.mod" || base == "go.sum" {
		return true
	}
	return strings.EqualFold(filepath.Ext(path), ".go")
}

func lastArchitectVerificationStamp(items []runtime.VerificationRecord) string {
	for index := len(items) - 1; index >= 0; index-- {
		item := items[index]
		if strings.TrimSpace(item.Scope) != architectVerificationScope {
			continue
		}
		if value := architectRepoStampFromNotes(item.Notes); value != "" {
			return value
		}
	}
	return ""
}

func architectRepoStampFromNotes(notes []string) string {
	for _, note := range notes {
		value := strings.TrimSpace(note)
		if strings.HasPrefix(value, architectRepoStampPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(value, architectRepoStampPrefix))
		}
	}
	return ""
}

func ensureArchitectVerificationRepairJob(service *runtime.Service, verification runtime.VerificationRecord, autoDispatch bool) (*runtime.AgentJob, *runtime.AgentDispatchResult, bool, bool, error) {
	jobID := deriveArchitectVerificationJobID(verification.RecordID)
	for _, existing := range service.ListAgentJobs() {
		if existing.JobID != jobID {
			continue
		}
		job := existing
		if autoDispatch && (existing.Status == "queued" || existing.Status == "failed") {
			if !architectRepairAutoDispatchAllowed(service, existing.Role) {
				return &job, nil, false, true, nil
			}
			dispatch, err := service.DispatchAgentJob(jobID)
			if err != nil {
				return &job, nil, false, true, err
			}
			job = dispatch.Job
			return &job, &dispatch, false, true, nil
		}
		return &job, nil, false, true, nil
	}
	stamp := architectRepoStampFromNotes(verification.Notes)
	payload := map[string]any{
		"autogenerated":        true,
		"source_kind":          "architect_verification",
		"verification_id":      verification.RecordID,
		"verification_scope":   verification.Scope,
		"verification_command": append([]string{}, verification.Command...),
	}
	if stamp != "" {
		payload["repo_stamp"] = stamp
	}
	if excerpt := architectVerificationOutputExcerpt(verification.Notes); excerpt != "" {
		payload["verification_output_excerpt"] = excerpt
	}
	if verificationError := architectVerificationNoteValue(verification.Notes, "error:"); verificationError != "" {
		payload["verification_error"] = verificationError
	}
	if instructions := architectVerificationRepairInstructions(verification); instructions != "" {
		payload["instructions"] = instructions
	}
	dispatchGuidance := ""
	if autoDispatch && !architectRepairAutoDispatchAllowed(service, "implementer") {
		dispatchGuidance = "implementer direct_llm dispatch is advisory-only; keep this repair job queued for a packet-capable worker"
		payload["dispatch_guidance"] = dispatchGuidance
	}
	ref := verification.RecordID
	now := time.Now().UTC()
	notes := []string{"architect_loop", "verification_failed", "verification:" + verification.RecordID}
	if dispatchGuidance != "" {
		notes = append(notes, "dispatch_skipped:direct_llm_requires_worker")
	}
	job := runtime.AgentJob{
		JobID:     jobID,
		Kind:      "implement",
		Role:      "implementer",
		Summary:   architectVerificationRepairSummary(verification),
		Status:    "queued",
		Source:    "architect.verification_failed",
		Surface:   runtime.SurfaceAPI,
		Stage:     runtime.StageCompose,
		Ref:       &ref,
		Payload:   payload,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.CreateAgentJob(job); err != nil {
		return nil, nil, false, false, err
	}
	if autoDispatch && dispatchGuidance == "" {
		dispatch, err := service.DispatchAgentJob(jobID)
		if err != nil {
			copy := job
			return &copy, nil, true, false, err
		}
		copy := dispatch.Job
		return &copy, &dispatch, true, false, nil
	}
	copy := job
	return &copy, nil, true, false, nil
}

func deriveArchitectVerificationJobID(verificationID string) string {
	sum := crc32.ChecksumIEEE([]byte(strings.TrimSpace(verificationID)))
	return fmt.Sprintf("job_architect_verify_%08x", sum)
}

func architectVerificationRepairSummary(item runtime.VerificationRecord) string {
	scope := strings.TrimSpace(item.Scope)
	if scope == "" {
		scope = architectVerificationScope
	}
	return "Fix failing verification lane: " + scope
}

func architectRepairAutoDispatchAllowed(service *runtime.Service, role string) bool {
	if service == nil {
		return true
	}
	role = strings.TrimSpace(role)
	if role == "" {
		return true
	}
	provider := ""
	for _, profile := range service.AgentDispatchProfiles() {
		if strings.TrimSpace(profile.Role) == role {
			provider = strings.TrimSpace(profile.Provider)
			break
		}
	}
	if provider == "" {
		return true
	}
	for _, summary := range service.Providers() {
		if strings.TrimSpace(summary.Provider) != provider {
			continue
		}
		return strings.TrimSpace(summary.Semantics) != "direct_llm"
	}
	return true
}

func architectVerificationRepairInstructions(item runtime.VerificationRecord) string {
	lines := []string{
		"You are fixing a failing Layer OS verification lane in the local repository.",
		"Make the smallest code or test change needed to get the verification command passing.",
	}
	if scope := strings.TrimSpace(item.Scope); scope != "" {
		lines = append(lines, "Verification scope: "+scope)
	}
	if len(item.Command) > 0 {
		lines = append(lines, "Verification command: "+strings.Join(item.Command, " "))
	}
	if verificationError := architectVerificationNoteValue(item.Notes, "error:"); verificationError != "" {
		lines = append(lines, "Verifier error: "+verificationError)
	}
	if excerpt := architectVerificationOutputExcerpt(item.Notes); excerpt != "" {
		lines = append(lines, "Failure evidence:\n"+excerpt)
	}
	lines = append(lines,
		"After the fix, rerun the verification command and report the exact result.",
		"If you are blocked, name the failing package/test and the concrete blocker instead of asking for generic more context.",
	)
	return strings.Join(lines, "\n\n")
}

func architectVerificationOutputExcerpt(notes []string) string {
	return architectBoundedText(architectVerificationNoteValue(notes, "output:"), architectEvidenceMaxChars)
}

func architectVerificationNoteValue(notes []string, prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return ""
	}
	for _, note := range notes {
		value := strings.TrimSpace(note)
		if strings.HasPrefix(value, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(value, prefix))
		}
	}
	return ""
}

func architectBoundedText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 {
		return value
	}
	if len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit]) + "..."
}

type architectTickWarning string

func (w architectTickWarning) Error() string {
	return string(w)
}

func logError(message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}
	return architectTickWarning(message)
}

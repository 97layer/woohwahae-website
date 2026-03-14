package runtime

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AntigravityScan struct {
	Root string           `json:"root"`
	Runs []AntigravityRun `json:"runs"`
}

type AntigravityRun struct {
	RunID      string                `json:"run_id"`
	CapturedAt time.Time             `json:"captured_at"`
	SyncKey    string                `json:"sync_key"`
	Artifacts  []AntigravityArtifact `json:"artifacts"`
	MediaPaths []string              `json:"media_paths"`
}

type AntigravityArtifact struct {
	RunID        string    `json:"run_id"`
	Stem         string    `json:"stem"`
	Topic        string    `json:"topic"`
	ArtifactType string    `json:"artifact_type"`
	Summary      string    `json:"summary"`
	RawExcerpt   string    `json:"raw_excerpt"`
	Confidence   string    `json:"confidence"`
	CapturedAt   time.Time `json:"captured_at"`
	SyncKey      string    `json:"sync_key"`
	PrimaryPath  string    `json:"primary_path"`
	RelatedPaths []string  `json:"related_paths"`
	MetadataPath *string   `json:"metadata_path,omitempty"`
	Version      *string   `json:"version,omitempty"`
	Hash         string    `json:"hash"`
	Refs         []string  `json:"refs"`
	Severity     string    `json:"severity"`
}

type antigravityMetadata struct {
	ArtifactType string `json:"artifactType"`
	Summary      string `json:"summary"`
	UpdatedAt    string `json:"updatedAt"`
	Version      string `json:"version"`
}

type antigravityArtifactGroup struct {
	stem        string
	metadata    string
	candidates  []string
	related     []string
	metadataRaw antigravityMetadata
}

var antigravityRefPattern = regexp.MustCompile(`\b(proposal_[A-Za-z0-9_]+|thread_[A-Za-z0-9_]+|intel_[0-9]+|job_[A-Za-z0-9_]+|infra_[0-9]+|security_[0-9]+|approval_[A-Za-z0-9_]+|release_[A-Za-z0-9_]+|deploy_[A-Za-z0-9_]+|rollback_[A-Za-z0-9_]+)\b`)

func DefaultAntigravityBrainRoot() string {
	if value := strings.TrimSpace(os.Getenv("LAYER_OS_ANTIGRAVITY_BRAIN_ROOT")); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gemini", "antigravity", "brain")
}

func ScanAntigravityRuns(root string, runID string, limit int) (AntigravityScan, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = DefaultAntigravityBrainRoot()
	}
	if root == "" {
		return AntigravityScan{}, errors.New("antigravity root is required")
	}
	info, err := os.Stat(root)
	if err != nil {
		return AntigravityScan{}, err
	}
	if !info.IsDir() {
		return AntigravityScan{}, errors.New("antigravity root must be a directory")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return AntigravityScan{}, err
	}
	runs := make([]AntigravityRun, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		if runID != "" && name != strings.TrimSpace(runID) {
			continue
		}
		run, err := scanAntigravityRun(root, name)
		if err != nil {
			return AntigravityScan{}, err
		}
		if len(run.Artifacts) == 0 && len(run.MediaPaths) == 0 {
			continue
		}
		runs = append(runs, run)
	}
	sort.SliceStable(runs, func(i, j int) bool {
		if runs[i].CapturedAt.Equal(runs[j].CapturedAt) {
			return runs[i].RunID > runs[j].RunID
		}
		return runs[i].CapturedAt.After(runs[j].CapturedAt)
	})
	if limit > 0 && len(runs) > limit {
		runs = runs[:limit]
	}
	return AntigravityScan{Root: root, Runs: runs}, nil
}

func AntigravityObservation(item AntigravityArtifact) ObservationRecord {
	refs := append([]string{}, item.Refs...)
	for _, ref := range ObservationMetadataRefs("antigravity", item.Topic, map[string]string{
		"artifact_type":  item.ArtifactType,
		"artifact_stem":  item.Stem,
		"ingest_adapter": "antigravity",
		"severity":       item.Severity,
	}) {
		refs = appendUniqueString(refs, ref)
	}
	refs = appendUniqueString(refs, "antigravity_sync:"+strings.TrimSpace(item.SyncKey))
	refs = appendUniqueString(refs, "antigravity_run:"+strings.TrimSpace(item.RunID))
	return ObservationRecord{
		ObservationID:     antigravityObservationID(item.SyncKey),
		SourceChannel:     "antigravity",
		CapturedAt:        item.CapturedAt,
		Actor:             "antigravity",
		Topic:             item.Topic,
		Refs:              refs,
		Confidence:        item.Confidence,
		RawExcerpt:        item.RawExcerpt,
		NormalizedSummary: item.Summary,
	}
}

func AntigravityRunPrimaryArtifact(run AntigravityRun) (AntigravityArtifact, bool) {
	if len(run.Artifacts) == 0 {
		return AntigravityArtifact{}, false
	}
	items := append([]AntigravityArtifact{}, run.Artifacts...)
	sort.SliceStable(items, func(i, j int) bool {
		ri := antigravityArtifactPriority(items[i])
		rj := antigravityArtifactPriority(items[j])
		if ri != rj {
			return ri > rj
		}
		if items[i].CapturedAt.Equal(items[j].CapturedAt) {
			return items[i].PrimaryPath < items[j].PrimaryPath
		}
		return items[i].CapturedAt.After(items[j].CapturedAt)
	})
	return items[0], true
}

func scanAntigravityRun(root string, runID string) (AntigravityRun, error) {
	runRoot := filepath.Join(root, runID)
	groups := map[string]*antigravityArtifactGroup{}
	mediaPaths := []string{}

	err := filepath.WalkDir(runRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(runRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		base := filepath.Base(rel)
		if strings.HasPrefix(base, "media__") {
			mediaPaths = append(mediaPaths, path)
			return nil
		}
		stem := normalizeAntigravityStem(rel)
		if stem == "" {
			return nil
		}
		group := groups[stem]
		if group == nil {
			group = &antigravityArtifactGroup{stem: stem}
			groups[stem] = group
		}
		group.related = append(group.related, path)
		if strings.HasSuffix(rel, ".metadata.json") {
			group.metadata = path
			metadata, err := loadAntigravityMetadata(path)
			if err == nil {
				group.metadataRaw = metadata
			}
			return nil
		}
		group.candidates = append(group.candidates, path)
		return nil
	})
	if err != nil {
		return AntigravityRun{}, err
	}

	artifacts := make([]AntigravityArtifact, 0, len(groups))
	latest := time.Time{}
	for _, group := range groups {
		artifact, ok := buildAntigravityArtifact(runID, runRoot, group)
		if !ok {
			continue
		}
		artifacts = append(artifacts, artifact)
		if latest.Before(artifact.CapturedAt) {
			latest = artifact.CapturedAt
		}
	}
	if latest.IsZero() {
		for _, mediaPath := range mediaPaths {
			info, err := os.Stat(mediaPath)
			if err != nil {
				continue
			}
			if latest.Before(info.ModTime().UTC()) {
				latest = info.ModTime().UTC()
			}
		}
	}
	sort.Strings(mediaPaths)
	sort.SliceStable(artifacts, func(i, j int) bool {
		if artifacts[i].CapturedAt.Equal(artifacts[j].CapturedAt) {
			return artifacts[i].PrimaryPath < artifacts[j].PrimaryPath
		}
		return artifacts[i].CapturedAt.After(artifacts[j].CapturedAt)
	})
	runSyncParts := []string{runID}
	for _, item := range artifacts {
		runSyncParts = append(runSyncParts, item.SyncKey)
	}
	for _, mediaPath := range mediaPaths {
		runSyncParts = append(runSyncParts, filepath.ToSlash(mediaPath))
	}
	return AntigravityRun{
		RunID:      runID,
		CapturedAt: latest,
		SyncKey:    antigravityHash(runSyncParts...),
		Artifacts:  artifacts,
		MediaPaths: mediaPaths,
	}, nil
}

func buildAntigravityArtifact(runID string, runRoot string, group *antigravityArtifactGroup) (AntigravityArtifact, bool) {
	primary := selectAntigravityPrimary(group.stem, group.candidates)
	if primary == "" {
		return AntigravityArtifact{}, false
	}
	raw, err := os.ReadFile(primary)
	if err != nil {
		return AntigravityArtifact{}, false
	}
	info, err := os.Stat(primary)
	if err != nil {
		return AntigravityArtifact{}, false
	}
	capturedAt := info.ModTime().UTC()
	summary := strings.TrimSpace(group.metadataRaw.Summary)
	if parsed, ok := parseAntigravityTime(group.metadataRaw.UpdatedAt); ok {
		capturedAt = parsed
	}
	content := string(raw)
	if summary == "" {
		summary = antigravitySummaryFromContent(content, group.stem)
	}
	topic := antigravityTopic(group.stem, group.metadataRaw.ArtifactType, summary)
	relatedPaths := append([]string{}, group.related...)
	sort.Strings(relatedPaths)
	metadataPath := optionalAntigravityString(group.metadata)
	version := optionalAntigravityString(strings.TrimSpace(group.metadataRaw.Version))
	relPrimary, err := filepath.Rel(runRoot, primary)
	if err == nil {
		relPrimary = filepath.ToSlash(relPrimary)
	} else {
		relPrimary = filepath.ToSlash(primary)
	}
	hash := antigravityHash(runID, group.stem, string(raw))
	syncParts := []string{runID, group.stem, hash}
	if version != nil {
		syncParts = append(syncParts, *version)
	}
	if !capturedAt.IsZero() {
		syncParts = append(syncParts, capturedAt.Format(time.RFC3339Nano))
	}
	severity := antigravitySeverity(summary + "\n" + content)
	refs := []string{
		"antigravity_run:" + runID,
		"antigravity_artifact:" + group.stem,
		"antigravity_path:" + relPrimary,
	}
	for _, ref := range ObservationMetadataRefs("antigravity", topic, map[string]string{
		"artifact_type":  strings.TrimSpace(group.metadataRaw.ArtifactType),
		"artifact_stem":  group.stem,
		"ingest_adapter": "antigravity",
		"severity":       severity,
	}) {
		refs = appendUniqueString(refs, ref)
	}
	for _, match := range antigravityRefPattern.FindAllString(content+"\n"+summary, -1) {
		refs = appendUniqueString(refs, strings.TrimSpace(match))
	}
	return AntigravityArtifact{
		RunID:        runID,
		Stem:         group.stem,
		Topic:        topic,
		ArtifactType: strings.TrimSpace(group.metadataRaw.ArtifactType),
		Summary:      limitText(summary, 180),
		RawExcerpt:   antigravityExcerpt(content),
		Confidence:   antigravityConfidence(group.metadata != "", primary != "" && strings.Contains(filepath.Base(primary), ".resolved")),
		CapturedAt:   capturedAt,
		SyncKey:      antigravityHash(syncParts...),
		PrimaryPath:  primary,
		RelatedPaths: relatedPaths,
		MetadataPath: metadataPath,
		Version:      version,
		Hash:         hash,
		Refs:         refs,
		Severity:     severity,
	}, true
}

func loadAntigravityMetadata(path string) (antigravityMetadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return antigravityMetadata{}, err
	}
	var metadata antigravityMetadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return antigravityMetadata{}, err
	}
	return metadata, nil
}

func normalizeAntigravityStem(rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return ""
	}
	if strings.HasSuffix(rel, ".metadata.json") {
		return strings.TrimSuffix(rel, ".metadata.json")
	}
	if index := strings.Index(rel, ".resolved"); index >= 0 {
		return rel[:index]
	}
	return rel
}

func selectAntigravityPrimary(stem string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	type scored struct {
		path   string
		rank   int
		serial int
	}
	items := make([]scored, 0, len(candidates))
	for _, candidate := range candidates {
		base := filepath.ToSlash(filepath.Base(candidate))
		stemBase := filepath.Base(stem)
		rank := 0
		serial := 0
		switch {
		case base == stemBase+".resolved":
			rank = 3
		case strings.HasPrefix(base, stemBase+".resolved."):
			rank = 2
			serial = antigravityResolvedSerial(base)
		case base == stemBase:
			rank = 1
		default:
			rank = 0
		}
		items = append(items, scored{path: candidate, rank: rank, serial: serial})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].rank != items[j].rank {
			return items[i].rank > items[j].rank
		}
		if items[i].serial != items[j].serial {
			return items[i].serial > items[j].serial
		}
		return items[i].path < items[j].path
	})
	return items[0].path
}

func antigravityResolvedSerial(base string) int {
	index := strings.LastIndex(base, ".resolved.")
	if index < 0 {
		return 0
	}
	value := strings.TrimSpace(base[index+len(".resolved."):])
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func antigravitySummaryFromContent(content string, fallback string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == "" {
			continue
		}
		return trimmed
	}
	return strings.TrimSpace(fallback)
}

func antigravityExcerpt(content string) string {
	lines := []string{}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
		if len(lines) >= 6 {
			break
		}
	}
	if len(lines) == 0 {
		return "(empty artifact)"
	}
	return limitText(strings.Join(lines, " | "), 600)
}

func antigravityConfidence(hasMetadata bool, hasResolved bool) string {
	if hasMetadata && hasResolved {
		return "high"
	}
	if hasMetadata || hasResolved {
		return "medium"
	}
	return "low"
}

func antigravityTopic(stem string, artifactType string, summary string) string {
	joined := strings.ToLower(strings.TrimSpace(artifactType + " " + stem + " " + summary))
	switch {
	case strings.Contains(joined, "walkthrough"):
		return "walkthrough"
	case strings.Contains(joined, "implementation_plan") || strings.Contains(joined, "implementation plan"):
		return "implementation_plan"
	case strings.Contains(joined, "diagnosis") || strings.Contains(joined, "audit"):
		return "diagnosis"
	case strings.Contains(joined, "task"):
		return "task"
	default:
		return "artifact"
	}
}

func antigravitySeverity(text string) string {
	text = strings.ToLower(text)
	switch {
	case strings.Contains(text, "critical"), strings.Contains(text, "🔴"), strings.Contains(text, "high severity"), strings.Contains(text, "action required"):
		return "high"
	case strings.Contains(text, "warning"), strings.Contains(text, "⚠️"), strings.Contains(text, "drift"):
		return "medium"
	default:
		return "low"
	}
}

func antigravityObservationID(syncKey string) string {
	return "observation_antigravity_" + antigravityHash(syncKey)
}

func antigravityHash(parts ...string) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(strings.TrimSpace(part)))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

func optionalAntigravityString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func parseAntigravityTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func antigravityArtifactPriority(item AntigravityArtifact) int {
	switch item.Topic {
	case "diagnosis":
		return 4
	case "walkthrough":
		return 3
	case "implementation_plan":
		return 2
	case "task":
		return 1
	default:
		return 0
	}
}

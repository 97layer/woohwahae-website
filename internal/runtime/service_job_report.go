package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (AgentJobReportResult, error) {
	if !validAgentJobTerminalStatus(status) {
		return AgentJobReportResult{}, errors.New("job report status must be terminal")
	}
	s.mu.Lock()
	current, ok := s.agentJob.get(jobID)
	if !ok {
		s.mu.Unlock()
		return AgentJobReportResult{}, errors.New("job_id not found")
	}
	normalizedResult, err := normalizeAgentJobReportResult(current, status, result, s.currentActor())
	if err != nil {
		s.mu.Unlock()
		return AgentJobReportResult{}, err
	}
	updated, event, err := s.updateAgentJobLocked(jobID, status, notes, normalizedResult, "agent_job."+status)
	if err != nil {
		s.mu.Unlock()
		return AgentJobReportResult{}, err
	}
	entry := newAgentJobCapitalizationEntry(updated, event, SummarizeReviewRoom(s.currentReviewRoomLocked()), s.currentActor())
	report := AgentJobReportResult{Job: updated, Event: event, Warnings: []string{}}
	if err := s.disk.appendCapitalizationEntry(entry); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("capitalization append failed: %v", err))
		warning := newSignalReviewRoomItem("`job.report` 이후 에이전트 작업 자본화 기록 추가가 실패했어. 이후 자동화를 믿기 전에 지식 코퍼스 경로를 확인해야 해.", "capitalization.failed", nil, "agent job capitalization append failure requires founder review before automation depends on the corpus", "review_room.auto.agent_job_capitalization_failed", []string{"entry:" + entry.EntryID, "job:" + updated.JobID, "error:" + err.Error()})
		if openErr := s.autoOpenReviewRoomItemLocked(warning); openErr == nil {
			_ = s.persistLocked()
		}
	} else {
		report.Capitalization = &entry
	}
	if err := s.appendAutoObservationLocked(agentJobReportObservation(updated, event, s.currentActor())); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("observation auto-ingest failed: %v", err))
		warning := newSignalReviewRoomItem("`job.report` 이후 observation 자동 적재가 실패했어. 크로스채널 합성을 믿기 전에 observation 레지스트리를 확인해야 해.", "observation.failed", nil, "observation append failure requires founder review before automation depends on the observation registry", "review_room.auto.agent_job_observation_failed", []string{"event:" + event.EventID, "job:" + updated.JobID, "error:" + err.Error()})
		if openErr := s.autoOpenReviewRoomItemLocked(warning); openErr == nil {
			_ = s.persistLocked()
		}
	}
	s.mu.Unlock()
	if chain, err := s.ApplyAgentJobChain(updated); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("agent job chain failed: %v", err))
		s.recordAutomationWarning(agentJobChainFailureReviewItem(updated, err))
	} else if chain.Considered > 0 || chain.Created > 0 || chain.Existing > 0 || chain.Dispatched > 0 || chain.TelegramRequested > 0 || len(chain.CreatedJobIDs) > 0 || len(chain.ExistingJobIDs) > 0 {
		report.Chain = &chain
	}
	if len(report.Warnings) == 0 {
		report.Warnings = nil
	}
	report.FollowUp = buildAgentJobReportFollowUp(updated, report.Warnings, report.Chain)
	s.mu.Lock()
	persisted, err := s.applyAgentJobResultPatchLocked(updated.JobID, map[string]any{
		"follow_up": agentJobFollowUpPayload(report.FollowUp),
	})
	s.mu.Unlock()
	if err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("follow-up persist failed: %v", err))
	} else {
		report.Job = persisted
	}
	if len(report.Warnings) == 0 {
		report.Warnings = nil
	}
	s.maybeNotifyFounderAttention()
	return report, nil
}

func buildAgentJobReportFollowUp(job AgentJob, warnings []string, chain *AgentJobChainResult) AgentJobFollowUp {
	jobIDs := collectAgentJobFollowUpIDs(chain)
	switch {
	case len(warnings) > 0:
		return AgentJobFollowUp{
			Mode:    "review_warnings",
			Summary: "job.report 경고부터 확인한 뒤 다음 자동화를 믿어야 해.",
			JobIDs:  jobIDs,
		}
	case chain != nil && chain.Dispatched > 0:
		return AgentJobFollowUp{
			Mode:    "monitor_dispatched_jobs",
			Summary: fmt.Sprintf("후속 작업 %d건이 자동 투입됐어. 새 수동 작업을 만들기 전에 그 레인부터 확인해야 해.", chain.Dispatched),
			JobIDs:  jobIDs,
		}
	case chain != nil && (chain.Created > 0 || chain.Existing > 0 || len(jobIDs) > 0):
		return AgentJobFollowUp{
			Mode:    "review_follow_up_jobs",
			Summary: "후속 작업은 준비됐지만 아직 자동 투입되진 않았어. 다음으로 그 레인을 검토하거나 투입해야 해.",
			JobIDs:  jobIDs,
		}
	case strings.TrimSpace(job.Status) == "failed":
		return AgentJobFollowUp{
			Mode:    "inspect_failure",
			Summary: "재시도하거나 대체 작업을 올리기 전에 실패 레인 증거부터 확인해야 해.",
		}
	case strings.TrimSpace(job.Status) == "canceled":
		return AgentJobFollowUp{
			Mode:    "decide_replacement",
			Summary: "취소된 레인을 다시 열지, 새 bounded 작업으로 바꿀지 먼저 결정해야 해.",
		}
	default:
		return AgentJobFollowUp{
			Mode:    "continue_loop",
			Summary: "자동 후속 작업은 안 열렸어. founder 검토나 `layer-osctl next`로 다음 레인을 잡아야 해.",
		}
	}
}

func collectAgentJobFollowUpIDs(chain *AgentJobChainResult) []string {
	if chain == nil {
		return nil
	}
	items := make([]string, 0, len(chain.CreatedJobIDs)+len(chain.ExistingJobIDs))
	seen := map[string]bool{}
	for _, jobID := range append(append([]string{}, chain.CreatedJobIDs...), chain.ExistingJobIDs...) {
		jobID = strings.TrimSpace(jobID)
		if jobID == "" || seen[jobID] {
			continue
		}
		seen[jobID] = true
		items = append(items, jobID)
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func agentJobFollowUpPayload(item AgentJobFollowUp) map[string]any {
	payload := map[string]any{
		"mode":    strings.TrimSpace(item.Mode),
		"summary": strings.TrimSpace(item.Summary),
	}
	if len(item.JobIDs) > 0 {
		payload["job_ids"] = append([]string{}, item.JobIDs...)
	}
	return payload
}

func normalizeAgentJobReportResult(job AgentJob, status string, result map[string]any, actor string) (map[string]any, error) {
	merged := cloneJSONObject(job.Result)
	if merged == nil {
		merged = map[string]any{}
	}
	for key, value := range cloneJSONObject(result) {
		merged[key] = value
	}

	summary := strings.TrimSpace(agentJobResultString(merged["summary"]))
	if summary == "" {
		summary = defaultAgentJobReportSummary(job, status)
	}
	merged["summary"] = summary
	if _, ok := merged["artifacts"]; !ok {
		merged["artifacts"] = []string{}
	}
	merged["report_status"] = status
	merged["reported_by_actor"] = normalizeActor(actor)
	merged["report_transport"] = "job_report"
	if _, ok := merged["completion_mode"]; !ok {
		merged["completion_mode"] = "explicit_job_report"
	}
	if status == "succeeded" {
		if err := validateAgentJobSucceededEvidence(job, merged); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func defaultAgentJobReportSummary(job AgentJob, status string) string {
	switch status {
	case "succeeded":
		return fmt.Sprintf("Agent job %s succeeded: %s", job.JobID, strings.TrimSpace(job.Summary))
	case "failed":
		return fmt.Sprintf("Agent job %s failed: %s", job.JobID, strings.TrimSpace(job.Summary))
	case "canceled":
		return fmt.Sprintf("Agent job %s canceled: %s", job.JobID, strings.TrimSpace(job.Summary))
	default:
		return strings.TrimSpace(job.Summary)
	}
}

func validateAgentJobSucceededEvidence(job AgentJob, result map[string]any) error {
	if strings.TrimSpace(agentJobResultString(result["summary"])) == "" {
		return errors.New("job report succeeded requires result.summary")
	}
	if agentJobRoleNeedsArtifacts(job.Role) && !agentJobResultHasArtifacts(result) {
		return fmt.Errorf("job report succeeded for role %s requires non-empty result.artifacts or changed path evidence", strings.TrimSpace(job.Role))
	}
	if err := validateAgentJobPathEvidence(job, result); err != nil {
		return err
	}
	return nil
}

func agentJobRoleNeedsArtifacts(role string) bool {
	switch strings.TrimSpace(role) {
	case "implementer", "verifier":
		return true
	default:
		return false
	}
}

func agentJobResultHasArtifacts(result map[string]any) bool {
	if result == nil {
		return false
	}
	for _, key := range []string{"artifacts", "changed_paths", "touched_paths", "artifact_relative_paths"} {
		if agentJobResultArrayLen(result[key]) > 0 {
			return true
		}
	}
	for _, key := range []string{"artifact_relative_path", "artifact_path", "verification_record_id"} {
		if strings.TrimSpace(agentJobResultString(result[key])) != "" {
			return true
		}
	}
	return false
}

func agentJobResultArrayLen(value any) int {
	switch typed := value.(type) {
	case []string:
		count := 0
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				count++
			}
		}
		return count
	case []any:
		count := 0
		for _, item := range typed {
			if strings.TrimSpace(agentJobResultString(item)) != "" {
				count++
			}
		}
		return count
	default:
		return 0
	}
}

func agentJobResultString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

func validateAgentJobPathEvidence(job AgentJob, result map[string]any) error {
	repoRoot := repoRootFromEnv()
	allowed := payloadStringList(job.Payload, "allowed_paths")
	blocked := payloadStringList(job.Payload, "blocked_paths")

	for _, ref := range collectAgentJobPathEvidence(result) {
		if ref.Kind == "repo" {
			if len(allowed) > 0 && !pathWithinAnyPrefix(ref.Path, allowed) {
				return fmt.Errorf("job report succeeded path %q is outside allowed_paths", ref.Raw)
			}
			if pathWithinAnyPrefix(ref.Path, blocked) {
				return fmt.Errorf("job report succeeded path %q is inside blocked_paths", ref.Raw)
			}
			full := filepath.Join(repoRoot, filepath.FromSlash(ref.Path))
			if _, err := os.Stat(full); err != nil {
				return fmt.Errorf("job report succeeded path %q does not exist under repo root", ref.Raw)
			}
			continue
		}
		if ref.Kind == "abs" {
			if len(allowed) > 0 && !absolutePathWithinAllowed(ref.Path, repoRoot, allowed) {
				return fmt.Errorf("job report succeeded absolute path %q is outside allowed_paths", ref.Raw)
			}
			if absolutePathMatchesBlocked(ref.Path, repoRoot, blocked) {
				return fmt.Errorf("job report succeeded absolute path %q is inside blocked_paths", ref.Raw)
			}
			if _, err := os.Stat(ref.Path); err != nil {
				return fmt.Errorf("job report succeeded absolute path %q does not exist", ref.Raw)
			}
		}
	}
	return nil
}

type agentJobPathEvidence struct {
	Raw  string
	Path string
	Kind string
}

func collectAgentJobPathEvidence(result map[string]any) []agentJobPathEvidence {
	if result == nil {
		return nil
	}
	items := []agentJobPathEvidence{}
	seen := map[string]bool{}
	add := func(raw string, repoScoped bool) {
		ref, ok := classifyAgentJobPathEvidence(raw, repoScoped)
		if !ok {
			return
		}
		key := ref.Kind + "::" + ref.Path
		if seen[key] {
			return
		}
		seen[key] = true
		items = append(items, ref)
	}

	for _, key := range []string{"changed_paths", "touched_paths", "artifact_relative_paths"} {
		for _, raw := range resultStringList(result[key]) {
			add(raw, true)
		}
	}
	for _, key := range []string{"artifact_relative_path"} {
		add(agentJobResultString(result[key]), true)
	}
	for _, key := range []string{"artifact_path"} {
		add(agentJobResultString(result[key]), false)
	}
	return items
}

func classifyAgentJobPathEvidence(raw string, repoScoped bool) (agentJobPathEvidence, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return agentJobPathEvidence{}, false
	}
	if strings.HasPrefix(raw, "gateway_call:") || strings.HasPrefix(raw, "verification:") || strings.HasPrefix(raw, "event:") || strings.HasPrefix(raw, "job:") {
		return agentJobPathEvidence{}, false
	}
	if filepath.IsAbs(raw) {
		return agentJobPathEvidence{Raw: raw, Path: filepath.Clean(raw), Kind: "abs"}, true
	}
	if repoScoped {
		clean := filepath.ToSlash(filepath.Clean(raw))
		if clean == "." || strings.HasPrefix(clean, "../") {
			return agentJobPathEvidence{}, false
		}
		return agentJobPathEvidence{Raw: raw, Path: clean, Kind: "repo"}, true
	}
	clean := filepath.ToSlash(filepath.Clean(raw))
	if strings.HasPrefix(clean, "./") {
		clean = strings.TrimPrefix(clean, "./")
	}
	if clean == "." || strings.HasPrefix(clean, "../") {
		return agentJobPathEvidence{}, false
	}
	if strings.Contains(clean, "/") {
		return agentJobPathEvidence{Raw: raw, Path: clean, Kind: "repo"}, true
	}
	return agentJobPathEvidence{}, false
}

func payloadStringList(payload map[string]any, key string) []string {
	if payload == nil {
		return nil
	}
	return resultStringList(payload[key])
}

func resultStringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if trimmed := strings.TrimSpace(agentJobResultString(item)); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	default:
		return nil
	}
}

func pathWithinAnyPrefix(path string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return false
	}
	normalizedPath := filepath.ToSlash(filepath.Clean(path))
	for _, prefix := range prefixes {
		normalizedPrefix := normalizePathPrefix(prefix)
		if normalizedPrefix == "" {
			continue
		}
		if normalizedPath == strings.TrimSuffix(normalizedPrefix, "/") || strings.HasPrefix(normalizedPath, normalizedPrefix) {
			return true
		}
	}
	return false
}

func normalizePathPrefix(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	if clean == "." {
		return ""
	}
	if !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean
}

func absolutePathWithinAllowed(path string, repoRoot string, allowed []string) bool {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return pathWithinAnyPrefix(filepath.ToSlash(rel), allowed)
}

func absolutePathMatchesBlocked(path string, repoRoot string, blocked []string) bool {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return pathWithinAnyPrefix(filepath.ToSlash(rel), blocked)
}

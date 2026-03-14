package runtime

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strings"
)

type AgentJobChainResult struct {
	ParentJobID       string   `json:"parent_job_id"`
	Considered        int      `json:"considered"`
	Created           int      `json:"created"`
	Existing          int      `json:"existing"`
	Dispatched        int      `json:"dispatched"`
	TelegramRequested int      `json:"telegram_requested"`
	CreatedJobIDs     []string `json:"created_job_ids"`
	ExistingJobIDs    []string `json:"existing_job_ids"`
}

func chainRulesFromPayload(payload map[string]any) (ChainRules, bool, error) {
	if payload == nil {
		return ChainRules{}, false, nil
	}
	raw, ok := payload["chain_rules"]
	if !ok || raw == nil {
		return ChainRules{}, false, nil
	}
	body, err := json.Marshal(raw)
	if err != nil {
		return ChainRules{}, true, err
	}
	var rules ChainRules
	if err := json.Unmarshal(body, &rules); err != nil {
		return ChainRules{}, true, err
	}
	if err := rules.Validate(); err != nil {
		return ChainRules{}, true, err
	}
	return rules, true, nil
}

func deriveChainedJobID(parentJobID string, rule ChainRule) string {
	parts := []string{
		strings.TrimSpace(parentJobID),
		strings.TrimSpace(rule.RuleID),
		strings.TrimSpace(rule.OnStatus),
		strings.TrimSpace(rule.NextKind),
		strings.TrimSpace(rule.NextRole),
		strings.TrimSpace(rule.Summary),
	}
	return fmt.Sprintf("job_chain_%08x", crc32.ChecksumIEEE([]byte(strings.Join(parts, "|"))))
}

func chainedAgentJob(parent AgentJob, rule ChainRule) AgentJob {
	now := zeroSafeNow()
	payload := cloneJSONObject(rule.Payload)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["chain_parent_job_id"] = parent.JobID
	payload["chain_parent_role"] = parent.Role
	payload["chain_parent_status"] = parent.Status
	payload["chain_rule_id"] = rule.RuleID
	payload["chain_parent_summary"] = parent.Summary
	if parent.Ref != nil {
		payload["chain_parent_ref"] = strings.TrimSpace(*parent.Ref)
	}
	if parent.Result != nil {
		payload["chain_parent_result"] = cloneJSONObject(parent.Result)
	}
	if rule.NotifyTelegram {
		payload["notify_telegram"] = true
	}
	notes := []string{"chain:child", "chain_parent:" + parent.JobID, "chain_rule:" + rule.RuleID}
	for _, note := range rule.Notes {
		notes = appendUniqueString(notes, note)
	}
	surface := rule.Surface
	if !validSurface(surface) {
		surface = parent.Surface
	}
	stage := rule.NextStage
	if !validStage(stage) {
		stage = parent.Stage
	}
	return AgentJob{
		JobID:     deriveChainedJobID(parent.JobID, rule),
		Kind:      strings.TrimSpace(rule.NextKind),
		Role:      strings.TrimSpace(rule.NextRole),
		Summary:   strings.TrimSpace(rule.Summary),
		Status:    "queued",
		Source:    "agent_job.chain",
		Surface:   surface,
		Stage:     stage,
		Ref:       normalizeRef(parent.Ref),
		Payload:   payload,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *Service) ApplyAgentJobChain(parent AgentJob) (AgentJobChainResult, error) {
	rules, ok, err := chainRulesFromPayload(parent.Payload)
	if err != nil {
		return AgentJobChainResult{ParentJobID: parent.JobID}, err
	}
	if !ok {
		return AgentJobChainResult{ParentJobID: parent.JobID}, nil
	}
	result := AgentJobChainResult{ParentJobID: parent.JobID, CreatedJobIDs: []string{}, ExistingJobIDs: []string{}}
	for _, rule := range rules.Rules {
		if strings.TrimSpace(rule.OnStatus) != strings.TrimSpace(parent.Status) {
			continue
		}
		result.Considered++
		child := chainedAgentJob(parent, rule)
		if err := s.CreateAgentJob(child); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "job_id already exists") {
				return result, fmt.Errorf("chain rule %s create job: %w", rule.RuleID, err)
			}
			result.Existing++
			result.ExistingJobIDs = append(result.ExistingJobIDs, child.JobID)
			if rule.AutoDispatch {
				existing, exists := findAgentJobByID(s.ListAgentJobs(), child.JobID)
				if exists && (existing.Status == "queued" || existing.Status == "failed") {
					if _, err := s.DispatchAgentJob(child.JobID); err != nil {
						return result, fmt.Errorf("chain rule %s dispatch existing job: %w", rule.RuleID, err)
					}
					result.Dispatched++
				}
			}
			continue
		}
		result.Created++
		result.CreatedJobIDs = append(result.CreatedJobIDs, child.JobID)
		if rule.NotifyTelegram {
			result.TelegramRequested++
		}
		if rule.AutoDispatch {
			if _, err := s.DispatchAgentJob(child.JobID); err != nil {
				return result, fmt.Errorf("chain rule %s dispatch child job: %w", rule.RuleID, err)
			}
			result.Dispatched++
		}
	}
	return result, nil
}

func (s *Service) recordAutomationWarning(item ReviewRoomItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.autoOpenReviewRoomItemLocked(item); err == nil {
		_ = s.persistLocked()
	}
}

func agentJobChainFailureReviewItem(job AgentJob, err error) ReviewRoomItem {
	ref := job.JobID
	evidence := []string{"job:" + job.JobID, "status:" + job.Status}
	if err != nil {
		evidence = append(evidence, "error:"+strings.TrimSpace(err.Error()))
	}
	return newSignalReviewRoomItem(
		"`job.report` 이후 에이전트 작업 체인이 `"+job.JobID+"`에서 실패했어. 이후 자동화를 믿기 전에 체인 규칙을 확인해야 해.",
		"agent_job.chain_failed",
		&ref,
		"post-report chain failure requires founder review before automated follow-up can be trusted",
		"review_room.auto.agent_job_chain_failed",
		evidence,
	)
}

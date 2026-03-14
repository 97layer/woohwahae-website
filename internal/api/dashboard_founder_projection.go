package api

import (
	"strconv"
	"strings"

	"layer-os/internal/runtime"
)

func summarizeOpenJobs(items []runtime.AgentJob) ([]runtime.AgentJob, cockpitJobCounts) {
	openJobs := make([]runtime.AgentJob, 0, len(items))
	counts := cockpitJobCounts{}
	for _, item := range items {
		switch item.Status {
		case "queued":
			counts.Open++
			counts.Queued++
			openJobs = append(openJobs, item)
		case "running", "failed":
			counts.Open++
			counts.Running++
			openJobs = append(openJobs, item)
		default:
			counts.Terminal++
		}
	}
	counts.SummaryNote = summarizeJobNote(openJobs)
	counts.SummaryMeta = "open " + intString(counts.Open) + " · queued " + intString(counts.Queued) + " · running " + intString(counts.Running)
	return openJobs, counts
}

func summarizeJobNote(items []runtime.AgentJob) string {
	for _, item := range items {
		mode := strings.TrimSpace(agentJobResultString(item.Result, "follow_up", "mode"))
		summary := strings.TrimSpace(agentJobResultString(item.Result, "follow_up", "summary"))
		if summary != "" && mode != "continue_loop" {
			return summary
		}
	}
	for _, item := range items {
		summary := strings.TrimSpace(agentJobResultString(item.Result, "follow_up", "summary"))
		if summary != "" {
			return summary
		}
	}
	for _, item := range items {
		if trimmed := strings.TrimSpace(item.Summary); trimmed != "" {
			return trimmed
		}
	}
	return "지금 열린 작업이 없습니다."
}

func summarizeJobAttention(items []runtime.AgentJob) cockpitFounderAttentionSnapshot {
	for _, item := range items {
		summary := strings.TrimSpace(agentJobResultString(item.Result, "follow_up", "summary"))
		if summary == "" {
			continue
		}
		mode := strings.TrimSpace(agentJobResultString(item.Result, "follow_up", "mode"))
		detail := strings.TrimSpace(item.Summary)
		if detail == "" {
			detail = "최근 작업 후속 액션을 먼저 확인하세요."
		}
		return cockpitFounderAttentionSnapshot{
			Mode:       nonEmptyString(mode, "job_follow_up"),
			Summary:    summary,
			Detail:     detail,
			Ref:        item.JobID,
			NextJobIDs: agentJobResultStrings(item.Result, "follow_up", "job_ids"),
		}
	}
	return cockpitFounderAttentionSnapshot{}
}

func agentJobResultString(result map[string]any, path ...string) string {
	current := any(result)
	for _, key := range path {
		node, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = node[key]
	}
	value, ok := current.(string)
	if !ok {
		return ""
	}
	return value
}

func agentJobResultStrings(result map[string]any, path ...string) []string {
	current := any(result)
	for _, key := range path {
		node, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = node[key]
	}
	items, ok := current.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func founderActionSummary(action string) string {
	switch strings.TrimSpace(action) {
	case "review_room":
		return "리뷰룸 안건을 먼저 봐야 합니다."
	case "rollback_or_fix":
		return "위험 lane을 먼저 고쳐야 합니다."
	case "shape_or_promote":
		return "proposal을 shaped work로 올릴 차례입니다."
	case "dispatch_job":
		return "queued job을 먼저 dispatch하거나 triage해야 합니다."
	case "approve":
		return "waiting approval을 founder가 먼저 결정해야 합니다."
	case "continue":
		return "이미 열린 실행 lane을 이어가면 됩니다."
	case "idle":
		return "긴급 blocker가 없어 다음 한 가지를 새로 열 수 있습니다."
	default:
		return ""
	}
}

func summarizeFounderAttention(summary runtime.FounderSummary, view runtime.FounderView, review runtime.ReviewRoomSummary, jobs []runtime.AgentJob) cockpitFounderAttentionSnapshot {
	primaryAction := strings.TrimSpace(summary.PrimaryAction)
	primaryRef := strings.TrimSpace(summary.PrimaryRef)
	rationale := summary.PriorityRationale
	reviewText := ""
	if len(summary.ReviewTopOpen) > 0 {
		reviewText = strings.TrimSpace(summary.ReviewTopOpen[0])
	}
	if reviewText == "" && len(review.TopOpen) > 0 {
		reviewText = strings.TrimSpace(review.TopOpen[0].Text)
	}
	lane := ""
	source := ""
	rule := ""
	reason := ""
	if rationale != nil {
		lane = strings.TrimSpace(rationale.Lane)
		source = strings.TrimSpace(rationale.Source)
		rule = strings.TrimSpace(rationale.Rule)
		reason = strings.TrimSpace(rationale.Reason)
	}
	lastChangeSummary := ""
	if summary.LastChange != nil {
		lastChangeSummary = strings.TrimSpace(summary.LastChange.Summary)
	}
	lastReleaseRef := ""
	if view.LastRelease != nil {
		lastReleaseRef = strings.TrimSpace(view.LastRelease.Ref)
	}
	lastRollbackRef := ""
	if view.LastRollback != nil {
		lastRollbackRef = strings.TrimSpace(view.LastRollback.Ref)
	}
	jobAttention := summarizeJobAttention(jobs)
	result := cockpitFounderAttentionSnapshot{
		Mode:            nonEmptyString(primaryAction, "idle"),
		Summary:         nonEmptyString(founderActionSummary(primaryAction), reviewText, lastChangeSummary, jobAttention.Summary, "지금 연결된 founder 흐름이 없습니다."),
		Ref:             nonEmptyString(primaryRef, jobAttention.Ref),
		Lane:            lane,
		Source:          source,
		Rule:            rule,
		WaitingCount:    summary.WaitingCount,
		RiskCount:       summary.RiskCount,
		ReviewOpenCount: summary.ReviewOpenCount,
		NowCount:        summary.NowCount,
		NextJobIDs:      append([]string{}, jobAttention.NextJobIDs...),
		LastReleaseRef:  lastReleaseRef,
		LastRollbackRef: lastRollbackRef,
	}
	detailParts := make([]string, 0, 3)
	if reason != "" {
		detailParts = append(detailParts, "rule · "+reason)
	}
	if reviewText != "" && reviewText != result.Summary {
		detailParts = append(detailParts, "agenda · "+reviewText)
	}
	if lastChangeSummary != "" && lastChangeSummary != result.Summary {
		detailParts = append(detailParts, "change · "+lastChangeSummary)
	}
	result.Detail = nonEmptyString(strings.Join(detailParts, " · "), "founder priority와 current runtime 상태를 합쳐 지금 먼저 볼 레인을 정리했습니다.")
	return result
}

func summarizePrimaryAttention(founderAttention cockpitFounderAttentionSnapshot, jobs []runtime.AgentJob, sourceIntake cockpitSourceIntakeSnapshot) cockpitFounderAttentionSnapshot {
	if founderAttention.Mode != "" && founderAttention.Mode != "idle" {
		return founderAttention
	}
	if jobAttention := summarizeJobAttention(jobs); jobAttention.Summary != "" {
		return jobAttention
	}
	if strings.TrimSpace(sourceIntake.Attention.Summary) != "" {
		return cockpitFounderAttentionSnapshot{
			Mode:    nonEmptyString(sourceIntake.Attention.Mode, "source_intake"),
			Summary: strings.TrimSpace(sourceIntake.Attention.Summary),
			Detail:  nonEmptyString(strings.TrimSpace(sourceIntake.Attention.Detail), "최근 source intake 후속 액션을 먼저 확인하세요."),
			Ref:     strings.TrimSpace(sourceIntake.Attention.Ref),
		}
	}
	return founderAttention
}

func summarizeFounderFlowDefaults(summary runtime.FounderSummary, founderAttention cockpitFounderAttentionSnapshot, primaryAttention cockpitFounderAttentionSnapshot, jobs []runtime.AgentJob) cockpitFounderFlowDefaultsSnapshot {
	founderRef := ""
	if strings.HasPrefix(strings.TrimSpace(summary.PrimaryRef), "flow_") {
		founderRef = strings.TrimSpace(summary.PrimaryRef)
	}
	jobAttention := summarizeJobAttention(jobs)
	return cockpitFounderFlowDefaultsSnapshot{
		FlowID:     founderRef,
		ApprovalID: "",
		Title:      nonEmptyString(strings.TrimSpace(summary.PrimaryAction), "운영 흐름"),
		Intent:     "보호된 관리자 화면에서 founder 흐름을 닫습니다.",
		Attention: cockpitFounderFlowAttentionSnapshot{
			Summary:      nonEmptyString(primaryAttention.Summary, strings.TrimSpace(summary.PrimaryAction), "지금 연결된 founder 흐름이 없습니다."),
			Ref:          nonEmptyString(primaryAttention.Ref, founderRef, jobAttention.Ref),
			WaitingCount: firstNonZero(founderAttention.WaitingCount, summary.WaitingCount),
			RiskCount:    firstNonZero(founderAttention.RiskCount, summary.RiskCount),
			Detail:       nonEmptyString(primaryAttention.Detail, "대표 액션과 대표 참조를 먼저 확인한 뒤 아래 founder flow 카드를 실행하면 됩니다."),
		},
	}
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func summarizeReviewNote(summary runtime.ReviewRoomSummary) string {
	if len(summary.TopOpen) > 0 && strings.TrimSpace(summary.TopOpen[0].Text) != "" {
		return summary.TopOpen[0].Text
	}
	if summary.OpenCount > 0 {
		return "현재 열린 안건을 바로 확인합니다."
	}
	return "리뷰룸이 비어 있습니다."
}

func summarizeReviewMeta(summary runtime.ReviewRoomSummary) string {
	return "열린 안건 " + intString(summary.OpenCount) + "건"
}

func recentEvents(items []runtime.EventEnvelope, limit int) []runtime.EventEnvelope {
	if limit <= 0 || len(items) <= limit {
		return append([]runtime.EventEnvelope{}, items...)
	}
	start := len(items) - limit
	return append([]runtime.EventEnvelope{}, items[start:]...)
}

func intString(value int) string {
	return strconv.Itoa(value)
}

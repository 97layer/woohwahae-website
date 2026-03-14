package runtime

import (
	"fmt"
	"hash/crc32"
	"strings"
)

func limitActionRoutes(items []ActionRoute, max int) []ActionRoute {
	if items == nil {
		return []ActionRoute{}
	}
	if max <= 0 || len(items) <= max {
		return append([]ActionRoute{}, items...)
	}
	return append([]ActionRoute{}, items[:max]...)
}

func appendUniqueKnowledgeText(items []string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return items
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), trimmed) {
			return items
		}
	}
	return append(items, trimmed)
}

func optionalKnowledgeString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func firstOpenThread(items []OpenThread) (OpenThread, bool) {
	for _, item := range items {
		if strings.TrimSpace(item.Question) != "" {
			return item, true
		}
	}
	return OpenThread{}, false
}

func firstOpenThreadByPatternRef(items []OpenThread, ref string) (OpenThread, bool) {
	ref = strings.TrimSpace(strings.ToLower(ref))
	for _, item := range items {
		for _, patternRef := range item.PatternRefs {
			if strings.TrimSpace(strings.ToLower(patternRef)) == ref {
				return item, true
			}
		}
	}
	return OpenThread{}, false
}

func proposalActionRouteSummary(candidate ProposalCandidate) string {
	if len(candidate.ThreadIDs) > 0 {
		return limitText("Triage "+candidate.ProposalID+"; use the create/promote helper only after resolving the linked thread.", 140)
	}
	if candidate.Source == "knowledge.review_gate" {
		return limitText("Triage "+candidate.ProposalID+"; use the create helper after recording review-room rationale.", 140)
	}
	return limitText("Triage "+candidate.ProposalID+"; use the create helper to capture repeated evidence.", 140)
}

func newKnowledgeActionRoute(kind string, summary string, targetLane string, targetRef *string, command *string, source string) ActionRoute {
	summary = strings.TrimSpace(summary)
	targetLane = strings.TrimSpace(targetLane)
	source = strings.TrimSpace(source)
	parts := []string{strings.TrimSpace(kind), summary, targetLane, source}
	if targetRef != nil {
		parts = append(parts, strings.TrimSpace(*targetRef))
	}
	return ActionRoute{
		RouteID:    fmt.Sprintf("route_%08x", crc32.ChecksumIEEE([]byte(strings.Join(parts, "|")))),
		Kind:       strings.TrimSpace(kind),
		Summary:    summary,
		TargetLane: targetLane,
		TargetRef:  copyStringRef(targetRef),
		Command:    copyStringRef(command),
		Source:     source,
	}
}

func deriveKnowledgeActionRoutes(memory SystemMemory, review ReviewRoomSummary, surprising []string, proposalCandidates []ProposalCandidate, openThreads []OpenThread) []ActionRoute {
	routes := []ActionRoute{}
	seen := map[string]struct{}{}
	appendRoute := func(route ActionRoute) {
		if _, ok := seen[route.RouteID]; ok {
			return
		}
		seen[route.RouteID] = struct{}{}
		routes = append(routes, route)
	}
	if len(proposalCandidates) > 0 {
		candidate := proposalCandidates[0]
		appendRoute(newKnowledgeActionRoute(
			"proposal_gate",
			proposalActionRouteSummary(candidate),
			"proposal",
			optionalKnowledgeString(candidate.ProposalID),
			optionalKnowledgeString(candidate.CreateCommand),
			candidate.Source,
		))
	}
	if review.OpenCount > 0 && len(memory.NextSteps) == 0 {
		if thread, ok := firstOpenThreadByPatternRef(openThreads, "next_steps_gap"); ok {
			appendRoute(newKnowledgeActionRoute("capture_next_step", limitText("Define next step: "+thread.Question, 140), "session_memory", optionalKnowledgeString(thread.ThreadID), nil, thread.Source))
		} else if thread, ok := firstOpenThread(openThreads); ok {
			appendRoute(newKnowledgeActionRoute("capture_next_step", limitText("Define next step: "+thread.Question, 140), "session_memory", optionalKnowledgeString(thread.ThreadID), nil, thread.Source))
		} else {
			appendRoute(newKnowledgeActionRoute("capture_next_step", "Record one concrete next step for the top open review tension.", "session_memory", nil, nil, "knowledge.surprising"))
		}
	}
	if review.OpenCount > 0 && len(memory.OpenRisks) == 0 {
		if thread, ok := firstOpenThreadByPatternRef(openThreads, "open_risks_gap"); ok {
			appendRoute(newKnowledgeActionRoute("capture_risk", limitText("State explicit risk: "+thread.Question, 140), "session_memory", optionalKnowledgeString(thread.ThreadID), nil, thread.Source))
		} else {
			appendRoute(newKnowledgeActionRoute("capture_risk", "Record one explicit risk for the current open review tension.", "session_memory", nil, nil, "knowledge.surprising"))
		}
	}
	if len(routes) == 0 {
		for _, signal := range surprising {
			switch {
			case strings.Contains(signal, "current_focus repeated"):
				appendRoute(newKnowledgeActionRoute("break_focus_stall", "Break focus stall: verify whether the current focus still reflects the highest-value question.", "session_memory", nil, nil, "knowledge.surprising"))
			case strings.Contains(signal, "agent_job.failed repeated") || strings.Contains(signal, "repeated_failure pattern"):
				var ref *string
				if thread, ok := firstOpenThreadByPatternRef(openThreads, "failure_pattern"); ok {
					ref = optionalKnowledgeString(thread.ThreadID)
				}
				appendRoute(newKnowledgeActionRoute("inspect_repeated_failure", "Inspect repeated agent-job failure for a shared root cause before adding new execution lanes.", "review_room", ref, nil, "knowledge.surprising"))
			case strings.Contains(signal, "succeeded entries=0"):
				appendRoute(newKnowledgeActionRoute("close_one_lane", "Close one small execution lane to green before expanding more surface area.", "job", nil, nil, "knowledge.surprising"))
			case strings.HasPrefix(signal, "review_backlog_high:"):
				appendRoute(newKnowledgeActionRoute("reduce_review_backlog", "Reduce the review backlog before adding more execution lanes.", "review_room", nil, nil, "knowledge.surprising"))
			case strings.Contains(signal, "dispatch_risk_detected"):
				appendRoute(newKnowledgeActionRoute("stabilize_dispatch_risk", "Stabilize dispatch/provider risk before expanding external execution lanes.", "review_room", nil, nil, "knowledge.surprising"))
			case strings.Contains(signal, "review_open carry-over"):
				appendRoute(newKnowledgeActionRoute("review_carry_over", "Review persistent carry-over and either promote it to a proposal or resolve it with rationale.", "review_room", nil, nil, "knowledge.surprising"))
			}
			if len(routes) >= 3 {
				break
			}
		}
	}
	if len(routes) == 0 {
		if thread, ok := firstOpenThread(openThreads); ok {
			appendRoute(newKnowledgeActionRoute("answer_open_thread", limitText("Answer: "+thread.Question, 140), "review_room", optionalKnowledgeString(thread.ThreadID), nil, thread.Source))
		}
	}
	return limitActionRoutes(routes, 3)
}

func deriveKnowledgeActionHints(routes []ActionRoute) []string {
	hints := []string{}
	for _, route := range routes {
		hints = appendUniqueKnowledgeText(hints, route.Summary)
	}
	return limitStrings(hints, 3)
}

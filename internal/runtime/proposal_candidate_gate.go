package runtime

import (
	"sort"
	"strings"
)

func proposalGatePasses(link ObservationLink) bool {
	if link.Count < 2 {
		return false
	}
	if hasFormalPlanningRef(link.Refs) {
		return false
	}
	if link.Kind == "ref" {
		return true
	}
	return len(link.Channels) >= 2 || link.Count >= 3
}

func hasFormalPlanningRef(refs []string) bool {
	for _, ref := range refs {
		trimmed := strings.ToLower(strings.TrimSpace(ref))
		if strings.HasPrefix(trimmed, "proposal_") || strings.HasPrefix(trimmed, "proposal:") || strings.HasPrefix(trimmed, "work_") || strings.HasPrefix(trimmed, "work:") {
			return true
		}
	}
	return false
}

func proposalCandidateThreads(link ObservationLink, threads []OpenThread) []OpenThread {
	refMatched := []OpenThread{}
	tokenMatched := []OpenThread{}
	for _, thread := range threads {
		if len(intersectStrings(link.Refs, thread.PatternRefs)) > 0 {
			refMatched = append(refMatched, thread)
			continue
		}
		if sharedTokenCount(link.Summary, thread.Question) > 0 {
			tokenMatched = append(tokenMatched, thread)
		}
	}
	items := refMatched
	if len(items) == 0 {
		items = tokenMatched
	}
	sort.SliceStable(items, func(i, j int) bool {
		if proposalPriorityRank(openThreadPriority(items[i])) == proposalPriorityRank(openThreadPriority(items[j])) {
			return items[i].ThreadID < items[j].ThreadID
		}
		return proposalPriorityRank(openThreadPriority(items[i])) > proposalPriorityRank(openThreadPriority(items[j]))
	})
	return items
}

func proposalThreadRaisesRisk(thread OpenThread) bool {
	switch openThreadKind(thread) {
	case openThreadKindPattern, openThreadKindContradict, openThreadKindStale:
		return true
	default:
		return false
	}
}

func proposalPriorityRank(value string) int {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func proposalRiskRank(value string) int {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func higherProposalPriority(left string, right string) string {
	if proposalPriorityRank(right) > proposalPriorityRank(left) {
		return strings.TrimSpace(strings.ToLower(right))
	}
	return strings.TrimSpace(strings.ToLower(left))
}

func sharedTokenCount(left string, right string) int {
	leftTokens := filteredObservationTokens(left)
	rightTokens := filteredObservationTokens(right)
	if len(leftTokens) == 0 || len(rightTokens) == 0 {
		return 0
	}
	rightSet := map[string]struct{}{}
	for _, token := range rightTokens {
		rightSet[token] = struct{}{}
	}
	count := 0
	for _, token := range leftTokens {
		if _, ok := rightSet[token]; ok {
			count++
		}
	}
	return count
}

func intersectStrings(left []string, right []string) []string {
	if len(left) == 0 || len(right) == 0 {
		return []string{}
	}
	rightSet := map[string]struct{}{}
	for _, item := range right {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		rightSet[trimmed] = struct{}{}
	}
	items := []string{}
	for _, item := range left {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := rightSet[trimmed]; ok {
			items = append(items, trimmed)
		}
	}
	return items
}

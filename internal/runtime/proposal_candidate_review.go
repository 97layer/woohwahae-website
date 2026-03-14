package runtime

import (
	"fmt"
	"sort"
	"strings"
)

func deriveReviewProposalCandidates(room ReviewRoom, threads []OpenThread) []ProposalCandidate {
	if len(room.Open) == 0 {
		return []ProposalCandidate{}
	}
	items := []ProposalCandidate{}
	seen := map[string]struct{}{}
	for _, item := range room.Open {
		refs := reviewProposalRefs(item)
		if !reviewProposalGatePasses(item, refs) {
			continue
		}
		candidate := newReviewProposalCandidate(item, refs, reviewProposalCandidateThreads(item, refs, threads))
		if _, ok := seen[candidate.CandidateID]; ok {
			continue
		}
		seen[candidate.CandidateID] = struct{}{}
		items = append(items, candidate)
	}
	return items
}

func mergeProposalCandidates(groups ...[]ProposalCandidate) []ProposalCandidate {
	items := []ProposalCandidate{}
	seen := map[string]struct{}{}
	for _, group := range groups {
		for _, candidate := range group {
			if _, ok := seen[candidate.CandidateID]; ok {
				continue
			}
			seen[candidate.CandidateID] = struct{}{}
			items = append(items, candidate)
		}
	}
	return items
}

func reviewProposalGatePasses(item ReviewRoomItem, refs []string) bool {
	item = normalizeReviewRoomItem(item)
	if len(refs) == 0 {
		return false
	}
	if hasFormalPlanningRef(refs) {
		return false
	}
	switch strings.TrimSpace(strings.ToLower(item.Severity)) {
	case "high", "critical":
		return strings.TrimSpace(item.Text) != ""
	default:
		return false
	}
}

func reviewProposalRefs(item ReviewRoomItem) []string {
	return reviewItemRefs(item)
}

func reviewProposalCandidateThreads(item ReviewRoomItem, refs []string, threads []OpenThread) []OpenThread {
	summary := reviewProposalSummary(item)
	refMatched := []OpenThread{}
	tokenMatched := []OpenThread{}
	for _, thread := range threads {
		if len(intersectStrings(refs, thread.PatternRefs)) > 0 {
			refMatched = append(refMatched, thread)
			continue
		}
		if sharedTokenCount(summary, thread.Question) > 0 {
			tokenMatched = append(tokenMatched, thread)
		}
	}
	matched := refMatched
	if len(matched) == 0 {
		matched = tokenMatched
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if proposalPriorityRank(openThreadPriority(matched[i])) == proposalPriorityRank(openThreadPriority(matched[j])) {
			return matched[i].ThreadID < matched[j].ThreadID
		}
		return proposalPriorityRank(openThreadPriority(matched[i])) > proposalPriorityRank(openThreadPriority(matched[j]))
	})
	return matched
}

func reviewProposalSummary(item ReviewRoomItem) string {
	item = normalizeReviewRoomItem(item)
	parts := []string{item.Text}
	if strings.TrimSpace(item.Why) != "" {
		parts = append(parts, item.Why)
	}
	if item.WhyUnresolved != nil {
		parts = append(parts, *item.WhyUnresolved)
	}
	parts = append(parts, item.Contradictions...)
	if item.Contradiction != nil {
		parts = append(parts, *item.Contradiction)
	}
	return limitText(strings.Join(normalizeReviewRoomEvidence(parts), " | "), 180)
}

func newReviewProposalCandidate(item ReviewRoomItem, refs []string, threads []OpenThread) ProposalCandidate {
	item = normalizeReviewRoomItem(item)
	threadIDs := []string{}
	questionBits := []string{}
	priority := strings.TrimSpace(strings.ToLower(item.Severity))
	if priority == "critical" {
		priority = "high"
	}
	if priority == "" {
		priority = "medium"
	}
	risk := "medium"
	if len(item.Contradictions) > 0 || item.Contradiction != nil || item.WhyUnresolved != nil || strings.TrimSpace(strings.ToLower(item.Severity)) == "critical" {
		risk = "high"
	}
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ThreadID)
		questionBits = append(questionBits, thread.Question)
		priority = higherProposalPriority(priority, openThreadPriority(thread))
		if proposalThreadRaisesRisk(thread) {
			risk = "high"
		}
	}
	anchorRef := refs[0]
	candidateID := deriveProposalCandidateID("review:"+anchorRef, threadIDs)
	proposalID := deriveProposalID(candidateID)
	workItemID := deriveProposalWorkItemID(candidateID)
	title := limitText("Triage review tension around "+anchorRef, 80)
	question := reviewProposalSummary(item)
	if len(questionBits) > 0 {
		question = strings.Join(limitStrings(questionBits, 2), " | ")
	}
	summary := limitText(fmt.Sprintf("Review tension %s; gate question: %s", limitText(item.Text, 100), limitText(question, 110)), 180)
	notes := []string{"review:" + anchorRef, "review_source:" + item.Source}
	if strings.TrimSpace(item.Why) != "" {
		notes = append(notes, "why:"+strings.TrimSpace(item.Why))
	}
	if item.WhyUnresolved != nil {
		notes = append(notes, "why:"+strings.TrimSpace(*item.WhyUnresolved))
	}
	for _, contradiction := range item.Contradictions {
		notes = append(notes, "contradiction:"+strings.TrimSpace(contradiction))
	}
	if item.Contradiction != nil {
		notes = append(notes, "contradiction:"+strings.TrimSpace(*item.Contradiction))
	}
	notes = append(notes, proposalCandidateNotes(candidateID, "review:"+anchorRef, threadIDs, refs)...)
	notes = normalizeReviewRoomEvidence(notes)
	if notes == nil {
		notes = []string{}
	}
	return ProposalCandidate{
		CandidateID:    candidateID,
		ProposalID:     proposalID,
		WorkItemID:     workItemID,
		Title:          title,
		Intent:         "convert unresolved review tension into an explicit proposal or reject it with rationale",
		Summary:        summary,
		Priority:       priority,
		Risk:           risk,
		Surface:        SurfaceAPI,
		Source:         "knowledge.review_gate",
		Notes:          notes,
		LinkIDs:        []string{"review:" + anchorRef},
		ThreadIDs:      threadIDs,
		Refs:           append([]string{}, refs...),
		CreatePath:     "/api/layer-os/proposals",
		CreateCommand:  proposalCreateCommand(proposalID, title, "convert unresolved review tension into an explicit proposal or reject it with rationale", summary, SurfaceAPI, priority, risk, notes),
		PromotePath:    "/api/layer-os/proposals/promote",
		PromoteCommand: proposalPromoteCommand(proposalID, workItemID),
	}
}

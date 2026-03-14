package runtime

import (
	"fmt"
	"sort"
	"strings"
)

func deriveProposalCandidates(links []ObservationLink, threads []OpenThread) []ProposalCandidate {
	if len(links) == 0 || len(threads) == 0 {
		return []ProposalCandidate{}
	}
	items := []ProposalCandidate{}
	seen := map[string]struct{}{}
	for _, link := range links {
		if !proposalGatePasses(link) {
			continue
		}
		matched := proposalCandidateThreads(link, threads)
		if len(matched) == 0 {
			continue
		}
		candidate := newProposalCandidate(link, matched)
		if _, ok := seen[candidate.CandidateID]; ok {
			continue
		}
		seen[candidate.CandidateID] = struct{}{}
		items = append(items, candidate)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if proposalPriorityRank(items[i].Priority) == proposalPriorityRank(items[j].Priority) {
			if proposalRiskRank(items[i].Risk) == proposalRiskRank(items[j].Risk) {
				return items[i].CandidateID < items[j].CandidateID
			}
			return proposalRiskRank(items[i].Risk) > proposalRiskRank(items[j].Risk)
		}
		return proposalPriorityRank(items[i].Priority) > proposalPriorityRank(items[j].Priority)
	})
	return items
}

func newProposalCandidate(link ObservationLink, threads []OpenThread) ProposalCandidate {
	threadIDs := make([]string, 0, len(threads))
	questionBits := make([]string, 0, len(threads))
	priority := "medium"
	risk := "medium"
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ThreadID)
		questionBits = append(questionBits, thread.Question)
		priority = higherProposalPriority(priority, openThreadPriority(thread))
		if proposalThreadRaisesRisk(thread) {
			risk = "high"
		}
	}
	if len(link.Refs) > 0 {
		for _, thread := range threads {
			if len(intersectStrings(link.Refs, thread.PatternRefs)) > 0 {
				priority = "high"
				break
			}
		}
	}
	if link.Count >= 3 && priority == "medium" {
		priority = "high"
	}
	candidateID := deriveProposalCandidateID(link.LinkID, threadIDs)
	proposalID := deriveProposalID(candidateID)
	workItemID := deriveProposalWorkItemID(candidateID)
	title := proposalCandidateTitle(link)
	summary := fmt.Sprintf("%s across %d observations / %d channels; gate question: %s", link.Summary, link.Count, len(link.Channels), limitText(strings.Join(limitStrings(questionBits, 2), " | "), 110))
	notes := proposalCandidateNotes(candidateID, link.LinkID, threadIDs, link.Refs)
	intent := "convert repeated evidence into an explicit proposal or reject it with rationale"
	summary = limitText(summary, 180)
	surface := SurfaceAPI
	return ProposalCandidate{
		CandidateID:    candidateID,
		ProposalID:     proposalID,
		WorkItemID:     workItemID,
		Title:          title,
		Intent:         intent,
		Summary:        summary,
		Priority:       priority,
		Risk:           risk,
		Surface:        surface,
		Source:         "knowledge.proposal_gate",
		Notes:          notes,
		LinkIDs:        []string{link.LinkID},
		ThreadIDs:      threadIDs,
		Refs:           append([]string{}, link.Refs...),
		CreatePath:     "/api/layer-os/proposals",
		CreateCommand:  proposalCreateCommand(proposalID, title, intent, summary, surface, priority, risk, notes),
		PromotePath:    "/api/layer-os/proposals/promote",
		PromoteCommand: proposalPromoteCommand(proposalID, workItemID),
	}
}

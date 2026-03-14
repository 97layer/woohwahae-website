package runtime

import (
	"fmt"
	"hash/crc32"
	"strings"
)

func proposalCandidateTitle(link ObservationLink) string {
	if len(link.Refs) > 0 {
		return limitText("Triage repeated signal around "+link.Refs[0], 80)
	}
	return limitText("Triage repeated signal cluster", 80)
}

func deriveProposalCandidateID(linkID string, threadIDs []string) string {
	payload := strings.Join(append([]string{strings.TrimSpace(linkID)}, threadIDs...), "|")
	return fmt.Sprintf("pcand_%08x", crc32.ChecksumIEEE([]byte(payload)))
}

func deriveProposalID(candidateID string) string {
	trimmed := strings.TrimSpace(candidateID)
	trimmed = strings.TrimPrefix(trimmed, "pcand_")
	if trimmed == "" {
		trimmed = "draft"
	}
	return "proposal_" + trimmed
}

func deriveProposalWorkItemID(candidateID string) string {
	trimmed := strings.TrimSpace(candidateID)
	trimmed = strings.TrimPrefix(trimmed, "pcand_")
	if trimmed == "" {
		trimmed = "draft"
	}
	return "work_" + trimmed
}

func proposalCandidateNotes(candidateID string, linkID string, threadIDs []string, refs []string) []string {
	items := []string{"candidate:" + candidateID, "link:" + linkID}
	for _, threadID := range threadIDs {
		items = append(items, "thread:"+strings.TrimSpace(threadID))
	}
	for _, ref := range refs {
		trimmed := strings.TrimSpace(ref)
		if trimmed == "" {
			continue
		}
		items = append(items, "ref:"+trimmed)
	}
	return normalizeReviewRoomEvidence(items)
}

func proposalPromoteCommand(proposalID string, workItemID string) string {
	parts := []string{
		"layer-osctl", "proposal", "promote",
		"--id", shellQuoteArg(proposalID),
		"--work", shellQuoteArg(workItemID),
	}
	return strings.Join(parts, " ")
}

func proposalCreateCommand(proposalID string, title string, intent string, summary string, surface Surface, priority string, risk string, notes []string) string {
	parts := []string{
		"layer-osctl", "proposal", "create",
		"--id", shellQuoteArg(proposalID),
		"--title", shellQuoteArg(title),
		"--intent", shellQuoteArg(intent),
		"--summary", shellQuoteArg(summary),
		"--surface", shellQuoteArg(string(surface)),
		"--priority", shellQuoteArg(priority),
		"--risk", shellQuoteArg(risk),
	}
	if len(notes) > 0 {
		parts = append(parts, "--notes", shellQuoteArg(strings.Join(notes, ",")))
	}
	return strings.Join(parts, " ")
}

func shellQuoteArg(value string) string {
	trimmed := strings.TrimSpace(value)
	return "'" + strings.ReplaceAll(trimmed, "'", "'\"'\"'") + "'"
}

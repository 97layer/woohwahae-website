package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestProposalPromotionAndFailedAgentJobSurfaceInKnowledgeAndHandoff(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	proposalID := "proposal_visibility_001"
	workID := "work_visibility_001"
	jobID := "job_visibility_001"
	if err := service.CreateProposal(ProposalItem{
		ProposalID: proposalID,
		Title:      "Visibility smoke lane",
		Intent:     "prove promotion and failed job surface together",
		Summary:    "Validate proposal promotion and review-room visibility",
		Surface:    SurfaceAPI,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      []string{"seed visibility smoke"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, _, err := service.PromoteProposal(proposalID, workID); err != nil {
		t.Fatalf("promote proposal: %v", err)
	}
	ref := proposalID
	if err := service.CreateAgentJob(AgentJob{
		JobID:     jobID,
		Kind:      "implement",
		Role:      "implementer",
		Summary:   "Patch the promoted visibility lane",
		Status:    "running",
		Source:    "proposal",
		Surface:   SurfaceAPI,
		Stage:     StageCompose,
		Ref:       &ref,
		Notes:     []string{"seed running lane"},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.ReportAgentJob(jobID, "failed", []string{"compile error"}, map[string]any{"summary": "Compile error in promoted lane", "artifacts": []string{"internal/runtime/visibility_smoke_test.go"}}); err != nil {
		t.Fatalf("report agent job: %v", err)
	}

	if got := service.ReviewRoomSummary().OpenCount; got != 1 {
		t.Fatalf("expected one open review-room item, got %d", got)
	}
	founder := service.FounderSummary()
	if founder.PrimaryAction != "review_room" {
		t.Fatalf("expected founder primary action review_room, got %+v", founder)
	}
	knowledge := service.Knowledge()
	if knowledge.ReviewOpenCount != 1 {
		t.Fatalf("expected knowledge review open count 1, got %+v", knowledge)
	}
	if len(knowledge.ReviewTopOpen) == 0 || !strings.Contains(knowledge.ReviewTopOpen[0], jobID) {
		t.Fatalf("expected knowledge top review item to mention failed job, got %+v", knowledge.ReviewTopOpen)
	}
	if knowledge.PrimaryAction != "review_room" {
		t.Fatalf("expected knowledge primary action review_room, got %+v", knowledge)
	}
	handoff := service.Handoff()
	if handoff.ReviewRoom.OpenCount != 1 || handoff.FounderSummary.PrimaryAction != "review_room" {
		t.Fatalf("expected handoff to surface review-room priority, got %+v", handoff)
	}
	if handoff.Counts.Proposals != 1 || handoff.Counts.AgentJobs != 1 || handoff.Counts.WorkItems != 1 {
		t.Fatalf("expected handoff counts to include proposal/job/work, got %+v", handoff.Counts)
	}
}

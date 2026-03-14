package runtime

import (
	"testing"
	"time"
)

func TestCreateProposalPersistsAndShowsInFounderWaiting(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateProposal(ProposalItem{ProposalID: "proposal_001", Title: "Plan queue", Intent: "close planning gap", Summary: "Plan queue", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListProposals()); got != 1 {
		t.Fatalf("expected 1 proposal, got %d", got)
	}
	view := reloaded.FounderView()
	if len(view.Waiting) == 0 || view.Waiting[0].Kind != "proposal" {
		t.Fatalf("expected proposal in founder waiting, got %+v", view.Waiting)
	}
	summary := reloaded.FounderSummary()
	if summary.PrimaryAction != "shape_or_promote" {
		t.Fatalf("expected proposal primary action, got %+v", summary)
	}
}

func TestPromoteProposalCreatesWorkAndMarksProposal(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateProposal(ProposalItem{ProposalID: "proposal_001", Title: "Plan queue", Intent: "close planning gap", Summary: "Plan queue", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	proposal, work, err := service.PromoteProposal("proposal_001", "work_001")
	if err != nil {
		t.Fatalf("promote proposal: %v", err)
	}
	if proposal.Status != "promoted" || proposal.PromotedWorkItemID == nil || *proposal.PromotedWorkItemID != "work_001" {
		t.Fatalf("unexpected proposal after promote: %+v", proposal)
	}
	if work.ID != "work_001" || work.Payload["proposal_id"] != "proposal_001" {
		t.Fatalf("unexpected work from proposal: %+v", work)
	}
	if got := len(service.ListWorkItems()); got != 1 {
		t.Fatalf("expected 1 work item, got %d", got)
	}
	if got := service.Handoff().Counts.Proposals; got != 1 {
		t.Fatalf("expected proposal count 1, got %d", got)
	}
}

package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestBranchLifecyclePersistsAcrossRuntimeViews(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	now := time.Now().UTC()
	if err := service.CreateBranch(Branch{
		BranchID:  "branch_root",
		Title:     "Root lane",
		Intent:    "hold canonical root",
		Summary:   "Root lane",
		Stage:     StageCompose,
		Surface:   SurfaceCockpit,
		Status:    "active",
		Notes:     []string{"seed"},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create root branch: %v", err)
	}

	parentID := "branch_root"
	if err := service.CreateBranch(Branch{
		BranchID:       "branch_design",
		ParentBranchID: &parentID,
		Title:          "Design lane",
		Intent:         "parallelize design work",
		Summary:        "Design lane",
		Stage:          StageCompose,
		Surface:        SurfaceCockpit,
		Status:         "active",
		Notes:          []string{"child"},
		CreatedAt:      now.Add(time.Second),
		UpdatedAt:      now.Add(time.Second),
	}); err != nil {
		t.Fatalf("create child branch: %v", err)
	}

	branchID := "branch_design"
	if err := service.CreateProposal(ProposalItem{
		ProposalID: "proposal_001",
		BranchID:   &branchID,
		Title:      "Design workspace",
		Intent:     "attach proposal to branch",
		Summary:    "Design workspace",
		Surface:    SurfaceCockpit,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      []string{"seed"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_001",
		BranchID:  &branchID,
		Kind:      "plan",
		Role:      "designer",
		Summary:   "Plan design lane",
		Status:    "queued",
		Source:    "manual",
		Surface:   SurfaceCockpit,
		Stage:     StageCompose,
		Notes:     []string{"seed"},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create job: %v", err)
	}
	if err := service.CreateWorkItem(WorkItem{
		ID:               "work_001",
		BranchID:         &branchID,
		Title:            "Build design lane",
		Intent:           "attach work to branch",
		Stage:            StageCompose,
		Surface:          SurfaceCockpit,
		Pack:             "design",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: false,
		Payload:          map[string]any{"proposal_id": "proposal_001"},
		CorrelationID:    "proposal_001",
		CreatedAt:        now,
	}); err != nil {
		t.Fatalf("create work: %v", err)
	}

	branches := service.ListBranches()
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %+v", branches)
	}
	if branches[1].RootBranchID != "branch_root" {
		t.Fatalf("expected child root branch_id branch_root, got %+v", branches[1])
	}

	handoff := service.Handoff()
	if handoff.Counts.Branches != 2 {
		t.Fatalf("expected handoff branch count 2, got %d", handoff.Counts.Branches)
	}
	if len(handoff.ActiveBranches) == 0 || handoff.ActiveBranches[0].BranchID != "branch_design" {
		t.Fatalf("expected active branch preview to include branch_design, got %+v", handoff.ActiveBranches)
	}

	knowledge := service.Knowledge()
	if len(knowledge.ActiveBranches) == 0 || knowledge.ActiveBranches[0].BranchID != "branch_design" {
		t.Fatalf("expected knowledge active branches to include branch_design, got %+v", knowledge.ActiveBranches)
	}

	packet := service.Snapshot()
	if len(packet.Branches) != 2 {
		t.Fatalf("expected snapshot branches, got %+v", packet.Branches)
	}
	if packet.Proposals[0].BranchID == nil || *packet.Proposals[0].BranchID != "branch_design" {
		t.Fatalf("expected snapshot proposal branch, got %+v", packet.Proposals[0])
	}
	if packet.AgentJobs[0].BranchID == nil || *packet.AgentJobs[0].BranchID != "branch_design" {
		t.Fatalf("expected snapshot job branch, got %+v", packet.AgentJobs[0])
	}
	if packet.WorkItems[0].BranchID == nil || *packet.WorkItems[0].BranchID != "branch_design" {
		t.Fatalf("expected snapshot work branch, got %+v", packet.WorkItems[0])
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListBranches()); got != 2 {
		t.Fatalf("expected reloaded branches 2, got %d", got)
	}

	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	if got := len(target.ListBranches()); got != 2 {
		t.Fatalf("expected imported branches 2, got %d", got)
	}
	imported := target.ListProposals()
	if len(imported) != 1 || imported[0].BranchID == nil || *imported[0].BranchID != "branch_design" {
		t.Fatalf("expected imported proposal branch, got %+v", imported)
	}
}

func TestMergeBranchOpensReviewRoomAgendaAndBlocksNewWrites(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	now := time.Now().UTC()
	if err := service.CreateBranch(Branch{
		BranchID:  "branch_root",
		Title:     "Root lane",
		Intent:    "hold canonical root",
		Summary:   "Root lane",
		Stage:     StageCompose,
		Surface:   SurfaceAPI,
		Status:    "active",
		Notes:     []string{"seed"},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create root branch: %v", err)
	}
	parentID := "branch_root"
	if err := service.CreateBranch(Branch{
		BranchID:       "branch_leaf",
		ParentBranchID: &parentID,
		Title:          "Leaf lane",
		Intent:         "run branch-bound work",
		Summary:        "Leaf lane",
		Stage:          StageCompose,
		Surface:        SurfaceAPI,
		Status:         "active",
		Notes:          []string{"seed"},
		CreatedAt:      now.Add(time.Second),
		UpdatedAt:      now.Add(time.Second),
	}); err != nil {
		t.Fatalf("create leaf branch: %v", err)
	}

	merged, err := service.MergeBranch("branch_leaf", "branch_root", []string{"close lane", "archive later"})
	if err != nil {
		t.Fatalf("merge branch: %v", err)
	}
	if merged.Status != "merged" || merged.MergeTargetBranchID == nil || *merged.MergeTargetBranchID != "branch_root" || merged.MergedAt == nil {
		t.Fatalf("unexpected merged branch: %+v", merged)
	}

	room := service.ReviewRoom()
	if len(room.Open) != 1 {
		t.Fatalf("expected one review room item, got %+v", room.Open)
	}
	if room.Open[0].Source != "branch.merge" || room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.branch_merge_reconcile" {
		t.Fatalf("unexpected branch merge review room item: %+v", room.Open[0])
	}
	if !strings.Contains(room.Open[0].Text, "branch_leaf") || !strings.Contains(room.Open[0].Text, "branch_root") {
		t.Fatalf("expected branch merge review room text, got %+v", room.Open[0])
	}

	branchID := "branch_leaf"
	err = service.CreateProposal(ProposalItem{
		ProposalID: "proposal_blocked",
		BranchID:   &branchID,
		Title:      "Blocked proposal",
		Intent:     "should fail on merged branch",
		Summary:    "Blocked proposal",
		Surface:    SurfaceAPI,
		Priority:   "low",
		Risk:       "low",
		Status:     "proposed",
		Notes:      []string{"blocked"},
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err == nil || !strings.Contains(err.Error(), "active branch") {
		t.Fatalf("expected merged branch rejection, got %v", err)
	}
}

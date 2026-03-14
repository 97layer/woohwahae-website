package runtime

import (
	"testing"
	"time"
)

func TestRejectedApprovalAutoOpensReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(WorkItem{ID: "work_001", Title: "Ship queue", Intent: "harden queue", Stage: StageDiscover, Surface: SurfaceAPI, Pack: "core", Priority: "high", Risk: "medium", RequiresApproval: false, Payload: map[string]any{}, CorrelationID: "corr_001", CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create work: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{ApprovalID: "approval_001", WorkItemID: "work_001", Stage: StageVerify, Summary: "Review queue", Risks: []string{}, RollbackPlan: "hold", DecisionSurface: SurfaceCockpit, Status: "pending", RequestedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	item, err := service.ResolveApproval("approval_001", "rejected")
	if err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if item.Status != "rejected" {
		t.Fatalf("expected rejected approval, got %+v", item)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Source != "approval.rejected" {
		t.Fatalf("expected approval rejection in review room, got %+v", room.Open)
	}
	if room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.approval_rejected" {
		t.Fatalf("unexpected rationale: %+v", room.Open[0].Rationale)
	}
}

func TestApprovedApprovalAutoResolvesRejectionReviewItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(WorkItem{ID: "work_001", Title: "Ship queue", Intent: "harden queue", Stage: StageDiscover, Surface: SurfaceAPI, Pack: "core", Priority: "high", Risk: "medium", RequiresApproval: false, Payload: map[string]any{}, CorrelationID: "corr_001", CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create work: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", newSignalReviewRoomItem("승인 `approval_001`이 작업 `work_001`에서 거절됐어. 계속 진행하기 전에 founder 검토나 제안 재정리가 필요해.", "approval.rejected", func() *string { value := "approval_001"; return &value }(), "rejected approval requires founder review before execution continues", "review_room.auto.approval_rejected", []string{"approval:approval_001", "work:work_001"})); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{ApprovalID: "approval_001", WorkItemID: "work_001", Stage: StageVerify, Summary: "Review queue", Risks: []string{}, RollbackPlan: "hold", DecisionSurface: SurfaceCockpit, Status: "pending", RequestedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	room := service.ReviewRoom()
	if len(room.Open) != 0 {
		t.Fatalf("expected rejection item resolved, got %+v", room.Open)
	}
}

func TestSessionFinishOpenRisksAutoOpenReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := service.FinishSession(SessionFinishInput{CurrentFocus: "Stabilize sync", NextSteps: []string{"resume"}, OpenRisks: []string{"agent drift", "lease bottleneck"}})
	if err != nil {
		t.Fatalf("finish session: %v", err)
	}
	if result.Event.Kind != "session.finished" {
		t.Fatalf("unexpected session event: %+v", result.Event)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Source != "session.finished" {
		t.Fatalf("expected session risk item in review room, got %+v", room.Open)
	}
	if room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.session_open_risks" {
		t.Fatalf("unexpected session rationale: %+v", room.Open[0].Rationale)
	}
}

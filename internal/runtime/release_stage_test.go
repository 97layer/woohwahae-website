package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestReleaseRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	approval := ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}
	if err := service.CreateApproval(approval); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}

	releasedAt := time.Now().UTC()
	release := ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}

	if err := service.CreateRelease(release); err != nil {
		t.Fatalf("create release: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListReleases()); got != 1 {
		t.Fatalf("expected 1 reloaded release, got %d", got)
	}
	if reloaded.Status().LastReleaseAt == nil {
		t.Fatal("expected last release timestamp to persist")
	}
}

func TestReleaseRequiresApprovedApproval(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	releasedAt := time.Now().UTC()
	err = service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	})
	if err == nil {
		t.Fatal("expected release gate failure, got nil")
	}
}

func TestDeployRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("put target: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	releasedAt := time.Now().UTC()
	if err := service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	finishedAt := time.Now().UTC()
	if err := service.CreateDeploy(DeployRun{
		DeployID:   "deploy_001",
		ReleaseID:  "release_001",
		Target:     "vm",
		Status:     "succeeded",
		Notes:      []string{"initial deploy"},
		StartedAt:  time.Now().UTC(),
		FinishedAt: &finishedAt,
	}); err != nil {
		t.Fatalf("create deploy: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListDeploys()); got != 1 {
		t.Fatalf("expected 1 deploy, got %d", got)
	}
	if reloaded.Status().DeployHealth != "ready" {
		t.Fatalf("expected deploy health ready, got %q", reloaded.Status().DeployHealth)
	}
}

func TestRollbackRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	approval := ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}
	if err := service.CreateApproval(approval); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	releasedAt := time.Now().UTC()
	if err := service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}
	if err := service.PutTarget(DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/true"}}); err != nil {
		t.Fatalf("put target: %v", err)
	}

	item, err := service.ExecuteRollback("rollback_001", "release_001", "", []string{"test"})
	if err != nil {
		t.Fatalf("execute rollback: %v", err)
	}
	if item.Status != "succeeded" {
		t.Fatalf("expected succeeded rollback, got %q", item.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListRollbacks()); got != 1 {
		t.Fatalf("expected 1 reloaded rollback, got %d", got)
	}
	if reloaded.Handoff().Counts.Rollbacks != 1 {
		t.Fatalf("expected handoff rollback count 1, got %d", reloaded.Handoff().Counts.Rollbacks)
	}
}

func TestDeployRequiresRelease(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	finishedAt := time.Now().UTC()
	err = service.CreateDeploy(DeployRun{
		DeployID:   "deploy_001",
		ReleaseID:  "release_missing",
		Target:     "vm",
		Status:     "succeeded",
		Notes:      []string{"initial deploy"},
		StartedAt:  time.Now().UTC(),
		FinishedAt: &finishedAt,
	})
	if err == nil {
		t.Fatal("expected deploy gate failure, got nil")
	}
}

func TestExecuteDeployRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("put target: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	releasedAt := time.Now().UTC()
	if err := service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	deploy, err := service.ExecuteDeploy("deploy_002", "release_001", []string{"adapter run"})
	if err != nil {
		t.Fatalf("execute deploy: %v", err)
	}
	if deploy.Status != "succeeded" {
		t.Fatalf("expected succeeded deploy, got %q", deploy.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListDeploys()); got != 1 {
		t.Fatalf("expected 1 deploy, got %d", got)
	}
}

func TestExecuteDeployFailureStillRecordsRun(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/false"},
	}); err != nil {
		t.Fatalf("put target: %v", err)
	}
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	releasedAt := time.Now().UTC()
	if err := service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	deploy, err := service.ExecuteDeploy("deploy_003", "release_001", []string{"adapter fail"})
	if err == nil {
		t.Fatal("expected execute deploy failure, got nil")
	}
	if deploy.Status != "failed" {
		t.Fatalf("expected failed deploy, got %q", deploy.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListDeploys()); got != 1 {
		t.Fatalf("expected 1 recorded failed deploy, got %d", got)
	}
	if reloaded.Status().DeployHealth != "degraded" {
		t.Fatalf("expected degraded deploy health, got %q", reloaded.Status().DeployHealth)
	}
	expected := "배포 `deploy_003`이 릴리스 `release_001`의 대상 `vm`에서 실패했어. 롤백 준비 상태를 먼저 검토해야 해."
	if len(reloaded.ReviewRoom().Open) != 1 || reloaded.ReviewRoom().Open[0].Text != expected {
		t.Fatalf("unexpected review room open items: %+v", reloaded.ReviewRoom().Open)
	}
	if reloaded.ReviewRoom().Open[0].Rationale == nil || reloaded.ReviewRoom().Open[0].Rationale.Rule != "review_room.auto.deploy_failed" {
		t.Fatalf("unexpected deploy rationale: %+v", reloaded.ReviewRoom().Open[0].Rationale)
	}
	if len(reloaded.ReviewRoom().Open[0].Evidence) != 3 || reloaded.ReviewRoom().Open[0].Evidence[0] != "deploy:deploy_003" {
		t.Fatalf("unexpected deploy evidence: %+v", reloaded.ReviewRoom().Open[0].Evidence)
	}
}

func TestDeployTargetRequiresAbsolutePaths(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"usr/bin/true"},
	}); err == nil {
		t.Fatal("expected absolute command path validation failure, got nil")
	}

	workdir := "relative-dir"
	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
		Workdir:  &workdir,
	}); err == nil {
		t.Fatal("expected absolute workdir validation failure, got nil")
	}
}

func TestReleaseRejectsUnsupportedMetricsValue(t *testing.T) {
	releasedAt := time.Now().UTC()
	err := ReleasePacket{
		ReleaseID:    "release_invalid_metrics",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "stable",
		Artifacts:    []string{"artifact.tar.gz"},
		Metrics:      map[string]any{"latency": []any{120, invalidJSONPayload{Message: "no structs"}}},
		RollbackPlan: "revert release",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}.Validate()
	if err == nil || !strings.Contains(err.Error(), "release metrics") {
		t.Fatalf("expected release metrics validation error, got %v", err)
	}
}

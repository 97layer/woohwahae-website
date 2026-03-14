package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestExecutionLoopClosesAcrossProposalVerifyApprovalReleaseDeploy(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/true"}}); err != nil {
		t.Fatalf("put target: %v", err)
	}

	verificationRoot := seedExecutionVerificationRoot(t)
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", verificationRoot)

	client := daemonClientForService(service)
	now := time.Now().UTC()
	proposalID := "proposal_exec_001"
	workID := "work_exec_001"
	approvalID := "approval_exec_001"
	verificationID := "verify_exec_001"
	releaseID := "release_exec_001"
	deployID := "deploy_exec_001"
	flowID := "flow_exec_001"

	if err := client.CreateProposal(runtime.ProposalItem{
		ProposalID: proposalID,
		Title:      "Ship execution vertical slice",
		Intent:     "close proposal to deploy loop",
		Summary:    "Founder can ship one slice end-to-end",
		Surface:    runtime.SurfaceAPI,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      []string{"seed execution loop"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}

	proposal, work, err := client.PromoteProposal(proposalID, workID)
	if err != nil {
		t.Fatalf("promote proposal: %v", err)
	}
	if proposal.Status != "promoted" {
		t.Fatalf("expected promoted proposal, got %+v", proposal)
	}
	if work.ID != workID || work.Payload["proposal_id"] != proposalID {
		t.Fatalf("unexpected promoted work item: %+v", work)
	}

	verification, err := client.RunVerification(verificationID, "kernel", "", nil, []string{"execution smoke"})
	if err != nil {
		t.Fatalf("run verification: %v", err)
	}
	if verification.RecordID != verificationID || verification.Status != "passed" {
		t.Fatalf("unexpected verification: %+v", verification)
	}

	if err := client.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      approvalID,
		WorkItemID:      workID,
		Stage:           runtime.StageVerify,
		Summary:         "approve execution slice",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "re-run previous release",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	flow, err := client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), nil, nil, nil, []string{"verification passed"})
	if err != nil {
		t.Fatalf("sync flow waiting: %v", err)
	}
	if flow.Status != "waiting" {
		t.Fatalf("expected waiting flow after verification/approval, got %+v", flow)
	}

	approval, err := client.ResolveApproval(approvalID, "approved")
	if err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if approval.Status != "approved" {
		t.Fatalf("unexpected approval resolution: %+v", approval)
	}

	if err := client.CreateRelease(runtime.ReleasePacket{
		ReleaseID:    releaseID,
		WorkItemID:   workID,
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"execution-slice.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "re-run previous release",
		ApprovalRefs: []string{approvalID},
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	deploy, err := client.ExecuteDeploy(deployID, releaseID, []string{"execution deploy"})
	if err != nil {
		t.Fatalf("execute deploy: %v", err)
	}
	if deploy.Status != "succeeded" {
		t.Fatalf("unexpected deploy: %+v", deploy)
	}

	flow, err = client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), stringPtr(releaseID), stringPtr(deployID), nil, []string{"released"})
	if err != nil {
		t.Fatalf("sync flow released: %v", err)
	}
	if flow.Status != "released" {
		t.Fatalf("expected released flow, got %+v", flow)
	}

	status := client.Status()
	if status.DeployHealth != "ready" {
		t.Fatalf("expected ready deploy health, got %+v", status)
	}
	founder := client.FounderSummary()
	if founder.LastRelease == nil || founder.LastRelease.Ref != releaseID {
		t.Fatalf("expected founder last release %q, got %+v", releaseID, founder)
	}

	events := client.ListEvents()
	requiredKinds := []string{
		"proposal.promoted",
		"verification.created",
		"approval.resolved",
		"release.created",
		"deploy.created",
		"flow.synced",
	}
	for _, kind := range requiredKinds {
		if !hasEventKind(events, kind) {
			t.Fatalf("expected event kind %q in %+v", kind, events)
		}
	}
}

func TestExecutionLoopClosesApprovalRejectIntoReviewRoom(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	verificationRoot := seedExecutionVerificationRoot(t)
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", verificationRoot)

	client := daemonClientForService(service)
	now := time.Now().UTC()
	proposalID := "proposal_exec_reject_001"
	workID := "work_exec_reject_001"
	approvalID := "approval_exec_reject_001"
	verificationID := "verify_exec_reject_001"
	flowID := "flow_exec_reject_001"

	if err := client.CreateProposal(runtime.ProposalItem{
		ProposalID: proposalID,
		Title:      "Reject execution slice",
		Intent:     "prove rejection blocks release and opens review tension",
		Summary:    "Founder can see approval rejection before release",
		Surface:    runtime.SurfaceAPI,
		Priority:   "high",
		Risk:       "high",
		Status:     "proposed",
		Notes:      []string{"seed rejection loop"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, _, err := client.PromoteProposal(proposalID, workID); err != nil {
		t.Fatalf("promote proposal: %v", err)
	}
	verification, err := client.RunVerification(verificationID, "kernel", "", nil, []string{"approval rejection smoke"})
	if err != nil {
		t.Fatalf("run verification: %v", err)
	}
	if verification.Status != "passed" {
		t.Fatalf("unexpected verification: %+v", verification)
	}
	if err := client.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      approvalID,
		WorkItemID:      workID,
		Stage:           runtime.StageVerify,
		Summary:         "reject execution slice",
		Risks:           []string{"founder concern persists"},
		RollbackPlan:    "hold release",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	flow, err := client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), nil, nil, nil, []string{"verification passed"})
	if err != nil {
		t.Fatalf("sync waiting flow: %v", err)
	}
	if flow.Status != "waiting" {
		t.Fatalf("expected waiting flow before rejection, got %+v", flow)
	}

	approval, err := client.ResolveApproval(approvalID, "rejected")
	if err != nil {
		t.Fatalf("resolve rejection: %v", err)
	}
	if approval.Status != "rejected" {
		t.Fatalf("expected rejected approval, got %+v", approval)
	}
	flow, err = client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), nil, nil, nil, []string{"approval rejected"})
	if err != nil {
		t.Fatalf("sync blocked flow: %v", err)
	}
	if flow.Status != "blocked" {
		t.Fatalf("expected blocked flow after rejection, got %+v", flow)
	}
	if got := len(client.ListReleases()); got != 0 {
		t.Fatalf("expected no releases after rejection, got %d", got)
	}
	if got := len(client.ListDeploys()); got != 0 {
		t.Fatalf("expected no deploys after rejection, got %d", got)
	}
	room := client.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Source != "approval.rejected" {
		t.Fatalf("expected approval rejection review-room item, got %+v", room.Open)
	}
	if room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.approval_rejected" {
		t.Fatalf("unexpected approval rejection rationale: %+v", room.Open[0].Rationale)
	}
	founder := client.FounderSummary()
	if founder.RiskCount == 0 {
		t.Fatalf("expected founder risk after rejection, got %+v", founder)
	}
	for _, kind := range []string{"approval.resolved", "flow.synced"} {
		if !hasEventKind(client.ListEvents(), kind) {
			t.Fatalf("expected event kind %q after rejection", kind)
		}
	}
}

func TestExecutionLoopClosesFailurePathWithRollbackRecovery(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/false"}}); err != nil {
		t.Fatalf("put failing target: %v", err)
	}

	verificationRoot := seedExecutionVerificationRoot(t)
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", verificationRoot)

	client := daemonClientForService(service)
	now := time.Now().UTC()
	proposalID := "proposal_exec_fail_001"
	workID := "work_exec_fail_001"
	approvalID := "approval_exec_fail_001"
	verificationID := "verify_exec_fail_001"
	releaseID := "release_exec_fail_001"
	deployID := "deploy_exec_fail_001"
	rollbackID := "rollback_exec_fail_001"
	flowID := "flow_exec_fail_001"

	if err := client.CreateProposal(runtime.ProposalItem{
		ProposalID: proposalID,
		Title:      "Recover failed execution slice",
		Intent:     "prove rollback closes deploy failure",
		Summary:    "Founder can recover one failed release lane",
		Surface:    runtime.SurfaceAPI,
		Priority:   "high",
		Risk:       "high",
		Status:     "proposed",
		Notes:      []string{"seed failure loop"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}

	if _, _, err := client.PromoteProposal(proposalID, workID); err != nil {
		t.Fatalf("promote proposal: %v", err)
	}
	verification, err := client.RunVerification(verificationID, "kernel", "", nil, []string{"execution failure smoke"})
	if err != nil {
		t.Fatalf("run verification: %v", err)
	}
	if verification.Status != "passed" {
		t.Fatalf("unexpected verification: %+v", verification)
	}
	if err := client.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      approvalID,
		WorkItemID:      workID,
		Stage:           runtime.StageVerify,
		Summary:         "approve failure recovery slice",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "switch target and execute rollback",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), nil, nil, nil, []string{"verification passed"}); err != nil {
		t.Fatalf("sync initial flow: %v", err)
	}
	if _, err := client.ResolveApproval(approvalID, "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if err := client.CreateRelease(runtime.ReleasePacket{
		ReleaseID:    releaseID,
		WorkItemID:   workID,
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"execution-failure-slice.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 180},
		RollbackPlan: "switch target and execute rollback",
		ApprovalRefs: []string{approvalID},
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	deploy, err := client.ExecuteDeploy(deployID, releaseID, []string{"deploy should fail"})
	if err == nil {
		t.Fatalf("expected deploy failure, got deploy=%+v", deploy)
	}
	if deploy.DeployID != deployID || deploy.Status != "failed" {
		t.Fatalf("expected failed deploy with partial payload, got deploy=%+v err=%v", deploy, err)
	}
	flow, err := client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), stringPtr(releaseID), stringPtr(deployID), nil, []string{"deploy failed"})
	if err != nil {
		t.Fatalf("sync blocked flow: %v", err)
	}
	if flow.Status != "blocked" {
		t.Fatalf("expected blocked flow after failed deploy, got %+v", flow)
	}
	status := client.Status()
	if status.DeployHealth != "degraded" {
		t.Fatalf("expected degraded deploy health after failed deploy, got %+v", status)
	}

	if err := client.PutTarget(runtime.DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/true"}}); err != nil {
		t.Fatalf("put rollback recovery target: %v", err)
	}
	rollback, err := client.ExecuteRollback(rollbackID, releaseID, deployID, []string{"rollback recovery"})
	if err != nil {
		t.Fatalf("execute rollback: %v", err)
	}
	if rollback.RollbackID != rollbackID || rollback.Status != "succeeded" {
		t.Fatalf("unexpected rollback result: %+v", rollback)
	}
	flow, err = client.SyncFlow(flowID, workID, stringPtr(approvalID), nil, nil, stringPtr(verificationID), stringPtr(releaseID), stringPtr(deployID), stringPtr(rollbackID), []string{"rollback recovered"})
	if err != nil {
		t.Fatalf("sync rolled_back flow: %v", err)
	}
	if flow.Status != "rolled_back" {
		t.Fatalf("expected rolled_back flow, got %+v", flow)
	}
	status = client.Status()
	if status.DeployHealth != "ready" {
		t.Fatalf("expected ready deploy health after rollback, got %+v", status)
	}
	founder := client.FounderSummary()
	if founder.LastRollback == nil || founder.LastRollback.Ref != rollbackID {
		t.Fatalf("expected founder last rollback %q, got %+v", rollbackID, founder)
	}

	events := client.ListEvents()
	for _, kind := range []string{"deploy.created", "rollback.created", "flow.synced"} {
		if !hasEventKind(events, kind) {
			t.Fatalf("expected event kind %q in %+v", kind, events)
		}
	}
}

func seedExecutionVerificationRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module executionverify\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "execution_verify.go"), []byte("package executionverify\n\nfunc Ready() bool { return true }\n"), 0o644); err != nil {
		t.Fatalf("write verify.go: %v", err)
	}
	return root
}

func hasEventKind(events []runtime.EventEnvelope, kind string) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

func stringPtr(value string) *string {
	return &value
}

package main

import (
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type workFlowApprovalServiceStubDedicated struct {
	auth        runtime.AuthStatus
	workItems   []runtime.WorkItem
	flows       []runtime.FlowRun
	approvals   []runtime.ApprovalItem
	createdWork runtime.WorkItem
	createdFlow runtime.FlowRun
	syncedFlow  runtime.FlowRun
	syncArgs    struct {
		flowID     string
		workID     string
		approvalID *string
		policyID   *string
		executeID  *string
		verifyID   *string
		releaseID  *string
		deployID   *string
		rollbackID *string
		notes      []string
	}
	createdApproval  runtime.ApprovalItem
	resolvedApproval runtime.ApprovalItem
}

func (s *workFlowApprovalServiceStubDedicated) AuthStatus() runtime.AuthStatus    { return s.auth }
func (s *workFlowApprovalServiceStubDedicated) ListWorkItems() []runtime.WorkItem { return s.workItems }
func (s *workFlowApprovalServiceStubDedicated) CreateWorkItem(item runtime.WorkItem) error {
	s.createdWork = item
	return nil
}
func (s *workFlowApprovalServiceStubDedicated) ListFlows() []runtime.FlowRun { return s.flows }
func (s *workFlowApprovalServiceStubDedicated) CreateFlow(item runtime.FlowRun) error {
	s.createdFlow = item
	return nil
}
func (s *workFlowApprovalServiceStubDedicated) SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (runtime.FlowRun, error) {
	s.syncArgs.flowID = flowID
	s.syncArgs.workID = workItemID
	s.syncArgs.approvalID = approvalID
	s.syncArgs.policyID = policyDecisionID
	s.syncArgs.executeID = executeID
	s.syncArgs.verifyID = verificationID
	s.syncArgs.releaseID = releaseID
	s.syncArgs.deployID = deployID
	s.syncArgs.rollbackID = rollbackID
	s.syncArgs.notes = append([]string{}, notes...)
	s.syncedFlow = runtime.FlowRun{FlowID: flowID, WorkItemID: workItemID, Status: "active", UpdatedAt: time.Now().UTC()}
	return s.syncedFlow, nil
}
func (s *workFlowApprovalServiceStubDedicated) ListApprovals() []runtime.ApprovalItem {
	return s.approvals
}
func (s *workFlowApprovalServiceStubDedicated) CreateApproval(item runtime.ApprovalItem) error {
	s.createdApproval = item
	return nil
}
func (s *workFlowApprovalServiceStubDedicated) ResolveApproval(approvalID string, status string) (runtime.ApprovalItem, error) {
	s.resolvedApproval = runtime.ApprovalItem{ApprovalID: approvalID, Status: status}
	return s.resolvedApproval, nil
}

func TestRunWorkCreateParsesFields_Dedicated(t *testing.T) {
	service := &workFlowApprovalServiceStubDedicated{}
	raw := captureStdout(t, func() {
		runWork(service, []string{"create", "--id", "work_123", "--title", "Ship", "--intent", "release", "--stage", "compose", "--surface", "api", "--pack", "release", "--priority", "high", "--risk", "medium", "--correlation", "proposal_9", "--requires-approval"})
	})
	if service.createdWork.ID != "work_123" || service.createdWork.Stage != runtime.StageCompose || !service.createdWork.RequiresApproval {
		t.Fatalf("unexpected work item: %+v", service.createdWork)
	}
	if !strings.Contains(raw, "work_123") {
		t.Fatalf("missing work id in output: %q", raw)
	}
}

func TestRunWorkListWritesItems_Dedicated(t *testing.T) {
	now := time.Now().UTC()
	service := &workFlowApprovalServiceStubDedicated{
		workItems: []runtime.WorkItem{{ID: "work_list_1", Title: "List me", Stage: runtime.StageDiscover, Surface: runtime.SurfaceAPI, CreatedAt: now}},
	}
	raw := captureStdout(t, func() {
		runWork(service, []string{"list"})
	})
	if !strings.Contains(raw, "work_list_1") {
		t.Fatalf("missing work item in output: %q", raw)
	}
}

func TestRunFlowCreateParsesFields_Dedicated(t *testing.T) {
	service := &workFlowApprovalServiceStubDedicated{}
	raw := captureStdout(t, func() {
		runFlow(service, []string{"create", "--id", "flow_9", "--work", "work_9", "--status", "active", "--notes", "n1,n2"})
	})
	if service.createdFlow.FlowID != "flow_9" || service.createdFlow.WorkItemID != "work_9" || len(service.createdFlow.Notes) != 2 {
		t.Fatalf("unexpected flow create: %+v", service.createdFlow)
	}
	if !strings.Contains(raw, "flow_9") {
		t.Fatalf("missing flow id in output: %q", raw)
	}
}

func TestRunFlowListWritesFlows_Dedicated(t *testing.T) {
	now := time.Now().UTC()
	service := &workFlowApprovalServiceStubDedicated{
		flows: []runtime.FlowRun{{FlowID: "flow_list_1", WorkItemID: "work_1", Status: "active", UpdatedAt: now}},
	}
	raw := captureStdout(t, func() {
		runFlow(service, []string{"list"})
	})
	if !strings.Contains(raw, "flow_list_1") || !strings.Contains(raw, "work_1") {
		t.Fatalf("missing flow list data in output: %q", raw)
	}
}

func TestRunFlowSyncWiresRefs_Dedicated(t *testing.T) {
	service := &workFlowApprovalServiceStubDedicated{}
	raw := captureStdout(t, func() {
		runFlow(service, []string{"sync", "--id", "flow_1", "--work", "work_1", "--approval", "approval_1", "--policy", "policy_1", "--execute", "exec_1", "--verify", "verify_1", "--release", "release_1", "--deploy", "deploy_1", "--rollback", "rollback_1", "--notes", "one,two"})
	})
	if service.syncArgs.flowID != "flow_1" || service.syncArgs.workID != "work_1" {
		t.Fatalf("unexpected sync ids: %+v", service.syncArgs)
	}
	if service.syncArgs.approvalID == nil || *service.syncArgs.approvalID != "approval_1" || service.syncArgs.deployID == nil || *service.syncArgs.deployID != "deploy_1" {
		t.Fatalf("unexpected sync refs: %+v", service.syncArgs)
	}
	if len(service.syncArgs.notes) != 2 || service.syncArgs.notes[1] != "two" {
		t.Fatalf("unexpected notes: %+v", service.syncArgs.notes)
	}
	if !strings.Contains(raw, "flow_1") {
		t.Fatalf("missing flow in output: %q", raw)
	}
}

func TestRunApprovalCreateAndResolve_Dedicated(t *testing.T) {
	service := &workFlowApprovalServiceStubDedicated{}
	rawCreate := captureStdout(t, func() {
		runApproval(service, []string{"create", "--id", "approval_1", "--work", "work_1", "--summary", "approve release", "--stage", "verify", "--surface", "cockpit", "--status", "pending", "--rollback", "restore", "--risks", "drift"})
	})
	if service.createdApproval.ApprovalID != "approval_1" || service.createdApproval.RollbackPlan != "restore" || len(service.createdApproval.Risks) != 1 {
		t.Fatalf("unexpected approval create: %+v", service.createdApproval)
	}
	if !strings.Contains(rawCreate, "approval_1") {
		t.Fatalf("missing approval id in output: %q", rawCreate)
	}

	rawResolve := captureStdout(t, func() {
		runApproval(service, []string{"resolve", "--id", "approval_1", "--status", "approved"})
	})
	if service.resolvedApproval.Status != "approved" {
		t.Fatalf("unexpected resolved status: %+v", service.resolvedApproval)
	}
	if !strings.Contains(rawResolve, "approved") {
		t.Fatalf("missing resolved status in output: %q", rawResolve)
	}
}

func TestRunApprovalListWritesApprovals_Dedicated(t *testing.T) {
	now := time.Now().UTC()
	service := &workFlowApprovalServiceStubDedicated{
		approvals: []runtime.ApprovalItem{{ApprovalID: "approval_list_1", WorkItemID: "work_1", Stage: runtime.StageVerify, Summary: "check", DecisionSurface: runtime.SurfaceCockpit, Status: "pending", RequestedAt: now}},
	}
	raw := captureStdout(t, func() {
		runApproval(service, []string{"list"})
	})
	if !strings.Contains(raw, "approval_list_1") || !strings.Contains(raw, "work_1") {
		t.Fatalf("missing approval list data in output: %q", raw)
	}
}

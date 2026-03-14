package main

import (
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type branchServiceStub struct {
	auth         runtime.AuthStatus
	branches     []runtime.Branch
	created      runtime.Branch
	merged       runtime.Branch
	mergedID     string
	mergedTarget string
	mergedNotes  []string
}

func (s *branchServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *branchServiceStub) ListBranches() []runtime.Branch { return s.branches }
func (s *branchServiceStub) CreateBranch(item runtime.Branch) error {
	s.created = item
	s.branches = append(s.branches, item)
	return nil
}
func (s *branchServiceStub) MergeBranch(branchID string, targetBranchID string, notes []string) (runtime.Branch, error) {
	s.mergedID = branchID
	s.mergedTarget = targetBranchID
	s.mergedNotes = append([]string{}, notes...)
	now := time.Now().UTC()
	s.merged = runtime.Branch{
		BranchID:            branchID,
		RootBranchID:        targetBranchID,
		Title:               "Merged branch",
		Intent:              "merge",
		Summary:             "Merged branch",
		Stage:               runtime.StageCompose,
		Surface:             runtime.SurfaceAPI,
		Status:              "merged",
		MergeTargetBranchID: &targetBranchID,
		MergedAt:            &now,
		Notes:               append([]string{}, notes...),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return s.merged, nil
}

func TestRunBranchCreateParsesFields(t *testing.T) {
	service := &branchServiceStub{}
	raw := captureStdout(t, func() {
		runBranch(service, []string{"create", "--id", "branch_design", "--title", "Design lane", "--intent", "parallelize design", "--summary", "Design lane summary", "--stage", "compose", "--surface", "cockpit", "--parent", "branch_root", "--basis-ref", "proposal:design", "--notes", "one,two"})
	})
	if service.created.BranchID != "branch_design" || service.created.ParentBranchID == nil || *service.created.ParentBranchID != "branch_root" {
		t.Fatalf("unexpected branch create payload: %+v", service.created)
	}
	if service.created.BasisRef == nil || *service.created.BasisRef != "proposal:design" {
		t.Fatalf("unexpected basis ref: %+v", service.created.BasisRef)
	}
	if len(service.created.Notes) != 2 || service.created.Notes[1] != "two" {
		t.Fatalf("unexpected branch notes: %+v", service.created.Notes)
	}
	if !strings.Contains(raw, "branch_design") {
		t.Fatalf("expected branch output, got %s", raw)
	}
}

func TestRunBranchMergeCallsService(t *testing.T) {
	service := &branchServiceStub{}
	raw := captureStdout(t, func() {
		runBranch(service, []string{"merge", "--id", "branch_design", "--target", "branch_root", "--notes", "close,out"})
	})
	if service.mergedID != "branch_design" || service.mergedTarget != "branch_root" {
		t.Fatalf("unexpected merge call: id=%q target=%q", service.mergedID, service.mergedTarget)
	}
	if len(service.mergedNotes) != 2 || service.mergedNotes[0] != "close" {
		t.Fatalf("unexpected merge notes: %+v", service.mergedNotes)
	}
	if !strings.Contains(raw, "branch_root") || !strings.Contains(raw, "\"status\": \"merged\"") {
		t.Fatalf("unexpected merge output: %s", raw)
	}
}

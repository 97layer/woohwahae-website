package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type workFlowApprovalService interface {
	authStatusProvider
	ListWorkItems() []runtime.WorkItem
	CreateWorkItem(item runtime.WorkItem) error
	ListFlows() []runtime.FlowRun
	CreateFlow(item runtime.FlowRun) error
	SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (runtime.FlowRun, error)
	ListApprovals() []runtime.ApprovalItem
	CreateApproval(item runtime.ApprovalItem) error
	ResolveApproval(approvalID string, status string) (runtime.ApprovalItem, error)
}

func runWork(service workFlowApprovalService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl work <list|create>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListWorkItems())
	case "create":
		createWork(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl work <list|create>")
	}
}

func runFlow(service workFlowApprovalService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl flow <list|create|sync>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListFlows())
	case "create":
		createFlow(service, args[1:])
	case "sync":
		syncFlow(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl flow <list|create|sync>")
	}
}

func runApproval(service workFlowApprovalService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl approval <list|create|resolve>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListApprovals())
	case "create":
		createApproval(service, args[1:])
	case "resolve":
		resolveApproval(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl approval <list|create|resolve>")
	}
}

func createWork(service workFlowApprovalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("work create", flag.ExitOnError)
	id := cmd.String("id", "", "work item id")
	branch := cmd.String("branch", "", "branch id")
	title := cmd.String("title", "", "work item title")
	intent := cmd.String("intent", "", "work item intent")
	stage := cmd.String("stage", string(runtime.StageDiscover), "stage")
	surface := cmd.String("surface", string(runtime.SurfaceAPI), "surface")
	pack := cmd.String("pack", "", "pack")
	priority := cmd.String("priority", "", "priority")
	risk := cmd.String("risk", "", "risk")
	correlation := cmd.String("correlation", "", "correlation id")
	requiresApproval := cmd.Bool("requires-approval", false, "requires approval")
	parseArgs(cmd, args)

	item := runtime.WorkItem{
		ID:               strings.TrimSpace(*id),
		BranchID:         normalizeCLIRef(*branch),
		Title:            strings.TrimSpace(*title),
		Intent:           strings.TrimSpace(*intent),
		Stage:            runtime.Stage(strings.TrimSpace(*stage)),
		Surface:          runtime.Surface(strings.TrimSpace(*surface)),
		Pack:             strings.TrimSpace(*pack),
		Priority:         strings.TrimSpace(*priority),
		Risk:             strings.TrimSpace(*risk),
		RequiresApproval: *requiresApproval,
		Payload:          map[string]any{},
		CorrelationID:    strings.TrimSpace(*correlation),
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func createFlow(service workFlowApprovalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("flow create", flag.ExitOnError)
	id := cmd.String("id", "", "flow id")
	workID := cmd.String("work", "", "work item id")
	status := cmd.String("status", "active", "flow status")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item := runtime.FlowRun{
		FlowID:     strings.TrimSpace(*id),
		WorkItemID: strings.TrimSpace(*workID),
		Status:     strings.TrimSpace(*status),
		Notes:      splitCSV(*notes),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := service.CreateFlow(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func syncFlow(service workFlowApprovalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("flow sync", flag.ExitOnError)
	id := cmd.String("id", "", "flow id")
	workID := cmd.String("work", "", "work item id")
	approvalID := cmd.String("approval", "", "approval id")
	policyID := cmd.String("policy", "", "policy decision id")
	executeID := cmd.String("execute", "", "execute id")
	verifyID := cmd.String("verify", "", "verification id")
	releaseID := cmd.String("release", "", "release id")
	deployID := cmd.String("deploy", "", "deploy id")
	rollbackID := cmd.String("rollback", "", "rollback id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.SyncFlow(
		strings.TrimSpace(*id),
		strings.TrimSpace(*workID),
		normalizeCLIRef(*approvalID),
		normalizeCLIRef(*policyID),
		normalizeCLIRef(*executeID),
		normalizeCLIRef(*verifyID),
		normalizeCLIRef(*releaseID),
		normalizeCLIRef(*deployID),
		normalizeCLIRef(*rollbackID),
		splitCSV(*notes),
	)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func createApproval(service workFlowApprovalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("approval create", flag.ExitOnError)
	id := cmd.String("id", "", "approval id")
	workID := cmd.String("work", "", "work item id")
	summary := cmd.String("summary", "", "approval summary")
	stage := cmd.String("stage", string(runtime.StageVerify), "stage")
	surface := cmd.String("surface", string(runtime.SurfaceCockpit), "decision surface")
	status := cmd.String("status", "pending", "approval status")
	rollback := cmd.String("rollback", "", "rollback plan")
	risks := cmd.String("risks", "", "comma-separated risks")
	parseArgs(cmd, args)

	item := runtime.ApprovalItem{
		ApprovalID:      strings.TrimSpace(*id),
		WorkItemID:      strings.TrimSpace(*workID),
		Stage:           runtime.Stage(strings.TrimSpace(*stage)),
		Summary:         strings.TrimSpace(*summary),
		Risks:           splitCSV(*risks),
		RollbackPlan:    strings.TrimSpace(*rollback),
		DecisionSurface: runtime.Surface(strings.TrimSpace(*surface)),
		Status:          strings.TrimSpace(*status),
		RequestedAt:     time.Now().UTC(),
	}
	if err := service.CreateApproval(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func resolveApproval(service workFlowApprovalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("approval resolve", flag.ExitOnError)
	id := cmd.String("id", "", "approval id")
	status := cmd.String("status", "", "approval status")
	parseArgs(cmd, args)

	item, err := service.ResolveApproval(strings.TrimSpace(*id), strings.TrimSpace(*status))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type branchService interface {
	authStatusProvider
	ListBranches() []runtime.Branch
	CreateBranch(item runtime.Branch) error
	MergeBranch(branchID string, targetBranchID string, notes []string) (runtime.Branch, error)
}

func runBranch(service branchService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl branch <list|create|merge>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListBranches())
	case "create":
		createBranch(service, args[1:])
	case "merge":
		mergeBranch(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl branch <list|create|merge>")
	}
}

func createBranch(service branchService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("branch create", flag.ExitOnError)
	id := cmd.String("id", "", "branch id")
	title := cmd.String("title", "", "branch title")
	intent := cmd.String("intent", "", "branch intent")
	summary := cmd.String("summary", "", "branch summary")
	stage := cmd.String("stage", string(runtime.StageCompose), "branch stage")
	surface := cmd.String("surface", string(runtime.SurfaceAPI), "branch surface")
	parent := cmd.String("parent", "", "parent branch id")
	basisRef := cmd.String("basis-ref", "", "basis ref")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	now := time.Now().UTC()
	item := runtime.Branch{
		BranchID:       strings.TrimSpace(*id),
		ParentBranchID: normalizeCLIRef(*parent),
		Title:          strings.TrimSpace(*title),
		Intent:         strings.TrimSpace(*intent),
		Summary:        strings.TrimSpace(*summary),
		Stage:          runtime.Stage(strings.TrimSpace(*stage)),
		Surface:        runtime.Surface(strings.TrimSpace(*surface)),
		Status:         "active",
		BasisRef:       normalizeCLIRef(*basisRef),
		Notes:          splitCSV(*notes),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if item.Summary == "" {
		item.Summary = item.Title
	}
	if err := service.CreateBranch(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func mergeBranch(service branchService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("branch merge", flag.ExitOnError)
	id := cmd.String("id", "", "branch id")
	target := cmd.String("target", "", "target branch id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.MergeBranch(strings.TrimSpace(*id), strings.TrimSpace(*target), splitCSV(*notes))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

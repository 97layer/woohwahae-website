package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type proposalService interface {
	authStatusProvider
	ListProposals() []runtime.ProposalItem
	CreateProposal(item runtime.ProposalItem) error
	PromoteProposal(proposalID string, workID string) (runtime.ProposalItem, runtime.WorkItem, error)
}

func runProposal(service proposalService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl proposal <list|create|promote>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListProposals())
	case "create":
		createProposal(service, args[1:])
	case "promote":
		promoteProposal(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl proposal <list|create|promote>")
	}
}

func createProposal(service proposalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("proposal create", flag.ExitOnError)
	id := cmd.String("id", "", "proposal id")
	branch := cmd.String("branch", "", "branch id")
	title := cmd.String("title", "", "proposal title")
	intent := cmd.String("intent", "", "proposal intent")
	summary := cmd.String("summary", "", "proposal summary")
	surface := cmd.String("surface", string(runtime.SurfaceAPI), "proposal surface")
	priority := cmd.String("priority", "", "proposal priority")
	risk := cmd.String("risk", "", "proposal risk")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	now := time.Now().UTC()
	item := runtime.ProposalItem{
		ProposalID: strings.TrimSpace(*id),
		BranchID:   normalizeCLIRef(*branch),
		Title:      strings.TrimSpace(*title),
		Intent:     strings.TrimSpace(*intent),
		Summary:    strings.TrimSpace(*summary),
		Surface:    runtime.Surface(strings.TrimSpace(*surface)),
		Priority:   strings.TrimSpace(*priority),
		Risk:       strings.TrimSpace(*risk),
		Status:     "proposed",
		Notes:      splitCSV(*notes),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if item.Summary == "" {
		item.Summary = item.Title
	}
	if err := service.CreateProposal(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func promoteProposal(service proposalService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("proposal promote", flag.ExitOnError)
	id := cmd.String("id", "", "proposal id")
	workID := cmd.String("work", "", "work item id")
	parseArgs(cmd, args)

	proposal, work, err := service.PromoteProposal(strings.TrimSpace(*id), strings.TrimSpace(*workID))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(map[string]any{"proposal": proposal, "work_item": work})
}

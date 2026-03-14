package main

import (
	"flag"
	"log"
	"strings"

	"layer-os/internal/runtime"
)

type founderService interface {
	authStatusProvider
	FounderSummary() runtime.FounderSummary
	StartFounderFlow(flowID string, workID string, approvalID string, title string, intent string, notes []string) (runtime.FlowRun, error)
	ApproveFounderFlow(flowID string, notes []string) (runtime.FlowRun, error)
	ReleaseFounderFlow(flowID string, releaseID string, deployID string, target string, channel string, notes []string) (runtime.FlowRun, error)
	RollbackFounderFlow(flowID string, rollbackID string, notes []string) (runtime.FlowRun, error)
}

func runFounder(service founderService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl founder <summary|start|approve|release|rollback>")
	}
	switch args[0] {
	case "summary":
		runFounderSummary(service, args[1:])
	case "start":
		startFounder(service, args[1:])
	case "approve":
		approveFounder(service, args[1:])
	case "release":
		releaseFounder(service, args[1:])
	case "rollback":
		rollbackFounder(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl founder <summary|start|approve|release|rollback>")
	}
}

func startFounder(service founderService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("founder start", flag.ExitOnError)
	flowID := cmd.String("flow", "", "flow id")
	workID := cmd.String("work", "", "work item id")
	approvalID := cmd.String("approval", "", "approval id")
	title := cmd.String("title", "", "founder work title")
	intent := cmd.String("intent", "", "founder work intent")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.StartFounderFlow(strings.TrimSpace(*flowID), strings.TrimSpace(*workID), strings.TrimSpace(*approvalID), strings.TrimSpace(*title), strings.TrimSpace(*intent), splitCSV(*notes))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func approveFounder(service founderService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("founder approve", flag.ExitOnError)
	flowID := cmd.String("flow", "", "flow id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.ApproveFounderFlow(strings.TrimSpace(*flowID), splitCSV(*notes))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func releaseFounder(service founderService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("founder release", flag.ExitOnError)
	flowID := cmd.String("flow", "", "flow id")
	releaseID := cmd.String("release", "", "release id")
	deployID := cmd.String("deploy", "", "deploy id")
	target := cmd.String("target", "vm", "release target")
	channel := cmd.String("channel", "cockpit", "release channel")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.ReleaseFounderFlow(strings.TrimSpace(*flowID), strings.TrimSpace(*releaseID), strings.TrimSpace(*deployID), strings.TrimSpace(*target), strings.TrimSpace(*channel), splitCSV(*notes))
	if item.FlowID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func rollbackFounder(service founderService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("founder rollback", flag.ExitOnError)
	flowID := cmd.String("flow", "", "flow id")
	rollbackID := cmd.String("rollback", "", "rollback id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.RollbackFounderFlow(strings.TrimSpace(*flowID), strings.TrimSpace(*rollbackID), splitCSV(*notes))
	if item.FlowID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

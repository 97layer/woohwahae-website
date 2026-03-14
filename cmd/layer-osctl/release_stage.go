package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type releaseStageService interface {
	authStatusProvider
	ListReleases() []runtime.ReleasePacket
	ListDeploys() []runtime.DeployRun
	ListRollbacks() []runtime.RollbackRun
	ListTargets() []runtime.DeployTarget
	CreateRelease(item runtime.ReleasePacket) error
	CreateDeploy(item runtime.DeployRun) error
	CreateRollback(item runtime.RollbackRun) error
	PutTarget(item runtime.DeployTarget) error
	ExecuteDeploy(deployID string, releaseID string, baseNotes []string) (runtime.DeployRun, error)
	ExecuteRollback(rollbackID string, releaseID string, deployID string, baseNotes []string) (runtime.RollbackRun, error)
}

func runRelease(service releaseStageService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl release <list|create>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListReleases())
	case "create":
		createRelease(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl release <list|create>")
	}
}

func runDeploy(service releaseStageService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl deploy <list|create|execute>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListDeploys())
	case "create":
		createDeploy(service, args[1:])
	case "execute":
		executeDeploy(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl deploy <list|create|execute>")
	}
}

func runRollback(service releaseStageService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl rollback <list|create|execute>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListRollbacks())
	case "create":
		createRollback(service, args[1:])
	case "execute":
		executeRollback(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl rollback <list|create|execute>")
	}
}

func runTarget(service releaseStageService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl target <list|put>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListTargets())
	case "put":
		putTarget(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl target <list|put>")
	}
}

func createRelease(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("release create", flag.ExitOnError)
	id := cmd.String("id", "", "release id")
	workID := cmd.String("work", "", "work item id")
	target := cmd.String("target", "", "release target")
	channel := cmd.String("channel", "", "release channel")
	artifacts := cmd.String("artifacts", "", "comma-separated artifacts")
	rollback := cmd.String("rollback", "", "rollback plan")
	approvals := cmd.String("approvals", "", "comma-separated approval refs")
	metrics := cmd.String("metrics", "", "comma-separated key=value metrics")
	parseArgs(cmd, args)

	releasedAt := time.Now().UTC()
	item := runtime.ReleasePacket{
		ReleaseID:    strings.TrimSpace(*id),
		WorkItemID:   strings.TrimSpace(*workID),
		Target:       strings.TrimSpace(*target),
		Channel:      strings.TrimSpace(*channel),
		Artifacts:    splitCSV(*artifacts),
		Metrics:      splitPairs(*metrics),
		RollbackPlan: strings.TrimSpace(*rollback),
		ApprovalRefs: splitCSV(*approvals),
		ReleasedAt:   &releasedAt,
	}
	if err := service.CreateRelease(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func createDeploy(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("deploy create", flag.ExitOnError)
	id := cmd.String("id", "", "deploy id")
	releaseID := cmd.String("release", "", "release id")
	target := cmd.String("target", "", "deploy target")
	status := cmd.String("status", "succeeded", "deploy status")
	notes := cmd.String("notes", "", "comma-separated deploy notes")
	parseArgs(cmd, args)

	startedAt := time.Now().UTC()
	finishedAt := time.Now().UTC()
	item := runtime.DeployRun{
		DeployID:   strings.TrimSpace(*id),
		ReleaseID:  strings.TrimSpace(*releaseID),
		Target:     strings.TrimSpace(*target),
		Status:     strings.TrimSpace(*status),
		Notes:      splitCSV(*notes),
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}
	if err := service.CreateDeploy(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func executeDeploy(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("deploy execute", flag.ExitOnError)
	id := cmd.String("id", "", "deploy id")
	releaseID := cmd.String("release", "", "release id")
	notes := cmd.String("notes", "", "comma-separated deploy notes")
	parseArgs(cmd, args)

	item, err := service.ExecuteDeploy(strings.TrimSpace(*id), strings.TrimSpace(*releaseID), splitCSV(*notes))
	if item.DeployID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func putTarget(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("target put", flag.ExitOnError)
	id := cmd.String("id", "", "target id")
	command := cmd.String("command", "", "comma-separated command parts")
	workdir := cmd.String("workdir", "", "command workdir")
	parseArgs(cmd, args)

	item := runtime.DeployTarget{
		TargetID: strings.TrimSpace(*id),
		Command:  splitCSV(*command),
	}
	if strings.TrimSpace(*workdir) != "" {
		item.Workdir = workdir
	}
	if err := service.PutTarget(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func createRollback(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("rollback create", flag.ExitOnError)
	id := cmd.String("id", "", "rollback id")
	releaseID := cmd.String("release", "", "release id")
	deployID := cmd.String("deploy", "", "deploy id")
	target := cmd.String("target", "", "rollback target")
	status := cmd.String("status", "recorded", "rollback status")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	var deployRef *string
	if strings.TrimSpace(*deployID) != "" {
		value := strings.TrimSpace(*deployID)
		deployRef = &value
	}
	item := runtime.RollbackRun{
		RollbackID: strings.TrimSpace(*id),
		ReleaseID:  strings.TrimSpace(*releaseID),
		DeployID:   deployRef,
		Target:     strings.TrimSpace(*target),
		Status:     strings.TrimSpace(*status),
		Notes:      splitCSV(*notes),
		StartedAt:  time.Now().UTC(),
	}
	if item.Status != "recorded" {
		finishedAt := time.Now().UTC()
		item.FinishedAt = &finishedAt
	}
	if err := service.CreateRollback(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func executeRollback(service releaseStageService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("rollback execute", flag.ExitOnError)
	id := cmd.String("id", "", "rollback id")
	releaseID := cmd.String("release", "", "release id")
	deployID := cmd.String("deploy", "", "deploy id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.ExecuteRollback(strings.TrimSpace(*id), strings.TrimSpace(*releaseID), strings.TrimSpace(*deployID), splitCSV(*notes))
	if item.RollbackID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

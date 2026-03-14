package main

import (
	"flag"
	"log"
	"strings"

	"layer-os/internal/runtime"
)

type threadsService interface {
	authStatusProvider
	ThreadsStatus() runtime.ThreadsStatus
	PublishThreads(approvalID string) (runtime.ThreadsPublishReceipt, error)
}

func runThreads(service threadsService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl threads <status|publish>")
	}
	switch args[0] {
	case "status":
		writeJSON(service.ThreadsStatus())
	case "publish":
		publishThreads(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl threads <status|publish>")
	}
}

func publishThreads(service threadsService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("threads publish", flag.ExitOnError)
	approvalID := cmd.String("approval", "", "approved brand publish approval id")
	parseArgs(cmd, args)

	trimmedApprovalID := strings.TrimSpace(*approvalID)
	if trimmedApprovalID == "" {
		log.Fatal("usage: layer-osctl threads publish --approval <approval-id>")
	}

	receipt, err := service.PublishThreads(trimmedApprovalID)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(receipt)
}

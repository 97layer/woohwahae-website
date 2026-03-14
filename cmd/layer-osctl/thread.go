package main

import (
	"flag"
	"log"

	"layer-os/internal/runtime"
)

type threadService interface {
	ListOpenThreads(limit int) []runtime.OpenThread
}

func runThread(service threadService, args []string) {
	if len(args) == 0 || args[0] == "list" {
		listOpenThreads(service, args[minInt(len(args), 1):])
		return
	}
	log.Fatal("usage: layer-osctl thread list [--limit <n>]")
}

func listOpenThreads(service threadService, args []string) {
	cmd := flag.NewFlagSet("thread list", flag.ExitOnError)
	limit := cmd.Int("limit", 0, "max open threads")
	parseArgs(cmd, args)
	if *limit < 0 {
		log.Fatal("limit must be >= 0")
	}
	writeJSON(service.ListOpenThreads(*limit))
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

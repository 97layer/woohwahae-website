package main

import (
	"fmt"
	"log"
	"os"

	"layer-os/internal/runtime"
)

type nextService interface {
	SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error)
	ListAgentJobsByStatus(status string, limit int) []runtime.AgentJob
}

func runNext(service nextService, args []string) {
	if len(args) != 0 {
		log.Fatal("usage: layer-osctl next")
	}
	packet, err := service.SessionBootstrap(false)
	if err != nil {
		log.Fatal(err)
	}
	items := service.ListAgentJobsByStatus("open", 1)
	fmt.Fprintln(os.Stdout, resolveNextAction(packet, items))
}

package main

import (
	"log"

	"layer-os/internal/runtime"
)

type providerService interface {
	Providers() []runtime.ProviderSummary
}

func runProvider(service providerService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl provider <list>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.Providers())
	default:
		log.Fatal("usage: layer-osctl provider <list>")
	}
}

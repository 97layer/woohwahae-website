package main

import (
	"flag"
	"log"
	"strings"

	"layer-os/internal/runtime"
)

type corpusService interface {
	ListCapitalizationEntries() []runtime.CapitalizationEntry
	SearchKnowledge(query string) runtime.KnowledgeSearchResponse
}

func runCorpus(service corpusService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl corpus <list|search>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListCapitalizationEntries())
	case "search":
		runCorpusSearch(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl corpus <list|search>")
	}
}

func runCorpusSearch(service corpusService, args []string) {
	cmd := flag.NewFlagSet("corpus search", flag.ExitOnError)
	query := cmd.String("query", "", "knowledge search query")
	parseArgs(cmd, args)
	value := strings.TrimSpace(*query)
	if value == "" && cmd.NArg() > 0 {
		value = strings.TrimSpace(strings.Join(cmd.Args(), " "))
	}
	if value == "" {
		log.Fatal("usage: layer-osctl corpus search --query <text>")
	}
	writeJSON(service.SearchKnowledge(value))
}

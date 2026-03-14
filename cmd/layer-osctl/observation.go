package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type observationService interface {
	authStatusProvider
	ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord
	CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error)
}

func runObservation(service observationService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl observation <list|create>")
	}
	switch args[0] {
	case "list":
		listObservations(service, args[1:])
	case "create":
		createObservation(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl observation <list|create>")
	}
}

func listObservations(service observationService, args []string) {
	cmd := flag.NewFlagSet("observation list", flag.ExitOnError)
	channel := cmd.String("channel", "", "source channel")
	topic := cmd.String("topic", "", "topic filter")
	actor := cmd.String("actor", "", "actor filter")
	ref := cmd.String("ref", "", "reference filter")
	text := cmd.String("text", "", "text filter")
	limit := cmd.Int("limit", 0, "max observations to return")
	parseArgs(cmd, args)
	if *limit < 0 {
		log.Fatal("limit must not be negative")
	}
	writeJSON(service.ListObservations(runtime.ObservationQuery{
		SourceChannel: strings.TrimSpace(*channel),
		Topic:         strings.TrimSpace(*topic),
		Actor:         strings.TrimSpace(*actor),
		Ref:           strings.TrimSpace(*ref),
		Text:          strings.TrimSpace(*text),
		Limit:         *limit,
	}))
}

func createObservation(service observationService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("observation create", flag.ExitOnError)
	id := cmd.String("id", "", "observation id")
	channel := cmd.String("channel", "", "source channel")
	topic := cmd.String("topic", "", "topic")
	summary := cmd.String("summary", "", "normalized summary")
	excerpt := cmd.String("excerpt", "", "raw excerpt")
	actor := cmd.String("actor", "", "actor")
	confidence := cmd.String("confidence", "medium", "confidence")
	refs := cmd.String("refs", "", "comma-separated refs")
	capturedAt := cmd.String("captured-at", "", "captured-at RFC3339 timestamp")
	parseArgs(cmd, args)

	item := runtime.ObservationRecord{
		ObservationID:     strings.TrimSpace(*id),
		SourceChannel:     strings.TrimSpace(*channel),
		Actor:             strings.TrimSpace(*actor),
		Topic:             strings.TrimSpace(*topic),
		Refs:              splitCSV(*refs),
		Confidence:        strings.TrimSpace(*confidence),
		RawExcerpt:        strings.TrimSpace(*excerpt),
		NormalizedSummary: strings.TrimSpace(*summary),
	}
	if value := strings.TrimSpace(*capturedAt); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			log.Fatalf("invalid captured-at: %v", err)
		}
		item.CapturedAt = parsed.UTC()
	}
	created, err := service.CreateObservation(item)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(created)
}

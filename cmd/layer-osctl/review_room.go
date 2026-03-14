package main

import (
	"flag"
	"log"
	"strings"

	"layer-os/internal/runtime"
)

type reviewRoomService interface {
	authStatusProvider
	ReviewRoom() runtime.ReviewRoom
	ReviewRoomSummary() runtime.ReviewRoomSummary
	AddReviewRoomItem(section string, item runtime.ReviewRoomItem) (runtime.ReviewRoom, error)
	TransitionStructuredReviewRoomItem(action string, item string, resolution *runtime.ReviewRoomResolution) (runtime.ReviewRoom, error)
}

func runReviewRoom(service reviewRoomService, args []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		runReviewRoomRead(service, args)
		return
	}
	switch args[0] {
	case "add":
		addReviewRoomItem(service, args[1:])
	case "accept", "defer", "resolve":
		transitionReviewRoomItem(service, args[0], args[1:])
	default:
		log.Fatal("usage: layer-osctl review-room [--allow-local-fallback] | [add|accept|defer|resolve] ...")
	}
}

func addReviewRoomItem(service reviewRoomService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("review-room add", flag.ExitOnError)
	section := cmd.String("section", "open", "review room section")
	item := cmd.String("item", "", "review room item")
	kind := cmd.String("kind", "agenda", "review room item kind")
	severity := cmd.String("severity", "medium", "review room item severity")
	source := cmd.String("source", "manual", "review room item source")
	ref := cmd.String("ref", "", "review room item ref")
	why := cmd.String("why", "", "decision basis for this agenda item")
	whyUnresolved := cmd.String("why-unresolved", "", "why this item remains unresolved")
	contradiction := cmd.String("contradiction", "", "described unresolved contradiction")
	contradictions := cmd.String("contradictions", "", "contradicting accepted decisions csv")
	patternRefs := cmd.String("pattern-refs", "", "pattern ref csv")
	parseArgs(cmd, args)
	mergedContradictions := splitCSV(*contradictions)
	if value := strings.TrimSpace(*contradiction); value != "" {
		mergedContradictions = append(mergedContradictions, value)
	}

	room, err := service.AddReviewRoomItem(strings.TrimSpace(*section), runtime.ReviewRoomItem{
		Text:           strings.TrimSpace(*item),
		Kind:           strings.TrimSpace(*kind),
		Severity:       strings.TrimSpace(*severity),
		Source:         strings.TrimSpace(*source),
		Ref:            normalizeCLIRef(*ref),
		Why:            strings.TrimSpace(*why),
		WhyUnresolved:  normalizeCLIRef(*whyUnresolved),
		Contradiction:  normalizeCLIRef(*contradiction),
		Contradictions: mergedContradictions,
		PatternRefs:    splitCSV(*patternRefs),
	})
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(map[string]any{
		"review_room": room,
		"summary":     service.ReviewRoomSummary(),
	})
}

func transitionReviewRoomItem(service reviewRoomService, action string, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("review-room "+action, flag.ExitOnError)
	id := cmd.String("id", "", "review room item ref")
	item := cmd.String("item", "", "review room item text")
	reason := cmd.String("reason", "", "resolution reason")
	rule := cmd.String("rule", "", "resolution rule")
	evidence := cmd.String("evidence", "", "resolution evidence csv")
	parseArgs(cmd, args)

	target := strings.TrimSpace(*item)
	if target == "" {
		resolved, err := reviewRoomItemTextByID(service.ReviewRoom(), *id)
		if err != nil {
			log.Fatal(err)
		}
		target = resolved
	}

	room, err := service.TransitionStructuredReviewRoomItem(strings.TrimSpace(action), target, &runtime.ReviewRoomResolution{
		Action:   strings.TrimSpace(action),
		Reason:   strings.TrimSpace(*reason),
		Rule:     strings.TrimSpace(*rule),
		Evidence: splitCSV(*evidence),
	})
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(map[string]any{
		"review_room": room,
		"summary":     service.ReviewRoomSummary(),
	})
}

func reviewRoomItemTextByID(room runtime.ReviewRoom, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", runtime.ErrMissingReviewRoomItem
	}
	for _, item := range append(append([]runtime.ReviewRoomItem{}, room.Open...), room.Deferred...) {
		if item.Ref == nil {
			continue
		}
		if strings.TrimSpace(*item.Ref) == id {
			return item.Text, nil
		}
	}
	for _, item := range room.Accepted {
		if item.Ref == nil {
			continue
		}
		if strings.TrimSpace(*item.Ref) == id {
			return item.Text, nil
		}
	}
	return "", runtime.ErrReviewRoomItemNotFound
}

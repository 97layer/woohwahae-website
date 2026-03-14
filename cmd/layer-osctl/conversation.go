package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type conversationService interface {
	authStatusProvider
	ConversationStatus() runtime.ConversationAutomationStatus
	CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error)
}

func runConversation(service conversationService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl conversation <status|note>")
	}
	switch args[0] {
	case "status":
		writeJSON(service.ConversationStatus())
	case "note":
		runConversationNote(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl conversation <status|note>")
	}
}

func runConversationNote(service conversationService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("conversation note", flag.ExitOnError)
	id := cmd.String("id", "", "conversation id")
	text := cmd.String("text", "", "conversation text")
	source := cmd.String("source", "terminal", "source channel")
	actor := cmd.String("actor", "", "actor override")
	summary := cmd.String("summary", "", "optional summary override")
	tags := cmd.String("tags", "", "comma-separated tags")
	refs := cmd.String("refs", "", "comma-separated refs")
	confidence := cmd.String("confidence", "medium", "confidence")
	capturedAt := cmd.String("captured-at", "", "captured-at RFC3339 timestamp")
	autoPlan := cmd.String("auto-plan", "", "override auto plan toggle (true|false)")
	autoRisk := cmd.String("auto-risk", "", "override auto risk toggle (true|false)")
	autoDispatch := cmd.String("auto-dispatch", "", "override auto dispatch toggle (true|false)")
	llmTag := cmd.String("llm-tag", "", "override llm tagging toggle (true|false)")
	parseArgs(cmd, args)

	input := runtime.ConversationNoteInput{
		ConversationID: strings.TrimSpace(*id),
		Text:           strings.TrimSpace(*text),
		Actor:          strings.TrimSpace(*actor),
		SourceChannel:  strings.TrimSpace(*source),
		Tags:           splitCSV(*tags),
		Summary:        strings.TrimSpace(*summary),
		Refs:           splitCSV(*refs),
		Confidence:     strings.TrimSpace(*confidence),
		AutoPlan:       parseOptionalBoolFlag(*autoPlan, "--auto-plan"),
		AutoRisk:       parseOptionalBoolFlag(*autoRisk, "--auto-risk"),
		AutoDispatch:   parseOptionalBoolFlag(*autoDispatch, "--auto-dispatch"),
		LLMTag:         parseOptionalBoolFlag(*llmTag, "--llm-tag"),
	}
	if value := strings.TrimSpace(*capturedAt); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			log.Fatalf("invalid captured-at: %v", err)
		}
		input.CapturedAt = &parsed
	}

	result, err := service.CreateConversationNote(input)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func parseOptionalBoolFlag(raw string, flagName string) *bool {
	switch value := strings.TrimSpace(strings.ToLower(raw)); value {
	case "":
		return nil
	case "1", "true", "yes", "on":
		enabled := true
		return &enabled
	case "0", "false", "no", "off":
		enabled := false
		return &enabled
	default:
		log.Fatalf("invalid %s: %q", flagName, raw)
		return nil
	}
}

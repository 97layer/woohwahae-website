package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"layer-os/internal/runtime"
)

type sessionService interface {
	authStatusProvider
	SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error)
	CheckpointSession(input runtime.SessionCheckpointInput) (runtime.SessionCheckpoint, error)
	SessionNote(input runtime.SessionNoteInput) (runtime.SessionNoteResult, error)
	SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error)
}

func runSession(service sessionService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl session <bootstrap|checkpoint|note|finish>")
	}
	switch args[0] {
	case "bootstrap":
		runSessionBootstrap(service, args[1:])
	case "checkpoint":
		runSessionCheckpoint(service, args[1:])
	case "note":
		runSessionNote(service, args[1:])
	case "finish":
		runSessionFinish(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl session <bootstrap|checkpoint|note|finish>")
	}
}

func runSessionBootstrap(service sessionService, args []string) {
	cmd := flag.NewFlagSet("session bootstrap", flag.ExitOnError)
	full := cmd.Bool("full", false, "include handoff, review-room, and capabilities")
	allowLocalFallback := cmd.Bool("allow-local-fallback", false, "use read-only local fallback when daemon is unavailable")
	parseArgs(cmd, args)

	packet, err := sessionBootstrapPacket(service, *full, *allowLocalFallback)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(packet)
}

func runSessionCheckpoint(service sessionService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("session checkpoint", flag.ExitOnError)
	source := cmd.String("source", "terminal", "checkpoint source")
	focus := cmd.String("focus", "", "current focus")
	goal := cmd.String("goal", "", "current goal")
	steps := cmd.String("steps", "", "comma-separated next steps")
	risks := cmd.String("risks", "", "comma-separated open risks")
	handoff := cmd.String("handoff", "", "handoff note")
	note := cmd.String("note", "", "checkpoint note")
	refs := cmd.String("refs", "", "comma-separated refs")
	parseArgs(cmd, args)

	input := runtime.SessionCheckpointInput{
		Source:       strings.TrimSpace(*source),
		CurrentFocus: strings.TrimSpace(*focus),
		CurrentGoal:  normalizeCLIRef(*goal),
		NextSteps:    splitCSV(*steps),
		OpenRisks:    splitCSV(*risks),
		HandoffNote:  normalizeCLIRef(*handoff),
		Note:         normalizeCLIRef(*note),
		Refs:         splitCSV(*refs),
	}
	result, err := sessionCheckpointResult(service, input)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func runSessionNote(service sessionService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("session note", flag.ExitOnError)
	source := cmd.String("source", "terminal", "note source")
	kind := cmd.String("kind", "note", "note kind")
	text := cmd.String("text", "", "note text")
	refs := cmd.String("refs", "", "comma-separated refs")
	parseArgs(cmd, args)

	input := runtime.SessionNoteInput{
		Source: strings.TrimSpace(*source),
		Kind:   strings.TrimSpace(*kind),
		Text:   strings.TrimSpace(*text),
		Refs:   splitCSV(*refs),
	}
	result, err := sessionNoteResult(service, input)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func runSessionFinish(service sessionService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("session finish", flag.ExitOnError)
	focus := cmd.String("focus", "", "current focus")
	goal := cmd.String("goal", "", "current goal")
	steps := cmd.String("steps", "", "comma-separated next steps")
	risks := cmd.String("risks", "", "comma-separated open risks")
	handoff := cmd.String("handoff", "", "handoff note")
	note := cmd.String("note", "", "session note")
	parseArgs(cmd, args)

	input := runtime.SessionFinishInput{
		CurrentFocus: strings.TrimSpace(*focus),
		CurrentGoal:  normalizeCLIRef(*goal),
		NextSteps:    splitCSV(*steps),
		OpenRisks:    splitCSV(*risks),
		HandoffNote:  normalizeCLIRef(*handoff),
		Note:         normalizeCLIRef(*note),
	}
	result, err := sessionFinishResult(service, input)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

type sessionFinishRequester interface {
	request(method string, path string, payload any, out any) error
}

func sessionCheckpointResult(service sessionService, input runtime.SessionCheckpointInput) (runtime.SessionCheckpoint, error) {
	requester, ok := service.(sessionFinishRequester)
	if !ok {
		return service.CheckpointSession(input)
	}
	var result runtime.SessionCheckpoint
	err := requester.request(http.MethodPost, "/api/layer-os/session/checkpoint", input, &result)
	return result, err
}

func sessionNoteResult(service sessionService, input runtime.SessionNoteInput) (runtime.SessionNoteResult, error) {
	requester, ok := service.(sessionFinishRequester)
	if !ok {
		return service.SessionNote(input)
	}
	var result runtime.SessionNoteResult
	err := requester.request(http.MethodPost, "/api/layer-os/session/note", input, &result)
	return result, err
}

func sessionFinishResult(service sessionService, input runtime.SessionFinishInput) (runtime.SessionFinishResult, error) {
	requester, ok := service.(sessionFinishRequester)
	if !ok {
		return service.SessionFinish(input)
	}
	var result runtime.SessionFinishResult
	err := requester.request(http.MethodPost, "/api/layer-os/session/finish", input, &result)
	return result, err
}

func localSessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error) {
	service, err := runtime.NewService(localRuntimeDataDir())
	if err != nil {
		return runtime.SessionBootstrapPacket{}, err
	}
	packet := service.SessionBootstrap("local_fallback", true, full)
	return packet, packet.Validate()
}

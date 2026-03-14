package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type systemStateService interface {
	authStatusProvider
	AuthStatus() runtime.AuthStatus
	Memory() runtime.SystemMemory
	SetWriteToken(token string) (runtime.AuthStatus, error)
	ClearWriteToken() (runtime.AuthStatus, error)
	ReplaceMemory(memory runtime.SystemMemory) error
	Snapshot() runtime.SnapshotPacket
	ImportSnapshot(packet runtime.SnapshotPacket) error
}

func runAuth(service systemStateService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl auth <status|set|clear>")
	}
	switch args[0] {
	case "status":
		writeJSON(service.AuthStatus())
	case "set":
		setAuth(service, args[1:])
	case "clear":
		clearAuth(service)
	default:
		log.Fatal("usage: layer-osctl auth <status|set|clear>")
	}
}

func runSnapshot(service systemStateService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl snapshot <export|import>")
	}
	switch args[0] {
	case "export":
		exportSnapshot(service, args[1:])
	case "import":
		importSnapshot(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl snapshot <export|import>")
	}
}

func runMemory(service systemStateService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl memory <get|set>")
	}
	switch args[0] {
	case "get":
		writeJSON(service.Memory())
	case "set":
		setMemory(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl memory <get|set>")
	}
}

func setAuth(service systemStateService, args []string) {
	cmd := flag.NewFlagSet("auth set", flag.ExitOnError)
	token := cmd.String("token", "", "write token")
	parseArgs(cmd, args)

	plain := strings.TrimSpace(*token)
	if plain == "" {
		generated, err := runtime.GenerateToken()
		if err != nil {
			log.Fatal(err)
		}
		plain = generated
	}
	if service.AuthStatus().WriteAuthEnabled {
		requireCLIWriteAuth(service)
	}

	status, err := service.SetWriteToken(plain)
	if err != nil {
		log.Fatal(err)
	}
	response := map[string]any{"auth": status}
	if strings.TrimSpace(*token) == "" {
		response["plain_token"] = plain
	}
	writeJSON(response)
}

func clearAuth(service systemStateService) {
	requireCLIWriteAuth(service)
	status, err := service.ClearWriteToken()
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(status)
}

func setMemory(service systemStateService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("memory set", flag.ExitOnError)
	focus := cmd.String("focus", "", "current focus")
	goal := cmd.String("goal", "", "current goal")
	steps := cmd.String("steps", "", "comma-separated next steps")
	risks := cmd.String("risks", "", "comma-separated open risks")
	handoff := cmd.String("handoff", "", "handoff note")
	operator := cmd.String("operator", runtime.DefaultActor(), "last operator")
	parseArgs(cmd, args)

	if strings.TrimSpace(*focus) == "" {
		log.Fatal("memory set requires --focus")
	}

	memory := runtime.SystemMemory{
		CurrentFocus: *focus,
		NextSteps:    splitCSV(*steps),
		OpenRisks:    splitCSV(*risks),
		UpdatedAt:    time.Now().UTC(),
	}
	if strings.TrimSpace(*goal) != "" {
		memory.CurrentGoal = goal
	}
	if strings.TrimSpace(*handoff) != "" {
		memory.HandoffNote = handoff
	}
	if strings.TrimSpace(*operator) != "" {
		memory.LastOperator = operator
	}

	if err := service.ReplaceMemory(memory); err != nil {
		log.Fatal(err)
	}
	writeJSON(service.Memory())
}

func exportSnapshot(service systemStateService, args []string) {
	cmd := flag.NewFlagSet("snapshot export", flag.ExitOnError)
	file := cmd.String("file", "", "snapshot file path")
	parseArgs(cmd, args)

	packet := service.Snapshot()
	if strings.TrimSpace(*file) == "" {
		writeJSON(packet)
		return
	}
	raw, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(strings.TrimSpace(*file), raw, 0o644); err != nil {
		log.Fatal(err)
	}
	writeJSON(map[string]any{"file": strings.TrimSpace(*file)})
}

func importSnapshot(service systemStateService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("snapshot import", flag.ExitOnError)
	file := cmd.String("file", "", "snapshot file path")
	parseArgs(cmd, args)

	if strings.TrimSpace(*file) == "" {
		log.Fatal("snapshot import requires --file")
	}
	raw, err := os.ReadFile(strings.TrimSpace(*file))
	if err != nil {
		log.Fatal(err)
	}
	var packet runtime.SnapshotPacket
	if err := json.Unmarshal(raw, &packet); err != nil {
		log.Fatal(err)
	}
	if err := service.ImportSnapshot(packet); err != nil {
		log.Fatal(err)
	}
	writeJSON(service.Snapshot())
}

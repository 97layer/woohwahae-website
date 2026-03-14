package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"layer-os/internal/api"
	"layer-os/internal/runtime"
)

func main() {
	if strings.TrimSpace(os.Getenv("LAYER_OS_WRITER_ID")) == "" {
		os.Setenv("LAYER_OS_WRITER_ID", "layer-osctl")
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	handler, ok := rootCommandRegistry[os.Args[1]]
	if !ok {
		usage()
		os.Exit(2)
	}

	handler(newDaemonClient(), os.Args[2:])
}

func runAudit(args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl audit <structure|contracts|residue|gemini|corpus|authority|surface|security|review-room|runtime-data|mcp|docs> [--strict]")
	}
	kind := strings.TrimSpace(args[0])
	cmd := flag.NewFlagSet("audit "+kind, flag.ExitOnError)
	cmd.SetOutput(io.Discard)
	strict := cmd.Bool("strict", false, "exit non-zero when audit finds issues")
	if err := cmd.Parse(args[1:]); err != nil {
		log.Fatal(err)
	}
	result, hasIssues, err := selectAudit(kind, repoRootForLocal(), localRuntimeDataDir())
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
	if *strict && hasIssues {
		os.Exit(1)
	}
}

func runSmoke(args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl smoke <cockpit>")
	}
	switch args[0] {
	case "cockpit":
		writeJSON(map[string]any{"issues": api.SmokeCockpit()})
	default:
		log.Fatal("usage: layer-osctl smoke <cockpit>")
	}
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

func splitPairs(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	items := map[string]any{}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			continue
		}
		items[key] = value
	}
	return items
}

func normalizeCLIRef(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func repoRootForLocal() string {
	root := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT"))
	if root != "" {
		return root
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func localRuntimeDataDir() string {
	if raw := strings.TrimSpace(os.Getenv("LAYER_OS_DATA_DIR")); raw != "" {
		if filepath.IsAbs(raw) {
			return filepath.Clean(raw)
		}
		return filepath.Join(repoRootForLocal(), raw)
	}
	return filepath.Join(repoRootForLocal(), ".layer-os")
}

func requireCLIWriteAuth(service authStatusProvider) {
	status := service.AuthStatus()
	if !status.WriteAuthEnabled {
		return
	}
	token := strings.TrimSpace(os.Getenv("LAYER_OS_WRITE_TOKEN"))
	if token == "" {
		log.Fatal("write auth enabled: set LAYER_OS_WRITE_TOKEN")
	}
}

func writeJSON(value any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	for _, line := range runtime.CanonicalSurfaceSpec().CLIUsage {
		fmt.Fprintln(os.Stderr, "  "+line)
	}
}

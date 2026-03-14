package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RuntimeDataFileCheck struct {
	Path   string   `json:"path"`
	Format string   `json:"format"`
	Status string   `json:"status"`
	Issues []string `json:"issues"`
}

type RuntimeDataAudit struct {
	Status            string                 `json:"status"`
	DataDir           string                 `json:"data_dir"`
	CheckedFiles      []RuntimeDataFileCheck `json:"checked_files"`
	Issues            []string               `json:"issues"`
	SuggestedCommands []string               `json:"suggested_commands"`
}

func AuditRuntimeData(_ string, dataDir string) RuntimeDataAudit {
	audit := RuntimeDataAudit{
		Status:       "ok",
		DataDir:      strings.TrimSpace(dataDir),
		CheckedFiles: []RuntimeDataFileCheck{},
		Issues:       []string{},
		SuggestedCommands: []string{
			"layer-osctl audit runtime-data --strict",
			"layer-osctl session bootstrap --allow-local-fallback",
		},
	}
	for _, spec := range runtimeDataAuditSpecs(dataDir) {
		check := RuntimeDataFileCheck{
			Path:   spec.path,
			Format: spec.format,
			Status: "ok",
			Issues: []string{},
		}
		switch spec.format {
		case "json":
			raw, err := os.ReadFile(spec.path)
			if err != nil {
				if os.IsNotExist(err) {
					check.Status = "missing"
					break
				}
				check.Status = "degraded"
				check.Issues = append(check.Issues, err.Error())
			} else if err := json.Unmarshal(raw, new(any)); err != nil {
				check.Status = "degraded"
				check.Issues = append(check.Issues, err.Error())
			}
		case "ndjson":
			if err := validateNDJSONFile(spec.path); err != nil {
				check.Status = "degraded"
				check.Issues = append(check.Issues, err.Error())
			}
		}
		if check.Status == "degraded" {
			audit.Status = "degraded"
			audit.Issues = append(audit.Issues, fmt.Sprintf("%s: %s", filepath.Base(check.Path), strings.Join(check.Issues, "; ")))
		}
		audit.CheckedFiles = append(audit.CheckedFiles, check)
	}
	sort.Slice(audit.CheckedFiles, func(i, j int) bool { return audit.CheckedFiles[i].Path < audit.CheckedFiles[j].Path })
	sort.Strings(audit.Issues)
	return audit
}

type runtimeDataAuditSpec struct {
	path   string
	format string
}

func runtimeDataAuditSpecs(dataDir string) []runtimeDataAuditSpec {
	return []runtimeDataAuditSpec{
		{path: filepath.Join(dataDir, "proposals.json"), format: "json"},
		{path: filepath.Join(dataDir, "agent_jobs.json"), format: "json"},
		{path: filepath.Join(dataDir, "work_items.json"), format: "json"},
		{path: filepath.Join(dataDir, "flows.json"), format: "json"},
		{path: filepath.Join(dataDir, "approval_inbox.json"), format: "json"},
		{path: filepath.Join(dataDir, "releases.json"), format: "json"},
		{path: filepath.Join(dataDir, "events.json"), format: "json"},
		{path: filepath.Join(dataDir, "events_archive.json"), format: "ndjson"},
		{path: filepath.Join(dataDir, "deploys.json"), format: "json"},
		{path: filepath.Join(dataDir, "rollbacks.json"), format: "json"},
		{path: filepath.Join(dataDir, "deploy_targets.json"), format: "json"},
		{path: filepath.Join(dataDir, "preflights.json"), format: "json"},
		{path: filepath.Join(dataDir, "policies.json"), format: "json"},
		{path: filepath.Join(dataDir, "gateway_calls.json"), format: "json"},
		{path: filepath.Join(dataDir, "executes.json"), format: "json"},
		{path: filepath.Join(dataDir, "verifications.json"), format: "json"},
		{path: filepath.Join(dataDir, "observations.json"), format: "json"},
		{path: filepath.Join(dataDir, "review_room.json"), format: "json"},
		{path: filepath.Join(dataDir, "review_room.seal.json"), format: "json"},
		{path: filepath.Join(dataDir, "system_memory.json"), format: "json"},
		{path: filepath.Join(dataDir, "company_state.json"), format: "json"},
		{path: filepath.Join(dataDir, "write_lease.json"), format: "json"},
	}
}

func validateNDJSONFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), new(any)); err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
	}
	return scanner.Err()
}

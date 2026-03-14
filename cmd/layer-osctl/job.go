package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type jobService interface {
	authStatusProvider
	ListAgentJobs() []runtime.AgentJob
	CreateAgentJob(item runtime.AgentJob) error
	PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error)
	UpdateAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJob, error)
	ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error)
	DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error)
	AgentDispatchProfiles() []runtime.AgentDispatchProfile
	AgentRunPacket(jobID string) (runtime.AgentRunPacket, error)
	SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error)
}

func runJob(service jobService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl job <list|create|promote|update|report|dispatch|profiles|packet|work>")
	}
	switch args[0] {
	case "list":
		listJobs(service, args[1:])
	case "create":
		createJob(service, args[1:])
	case "promote":
		promoteJobs(service, args[1:])
	case "update":
		updateJob(service, args[1:])
	case "report":
		reportJob(service, args[1:])
	case "dispatch":
		dispatchJob(service, args[1:])
	case "profiles":
		writeJSON(service.AgentDispatchProfiles())
	case "packet":
		jobPacket(service, args[1:])
	case "work":
		workJobs(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl job <list|create|promote|update|report|dispatch|profiles|packet|work>")
	}
}

func listJobs(service jobService, args []string) {
	cmd := flag.NewFlagSet("job list", flag.ExitOnError)
	status := cmd.String("status", "", "filter by status (queued|running|succeeded|failed|canceled|open)")
	limit := cmd.Int("limit", 0, "max jobs to return")
	parseArgs(cmd, args)

	items := filterAgentJobs(service.ListAgentJobs(), strings.TrimSpace(*status), *limit)
	writeJSON(items)
}

func createJob(service jobService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("job create", flag.ExitOnError)
	id := cmd.String("id", "", "job id")
	branch := cmd.String("branch", "", "branch id")
	kind := cmd.String("kind", "", "job kind")
	role := cmd.String("role", "", "job role")
	summary := cmd.String("summary", "", "job summary")
	source := cmd.String("source", "manual", "job source")
	surface := cmd.String("surface", string(runtime.SurfaceAPI), "job surface")
	stage := cmd.String("stage", string(runtime.StageDiscover), "job stage")
	ref := cmd.String("ref", "", "job ref")
	notes := cmd.String("notes", "", "comma-separated notes")
	payloadJSON := cmd.String("payload-json", "", "JSON object payload for the job")
	allowedPaths := cmd.String("allowed-paths", "", "comma-separated allowed_paths payload value")
	council := cmd.String("council", "", "comma-separated council providers (claude,openai,gemini)")
	councilPrimary := cmd.String("council-primary", "", "primary provider for council mode")
	parseArgs(cmd, args)

	payload, err := parseJobCreatePayloadInput(*payloadJSON, *allowedPaths, *council, *councilPrimary)
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now().UTC()
	jobID := strings.TrimSpace(*id)
	if jobID == "" {
		jobID = fmt.Sprintf("job_%d", now.UnixMilli())
	}
	item := runtime.AgentJob{
		JobID:     jobID,
		BranchID:  normalizeCLIRef(*branch),
		Kind:      strings.TrimSpace(*kind),
		Role:      strings.TrimSpace(*role),
		Summary:   strings.TrimSpace(*summary),
		Status:    "queued",
		Source:    strings.TrimSpace(*source),
		Surface:   runtime.Surface(strings.TrimSpace(*surface)),
		Stage:     runtime.Stage(strings.TrimSpace(*stage)),
		Notes:     splitCSV(*notes),
		Payload:   payload,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if value := strings.TrimSpace(*ref); value != "" {
		item.Ref = &value
	}
	if err := service.CreateAgentJob(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func parseJobCreatePayloadInput(payloadJSON string, allowedPaths string, councilProviders string, councilPrimary string) (map[string]any, error) {
	payload := map[string]any{}
	if strings.TrimSpace(payloadJSON) != "" {
		parsed, err := parseCLIJSONObject([]byte(payloadJSON))
		if err != nil {
			return nil, fmt.Errorf("parse --payload-json: %w", err)
		}
		payload = parsed
	}
	if paths := splitCSV(allowedPaths); len(paths) > 0 {
		normalized := make([]string, 0, len(paths))
		for _, path := range paths {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			normalized = append(normalized, path)
		}
		if len(normalized) > 0 {
			payload["allowed_paths"] = normalized
		}
	}
	normalizedCouncil, err := normalizeCouncilPayloadInput(councilProviders, councilPrimary)
	if err != nil {
		return nil, err
	}
	if normalizedCouncil != nil {
		payload["council"] = normalizedCouncil
	}
	if len(payload) == 0 {
		return nil, nil
	}
	return payload, nil
}

func normalizeCouncilPayloadInput(councilProviders string, councilPrimary string) (map[string]any, error) {
	providers, invalid := normalizeCouncilProvidersInput(councilProviders)
	if len(invalid) > 0 {
		return nil, fmt.Errorf("council providers must be drawn from claude, openai, or gemini: %s", strings.Join(invalid, ", "))
	}
	primary := strings.ToLower(strings.TrimSpace(councilPrimary))
	if primary != "" && !cliCouncilProviderAllowed(primary) {
		return nil, fmt.Errorf("council primary must be one of claude, openai, or gemini")
	}
	if len(providers) == 0 {
		if primary != "" {
			return nil, fmt.Errorf("council primary requires --council providers")
		}
		return nil, nil
	}
	if primary != "" && !containsCLIProvider(providers, primary) {
		providers = append([]string{primary}, providers...)
	}
	council := map[string]any{
		"providers": providers,
	}
	if primary != "" {
		council["primary_provider"] = primary
	}
	return council, nil
}

func normalizeCouncilProvidersInput(raw string) ([]string, []string) {
	items := []string{}
	invalid := []string{}
	seen := map[string]struct{}{}
	for _, item := range splitCSV(raw) {
		provider := strings.ToLower(strings.TrimSpace(item))
		if provider == "" {
			continue
		}
		if !cliCouncilProviderAllowed(provider) {
			invalid = append(invalid, provider)
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		items = append(items, provider)
	}
	return items, invalid
}

func cliCouncilProviderAllowed(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "claude", "openai", "gemini":
		return true
	default:
		return false
	}
}

func containsCLIProvider(items []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item)) == want {
			return true
		}
	}
	return false
}

func promoteJobs(service jobService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("job promote", flag.ExitOnError)
	limit := cmd.Int("limit", 3, "max derived jobs to promote")
	dispatch := cmd.Bool("dispatch", false, "dispatch promoted jobs immediately")
	parseArgs(cmd, args)

	result, err := service.PromoteContextJobs(*limit, *dispatch)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func updateJob(service jobService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("job update", flag.ExitOnError)
	id := cmd.String("id", "", "job id")
	status := cmd.String("status", "", "job status")
	notes := cmd.String("notes", "", "comma-separated notes")
	result := cmd.String("result", "", "comma-separated key=value result pairs")
	resultJSON := cmd.String("result-json", "", "JSON object result payload")
	resultFile := cmd.String("result-file", "", "path to a JSON object result payload")
	parseArgs(cmd, args)

	parsedResult, err := parseJobResultInput(*result, *resultJSON, *resultFile)
	if err != nil {
		log.Fatal(err)
	}
	item, err := service.UpdateAgentJob(strings.TrimSpace(*id), strings.TrimSpace(*status), splitCSV(*notes), parsedResult)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func reportJob(service jobService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("job report", flag.ExitOnError)
	id := cmd.String("id", "", "job id")
	status := cmd.String("status", "", "terminal job status")
	notes := cmd.String("notes", "", "comma-separated notes")
	result := cmd.String("result", "", "comma-separated key=value result pairs")
	resultJSON := cmd.String("result-json", "", "JSON object result payload")
	resultFile := cmd.String("result-file", "", "path to a JSON object result payload")
	parseArgs(cmd, args)

	parsedResult, err := parseJobResultInput(*result, *resultJSON, *resultFile)
	if err != nil {
		log.Fatal(err)
	}
	item, err := service.ReportAgentJob(strings.TrimSpace(*id), strings.TrimSpace(*status), splitCSV(*notes), parsedResult)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func dispatchJob(service jobService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("job dispatch", flag.ExitOnError)
	id := cmd.String("id", "", "job id")
	parseArgs(cmd, args)
	result, err := service.DispatchAgentJob(strings.TrimSpace(*id))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(result)
}

func jobPacket(service jobService, args []string) {
	cmd := flag.NewFlagSet("job packet", flag.ExitOnError)
	id := cmd.String("id", "", "job id")
	parseArgs(cmd, args)
	packet, err := service.AgentRunPacket(strings.TrimSpace(*id))
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(packet)
}

func parseJobResultInput(resultPairs string, resultJSON string, resultFile string) (map[string]any, error) {
	if strings.TrimSpace(resultJSON) != "" && strings.TrimSpace(resultFile) != "" {
		return nil, fmt.Errorf("job result accepts only one of --result-json or --result-file")
	}

	result := map[string]any{}
	switch {
	case strings.TrimSpace(resultJSON) != "":
		parsed, err := parseCLIJSONObject([]byte(resultJSON))
		if err != nil {
			return nil, fmt.Errorf("parse --result-json: %w", err)
		}
		result = parsed
	case strings.TrimSpace(resultFile) != "":
		raw, err := os.ReadFile(strings.TrimSpace(resultFile))
		if err != nil {
			return nil, fmt.Errorf("read --result-file: %w", err)
		}
		parsed, err := parseCLIJSONObject(raw)
		if err != nil {
			return nil, fmt.Errorf("parse --result-file: %w", err)
		}
		result = parsed
	}

	for key, value := range splitPairs(resultPairs) {
		result[key] = value
	}
	return result, nil
}

func parseCLIJSONObject(raw []byte) (map[string]any, error) {
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	if item == nil {
		return map[string]any{}, nil
	}
	return item, nil
}

func filterAgentJobs(items []runtime.AgentJob, status string, limit int) []runtime.AgentJob {
	filtered := make([]runtime.AgentJob, 0, len(items))
	for _, item := range items {
		if status != "" {
			if status == "open" {
				if item.Status != "queued" && item.Status != "running" {
					continue
				}
			} else if item.Status != status {
				continue
			}
		}
		filtered = append(filtered, item)
		if limit > 0 && len(filtered) >= limit {
			break
		}
	}
	return filtered
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type quickWorkService interface {
	authStatusProvider
	Adapters() runtime.AdapterSummary
	Providers() []runtime.ProviderSummary
	SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error)
	SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error)
	CreateAgentJob(item runtime.AgentJob) error
	PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error)
	DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error)
	ListAgentJobs() []runtime.AgentJob
}

type quickWorkCommandRunner struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	exec   func(path string, args []string, stdout io.Writer, stderr io.Writer) error
}

type quickWorkSeed struct {
	Summary      string
	Role         string
	AllowedPaths []string
	Source       string
}

type quickWorkTarget struct {
	Seed        *quickWorkSeed
	DispatchJob *runtime.AgentJob
	WaitJob     *runtime.AgentJob
}

var quickWorkRunner = quickWorkCommandRunner{
	stdin:  os.Stdin,
	stdout: os.Stdout,
	stderr: os.Stderr,
	exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
		cmd := exec.Command(path, args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	},
}

func runQuickWork(service quickWorkService, args []string) {
	cmd := flag.NewFlagSet("quickwork", flag.ExitOnError)
	summary := cmd.String("summary", "", "one-line job summary")
	role := cmd.String("role", "", "job role override (implementer|verifier|planner|designer)")
	allowedPaths := cmd.String("allowed-paths", "", "comma-separated allowed_paths payload value")
	payloadJSON := cmd.String("payload-json", "", "JSON object payload for the job")
	council := cmd.String("council", "", "comma-separated council providers (claude,openai,gemini)")
	councilPrimary := cmd.String("council-primary", "", "primary provider for council mode")
	statusOnly := cmd.Bool("status", false, "show orchestrator status")
	down := cmd.Bool("down", false, "stop background workers")
	upOnly := cmd.Bool("up", false, "start/reuse daemon and workers only")
	loop := cmd.Bool("loop", false, "keep resolving and dispatching one bounded next job at a time")
	poll := cmd.Duration("poll", 10*time.Second, "poll interval while waiting for loop progress")
	exitOnIdle := cmd.Duration("exit-on-idle", 10*time.Minute, "stop loop mode after this much idle time without a new suggested action (0 disables)")
	parseArgs(cmd, args)

	orchestrator := filepath.Join(repoRootForLocal(), "scripts", "worker_orchestrator.sh")
	switch {
	case *statusOnly:
		mustRunQuickWork(orchestrator, []string{"status"})
		return
	case *down:
		mustRunQuickWork(orchestrator, []string{"down"})
		return
	case *upOnly:
		mustRunQuickWork(orchestrator, []string{"up"})
		return
	}

	requireCLIWriteAuth(service)

	allowedPathItems := normalizeQuickWorkAllowedPaths(*allowedPaths)
	payload, err := parseJobCreatePayloadInput(*payloadJSON, *allowedPaths, *council, *councilPrimary)
	if err != nil {
		log.Fatal(err)
	}

	var initialTarget *quickWorkTarget
	var initialBootstrap *runtime.SessionBootstrapPacket
	if strings.TrimSpace(*summary) != "" {
		resolvedRole, err := normalizeQuickWorkRole(strings.TrimSpace(*role))
		if err != nil {
			log.Fatal(err)
		}
		initialTarget = &quickWorkTarget{Seed: &quickWorkSeed{
			Summary:      strings.TrimSpace(*summary),
			Role:         resolvedRole,
			AllowedPaths: allowedPathItems,
			Source:       "founder.manual",
		}}
	} else {
		packet, err := service.SessionBootstrap(false)
		if err == nil {
			initialBootstrap = &packet
			openJobs := listOpenAgentJobs(service.ListAgentJobs(), 1)
			if target, ok := quickWorkTargetFromBootstrap(packet, openJobs, strings.TrimSpace(*role), allowedPathItems); ok {
				initialTarget = &target
			}
		}
		if initialTarget == nil {
			target, ok, promoteErr := quickWorkPromoteTarget(service)
			if promoteErr != nil {
				log.Fatal(promoteErr)
			}
			if ok {
				initialTarget = &target
			}
		}
		if initialTarget == nil && !*loop {
			resolvedSummary, resolvedRole, resolvedAllowedPaths, promptErr := promptQuickWork(strings.TrimSpace(*role), strings.TrimSpace(*allowedPaths))
			if promptErr != nil {
				log.Fatal(promptErr)
			}
			resolvedRole, err = normalizeQuickWorkRole(resolvedRole)
			if err != nil {
				log.Fatal(err)
			}
			initialTarget = &quickWorkTarget{Seed: &quickWorkSeed{
				Summary:      resolvedSummary,
				Role:         resolvedRole,
				AllowedPaths: normalizeQuickWorkAllowedPaths(resolvedAllowedPaths),
				Source:       "founder.manual",
			}}
		}
	}

	if initialTarget == nil && !*loop {
		log.Fatal("quickwork could not derive a next action; try layer-osctl next or provide --summary")
	}

	if !*loop {
		switch {
		case initialTarget.Seed != nil:
			dispatch := quickWorkSubmit(service, orchestrator, *initialTarget.Seed, payload)
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "submitted %s (%s)\n", dispatch.Job.JobID, dispatch.Job.Role)
		case initialTarget.DispatchJob != nil:
			dispatch := quickWorkDispatchExisting(service, orchestrator, *initialTarget.DispatchJob)
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "submitted %s (%s)\n", dispatch.Job.JobID, dispatch.Job.Role)
		case initialTarget.WaitJob != nil:
			status := strings.TrimSpace(initialTarget.WaitJob.Status)
			if status == "succeeded" || status == "failed" || status == "canceled" {
				_, _ = fmt.Fprintf(quickWorkRunner.stdout, "already completed %s (%s)\n", initialTarget.WaitJob.JobID, status)
			} else {
				_, _ = fmt.Fprintf(quickWorkRunner.stdout, "already running %s (%s)\n", initialTarget.WaitJob.JobID, initialTarget.WaitJob.Role)
			}
		default:
			log.Fatal("quickwork could not derive a runnable target")
		}
		return
	}

	if err := runQuickWorkLoop(service, orchestrator, initialTarget, payload, initialBootstrap, quickWorkLoopConfig{
		pollInterval: positiveDuration(*poll, 10*time.Second),
		exitOnIdle:   *exitOnIdle,
		roleOverride: strings.TrimSpace(*role),
		allowedPaths: allowedPathItems,
	}); err != nil {
		log.Fatal(err)
	}
}

type quickWorkLoopConfig struct {
	pollInterval time.Duration
	exitOnIdle   time.Duration
	roleOverride string
	allowedPaths []string
}

type quickWorkAgendaState struct {
	OpenCount int
	TopText   string
}

func runQuickWorkLoop(service quickWorkService, orchestrator string, initialTarget *quickWorkTarget, initialPayload map[string]any, initialBootstrap *runtime.SessionBootstrapPacket, config quickWorkLoopConfig) error {
	seenSeeds := map[string]bool{}
	var idleSince time.Time
	target := initialTarget
	payload := cloneQuickWorkPayload(initialPayload)
	var lastBootstrap *runtime.SessionBootstrapPacket
	var agendaState quickWorkAgendaState
	if initialBootstrap != nil {
		lastBootstrap = initialBootstrap
		announceQuickWorkAgenda(*initialBootstrap, &agendaState)
	}

	for {
		if target == nil {
			packet, err := service.SessionBootstrap(false)
			if err != nil {
				return err
			}
			lastBootstrap = &packet
			announceQuickWorkAgenda(packet, &agendaState)
			openJobs := listOpenAgentJobs(service.ListAgentJobs(), 1)
			nextTarget, ok := quickWorkTargetFromBootstrap(packet, openJobs, config.roleOverride, config.allowedPaths)
			if !ok || (nextTarget.Seed != nil && quickWorkSeedSeen(seenSeeds, *nextTarget.Seed)) {
				if promotedTarget, promoted, promoteErr := quickWorkPromoteTarget(service); promoteErr != nil {
					return promoteErr
				} else if promoted {
					target = &promotedTarget
					payload = nil
					idleSince = time.Time{}
					continue
				}
				if config.exitOnIdle <= 0 {
					return nil
				}
				if idleSince.IsZero() {
					idleSince = time.Now()
				}
				if time.Since(idleSince) >= config.exitOnIdle {
					_, _ = fmt.Fprintln(quickWorkRunner.stdout, "idle exit reached")
					if lastBootstrap != nil {
						if err := finishIdleQuickWorkSession(service, *lastBootstrap); err != nil {
							_, _ = fmt.Fprintf(quickWorkRunner.stdout, "session finish failed: %v\n", err)
						}
					}
					return nil
				}
				time.Sleep(config.pollInterval)
				continue
			}
			target = &nextTarget
			payload = nil
			idleSince = time.Time{}
		}

		switch {
		case target.Seed != nil:
			dispatch := quickWorkSubmit(service, orchestrator, *target.Seed, payload)
			seenSeeds[quickWorkSeedKey(*target.Seed)] = true
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "submitted %s (%s)\n", dispatch.Job.JobID, dispatch.Job.Role)
			finalJob, err := waitForQuickWorkTerminalJob(service, dispatch.Job.JobID, config.pollInterval)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "completed %s (%s)\n", finalJob.JobID, finalJob.Status)
		case target.DispatchJob != nil:
			dispatch := quickWorkDispatchExisting(service, orchestrator, *target.DispatchJob)
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "submitted %s (%s)\n", dispatch.Job.JobID, dispatch.Job.Role)
			finalJob, err := waitForQuickWorkTerminalJob(service, dispatch.Job.JobID, config.pollInterval)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(quickWorkRunner.stdout, "completed %s (%s)\n", finalJob.JobID, finalJob.Status)
		case target.WaitJob != nil:
			if status := strings.TrimSpace(target.WaitJob.Status); status == "succeeded" || status == "failed" || status == "canceled" {
				_, _ = fmt.Fprintf(quickWorkRunner.stdout, "already completed %s (%s)\n", target.WaitJob.JobID, target.WaitJob.Status)
			} else {
				finalJob, err := waitForQuickWorkTerminalJob(service, target.WaitJob.JobID, config.pollInterval)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(quickWorkRunner.stdout, "completed %s (%s)\n", finalJob.JobID, finalJob.Status)
			}
		default:
			return nil
		}
		target = nil
		payload = nil
	}
}

func quickWorkSubmit(service quickWorkService, orchestrator string, seed quickWorkSeed, payload map[string]any) runtime.AgentDispatchResult {
	mustRunQuickWork(orchestrator, quickWorkUpArgs(seed.Role))
	now := time.Now().UTC()
	finalPayload, autoCouncil := prepareQuickWorkPayload(service, seed, payload)
	if autoCouncil != nil {
		_, _ = fmt.Fprintf(quickWorkRunner.stdout, "council auto[%s] primary=%s\n", strings.Join(autoCouncil.Providers, ","), autoCouncil.Primary)
	}
	item := runtime.AgentJob{
		JobID:     fmt.Sprintf("job_%d", now.UnixMilli()),
		Kind:      defaultQuickWorkKind(seed.Role),
		Role:      seed.Role,
		Summary:   strings.TrimSpace(seed.Summary),
		Status:    "queued",
		Source:    quickWorkSource(seed.Source),
		Surface:   runtime.SurfaceAPI,
		Stage:     defaultQuickWorkStage(seed.Role),
		Payload:   finalPayload,
		Notes:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.CreateAgentJob(item); err != nil {
		log.Fatal(err)
	}
	dispatch, err := service.DispatchAgentJob(item.JobID)
	if err != nil {
		log.Fatal(err)
	}
	return dispatch
}

func quickWorkDispatchExisting(service quickWorkService, orchestrator string, job runtime.AgentJob) runtime.AgentDispatchResult {
	mustRunQuickWork(orchestrator, quickWorkUpArgs(job.Role))
	dispatch, err := service.DispatchAgentJob(job.JobID)
	if err != nil {
		log.Fatal(err)
	}
	return dispatch
}

type quickWorkJobReader interface {
	ListAgentJobs() []runtime.AgentJob
}

type quickWorkAutoCouncil struct {
	Providers []string
	Primary   string
}

func waitForQuickWorkTerminalJob(service quickWorkJobReader, jobID string, pollInterval time.Duration) (runtime.AgentJob, error) {
	interval := positiveDuration(pollInterval, 10*time.Second)
	for {
		for _, job := range service.ListAgentJobs() {
			if strings.TrimSpace(job.JobID) != strings.TrimSpace(jobID) {
				continue
			}
			switch strings.TrimSpace(job.Status) {
			case "succeeded", "failed", "canceled":
				return job, nil
			}
			break
		}
		time.Sleep(interval)
	}
}

func quickWorkSeedFromBootstrap(packet runtime.SessionBootstrapPacket, roleOverride string, allowedPaths []string) (quickWorkSeed, bool) {
	target, ok := quickWorkTargetFromBootstrap(packet, nil, roleOverride, allowedPaths)
	if !ok || target.Seed == nil {
		return quickWorkSeed{}, false
	}
	return *target.Seed, true
}

func quickWorkTargetFromBootstrap(packet runtime.SessionBootstrapPacket, items []runtime.AgentJob, roleOverride string, allowedPaths []string) (quickWorkTarget, bool) {
	for _, candidate := range nextActionCandidates(packet, items) {
		switch candidate.Kind {
		case nextActionRoute:
			role, ok := quickWorkRoleForTargetLane(candidate.TargetLane, roleOverride)
			if !ok {
				continue
			}
			source := strings.TrimSpace(candidate.Source)
			if source == "" {
				source = "quickwork.loop"
			}
			seed := quickWorkSeed{
				Summary:      candidate.Summary,
				Role:         role,
				AllowedPaths: append([]string{}, allowedPaths...),
				Source:       source,
			}
			return quickWorkTarget{Seed: &seed}, true
		case nextActionAgenda:
			role, ok := quickWorkRoleForTargetLane(candidate.TargetLane, roleOverride)
			if !ok {
				continue
			}
			source := strings.TrimSpace(candidate.Source)
			if source == "" {
				source = "review_room.open"
			}
			seed := quickWorkSeed{
				Summary:      candidate.Summary,
				Role:         role,
				AllowedPaths: append([]string{}, allowedPaths...),
				Source:       source,
			}
			return quickWorkTarget{Seed: &seed}, true
		case nextActionPrimary:
			role := "implementer"
			if strings.TrimSpace(roleOverride) != "" {
				var err error
				role, err = normalizeQuickWorkRole(roleOverride)
				if err != nil {
					continue
				}
			}
			seed := quickWorkSeed{
				Summary:      candidate.Summary,
				Role:         role,
				AllowedPaths: append([]string{}, allowedPaths...),
				Source:       "quickwork.loop",
			}
			return quickWorkTarget{Seed: &seed}, true
		case nextActionProposal:
			role := "planner"
			if strings.TrimSpace(roleOverride) != "" {
				var err error
				role, err = normalizeQuickWorkRole(roleOverride)
				if err != nil {
					continue
				}
			}
			source := strings.TrimSpace(candidate.Source)
			if source == "" {
				source = "quickwork.loop"
			}
			seed := quickWorkSeed{
				Summary:      candidate.Summary,
				Role:         role,
				AllowedPaths: append([]string{}, allowedPaths...),
				Source:       source,
			}
			return quickWorkTarget{Seed: &seed}, true
		case nextActionOpenJob:
			if candidate.Job == nil {
				continue
			}
			job := *candidate.Job
			switch strings.TrimSpace(job.Status) {
			case "queued", "failed":
				return quickWorkTarget{DispatchJob: &job}, true
			case "running":
				return quickWorkTarget{WaitJob: &job}, true
			}
		}
	}
	return quickWorkTarget{}, false
}

func quickWorkPromoteTarget(service quickWorkService) (quickWorkTarget, bool, error) {
	result, err := service.PromoteContextJobs(1, true)
	if err != nil {
		return quickWorkTarget{}, false, err
	}
	target, message, ok := quickWorkTargetFromPromotion(result)
	if !ok {
		return quickWorkTarget{}, false, nil
	}
	if strings.TrimSpace(message) != "" {
		_, _ = fmt.Fprintln(quickWorkRunner.stdout, message)
	}
	return target, true, nil
}

func quickWorkTargetFromPromotion(result runtime.AgentJobPromotionResult) (quickWorkTarget, string, bool) {
	for _, item := range result.Items {
		if item.Job == nil {
			continue
		}
		job := *item.Job
		summary := strings.TrimSpace(item.Summary)
		if summary == "" {
			summary = strings.TrimSpace(job.Summary)
		}
		message := ""
		if summary != "" {
			message = fmt.Sprintf("promoted[%s]: %s", strings.TrimSpace(item.SourceKind), summary)
		}
		switch strings.TrimSpace(item.Status) {
		case "dispatched":
			return quickWorkTarget{WaitJob: &job}, message, true
		case "created", "existing":
			switch strings.TrimSpace(job.Status) {
			case "queued", "failed":
				return quickWorkTarget{DispatchJob: &job}, message, true
			case "running":
				return quickWorkTarget{WaitJob: &job}, message, true
			}
		}
	}
	return quickWorkTarget{}, "", false
}

func quickWorkRoleForTargetLane(targetLane string, override string) (string, bool) {
	if strings.TrimSpace(override) != "" {
		role, err := normalizeQuickWorkRole(override)
		return role, err == nil
	}
	switch strings.TrimSpace(targetLane) {
	case "planner", "implementer", "verifier", "designer":
		return strings.TrimSpace(targetLane), true
	default:
		return "", false
	}
}

func quickWorkSeedSeen(seen map[string]bool, seed quickWorkSeed) bool {
	return seen[quickWorkSeedKey(seed)]
}

func quickWorkSeedKey(seed quickWorkSeed) string {
	return strings.TrimSpace(seed.Role) + "|" + strings.TrimSpace(seed.Summary)
}

func normalizeQuickWorkRole(role string) (string, error) {
	switch strings.TrimSpace(role) {
	case "", "implementer":
		return "implementer", nil
	case "verifier":
		return "verifier", nil
	case "planner":
		return "planner", nil
	case "designer":
		return "designer", nil
	default:
		return "", fmt.Errorf("quickwork role must be implementer, verifier, planner, or designer")
	}
}

func quickWorkUpArgs(role string) []string {
	switch strings.TrimSpace(role) {
	case "planner":
		return []string{"up", "--roles", "implementer,verifier,planner"}
	case "designer":
		return []string{"up", "--roles", "implementer,verifier,designer"}
	default:
		return []string{"up"}
	}
}

func listOpenAgentJobs(items []runtime.AgentJob, limit int) []runtime.AgentJob {
	open := make([]runtime.AgentJob, 0, len(items))
	for _, item := range items {
		switch strings.TrimSpace(item.Status) {
		case "queued", "running", "failed":
			open = append(open, item)
		}
		if limit > 0 && len(open) >= limit {
			break
		}
	}
	return open
}

func finishIdleQuickWorkSession(service quickWorkService, packet runtime.SessionBootstrapPacket) error {
	note := "quickwork idle exit"
	focus := strings.TrimSpace(packet.Knowledge.CurrentFocus)
	if focus == "" {
		focus = "Quickwork idle exit"
	}
	var goal *string
	if packet.Knowledge.CurrentGoal != nil {
		value := strings.TrimSpace(*packet.Knowledge.CurrentGoal)
		if value != "" {
			goal = &value
		}
	}
	_, err := service.SessionFinish(runtime.SessionFinishInput{
		CurrentFocus: focus,
		CurrentGoal:  goal,
		NextSteps:    append([]string{}, packet.Knowledge.NextSteps...),
		OpenRisks:    append([]string{}, packet.Knowledge.OpenRisks...),
		Note:         &note,
	})
	return err
}

func announceQuickWorkAgenda(packet runtime.SessionBootstrapPacket, state *quickWorkAgendaState) {
	count, text, _ := reviewAgendaSummary(packet)
	trimmed := strings.TrimSpace(text)
	if state != nil && state.OpenCount == count && strings.TrimSpace(state.TopText) == trimmed {
		return
	}
	switch {
	case count > 0 && trimmed != "":
		_, _ = fmt.Fprintf(quickWorkRunner.stdout, "agenda[%d]: %s\n", count, trimmed)
	case state != nil && state.OpenCount > 0 && count == 0:
		_, _ = fmt.Fprintln(quickWorkRunner.stdout, "agenda cleared")
	}
	if state != nil {
		state.OpenCount = count
		state.TopText = trimmed
	}
}

func prepareQuickWorkPayload(service quickWorkService, seed quickWorkSeed, payload map[string]any) (map[string]any, *quickWorkAutoCouncil) {
	merged := mergeQuickWorkPayload(payload, seed.AllowedPaths)
	council := deriveQuickWorkAutoCouncil(service, seed, merged)
	if council == nil {
		return merged, nil
	}
	if merged == nil {
		merged = map[string]any{}
	}
	merged["council"] = map[string]any{
		"providers": append([]string{}, council.Providers...),
	}
	if strings.TrimSpace(council.Primary) != "" {
		merged["council"].(map[string]any)["primary_provider"] = council.Primary
	}
	return merged, council
}

func deriveQuickWorkAutoCouncil(service quickWorkService, seed quickWorkSeed, payload map[string]any) *quickWorkAutoCouncil {
	if quickWorkPayloadHasCouncil(payload) || !quickWorkAutoCouncilEligible(seed) {
		return nil
	}
	providers := quickWorkAutoCouncilProviders(service.Adapters(), service.Providers(), seed.Role)
	if len(providers) < 2 {
		return nil
	}
	primary := providers[0]
	return &quickWorkAutoCouncil{
		Providers: providers,
		Primary:   primary,
	}
}

func quickWorkPayloadHasCouncil(payload map[string]any) bool {
	if payload == nil {
		return false
	}
	_, ok := payload["council"]
	return ok
}

func quickWorkAutoCouncilEligible(seed quickWorkSeed) bool {
	if strings.TrimSpace(seed.Source) == "founder.manual" {
		return false
	}
	switch strings.TrimSpace(seed.Role) {
	case "planner", "designer":
		return true
	default:
		return false
	}
}

func quickWorkAutoCouncilProviders(adapters runtime.AdapterSummary, providers []runtime.ProviderSummary, role string) []string {
	available := []string{}
	switch strings.TrimSpace(adapters.GatewaySemantics) {
	case "direct_llm":
		gateway := strings.ToLower(strings.TrimSpace(adapters.Gateway))
		if gateway == "multi" {
			for _, summary := range providers {
				if quickWorkProviderDispatchReady(summary) {
					available = appendUniqueProvider(available, summary.Provider)
				}
			}
		} else if cliCouncilProviderAllowed(gateway) && quickWorkDirectProviderReady(gateway) {
			available = appendUniqueProvider(available, gateway)
		}
	case "dispatch":
		for _, summary := range providers {
			if quickWorkProviderDispatchReady(summary) {
				available = appendUniqueProvider(available, summary.Provider)
			}
		}
	}
	ordered := orderCouncilProviders(available, councilPreferenceForRole(role))
	if len(ordered) > 2 {
		return append([]string{}, ordered[:2]...)
	}
	return ordered
}

func quickWorkProviderDispatchReady(summary runtime.ProviderSummary) bool {
	if !summary.DispatchEnabled {
		return false
	}
	provider := strings.ToLower(strings.TrimSpace(summary.Provider))
	switch strings.TrimSpace(summary.Semantics) {
	case "direct_llm":
		gateway := strings.ToLower(strings.TrimSpace(summary.GatewayAdapter))
		if gateway == "multi" {
			return cliCouncilProviderAllowed(provider) && quickWorkDirectProviderReady(provider)
		}
		return gateway == provider && quickWorkDirectProviderReady(provider)
	case "dispatch":
		return summary.Endpoint != nil && strings.TrimSpace(*summary.Endpoint) != ""
	default:
		return false
	}
}

func quickWorkDirectProviderReady(provider string) bool {
	return runtime.ProviderCredentialReady(provider)
}

func councilPreferenceForRole(role string) []string {
	items := []string{}
	if preferred := quickWorkRoleProviderBinding(role); cliCouncilProviderAllowed(preferred) {
		items = append(items, preferred)
	}
	for _, provider := range []string{"claude", "openai", "gemini"} {
		items = appendUniqueProvider(items, provider)
	}
	return items
}

func quickWorkRoleProviderBinding(role string) string {
	for _, item := range splitCSV(os.Getenv("LAYER_OS_AGENT_ROLE_PROVIDERS")) {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[0]) != strings.TrimSpace(role) {
			continue
		}
		return strings.ToLower(strings.TrimSpace(parts[1]))
	}
	return ""
}

func orderCouncilProviders(available []string, preferred []string) []string {
	ordered := []string{}
	for _, provider := range preferred {
		if containsCLIProvider(available, provider) {
			ordered = appendUniqueProvider(ordered, provider)
		}
	}
	for _, provider := range available {
		ordered = appendUniqueProvider(ordered, provider)
	}
	return ordered
}

func appendUniqueProvider(items []string, provider string) []string {
	if !containsCLIProvider(items, provider) {
		return append(items, strings.ToLower(strings.TrimSpace(provider)))
	}
	return items
}

func defaultQuickWorkKind(role string) string {
	switch strings.TrimSpace(role) {
	case "verifier":
		return "verify"
	case "planner":
		return "plan"
	case "designer":
		return "design"
	default:
		return "implement"
	}
}

func defaultQuickWorkStage(role string) runtime.Stage {
	switch strings.TrimSpace(role) {
	case "verifier":
		return runtime.StageVerify
	case "planner":
		return runtime.StageDiscover
	case "designer":
		return runtime.StageExperience
	default:
		return runtime.StageCompose
	}
}

func normalizeQuickWorkAllowedPaths(raw string) []string {
	items := splitCSV(raw)
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return normalized
}

func mergeQuickWorkPayload(payload map[string]any, allowedPaths []string) map[string]any {
	merged := cloneQuickWorkPayload(payload)
	if len(allowedPaths) > 0 {
		if merged == nil {
			merged = map[string]any{}
		}
		merged["allowed_paths"] = append([]string{}, allowedPaths...)
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func cloneQuickWorkPayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	cloned := map[string]any{}
	for key, value := range payload {
		cloned[key] = value
	}
	return cloned
}

func quickWorkSource(source string) string {
	if strings.TrimSpace(source) == "" {
		return "founder.manual"
	}
	return strings.TrimSpace(source)
}

func mustRunQuickWork(path string, args []string) {
	if err := quickWorkRunner.exec(path, args, quickWorkRunner.stdout, quickWorkRunner.stderr); err != nil {
		log.Fatal(err)
	}
}

func promptQuickWork(defaultRole string, defaultAllowedPaths string) (summary string, role string, allowedPaths string, err error) {
	reader := bufio.NewReader(quickWorkRunner.stdin)
	if strings.TrimSpace(defaultRole) == "" {
		defaultRole = "implementer"
	}
	if _, err = fmt.Fprintln(quickWorkRunner.stderr, "Layer OS quick work"); err != nil {
		return "", "", "", err
	}
	if _, err = fmt.Fprintln(quickWorkRunner.stderr, "-------------------"); err != nil {
		return "", "", "", err
	}
	if summary, err = quickWorkPromptLine(reader, "할 일 요약", ""); err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(summary) == "" {
		return "", "", "", fmt.Errorf("canceled: summary is empty")
	}
	if role, err = quickWorkPromptLine(reader, "역할", defaultRole); err != nil {
		return "", "", "", err
	}
	if allowedPaths, err = quickWorkPromptLine(reader, "허용 경로", defaultAllowedPaths); err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(allowedPaths) == "" && strings.TrimSpace(role) == "implementer" {
		allowedPaths = "cmd/layer-osctl/,scripts/,docs/"
	}
	return strings.TrimSpace(summary), strings.TrimSpace(role), strings.TrimSpace(allowedPaths), nil
}

func quickWorkPromptLine(reader *bufio.Reader, label string, defaultValue string) (string, error) {
	if defaultValue != "" {
		if _, err := fmt.Fprintf(quickWorkRunner.stderr, "%s [%s]: ", label, defaultValue); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(quickWorkRunner.stderr, "%s: ", label); err != nil {
			return "", err
		}
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = defaultValue
	}
	return line, nil
}

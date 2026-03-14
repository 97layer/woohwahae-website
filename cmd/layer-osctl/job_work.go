package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"layer-os/internal/runtime"
)

type jobWorkConfig struct {
	command        string
	shell          string
	roles          map[string]bool
	pollInterval   time.Duration
	once           bool
	dispatchQueued bool
	limit          int
	workRoot       string
	repoRoot       string
	stopOnFailure  bool
	idleExitAfter  time.Duration
	packetFile     string
}

type jobWorkSummary struct {
	Scanned         int      `json:"scanned"`
	Dispatched      int      `json:"dispatched"`
	Claimed         int      `json:"claimed"`
	Completed       int      `json:"completed"`
	Failed          int      `json:"failed"`
	Skipped         int      `json:"skipped"`
	IdlePolls       int      `json:"idle_polls,omitempty"`
	ProcessedJobIDs []string `json:"processed_job_ids"`
	Notes           []string `json:"notes"`
}

type workerRunContext struct {
	packetPath string
	promptPath string
	resultPath string
	stdoutPath string
	stderrPath string
	jobDir     string
}

func workJobs(service jobService, args []string) {
	cmd := flag.NewFlagSet("job work", flag.ExitOnError)
	command := cmd.String("command", "", "shell command to run for each claimed packet-ready job")
	shell := cmd.String("shell", os.Getenv("SHELL"), "shell used to execute --command")
	roles := cmd.String("roles", "", "comma-separated job roles to process (default: all packet-ready roles)")
	poll := cmd.Duration("poll", 30*time.Second, "poll interval in loop mode")
	once := cmd.Bool("once", false, "process available work once and exit")
	dispatchQueued := cmd.Bool("dispatch-queued", true, "dispatch queued jobs before looking for packet-ready work")
	limit := cmd.Int("limit", 1, "max jobs to process per poll")
	workRoot := cmd.String("work-root", filepath.Join(os.TempDir(), "layer-os-job-work"), "directory for packet/prompt/result artifacts")
	repoRoot := cmd.String("repo-root", repoRootForLocal(), "working directory for the external command")
	stopOnFailure := cmd.Bool("stop-on-failure", false, "exit after the first worker/report failure")
	idleExitAfter := cmd.Duration("idle-exit-after", 0, "exit after this much idle time in loop mode (0 keeps watching)")
	packetFile := cmd.String("packet-file", "", "offline AgentRunPacket JSON file for wrapper rehearsal without daemon writes")
	parseArgs(cmd, args)

	if strings.TrimSpace(*command) == "" {
		log.Fatal("job work requires --command")
	}
	resolvedShell := strings.TrimSpace(*shell)
	if resolvedShell == "" {
		resolvedShell = "/bin/sh"
	}
	resolvedRepoRoot := strings.TrimSpace(*repoRoot)
	if resolvedRepoRoot == "" {
		resolvedRepoRoot = repoRootForLocal()
	}
	resolvedRepoRoot = filepath.Clean(resolvedRepoRoot)
	if err := os.MkdirAll(strings.TrimSpace(*workRoot), 0o755); err != nil {
		log.Fatal(err)
	}
	resolvedPacketFile := filepath.Clean(strings.TrimSpace(*packetFile))
	if resolvedPacketFile != "." && strings.TrimSpace(*packetFile) != "" {
		if *dispatchQueued {
			log.Fatal("job work --packet-file does not support --dispatch-queued")
		}
		if !*once {
			log.Fatal("job work --packet-file requires --once")
		}
	} else {
		resolvedPacketFile = ""
		requireCLIWriteAuth(service)
	}

	config := jobWorkConfig{
		command:        strings.TrimSpace(*command),
		shell:          resolvedShell,
		roles:          workerRoleFilter(*roles),
		pollInterval:   positiveDuration(*poll, 30*time.Second),
		once:           *once,
		dispatchQueued: *dispatchQueued,
		limit:          positiveInt(*limit, 1),
		workRoot:       filepath.Clean(strings.TrimSpace(*workRoot)),
		repoRoot:       resolvedRepoRoot,
		stopOnFailure:  *stopOnFailure,
		idleExitAfter:  *idleExitAfter,
		packetFile:     resolvedPacketFile,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	summary, err := runJobWorkerLoop(ctx, service, config)
	if err != nil {
		if summary.Scanned > 0 || len(summary.Notes) > 0 {
			writeJSON(summary)
		}
		log.Fatal(err)
	}
	writeJSON(summary)
}

func runJobWorkerLoop(ctx context.Context, service jobService, config jobWorkConfig) (jobWorkSummary, error) {
	if config.packetFile != "" {
		return runJobWorkerPacketFile(ctx, config)
	}
	summary := jobWorkSummary{ProcessedJobIDs: []string{}, Notes: []string{}}
	lastProgress := time.Now()
	for {
		scan, err := runJobWorkerScan(ctx, service, config)
		summary = mergeJobWorkSummary(summary, scan)
		if err != nil {
			return summary, err
		}
		if scan.Completed > 0 || scan.Failed > 0 || scan.Claimed > 0 || scan.Dispatched > 0 {
			lastProgress = time.Now()
		}
		if config.once {
			return summary, nil
		}
		if config.idleExitAfter > 0 && time.Since(lastProgress) >= config.idleExitAfter {
			summary.Notes = append(summary.Notes, "idle_exit_after_reached")
			if err := finishIdleWorkerSession(service); err != nil {
				summary.Notes = append(summary.Notes, "session_finish_failed:"+err.Error())
			} else {
				summary.Notes = append(summary.Notes, "session_finished")
			}
			return summary, nil
		}
		summary.IdlePolls++
		select {
		case <-ctx.Done():
			summary.Notes = append(summary.Notes, "worker_loop_stopped")
			return summary, nil
		case <-time.After(config.pollInterval):
		}
	}
}

func runJobWorkerPacketFile(ctx context.Context, config jobWorkConfig) (jobWorkSummary, error) {
	packet, err := readWorkerPacketFile(config.packetFile)
	if err != nil {
		return jobWorkSummary{}, err
	}
	result, err := runSingleOfflineWorkerPacket(ctx, config, packet)
	summary := jobWorkSummary{
		Scanned:         1,
		Claimed:         1,
		ProcessedJobIDs: []string{packet.Job.JobID},
		Notes:           []string{"packet_file_mode"},
	}
	if result.ReportStatus == "succeeded" {
		summary.Completed = 1
	} else {
		summary.Failed = 1
	}
	return summary, err
}

func runJobWorkerScan(ctx context.Context, service jobService, config jobWorkConfig) (jobWorkSummary, error) {
	summary := jobWorkSummary{ProcessedJobIDs: []string{}, Notes: []string{}}
	jobs := append([]runtime.AgentJob{}, service.ListAgentJobs()...)
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		if config.limit > 0 && summary.Claimed >= config.limit {
			break
		}
		summary.Scanned++
		if !workerRoleAllowed(config.roles, job.Role) {
			summary.Skipped++
			continue
		}
		current := job
		if current.Status == "queued" && config.dispatchQueued {
			dispatch, err := service.DispatchAgentJob(current.JobID)
			if err != nil {
				summary.Failed++
				summary.Notes = append(summary.Notes, fmt.Sprintf("dispatch_failed:%s:%v", current.JobID, err))
				if config.stopOnFailure {
					return summary, err
				}
				continue
			}
			summary.Dispatched++
			current = dispatch.Job
		}
		if current.Status != "running" || workerDispatchState(current) != "packet_ready" {
			summary.Skipped++
			continue
		}
		result, err := runSingleWorkerJob(ctx, service, config, current)
		summary.Claimed++
		summary.ProcessedJobIDs = append(summary.ProcessedJobIDs, current.JobID)
		if result.ReportStatus == "succeeded" {
			summary.Completed++
		} else {
			summary.Failed++
		}
		if err != nil {
			summary.Notes = append(summary.Notes, fmt.Sprintf("job_failed:%s:%v", current.JobID, err))
			if config.stopOnFailure {
				return summary, err
			}
			continue
		}
	}
	return summary, nil
}

type workerJobOutcome struct {
	ReportStatus string
}

func runSingleWorkerJob(ctx context.Context, service jobService, config jobWorkConfig, job runtime.AgentJob) (workerJobOutcome, error) {
	packet, err := service.AgentRunPacket(job.JobID)
	if err != nil {
		return workerJobOutcome{ReportStatus: "failed"}, err
	}
	runCtx, err := prepareWorkerRunContext(config, packet)
	if err != nil {
		return workerJobOutcome{ReportStatus: "failed"}, err
	}
	status, notes, result, reportErr := executeWorkerCommand(ctx, config, packet, runCtx)
	if reportErr != nil {
		fallback := workerFailureResult(packet.Job, runCtx, reportErr.Error())
		if _, err := service.ReportAgentJob(job.JobID, "failed", []string{"worker_command_failed"}, fallback); err != nil {
			return workerJobOutcome{ReportStatus: "failed"}, fmt.Errorf("worker command failed (%v) and fallback report failed (%v)", reportErr, err)
		}
		return workerJobOutcome{ReportStatus: "failed"}, reportErr
	}
	if _, err := service.ReportAgentJob(job.JobID, status, notes, result); err != nil {
		if status == "succeeded" {
			fallback := workerFailureResult(packet.Job, runCtx, "worker result rejected: "+err.Error())
			fallback["verification"] = result["verification"]
			fallback["open_risks"] = appendStringList(resultArrayStrings(result["open_risks"]), "Layer OS rejected the success report; inspect the saved result payload before rerunning the job.")
			if _, fallbackErr := service.ReportAgentJob(job.JobID, "failed", append(notes, "worker_result_rejected"), fallback); fallbackErr != nil {
				return workerJobOutcome{ReportStatus: "failed"}, fmt.Errorf("report worker result: %w; fallback report: %v", err, fallbackErr)
			}
			return workerJobOutcome{ReportStatus: "failed"}, err
		}
		return workerJobOutcome{ReportStatus: status}, err
	}
	return workerJobOutcome{ReportStatus: status}, nil
}

func runSingleOfflineWorkerPacket(ctx context.Context, config jobWorkConfig, packet runtime.AgentRunPacket) (workerJobOutcome, error) {
	runCtx, err := prepareWorkerRunContext(config, packet)
	if err != nil {
		return workerJobOutcome{ReportStatus: "failed"}, err
	}
	status, _, _, err := executeWorkerCommand(ctx, config, packet, runCtx)
	return workerJobOutcome{ReportStatus: status}, err
}

func readWorkerPacketFile(path string) (runtime.AgentRunPacket, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return runtime.AgentRunPacket{}, err
	}
	var packet runtime.AgentRunPacket
	if err := json.Unmarshal(raw, &packet); err != nil {
		return runtime.AgentRunPacket{}, err
	}
	if strings.TrimSpace(packet.Job.JobID) == "" {
		return runtime.AgentRunPacket{}, errors.New("packet file missing job.job_id")
	}
	return packet, nil
}

func prepareWorkerRunContext(config jobWorkConfig, packet runtime.AgentRunPacket) (workerRunContext, error) {
	timestamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	jobDir := filepath.Join(config.workRoot, packet.Job.JobID, timestamp)
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		return workerRunContext{}, err
	}
	runCtx := workerRunContext{
		jobDir:     jobDir,
		packetPath: filepath.Join(jobDir, "packet.json"),
		promptPath: filepath.Join(jobDir, "prompt.md"),
		resultPath: filepath.Join(jobDir, "result.json"),
		stdoutPath: filepath.Join(jobDir, "stdout.log"),
		stderrPath: filepath.Join(jobDir, "stderr.log"),
	}
	rawPacket, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return workerRunContext{}, err
	}
	if err := os.WriteFile(runCtx.packetPath, rawPacket, 0o644); err != nil {
		return workerRunContext{}, err
	}
	prompt := workerPrompt(packet, runCtx)
	if err := os.WriteFile(runCtx.promptPath, []byte(prompt), 0o644); err != nil {
		return workerRunContext{}, err
	}
	return runCtx, nil
}

func executeWorkerCommand(ctx context.Context, config jobWorkConfig, packet runtime.AgentRunPacket, runCtx workerRunContext) (string, []string, map[string]any, error) {
	stdoutFile, err := os.Create(runCtx.stdoutPath)
	if err != nil {
		return "failed", nil, nil, err
	}
	defer stdoutFile.Close()
	stderrFile, err := os.Create(runCtx.stderrPath)
	if err != nil {
		return "failed", nil, nil, err
	}
	defer stderrFile.Close()

	cmd := exec.CommandContext(ctx, config.shell, "-lc", config.command)
	cmd.Dir = config.repoRoot
	cmd.Env = append(os.Environ(), workerCommandEnv(config, packet, runCtx)...)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	allowedPaths := workerAllowedPaths(packet.Job.Payload)
	baselineDirtyPaths := map[string]bool{}
	scopeScanEnabled := len(allowedPaths) > 0
	if len(allowedPaths) > 0 {
		baselineDirtyPaths, err = repoDirtyPathSet(config.repoRoot)
		if err != nil {
			if isNonGitRepoError(err) {
				scopeScanEnabled = false
			} else {
				return "failed", []string{"worker_scope_scan_failed"}, workerFailureResult(packet.Job, runCtx, "failed to scan repository before worker run: "+err.Error()), err
			}
		}
	}

	cmdErr := cmd.Run()

	if scopeScanEnabled {
		afterDirtyPaths, scanErr := repoDirtyPathSet(config.repoRoot)
		if scanErr != nil {
			return "failed", []string{"worker_scope_scan_failed"}, workerFailureResult(packet.Job, runCtx, "failed to scan repository after worker run: "+scanErr.Error()), scanErr
		}
		scopeViolations := workerScopeViolations(baselineDirtyPaths, afterDirtyPaths, allowedPaths)
		if len(scopeViolations) > 0 {
			scopeErr := fmt.Errorf("worker changed paths outside allowed_paths: %s", strings.Join(scopeViolations, ", "))
			return "failed", []string{"worker_scope_violation"}, workerFailureResult(packet.Job, runCtx, scopeErr.Error()), scopeErr
		}
	}

	result, parseErr := readWorkerResult(runCtx)
	if parseErr != nil && cmdErr == nil {
		return "failed", []string{"worker_result_missing"}, workerFailureResult(packet.Job, runCtx, parseErr.Error()), parseErr
	}
	if parseErr != nil {
		return "failed", []string{"worker_command_failed"}, workerFailureResult(packet.Job, runCtx, commandFailureMessage(cmdErr, runCtx)), cmdErr
	}
	status, notes := workerStatusAndNotes(result, cmdErr)
	delete(result, "status")
	delete(result, "notes")
	result = normalizeWorkerResult(packet.Job, result, runCtx, status)
	if cmdErr != nil && status == "succeeded" {
		status = "failed"
		result = workerFailureResult(packet.Job, runCtx, commandFailureMessage(cmdErr, runCtx))
		notes = append(notes, "worker_command_failed")
	}
	if cmdErr != nil && status != "failed" && status != "canceled" {
		status = "failed"
		result = workerFailureResult(packet.Job, runCtx, commandFailureMessage(cmdErr, runCtx))
		notes = append(notes, "worker_command_failed")
	}
	return status, notes, result, nil
}

func readWorkerResult(runCtx workerRunContext) (map[string]any, error) {
	if raw, err := os.ReadFile(runCtx.resultPath); err == nil {
		return parseWorkerResultJSON(raw)
	}
	stdoutRaw, err := os.ReadFile(runCtx.stdoutPath)
	if err != nil {
		return nil, errors.New("worker did not produce result JSON")
	}
	trimmed := strings.TrimSpace(string(stdoutRaw))
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return nil, errors.New("worker did not produce result JSON")
	}
	return parseWorkerResultJSON([]byte(trimmed))
}

func parseWorkerResultJSON(raw []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("worker result JSON must be an object")
	}
	return result, nil
}

func workerPrompt(packet runtime.AgentRunPacket, runCtx workerRunContext) string {
	payloadRaw, _ := json.MarshalIndent(packet.Job.Payload, "", "  ")
	knowledgeRaw, _ := json.MarshalIndent(packet.Knowledge, "", "  ")
	allowedPaths := workerAllowedPaths(packet.Job.Payload)
	allowedPathsSummary := "not specified in payload"
	if len(allowedPaths) > 0 {
		allowedPathsSummary = strings.Join(allowedPaths, ", ")
	}
	promptingSummary := workerPromptingSummary(packet.Prompting, packet.Job.Role)
	proposalSummary := "none"
	if packet.Proposal != nil {
		proposalSummary = strings.TrimSpace(packet.Proposal.Title)
		if proposalSummary == "" {
			proposalSummary = strings.TrimSpace(packet.Proposal.Summary)
		}
		if proposalSummary == "" {
			proposalSummary = packet.Proposal.ProposalID
		}
	}
	return strings.TrimSpace(fmt.Sprintf(`# Layer OS Worker Prompt

You are processing one Layer OS agent job from the CLI worker loop.

- Job ID: %s
- Role: %s
- Kind: %s
- Summary: %s
- Proposal: %s
- Allowed paths: %s
- Packet JSON: %s
- Result JSON path: %s
- Stdout log: %s
- Stderr log: %s

Rules:
1. Use the packet JSON as the source of truth.
2. Make code changes only inside the repository worktree and stay inside payload.allowed_paths when that list is present.
3. When finished, write one JSON object to the result path.
4. Use the official result keys: summary, artifacts, verification, open_risks, follow_on, touched_paths, blocked_paths.
5. Write result values for a non-developer first: explain what changed, why it matters, and what still needs attention before internal detail.
6. The CLI worker loop will report through the official Layer OS path after reading your JSON; do not invent side-channel completion files.
7. If work cannot be completed, still write a JSON object and set status to failed or canceled.
8. Follow the prompting contract below; take at most one bounded next action.

Prompting contract:
%s

Job payload:
%s

Knowledge packet:
%s
`, packet.Job.JobID, packet.Job.Role, packet.Job.Kind, packet.Job.Summary, proposalSummary, allowedPathsSummary, runCtx.packetPath, runCtx.resultPath, runCtx.stdoutPath, runCtx.stderrPath, promptingSummary, string(payloadRaw), string(knowledgeRaw)))
}

func workerCommandEnv(config jobWorkConfig, packet runtime.AgentRunPacket, runCtx workerRunContext) []string {
	env := []string{
		"LAYER_OS_JOB_ID=" + packet.Job.JobID,
		"LAYER_OS_JOB_ROLE=" + stringValue(packet.Job.Role),
		"LAYER_OS_JOB_KIND=" + stringValue(packet.Job.Kind),
		"LAYER_OS_JOB_SUMMARY=" + packet.Job.Summary,
		"LAYER_OS_JOB_STAGE=" + string(packet.Job.Stage),
		"LAYER_OS_JOB_SOURCE=" + packet.Job.Source,
		"LAYER_OS_REPO_ROOT=" + config.repoRoot,
		"LAYER_OS_PACKET_PATH=" + runCtx.packetPath,
		"LAYER_OS_PROMPT_PATH=" + runCtx.promptPath,
		"LAYER_OS_RESULT_PATH=" + runCtx.resultPath,
		"LAYER_OS_STDOUT_PATH=" + runCtx.stdoutPath,
		"LAYER_OS_STDERR_PATH=" + runCtx.stderrPath,
		"LAYER_OS_JOB_WORK_DIR=" + runCtx.jobDir,
	}
	if allowedPaths := workerAllowedPaths(packet.Job.Payload); len(allowedPaths) > 0 {
		env = append(env, "LAYER_OS_ALLOWED_PATHS="+strings.Join(allowedPaths, ","))
	}
	if packet.Prompting != nil {
		env = append(env,
			"LAYER_OS_PROMPTING_COGNITION_MODE="+packet.Prompting.CognitionMode,
			"LAYER_OS_PROMPTING_DECISION_SCOPE="+packet.Prompting.DecisionScope,
			"LAYER_OS_PROMPTING_AUTONOMY_BUDGET="+packet.Prompting.AutonomyBudget,
			"LAYER_OS_PROMPTING_MUTATION_POLICY="+packet.Prompting.MutationPolicy,
		)
	}
	if packet.Runtime.ReportCommand != "" {
		env = append(env, "LAYER_OS_REPORT_COMMAND="+packet.Runtime.ReportCommand)
	}
	if packet.Runtime.ReportPath != "" {
		env = append(env, "LAYER_OS_REPORT_PATH="+packet.Runtime.ReportPath)
	}
	if packet.Runtime.WriteTokenEnv != nil && strings.TrimSpace(*packet.Runtime.WriteTokenEnv) != "" {
		env = append(env, "LAYER_OS_REPORT_TOKEN_ENV="+strings.TrimSpace(*packet.Runtime.WriteTokenEnv))
	}
	return env
}

func workerStatusAndNotes(result map[string]any, cmdErr error) (string, []string) {
	status := strings.TrimSpace(strings.ToLower(resultString(result["status"])))
	switch status {
	case "succeeded", "failed", "canceled":
	default:
		if cmdErr != nil {
			status = "failed"
		} else {
			status = "succeeded"
		}
	}
	return status, resultArrayStrings(result["notes"])
}

func normalizeWorkerResult(job runtime.AgentJob, result map[string]any, runCtx workerRunContext, status string) map[string]any {
	merged := map[string]any{}
	for key, value := range result {
		merged[key] = value
	}
	if strings.TrimSpace(resultString(merged["summary"])) == "" {
		merged["summary"] = defaultWorkerSummary(job, status)
	}
	if _, ok := merged["artifacts"]; !ok {
		merged["artifacts"] = []string{}
	}
	if _, ok := merged["verification"]; !ok {
		merged["verification"] = []string{}
	}
	if _, ok := merged["open_risks"]; !ok {
		merged["open_risks"] = []string{}
	}
	if _, ok := merged["follow_on"]; !ok {
		merged["follow_on"] = []string{}
	}
	if _, ok := merged["touched_paths"]; !ok {
		merged["touched_paths"] = []string{}
	}
	if _, ok := merged["blocked_paths"]; !ok {
		merged["blocked_paths"] = []string{}
	}
	merged["worker_prompt_path"] = runCtx.promptPath
	merged["worker_packet_path"] = runCtx.packetPath
	return merged
}

func workerFailureResult(job runtime.AgentJob, runCtx workerRunContext, detail string) map[string]any {
	stderrTail := readLogTail(runCtx.stderrPath)
	stdoutTail := readLogTail(runCtx.stdoutPath)
	allowedPaths := workerAllowedPaths(job.Payload)
	verification := []string{}
	if stderrTail != "" {
		verification = append(verification, "stderr tail: "+stderrTail)
	}
	if stdoutTail != "" {
		verification = append(verification, "stdout tail: "+stdoutTail)
	}
	openRisks := []string{"The local worker command did not complete cleanly, so the job needs review before retrying automation."}
	if len(allowedPaths) > 0 {
		openRisks = append(openRisks, "Retry inside allowed_paths only: "+strings.Join(allowedPaths, ", "))
	}
	return map[string]any{
		"summary":       fmt.Sprintf("CLI worker could not complete %s: %s", strings.TrimSpace(job.JobID), strings.TrimSpace(detail)),
		"artifacts":     []string{runCtx.packetPath, runCtx.promptPath, runCtx.stdoutPath, runCtx.stderrPath},
		"verification":  verification,
		"open_risks":    openRisks,
		"follow_on":     []string{"Fix the worker command or its prompt contract, then rerun the same job."},
		"touched_paths": []string{},
		"blocked_paths": []string{},
	}
}

func workerAllowedPaths(payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	return resultArrayStrings(payload["allowed_paths"])
}

func workerRoleGuardrail(role string) string {
	switch strings.TrimSpace(role) {
	case "verifier":
		return "Verifier jobs default to read-only inspection and test execution; do not edit code unless the packet explicitly asks for a repair."
	case "planner":
		return "Planner jobs default to analysis and orchestration notes; avoid product-code edits unless the packet explicitly asks for them."
	case "designer":
		return "Designer jobs may change experience-layer code, but keep the lane bounded to the assigned surface and avoid widening into runtime seams."
	default:
		return "Keep the lane bounded to the assigned role and acceptance target; do not widen scope just because adjacent work looks tempting."
	}
}

func workerPromptingSummary(prompting *runtime.PromptingContract, role string) string {
	if prompting == nil {
		return workerRoleGuardrail(role)
	}
	lines := []string{
		"- Directive: " + prompting.Directive,
		"- Cognition mode: " + prompting.CognitionMode,
		"- Decision scope: " + prompting.DecisionScope,
		"- Autonomy budget: " + prompting.AutonomyBudget,
		"- Mutation policy: " + prompting.MutationPolicy,
	}
	if len(prompting.EscalationTriggers) > 0 {
		lines = append(lines, "- Escalate when: "+strings.Join(prompting.EscalationTriggers, "; "))
	}
	if len(prompting.OpenQuestions) > 0 {
		lines = append(lines, "- Open questions: "+strings.Join(prompting.OpenQuestions, " | "))
	}
	return strings.Join(lines, "\n")
}

func finishIdleWorkerSession(service jobService) error {
	note := "job worker idle exit"
	_, err := service.SessionFinish(runtime.SessionFinishInput{
		CurrentFocus: "Worker idle exit",
		NextSteps:    []string{},
		OpenRisks:    []string{},
		Note:         &note,
	})
	return err
}

func readLogTail(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return ""
	}
	text := strings.TrimSpace(string(raw))
	if len(text) > 400 {
		text = text[len(text)-400:]
	}
	return strings.ReplaceAll(text, "\n", " ")
}

func workerDispatchState(job runtime.AgentJob) string {
	if job.Result == nil {
		return ""
	}
	return strings.TrimSpace(resultString(job.Result["dispatch_state"]))
}

func mergeJobWorkSummary(base jobWorkSummary, next jobWorkSummary) jobWorkSummary {
	base.Scanned += next.Scanned
	base.Dispatched += next.Dispatched
	base.Claimed += next.Claimed
	base.Completed += next.Completed
	base.Failed += next.Failed
	base.Skipped += next.Skipped
	base.IdlePolls += next.IdlePolls
	base.ProcessedJobIDs = append(base.ProcessedJobIDs, next.ProcessedJobIDs...)
	base.Notes = append(base.Notes, next.Notes...)
	return base
}

func workerRoleFilter(raw string) map[string]bool {
	items := splitCSV(raw)
	if len(items) == 0 {
		return nil
	}
	filter := map[string]bool{}
	for _, item := range items {
		filter[strings.TrimSpace(item)] = true
	}
	return filter
}

func workerRoleAllowed(filter map[string]bool, role string) bool {
	if len(filter) == 0 {
		return true
	}
	return filter[strings.TrimSpace(role)]
}

func positiveDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func positiveInt(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func resultArrayStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string{}, typed...)
	case []any:
		items := []string{}
		for _, item := range typed {
			text := strings.TrimSpace(resultString(item))
			if text != "" {
				items = append(items, text)
			}
		}
		return items
	case string:
		if strings.TrimSpace(typed) == "" {
			return []string{}
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return []string{}
	}
}

func appendStringList(base []string, extras ...string) []string {
	items := append([]string{}, base...)
	for _, extra := range extras {
		if strings.TrimSpace(extra) != "" {
			items = append(items, strings.TrimSpace(extra))
		}
	}
	return items
}

func repoDirtyPathSet(repoRoot string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain", "--untracked-files=all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, text)
	}
	items := map[string]bool{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path := parseGitStatusPath(line)
		if path == "" {
			continue
		}
		items[path] = true
	}
	return items, nil
}

func isNonGitRepoError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "not a git repository")
}

func parseGitStatusPath(line string) string {
	if len(line) < 4 {
		return ""
	}
	path := strings.TrimSpace(line[3:])
	if path == "" {
		return ""
	}
	if strings.Contains(path, " -> ") {
		parts := strings.Split(path, " -> ")
		path = strings.TrimSpace(parts[len(parts)-1])
	}
	return filepath.ToSlash(path)
}

func workerScopeViolations(before map[string]bool, after map[string]bool, allowed []string) []string {
	violations := []string{}
	for path := range after {
		if before[path] {
			continue
		}
		if pathWithinAllowedPrefixes(path, allowed) {
			continue
		}
		violations = append(violations, path)
	}
	sort.Strings(violations)
	return violations
}

func pathWithinAllowedPrefixes(path string, prefixes []string) bool {
	normalizedPath := filepath.ToSlash(filepath.Clean(path))
	for _, prefix := range prefixes {
		normalizedPrefix := normalizeAllowedPrefix(prefix)
		if normalizedPrefix == "" {
			continue
		}
		if normalizedPath == strings.TrimSuffix(normalizedPrefix, "/") || strings.HasPrefix(normalizedPath, normalizedPrefix) {
			return true
		}
	}
	return false
}

func normalizeAllowedPrefix(value string) string {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
	if clean == "." || clean == "" {
		return ""
	}
	if !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean
}

func resultString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

func defaultWorkerSummary(job runtime.AgentJob, status string) string {
	summary := strings.TrimSpace(job.Summary)
	if summary == "" {
		summary = strings.TrimSpace(job.JobID)
	}
	switch status {
	case "failed":
		return "CLI worker failed to complete: " + summary
	case "canceled":
		return "CLI worker canceled: " + summary
	default:
		return "CLI worker completed: " + summary
	}
}

func commandFailureMessage(err error, runCtx workerRunContext) string {
	if err == nil {
		return "worker command failed"
	}
	stderrTail := readLogTail(runCtx.stderrPath)
	if stderrTail == "" {
		return err.Error()
	}
	return err.Error() + ": " + stderrTail
}

func stringValue(value string) string {
	return strings.TrimSpace(value)
}

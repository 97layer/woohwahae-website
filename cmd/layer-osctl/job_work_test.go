package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestWorkerPromptUsesCanonicalFounderReadableReportContract(t *testing.T) {
	packet := runtime.AgentRunPacket{
		Job: runtime.AgentJob{
			JobID:   "job_001",
			Role:    "implementer",
			Kind:    "implement",
			Summary: "Patch the backend lane",
			Payload: map[string]any{"allowed_paths": []string{"cmd/layer-osctl/"}},
		},
		Knowledge: runtime.KnowledgePacket{CurrentFocus: "Unify agent worker reporting"},
	}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
		resultPath: "/tmp/job/result.json",
		stdoutPath: "/tmp/job/stdout.log",
		stderrPath: "/tmp/job/stderr.log",
	}

	prompt := workerPrompt(packet, runCtx)

	for _, want := range []string{
		"Use the official result keys: summary, artifacts, verification, open_risks, follow_on, touched_paths, blocked_paths.",
		"Write result values for a non-developer first",
		"do not invent side-channel completion files",
		"Allowed paths: cmd/layer-osctl/",
		"stay inside payload.allowed_paths when that list is present",
		"/tmp/job/result.json",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected worker prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestNormalizeWorkerResultAddsCanonicalDefaults(t *testing.T) {
	job := runtime.AgentJob{JobID: "job_002", Summary: "Stabilize worker lane"}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
	}

	result := normalizeWorkerResult(job, map[string]any{
		"summary": "Kept the worker report easy to understand.",
	}, runCtx, "succeeded")

	for _, key := range []string{"artifacts", "verification", "open_risks", "follow_on", "touched_paths", "blocked_paths"} {
		items, ok := result[key].([]string)
		if !ok {
			t.Fatalf("expected %s default to be []string, got %#v", key, result[key])
		}
		if len(items) != 0 {
			t.Fatalf("expected %s default to be empty, got %#v", key, items)
		}
	}
	if result["worker_prompt_path"] != "/tmp/job/prompt.md" || result["worker_packet_path"] != "/tmp/job/packet.json" {
		t.Fatalf("expected worker paths to be retained, got %#v", result)
	}
}

func TestWorkerFailureResultCarriesAllowedPathHint(t *testing.T) {
	job := runtime.AgentJob{
		JobID:   "job_003",
		Summary: "Respect backend scope",
		Payload: map[string]any{"allowed_paths": []string{"internal/runtime/"}},
	}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
		stdoutPath: "/tmp/job/stdout.log",
		stderrPath: "/tmp/job/stderr.log",
	}

	result := workerFailureResult(job, runCtx, "rate limited")
	openRisks, ok := result["open_risks"].([]string)
	if !ok {
		t.Fatalf("expected open_risks to be []string, got %#v", result["open_risks"])
	}
	if len(openRisks) < 2 || !strings.Contains(openRisks[1], "internal/runtime/") {
		t.Fatalf("expected allowed_paths retry hint, got %#v", openRisks)
	}
}

func TestWorkerCommandEnvCarriesAllowedPaths(t *testing.T) {
	packet := runtime.AgentRunPacket{
		Job: runtime.AgentJob{
			JobID:   "job_004",
			Role:    "implementer",
			Kind:    "implement",
			Summary: "Stay inside the runtime lane",
			Stage:   runtime.StageCompose,
			Source:  "founder.manual",
			Payload: map[string]any{"allowed_paths": []string{"internal/runtime/", "cmd/layer-osctl/"}},
		},
		Prompting: &runtime.PromptingContract{
			Directive:          "Take one bounded step.",
			CognitionMode:      "staff_advisor",
			DecisionScope:      "bounded",
			AutonomyBudget:     "single_step",
			MutationPolicy:     "scoped_write",
			EscalationTriggers: []string{"hot seam"},
			OpenQuestions:      []string{"What is blocking the lane?"},
		},
	}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
		resultPath: "/tmp/job/result.json",
		stdoutPath: "/tmp/job/stdout.log",
		stderrPath: "/tmp/job/stderr.log",
		jobDir:     "/tmp/job",
	}

	env := workerCommandEnv(jobWorkConfig{repoRoot: "/repo"}, packet, runCtx)
	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "LAYER_OS_ALLOWED_PATHS=internal/runtime/,cmd/layer-osctl/") {
		t.Fatalf("expected allowed paths env, got %#v", env)
	}
	for _, want := range []string{
		"LAYER_OS_PROMPTING_COGNITION_MODE=staff_advisor",
		"LAYER_OS_PROMPTING_DECISION_SCOPE=bounded",
		"LAYER_OS_PROMPTING_AUTONOMY_BUDGET=single_step",
		"LAYER_OS_PROMPTING_MUTATION_POLICY=scoped_write",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected prompting env %q, got %#v", want, env)
		}
	}
}

func TestWorkerPromptVerifierUsesReadOnlyGuardrail(t *testing.T) {
	packet := runtime.AgentRunPacket{
		Job: runtime.AgentJob{
			JobID:   "job_005",
			Role:    "verifier",
			Kind:    "verify",
			Summary: "Check worker safety",
		},
	}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
		resultPath: "/tmp/job/result.json",
		stdoutPath: "/tmp/job/stdout.log",
		stderrPath: "/tmp/job/stderr.log",
	}

	prompt := workerPrompt(packet, runCtx)
	if !strings.Contains(prompt, "Verifier jobs default to read-only inspection and test execution") {
		t.Fatalf("expected verifier read-only guardrail, got:\n%s", prompt)
	}
}

func TestWorkerPromptUsesPromptingContractWhenPresent(t *testing.T) {
	packet := runtime.AgentRunPacket{
		Job: runtime.AgentJob{
			JobID:   "job_006",
			Role:    "implementer",
			Kind:    "implement",
			Summary: "Ship one bounded patch",
		},
		Prompting: &runtime.PromptingContract{
			Directive:          "Act like a practical staff aide.",
			CognitionMode:      "staff_advisor",
			DecisionScope:      "bounded",
			AutonomyBudget:     "single_step",
			MutationPolicy:     "scoped_write",
			EscalationTriggers: []string{"contract change needed"},
			OpenQuestions:      []string{"Which lane is safest next?"},
		},
	}
	runCtx := workerRunContext{
		packetPath: "/tmp/job/packet.json",
		promptPath: "/tmp/job/prompt.md",
		resultPath: "/tmp/job/result.json",
		stdoutPath: "/tmp/job/stdout.log",
		stderrPath: "/tmp/job/stderr.log",
	}

	prompt := workerPrompt(packet, runCtx)
	for _, want := range []string{
		"Prompting contract:",
		"Directive: Act like a practical staff aide.",
		"Decision scope: bounded",
		"Autonomy budget: single_step",
		"Mutation policy: scoped_write",
		"Open questions: Which lane is safest next?",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompting contract line %q, got:\n%s", want, prompt)
		}
	}
}

func TestWorkerScopeViolationsDetectsNewOutOfScopePaths(t *testing.T) {
	before := map[string]bool{
		"cmd/layer-osctl/job.go": true,
	}
	after := map[string]bool{
		"cmd/layer-osctl/job.go":    true,
		"internal/runtime/audit.go": true,
		"docs/operator.md":          true,
	}
	violations := workerScopeViolations(before, after, []string{"cmd/layer-osctl/"})
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %#v", violations)
	}
	if violations[0] != "docs/operator.md" || violations[1] != "internal/runtime/audit.go" {
		t.Fatalf("unexpected violations order/content: %#v", violations)
	}
}

func TestWorkerScopeViolationsIgnoresPreExistingDirtyPaths(t *testing.T) {
	before := map[string]bool{
		"docs/operator.md": true,
	}
	after := map[string]bool{
		"docs/operator.md":       true,
		"cmd/layer-osctl/job.go": true,
	}
	violations := workerScopeViolations(before, after, []string{"cmd/layer-osctl/"})
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %#v", violations)
	}
}

func TestJobWorkerE2ESmokePacketToReport(t *testing.T) {
	repoRoot := t.TempDir()
	targetPath := filepath.Join(repoRoot, "internal", "runtime")
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		t.Fatalf("mkdir repo path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetPath, "e2e.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write repo evidence: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	job := runtime.AgentJob{
		JobID:     "job_e2e_001",
		Role:      "implementer",
		Kind:      "implement",
		Summary:   "E2E worker smoke",
		Status:    "running",
		Source:    "founder.manual",
		Surface:   runtime.SurfaceAPI,
		Stage:     runtime.StageCompose,
		Payload:   map[string]any{"allowed_paths": []string{"internal/runtime/"}},
		Result:    map[string]any{"dispatch_state": "packet_ready"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := service.CreateAgentJob(job); err != nil {
		t.Fatalf("create job: %v", err)
	}
	packet, err := service.AgentRunPacket(job.JobID)
	if err != nil {
		t.Fatalf("job packet: %v", err)
	}
	if packet.Job.JobID != job.JobID {
		t.Fatalf("expected packet job id %q, got %q", job.JobID, packet.Job.JobID)
	}
	if !strings.Contains(packet.Runtime.ReportCommand, job.JobID) {
		t.Fatalf("expected report command to include job id, got %q", packet.Runtime.ReportCommand)
	}

	result := map[string]any{
		"summary":       "CLI worker completed.",
		"artifacts":     []string{"internal/runtime/e2e.txt"},
		"verification":  []string{"smoke only"},
		"open_risks":    []string{},
		"follow_on":     []string{},
		"touched_paths": []string{"internal/runtime/e2e.txt"},
		"blocked_paths": []string{},
	}
	report, err := service.ReportAgentJob(job.JobID, "succeeded", []string{"worker_smoke"}, result)
	if err != nil {
		t.Fatalf("job report: %v", err)
	}
	if report.Job.Status != "succeeded" {
		t.Fatalf("expected succeeded status, got %q", report.Job.Status)
	}
	if summary, ok := report.Job.Result["summary"].(string); !ok || strings.TrimSpace(summary) == "" {
		t.Fatalf("expected summary in job report, got %#v", report.Job.Result)
	}
}

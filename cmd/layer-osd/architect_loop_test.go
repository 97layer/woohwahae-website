package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestRunArchitectTickPromotesDerivedJob(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_architect_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "architect.loop.test"}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: true, CleanupGemini: true})
	if err != nil {
		t.Fatalf("run architect tick: %v", err)
	}
	if result.Promotion.Created != 1 || result.Promotion.Dispatched != 0 || len(result.Promotion.Items) != 1 {
		t.Fatalf("unexpected architect tick result: %+v", result)
	}
	if result.Promotion.Items[0].SourceKind != "parallel_candidate" || result.Promotion.Items[0].Status != "created" {
		t.Fatalf("unexpected promotion item: %+v", result.Promotion.Items[0])
	}
	jobs := service.ListAgentJobs()
	if len(jobs) != 1 || jobs[0].Source != "knowledge.parallel_candidate" || jobs[0].Status != "queued" {
		t.Fatalf("unexpected jobs after architect tick: %+v", jobs)
	}
}

func TestRunArchitectTickDispatchesWhenEnabled(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_architect_dispatch_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "architect.loop.test"}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: true, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: true, CleanupGemini: true})
	if err != nil {
		t.Fatalf("run architect tick with dispatch: %v", err)
	}
	if result.Promotion.Created != 1 || result.Promotion.Dispatched != 1 || len(result.Promotion.Items) != 1 {
		t.Fatalf("unexpected architect tick dispatch result: %+v", result)
	}
	item := result.Promotion.Items[0]
	if item.Status != "dispatched" || item.Job == nil || item.Dispatch == nil {
		t.Fatalf("expected dispatched promotion item, got %+v", item)
	}
	if item.Job.Status != "running" || item.Dispatch.Job.Result == nil || item.Dispatch.Job.Result["dispatch_state"] != "packet_ready" {
		t.Fatalf("expected packet-ready running job, got %+v", item.Dispatch)
	}
}

func TestRunArchitectTickAutoRecoversGeminiArtifacts(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	mustArchitectWrite(t, filepath.Join(repoRoot, "task.md.resolved"), "# Task\n\nproposal_001 remains open.")
	mustArchitectWrite(t, filepath.Join(repoRoot, "task.md.metadata.json"), `{"summary":"Recovered Gemini task","updatedAt":"2026-03-08T10:29:56Z","version":"2"}`)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: true, CleanupGemini: true})
	if err != nil {
		t.Fatalf("run architect tick auto recover: %v", err)
	}
	if result.Recovery.Considered != 1 || result.Recovery.Created != 1 || result.Recovery.Cleaned < 2 {
		t.Fatalf("unexpected gemini recovery result: %+v", result.Recovery)
	}
	if len(service.ListObservations(runtime.ObservationQuery{SourceChannel: "gemini"})) != 1 {
		t.Fatalf("expected recovered gemini observation, got %+v", service.ListObservations(runtime.ObservationQuery{SourceChannel: "gemini"}))
	}
}

func TestRunArchitectTickRunsVerificationOncePerRepoStamp(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	seedArchitectVerificationRoot(t, repoRoot, false)
	seedArchitectGoCache(t)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false})
	if err != nil {
		t.Fatalf("run architect tick verify: %v", err)
	}
	if !result.VerificationRan || result.Verification == nil || result.Verification.Status != "passed" {
		t.Fatalf("expected passed architect verification, got %+v", result.Verification)
	}
	if got := len(service.ListVerifications()); got != 1 {
		t.Fatalf("expected 1 verification after first run, got %d", got)
	}

	second, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false})
	if err != nil {
		t.Fatalf("run architect tick verify again: %v", err)
	}
	if second.VerificationRan {
		t.Fatalf("expected unchanged repo stamp to skip verification, got %+v", second.Verification)
	}
	if got := len(service.ListVerifications()); got != 1 {
		t.Fatalf("expected unchanged repo stamp to keep 1 verification, got %d", got)
	}
}

func TestRunArchitectTickCreatesRepairJobOnVerificationFailure(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	seedArchitectVerificationRoot(t, repoRoot, true)
	seedArchitectGoCache(t)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: true, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false})
	if err == nil {
		t.Fatal("expected architect verification warning on failing test suite")
	}
	if !strings.Contains(err.Error(), "verification:") {
		t.Fatalf("expected verification warning, got %v", err)
	}
	if !result.VerificationRan || result.Verification == nil || result.Verification.Status != "failed" {
		t.Fatalf("expected failed architect verification, got %+v", result.Verification)
	}
	if result.RepairJob == nil || result.RepairJob.Source != "architect.verification_failed" {
		t.Fatalf("expected repair job, got %+v", result.RepairJob)
	}
	if result.RepairDispatch == nil || result.RepairDispatch.Job.Result == nil || result.RepairDispatch.Job.Result["dispatch_state"] != "packet_ready" {
		t.Fatalf("expected packet-ready repair dispatch, got %+v", result.RepairDispatch)
	}
	instructions, _ := result.RepairJob.Payload["instructions"].(string)
	if !strings.Contains(instructions, "Verification scope: architect_loop") || !strings.Contains(instructions, "After the fix, rerun the verification command") {
		t.Fatalf("expected repair instructions in payload, got %q", instructions)
	}
	excerpt, _ := result.RepairJob.Payload["verification_output_excerpt"].(string)
	if !strings.Contains(excerpt, "forced architect failure") {
		t.Fatalf("expected verification excerpt in payload, got %q", excerpt)
	}
	jobs := service.ListAgentJobs()
	if len(jobs) != 1 || jobs[0].Role != "implementer" {
		t.Fatalf("expected one implementer repair job, got %+v", jobs)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.verification_failed" {
		t.Fatalf("expected verification review-room escalation, got %+v", room.Open)
	}
}

func TestRunArchitectTickKeepsDirectLLMRepairQueued(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "gemini")
	t.Setenv("LAYER_OS_PROVIDERS", "gemini")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	seedArchitectVerificationRoot(t, repoRoot, true)
	seedArchitectGoCache(t)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: true, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false})
	if err == nil {
		t.Fatal("expected architect verification warning on failing test suite")
	}
	if !strings.Contains(err.Error(), "verification:") {
		t.Fatalf("expected verification warning, got %v", err)
	}
	if result.RepairJob == nil || result.RepairJob.Source != "architect.verification_failed" {
		t.Fatalf("expected repair job, got %+v", result.RepairJob)
	}
	if result.RepairDispatch != nil {
		t.Fatalf("expected direct-LLM repair to stay queued, got %+v", result.RepairDispatch)
	}
	if result.RepairJob.Status != "queued" {
		t.Fatalf("expected queued repair job, got %+v", result.RepairJob)
	}
	dispatchGuidance, _ := result.RepairJob.Payload["dispatch_guidance"].(string)
	if !strings.Contains(dispatchGuidance, "packet-capable worker") {
		t.Fatalf("expected dispatch guidance in payload, got %q", dispatchGuidance)
	}
	instructions, _ := result.RepairJob.Payload["instructions"].(string)
	if !strings.Contains(instructions, "If you are blocked, name the failing package/test") {
		t.Fatalf("expected scoped repair instructions, got %q", instructions)
	}
	foundNote := false
	for _, note := range result.RepairJob.Notes {
		if strings.TrimSpace(note) == "dispatch_skipped:direct_llm_requires_worker" {
			foundNote = true
			break
		}
	}
	if !foundNote {
		t.Fatalf("expected direct-LLM skip note, got %+v", result.RepairJob.Notes)
	}
}

func TestRunArchitectTickWithLoggingRecordsStatus(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_architect_status_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "architect.loop.test"}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	status := newArchitectLoopStatus(architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: true, CleanupGemini: true})
	runArchitectTickWithLogging(service, status, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: true, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: true, CleanupGemini: true})

	snapshot := status.Snapshot()
	if snapshot == nil {
		t.Fatal("expected architect status snapshot")
	}
	if snapshot.LastRunAt == nil {
		t.Fatalf("expected last_run_at, got %+v", snapshot)
	}
	if snapshot.LastConsidered == 0 || snapshot.LastCreated != 1 || snapshot.LastError != nil {
		t.Fatalf("unexpected architect status snapshot: %+v", snapshot)
	}
}

func mustArchitectWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func seedArchitectVerificationRoot(t *testing.T, root string, fail bool) {
	t.Helper()
	mustArchitectWrite(t, filepath.Join(root, "go.mod"), "module example.com/architecttest\n\ngo 1.25.0\n")
	mustArchitectWrite(t, filepath.Join(root, "pkg", "queue", "queue.go"), "package queue\n\nfunc Healthy() bool { return true }\n")
	testBody := "package queue\n\nimport \"testing\"\n\nfunc TestHealthy(t *testing.T) {\n\tif !Healthy() {\n\t\tt.Fatal(\"expected healthy queue\")\n\t}\n}\n"
	if fail {
		testBody = "package queue\n\nimport \"testing\"\n\nfunc TestHealthy(t *testing.T) {\n\tt.Fatal(\"forced architect failure\")\n}\n"
	}
	mustArchitectWrite(t, filepath.Join(root, "pkg", "queue", "queue_test.go"), testBody)
}

func seedArchitectGoCache(t *testing.T) {
	t.Helper()
	cacheRoot := t.TempDir()
	goCache := filepath.Join(cacheRoot, "go-build")
	goTmp := filepath.Join(cacheRoot, "go-tmp")
	if err := os.MkdirAll(goCache, 0o755); err != nil {
		t.Fatalf("mkdir gocache: %v", err)
	}
	if err := os.MkdirAll(goTmp, 0o755); err != nil {
		t.Fatalf("mkdir gotmpdir: %v", err)
	}
	t.Setenv("GOCACHE", goCache)
	t.Setenv("GOTMPDIR", goTmp)
}

func TestRunArchitectTickWithLoggingMarksIdleWhenNoWork(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	config := architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: false, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false, AutoRecoverCorpus: false, CleanupCorpus: false}
	status := newArchitectLoopStatus(config)
	runArchitectTickWithLogging(service, status, config)
	snapshot := status.Snapshot()
	if snapshot == nil || snapshot.LastRunAt == nil {
		t.Fatalf("expected architect snapshot, got %+v", snapshot)
	}
	if !snapshot.LastIdle || snapshot.LastConsidered != 0 || snapshot.LastCreated != 0 || snapshot.LastError != nil {
		t.Fatalf("expected idle architect snapshot, got %+v", snapshot)
	}
}

func TestRunArchitectTickRecoversCorpusMarkdown(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source := filepath.Join(repoRoot, "knowledge", "corpus", "entries", "analysis.md")
	mustArchitectWrite(t, source, "# Architecture Note\n\nKeep the corpus canonical.\n")
	result, err := runArchitectTick(service, architectLoopConfig{Enabled: true, AutoDispatch: false, AutoVerify: false, Interval: 15 * time.Second, PromoteLimit: 1, AutoRecoverGemini: false, CleanupGemini: false, AutoRecoverCorpus: true, CleanupCorpus: true})
	if err != nil {
		t.Fatalf("run architect tick: %v", err)
	}
	if result.CorpusRecovery.Created != 1 || result.CorpusRecovery.Cleaned == 0 {
		t.Fatalf("unexpected corpus recovery result: %+v", result.CorpusRecovery)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("expected corpus markdown cleanup, err=%v", err)
	}
}

func TestArchitectPromoteLimitDefaultsToThree(t *testing.T) {
	t.Setenv("LAYER_OS_ARCHITECT_PROMOTE_LIMIT", "")
	if got := architectPromoteLimit(); got != 3 {
		t.Fatalf("expected default architect promote limit 3, got %d", got)
	}

	t.Setenv("LAYER_OS_ARCHITECT_PROMOTE_LIMIT", "invalid")
	if got := architectPromoteLimit(); got != 3 {
		t.Fatalf("expected invalid architect promote limit to fall back to 3, got %d", got)
	}
}

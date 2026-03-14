package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewRoomLoadsLegacyStringPayload(t *testing.T) {
	dataDir := t.TempDir()
	legacy := map[string]any{
		"source":   "review_room.json",
		"accepted": []string{"Keep root budget tight."},
		"open":     []string{"Fix data race."},
		"deferred": []string{"Oracle Cloud ARM rollout."},
		"issues":   []string{},
	}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy review room: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "review_room.json"), raw, 0o644); err != nil {
		t.Fatalf("write legacy review room: %v", err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	room := service.ReviewRoom()
	if len(room.Accepted) != 1 || room.Accepted[0].Text != "Keep root budget tight." {
		t.Fatalf("unexpected accepted items: %+v", room.Accepted)
	}
	if room.Open[0].Kind != "legacy_note" || room.Open[0].Source != "legacy_import" {
		t.Fatalf("unexpected migrated open item: %+v", room.Open[0])
	}
}

func TestReviewRoomDoesNotSeedFromLegacyMarkdownScratchpad(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "@임시회의.md"), []byte("## Open\n- stale legacy note\n"), 0o644); err != nil {
		t.Fatalf("write legacy review room scratchpad: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	room := service.ReviewRoom()
	if got, want := room.Source, ".layer-os/review_room.json"; got != want {
		t.Fatalf("expected canonical runtime source %q, got %q", want, got)
	}
	for _, item := range room.Open {
		if item.Text == "stale legacy note" {
			t.Fatalf("expected legacy markdown note to stay ignored, got %+v", room.Open)
		}
		if item.Source == "legacy_markdown" {
			t.Fatalf("expected no legacy_markdown import, got %+v", room.Open)
		}
	}
}

func TestReviewRoomStructuredItemsPreserveMetadata(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "issue_001"
	why := "reduce deploy-risk before retry"
	whyUnresolved := "race source is still not isolated"
	contradiction := "accepted mitigation exists but runtime still races"
	room, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Fix data race.", Kind: "bug", Severity: "high", Source: "agent.codex", Ref: &ref, Why: why, WhyUnresolved: &whyUnresolved, Contradictions: []string{contradiction}, PatternRefs: []string{"issue_001", "session_003"}})
	if err != nil {
		t.Fatalf("add structured review room item: %v", err)
	}
	if len(room.Open) != 1 {
		t.Fatalf("expected one open item, got %+v", room.Open)
	}
	if room.Open[0].Kind != "bug" || room.Open[0].Severity != "high" || room.Open[0].Source != "agent.codex" {
		t.Fatalf("unexpected structured item: %+v", room.Open[0])
	}
	if room.Open[0].Ref == nil || *room.Open[0].Ref != "issue_001" {
		t.Fatalf("unexpected structured ref: %+v", room.Open[0].Ref)
	}
	if room.Open[0].Why != why {
		t.Fatalf("unexpected why: %+v", room.Open[0].Why)
	}
	if room.Open[0].WhyUnresolved == nil || *room.Open[0].WhyUnresolved != whyUnresolved {
		t.Fatalf("unexpected why_unresolved: %+v", room.Open[0].WhyUnresolved)
	}
	if len(room.Open[0].Contradictions) != 1 || room.Open[0].Contradictions[0] != contradiction {
		t.Fatalf("unexpected contradictions: %+v", room.Open[0].Contradictions)
	}
	if len(room.Open[0].PatternRefs) != 2 || room.Open[0].PatternRefs[1] != "session_003" {
		t.Fatalf("unexpected pattern refs: %+v", room.Open[0].PatternRefs)
	}
}

func TestReviewRoomReloadCanonicalizesRuntimeSource(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data dir: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	room := ReviewRoom{
		Source: "/var/folders/example/TestDaemonReadFlowStatusHealthzBootstrapAndFallback/review_room.json",
		Open: []ReviewRoomItem{{
			Text:      "Keep runtime source canonical.",
			Kind:      "agenda",
			Severity:  "medium",
			Source:    "test.seed",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}},
	}
	raw, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("marshal review room: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, reviewRoomRuntimeFile), raw, 0o644); err != nil {
		t.Fatalf("write review room: %v", err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	reloaded := service.ReviewRoom()
	if got, want := reloaded.Source, ".layer-os/review_room.json"; got != want {
		t.Fatalf("expected canonical runtime source %q, got %q", want, got)
	}
}

func TestReviewRoomOpenAgendaAutoDetectsContradictionsAgainstAcceptedDecisions(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("accepted", ReviewRoomItem{Text: "Keep root budget tight for deploy changes.", Kind: "decision", Severity: "medium", Source: "founder"}); err != nil {
		t.Fatalf("seed accepted decision: %v", err)
	}
	room, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Re-open deploy budget review before rollout.", Kind: "agenda", Severity: "high", Source: "layer-osd", Why: "deploy budget needs another decision"})
	if err != nil {
		t.Fatalf("add structured review room item: %v", err)
	}
	if len(room.Open) != 1 {
		t.Fatalf("expected one open item, got %+v", room.Open)
	}
	if len(room.Open[0].Contradictions) != 1 || room.Open[0].Contradictions[0] != "Keep root budget tight for deploy changes." {
		t.Fatalf("expected detected contradiction, got %+v", room.Open[0])
	}
}

func TestReviewRoomTransitionMovesItemsAcrossStates(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Fix data race."); err != nil {
		t.Fatalf("add open item: %v", err)
	}
	room, err := service.TransitionReviewRoomItem("accept", "Fix data race.")
	if err != nil {
		t.Fatalf("accept review room item: %v", err)
	}
	if len(room.Open) != 0 || len(room.Accepted) != 1 || room.Accepted[0].Text != "Fix data race." {
		t.Fatalf("unexpected room after accept: %+v", room)
	}
	if room.Accepted[0].Resolution == nil || room.Accepted[0].Resolution.Rule != "review_room.transition.accept" {
		t.Fatalf("unexpected accept resolution: %+v", room.Accepted[0].Resolution)
	}
	room, err = service.TransitionReviewRoomItem("defer", "Fix data race.")
	if err != nil {
		t.Fatalf("defer review room item: %v", err)
	}
	if len(room.Accepted) != 0 || len(room.Deferred) != 1 || room.Deferred[0].Text != "Fix data race." {
		t.Fatalf("unexpected room after defer: %+v", room)
	}
	if room.Deferred[0].Resolution == nil || room.Deferred[0].Resolution.Rule != "review_room.transition.defer" {
		t.Fatalf("unexpected defer resolution: %+v", room.Deferred[0].Resolution)
	}
	if room.Deferred[0].WhyUnresolved == nil || *room.Deferred[0].WhyUnresolved == "" {
		t.Fatalf("expected why_unresolved to be stamped on defer, got %+v", room.Deferred[0])
	}
	room, err = service.TransitionReviewRoomItem("resolve", "Fix data race.")
	if err != nil {
		t.Fatalf("resolve review room item: %v", err)
	}
	if len(room.Accepted) != 0 || len(room.Open) != 0 || len(room.Deferred) != 0 {
		t.Fatalf("unexpected room after resolve: %+v", room)
	}
	events := service.ListEvents()
	last := events[len(events)-1]
	resolution, _ := last.Data["resolution"].(map[string]any)
	if last.Kind != "review_room.item_transitioned" || resolution["rule"] != "review_room.transition.resolve" {
		t.Fatalf("unexpected resolve event: %+v", last)
	}
}

func TestReviewRoomSummaryReadsRuntimeState(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("accepted", "Keep root budget tight."); err != nil {
		t.Fatalf("add accepted item: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Replace multi-process file writes with single-writer discipline."); err != nil {
		t.Fatalf("add open item: %v", err)
	}
	if _, err := service.AddReviewRoomItem("deferred", "Oracle Cloud ARM rollout."); err != nil {
		t.Fatalf("add deferred item: %v", err)
	}

	summary := service.ReviewRoomSummary()
	if summary.AcceptedCount != 1 || summary.OpenCount != 1 || summary.DeferredCount != 1 {
		t.Fatalf("unexpected review room counts: %+v", summary)
	}
	if len(summary.TopOpen) != 1 || !strings.Contains(summary.TopOpen[0].Text, "single-writer") {
		t.Fatalf("unexpected top open: %+v", summary.TopOpen)
	}
}

func TestFounderViewCompression(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Founder loop",
		Intent:           "compress view",
		Stage:            StageDiscover,
		Surface:          SurfaceCockpit,
		Pack:             "founder",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work: %v", err)
	}
	if err := service.CreateFlow(FlowRun{
		FlowID:     "flow_001",
		WorkItemID: "work_001",
		Status:     "active",
		Notes:      []string{"seed"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	view := service.FounderView()
	if len(view.Now) != 1 {
		t.Fatalf("expected 1 now card, got %d", len(view.Now))
	}
	if view.Now[0].Summary != "Founder loop" {
		t.Fatalf("expected founder summary from work title, got %q", view.Now[0].Summary)
	}
	if service.Handoff().FounderView.Now[0].Ref != "flow_001" {
		t.Fatalf("expected handoff founder_view to include flow_001")
	}
}

func TestFounderSummaryCompression(t *testing.T) {
	t.Setenv("LAYER_OS_HOST_CLASS", "vm")
	t.Setenv("LAYER_OS_POWER_MODE", "normal")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	flow, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"})
	if err != nil {
		t.Fatalf("start founder flow: %v", err)
	}
	if flow.Status != "waiting" {
		t.Fatalf("expected waiting flow, got %q", flow.Status)
	}

	summary := service.FounderSummary()
	if summary.PrimaryAction != "approve" {
		t.Fatalf("expected primary action approve, got %q", summary.PrimaryAction)
	}
	if summary.WaitingCount != 1 {
		t.Fatalf("expected waiting_count 1, got %d", summary.WaitingCount)
	}
	if summary.PriorityRationale == nil || summary.PriorityRationale.Lane != "waiting" || summary.PriorityRationale.Rule != "founder_priority.waiting" {
		t.Fatalf("unexpected waiting rationale: %+v", summary.PriorityRationale)
	}
	if summary.EnvironmentAdvisory.ContinuityRole != "continuity_host" || summary.EnvironmentAdvisory.AgentMode != "full" {
		t.Fatalf("unexpected environment advisory: %+v", summary.EnvironmentAdvisory)
	}
	if service.Handoff().FounderSummary.PrimaryAction != "approve" {
		t.Fatalf("expected handoff founder summary approve")
	}
}

func TestHandoffSurfacesParallelCandidates(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"}); err != nil {
		t.Fatalf("start founder flow 1: %v", err)
	}
	if _, err := service.StartFounderFlow("flow_002", "work_002", "approval_002", "Founder loop 2", "close founder loop 2", []string{"start"}); err != nil {
		t.Fatalf("start founder flow 2: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Review blocked deploy before approval."); err != nil {
		t.Fatalf("add review-room item 1: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate flaky verification path."); err != nil {
		t.Fatalf("add review-room item 2: %v", err)
	}

	handoff := service.Handoff()
	if handoff.FounderSummary.PrimaryAction != "review_room" {
		t.Fatalf("expected primary action review_room, got %q", handoff.FounderSummary.PrimaryAction)
	}
	if len(handoff.ParallelCandidates) == 0 {
		t.Fatal("expected parallel candidates")
	}
	if handoff.ParallelCandidates[0].Text != "Investigate flaky verification path." {
		t.Fatalf("unexpected first parallel candidate: %+v", handoff.ParallelCandidates)
	}
	if handoff.ParallelCandidates[0].Reason != "secondary review-room agenda" {
		t.Fatalf("unexpected parallel candidate reason: %+v", handoff.ParallelCandidates[0])
	}
	if handoff.FounderSummary.PriorityRationale == nil || handoff.FounderSummary.PriorityRationale.Lane != "review_open" {
		t.Fatalf("unexpected handoff rationale: %+v", handoff.FounderSummary.PriorityRationale)
	}
}

func TestFounderSummaryPrioritizesReviewRoomOpenAgenda(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"}); err != nil {
		t.Fatalf("start founder flow: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Review blocked deploy before approval."); err != nil {
		t.Fatalf("add review-room item: %v", err)
	}

	summary := service.FounderSummary()
	if summary.PrimaryAction != "review_room" {
		t.Fatalf("expected primary action review_room, got %q", summary.PrimaryAction)
	}
	if summary.PrimaryRef != "Review blocked deploy before approval." {
		t.Fatalf("unexpected primary ref: %q", summary.PrimaryRef)
	}
	if summary.ReviewOpenCount != 1 {
		t.Fatalf("expected review_open_count 1, got %d", summary.ReviewOpenCount)
	}
	if summary.PriorityRationale == nil || summary.PriorityRationale.Lane != "review_open" || summary.PriorityRationale.Source != "review_room.open" {
		t.Fatalf("unexpected review rationale: %+v", summary.PriorityRationale)
	}
	if len(summary.ReviewTopOpen) != 1 || summary.ReviewTopOpen[0] != "Review blocked deploy before approval." {
		t.Fatalf("unexpected review_top_open: %+v", summary.ReviewTopOpen)
	}
	handoff := service.Handoff()
	if handoff.FounderSummary.PrimaryAction != "review_room" {
		t.Fatalf("expected handoff founder summary review_room, got %q", handoff.FounderSummary.PrimaryAction)
	}
	if handoff.ReviewRoom.OpenCount != 1 {
		t.Fatalf("expected handoff review room open_count 1, got %d", handoff.ReviewRoom.OpenCount)
	}
}

func TestFounderActionFlow(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("put target: %v", err)
	}

	flow, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"})
	if err != nil {
		t.Fatalf("start founder flow: %v", err)
	}
	if flow.Status != "waiting" {
		t.Fatalf("expected waiting flow after start, got %q", flow.Status)
	}

	flow, err = service.ApproveFounderFlow("flow_001", []string{"approve"})
	if err != nil {
		t.Fatalf("approve founder flow: %v", err)
	}
	if flow.Status != "active" {
		t.Fatalf("expected active flow after approve, got %q", flow.Status)
	}

	flow, err = service.ReleaseFounderFlow("flow_001", "release_001", "deploy_001", "vm", "cockpit", []string{"release"})
	if err != nil {
		t.Fatalf("release founder flow: %v", err)
	}
	if flow.Status != "released" {
		t.Fatalf("expected released flow after release, got %q", flow.Status)
	}

	flow, err = service.RollbackFounderFlow("flow_001", "rollback_001", []string{"rollback"})
	if err != nil {
		t.Fatalf("rollback founder flow: %v", err)
	}
	if flow.Status != "rolled_back" {
		t.Fatalf("expected rolled_back flow after rollback, got %q", flow.Status)
	}
	if service.Handoff().Counts.Rollbacks != 1 {
		t.Fatalf("expected handoff rollback count 1, got %d", service.Handoff().Counts.Rollbacks)
	}
}

func TestReviewRoomSurfacesStructureAuditIssues(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "tmp"))
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 {
		t.Fatalf("expected audit-promoted review room issue, got %+v", room)
	}
	found := false
	for _, item := range room.Open {
		if item.Source == "audit.structure" && strings.Contains(item.Text, "tmp") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected structure audit review room item, got %+v", room.Open)
	}
	for _, item := range room.Open {
		if item.Source == "audit.structure" && strings.Contains(item.Text, "tmp") {
			if item.Rationale == nil || item.Rationale.Rule != "review_room.audit.structure" {
				t.Fatalf("unexpected structure rationale: %+v", item.Rationale)
			}
			if len(item.Evidence) != 1 || !strings.Contains(item.Evidence[0], "tmp") {
				t.Fatalf("unexpected structure evidence: %+v", item.Evidence)
			}
		}
	}
	summary := service.FounderSummary()
	if summary.PrimaryAction != "review_room" {
		t.Fatalf("expected founder summary to prioritize review room, got %q", summary.PrimaryAction)
	}
}

func TestReviewRoomSurfacesSurfaceAuditIssues(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd", "layer-osctl"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "cmd", "layer-osctl", "main.go"), `layer-osctl review-room [add|accept|defer|resolve] ...
case "review-room":
`)
	mustRuntimeWrite(t, filepath.Join(repoRoot, "internal", "api", "router.go"), `/api/layer-os/review-room`)
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	room := service.ReviewRoom()
	found := false
	for _, item := range room.Open {
		if item.Source == "audit.surface" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected surface audit review room item, got %+v", room.Open)
	}
}

func TestReviewRoomAggregatesGeminiAuditDrift(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "task.md.resolved"), "artifact output")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	room := service.ReviewRoom()
	foundGemini := false
	for _, item := range room.Open {
		if item.Source == "audit.gemini" {
			foundGemini = true
			if item.Rationale == nil || item.Rationale.Rule != "review_room.audit.gemini" {
				t.Fatalf("unexpected gemini rationale: %+v", item.Rationale)
			}
			joined := strings.Join(item.Evidence, "\n")
			if !strings.Contains(joined, ".gemini/GEMINI.md") || !strings.Contains(joined, "task.md.resolved") {
				t.Fatalf("unexpected gemini evidence: %+v", item.Evidence)
			}
		}
		if item.Source == "audit.residue" && strings.Contains(strings.Join(item.Evidence, "\n"), "task.md.resolved") {
			t.Fatalf("expected gemini artifact to stay aggregated, got %+v", item)
		}
	}
	if !foundGemini {
		t.Fatalf("expected gemini audit review room item, got %+v", room.Open)
	}
}

func TestReviewRoomClearsResolvedAuditIssues(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	cachePath := filepath.Join(repoRoot, ".cache")
	mustRuntimeMkdir(t, cachePath)
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if len(service.ReviewRoom().Open) == 0 {
		t.Fatal("expected initial audit issue")
	}
	if err := os.RemoveAll(cachePath); err != nil {
		t.Fatalf("remove cache path: %v", err)
	}
	room := service.ReviewRoom()
	for _, item := range room.Open {
		if item.Source == "audit.structure" && strings.Contains(item.Text, "tmp") {
			t.Fatalf("expected resolved audit issue to clear, got %+v", room.Open)
		}
	}
}

func mustRuntimeMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustRuntimeWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

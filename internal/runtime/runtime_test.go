package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type runtimeFileSnapshot struct {
	contents string
	modTime  time.Time
}

func snapshotRuntimeFile(t *testing.T, path string) runtimeFileSnapshot {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", filepath.Base(path), err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", filepath.Base(path), err)
	}
	return runtimeFileSnapshot{
		contents: string(raw),
		modTime:  info.ModTime(),
	}
}

func assertRuntimeFileUnchanged(t *testing.T, path string, before runtimeFileSnapshot) {
	t.Helper()

	after := snapshotRuntimeFile(t, path)
	if before.contents != after.contents {
		t.Fatalf("expected %s content unchanged", filepath.Base(path))
	}
	if !before.modTime.Equal(after.modTime) {
		t.Fatalf("expected %s modtime unchanged; before=%s after=%s", filepath.Base(path), before.modTime.Format(time.RFC3339Nano), after.modTime.Format(time.RFC3339Nano))
	}
}

func TestStatusReflectsWorkAndApprovalCounts(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Initialize runtime",
		Intent:           "bootstrap",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	approval := ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}
	if err := service.CreateApproval(approval); err != nil {
		t.Fatalf("create approval item: %v", err)
	}

	status := service.Status()
	if status.WorkItemsActive != 1 {
		t.Fatalf("expected 1 active work item, got %d", status.WorkItemsActive)
	}
	if status.ApprovalsPending != 1 {
		t.Fatalf("expected 1 pending approval, got %d", status.ApprovalsPending)
	}

	resolved, err := service.ResolveApproval("approval_001", "approved")
	if err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if resolved.Status != "approved" {
		t.Fatalf("expected approved status, got %q", resolved.Status)
	}

	status = service.Status()
	if status.ApprovalsPending != 0 {
		t.Fatalf("expected 0 pending approvals after resolve, got %d", status.ApprovalsPending)
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Initialize runtime",
		Intent:           "bootstrap",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	approval := ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Review release readiness",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}
	if err := service.CreateApproval(approval); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListWorkItems()); got != 1 {
		t.Fatalf("expected 1 reloaded work item, got %d", got)
	}
	if got := len(reloaded.ListApprovals()); got != 1 {
		t.Fatalf("expected 1 reloaded approval item, got %d", got)
	}
}

func TestInvalidJSONFailsFast(t *testing.T) {
	dataDir := t.TempDir()
	path := filepath.Join(dataDir, "work_items.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("seed invalid json: %v", err)
	}

	if _, err := NewService(dataDir); err == nil {
		t.Fatal("expected invalid json error, got nil")
	}
}

func TestWriteLeaseBlocksDifferentWriter(t *testing.T) {
	dataDir := t.TempDir()
	oldWriter := os.Getenv("LAYER_OS_WRITER_ID")
	defer os.Setenv("LAYER_OS_WRITER_ID", oldWriter)

	os.Setenv("LAYER_OS_WRITER_ID", "layer-osd")
	writerA, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service A: %v", err)
	}
	if err := writerA.CreateWorkItem(WorkItem{
		ID:               "work_001",
		Title:            "Lock writer",
		Intent:           "lease",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("writer A create work: %v", err)
	}

	os.Setenv("LAYER_OS_WRITER_ID", "layer-osctl")
	writerB, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service B: %v", err)
	}
	err = writerB.CreateWorkItem(WorkItem{
		ID:               "work_002",
		Title:            "Blocked writer",
		Intent:           "lease",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_002",
		CreatedAt:        time.Now().UTC(),
	})
	if err == nil || !strings.Contains(err.Error(), "write lease held by") {
		t.Fatalf("expected writer lease error, got %v", err)
	}
}

func TestMemoryRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	goal := "stabilize runtime"
	note := "next agent should wire release path"
	operator := "codex"
	memory := SystemMemory{
		CurrentFocus: "approval state machine",
		CurrentGoal:  &goal,
		NextSteps:    []string{"add release flow", "add deploy contract"},
		OpenRisks:    []string{"no release persistence yet"},
		HandoffNote:  &note,
		LastOperator: &operator,
		UpdatedAt:    time.Now().UTC(),
	}

	if err := service.ReplaceMemory(memory); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if reloaded.Memory().CurrentFocus != "approval state machine" {
		t.Fatalf("expected memory focus to persist, got %q", reloaded.Memory().CurrentFocus)
	}
}

func TestReplaceMemoryPersistsOnlyChangedDomains(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Seed work persistence",
		Intent:           "verify selective persistence",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	workPath := filepath.Join(dataDir, "work_items.json")
	statePath := filepath.Join(dataDir, "company_state.json")
	eventsPath := filepath.Join(dataDir, "events.json")
	workBefore := snapshotRuntimeFile(t, workPath)
	stateBefore := snapshotRuntimeFile(t, statePath)
	eventsBefore := snapshotRuntimeFile(t, eventsPath)

	time.Sleep(1100 * time.Millisecond)

	operator := "codex"
	memory := SystemMemory{
		CurrentFocus: "verify selective persistence",
		NextSteps:    []string{"compare runtime file timestamps"},
		OpenRisks:    []string{},
		LastOperator: &operator,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := service.ReplaceMemory(memory); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	assertRuntimeFileUnchanged(t, workPath, workBefore)
	assertRuntimeFileUnchanged(t, statePath, stateBefore)

	eventsAfter := snapshotRuntimeFile(t, eventsPath)
	if eventsBefore.contents == eventsAfter.contents {
		t.Fatal("expected events.json content to change after memory update")
	}
	if !eventsAfter.modTime.After(eventsBefore.modTime) {
		t.Fatalf("expected events.json modtime to advance; before=%s after=%s", eventsBefore.modTime.Format(time.RFC3339Nano), eventsAfter.modTime.Format(time.RFC3339Nano))
	}

	memoryPath := filepath.Join(dataDir, "system_memory.json")
	memoryAfter := snapshotRuntimeFile(t, memoryPath)
	if !strings.Contains(memoryAfter.contents, memory.CurrentFocus) {
		t.Fatalf("expected system_memory.json to contain %q, got %s", memory.CurrentFocus, memoryAfter.contents)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := reloaded.Memory().CurrentFocus; got != memory.CurrentFocus {
		t.Fatalf("expected reloaded memory focus %q, got %q", memory.CurrentFocus, got)
	}
}

func TestFlowRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Initialize runtime",
		Intent:           "bootstrap",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work item: %v", err)
	}
	if err := service.CreateFlow(FlowRun{
		FlowID:     "flow_001",
		WorkItemID: "work_001",
		Status:     "active",
		Notes:      []string{"test"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListFlows()); got != 1 {
		t.Fatalf("expected 1 reloaded flow, got %d", got)
	}
	if reloaded.Handoff().Counts.Flows != 1 {
		t.Fatalf("expected handoff flow count 1, got %d", reloaded.Handoff().Counts.Flows)
	}
}

func TestAdapterSummaryDefaults(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	adapters := service.Adapters()
	if adapters.Gateway != "record" || adapters.Verify != "command" || adapters.Deploy != "command" || adapters.Rollback != "command" {
		t.Fatalf("unexpected adapter summary names: %+v", adapters)
	}
	if adapters.GatewaySemantics != "record_only" || adapters.GatewayDispatchEnabled || adapters.GatewayRequiredMode != "single" {
		t.Fatalf("unexpected gateway adapter capabilities: %+v", adapters)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"}); err != nil {
		t.Fatalf("start founder flow: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Investigate snapshot continuity.", Kind: "risk", Severity: "high", Source: "agent.codex"}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	packet := service.Snapshot()
	if len(packet.ReviewRoom.Open) != 1 || packet.ReviewRoom.Open[0].Text != "Investigate snapshot continuity." {
		t.Fatalf("expected snapshot to carry review room, got %+v", packet.ReviewRoom)
	}

	reloaded, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := reloaded.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	if reloaded.FounderSummary().PrimaryAction != "review_room" {
		t.Fatalf("expected imported snapshot to preserve review-room priority, got %+v", reloaded.FounderSummary())
	}
	room := reloaded.ReviewRoom()
	if len(room.Open) != 1 || room.Open[0].Text != "Investigate snapshot continuity." {
		t.Fatalf("expected imported snapshot to preserve review room, got %+v", room)
	}
}

func TestSnapshotImportNormalizesLegacyDualMode(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	packet := service.Snapshot()
	packet.Policies = append(packet.Policies, PolicyDecision{
		DecisionID:       "policy_legacy",
		Intent:           "legacy dual",
		Scope:            "kernel",
		Risk:             "high",
		Novelty:          "high",
		TokenClass:       "large",
		RequiresApproval: true,
		Mode:             "dual",
		Decision:         "go",
		Reasons:          []string{"legacy payload"},
		CreatedAt:        time.Now().UTC(),
	})
	packet.Executes = append(packet.Executes, ExecuteRun{
		ExecuteID:        "execute_legacy",
		WorkItemID:       "work_legacy",
		PolicyDecisionID: "policy_legacy",
		Mode:             "dual",
		Status:           "recorded",
		Notes:            []string{"legacy payload"},
		StartedAt:        time.Now().UTC(),
	})
	packet.WorkItems = append(packet.WorkItems, WorkItem{
		ID:               "work_legacy",
		Title:            "legacy work",
		Intent:           "legacy dual",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "high",
		RequiresApproval: true,
		Payload:          map[string]any{},
		CorrelationID:    "corr_legacy",
		CreatedAt:        time.Now().UTC(),
	})

	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	if got := target.ListPolicies()[0].Mode; got == "dual" {
		t.Fatalf("expected legacy dual policy to normalize, got %q", got)
	}
	if got := target.ListExecutes()[0].Mode; got == "dual" {
		t.Fatalf("expected legacy dual execute to normalize, got %q", got)
	}
}

func TestSnapshotImportRejectsActiveAuth(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	packet := service.Snapshot()
	packet.Auth.WriteAuthEnabled = true

	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err == nil {
		t.Fatal("expected auth-enabled snapshot import failure")
	}
}

func TestHandoffPacketReflectsRuntimeState(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	note := "next agent continues here"
	goal := "keep baton intact"
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "handoff contract",
		CurrentGoal:  &goal,
		NextSteps:    []string{"wire api route"},
		OpenRisks:    []string{"deploy adapter still local"},
		HandoffNote:  &note,
		Refs:         []string{"proposal_001"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	if _, err := service.SessionNote(SessionNoteInput{
		Source: "terminal",
		Kind:   continuityNoteKindTodo,
		Text:   "resume agent baton after bootstrap",
		Refs:   []string{"proposal_001"},
	}); err != nil {
		t.Fatalf("session note: %v", err)
	}

	handoff := service.Handoff()
	if err := handoff.Validate(); err != nil {
		t.Fatalf("validate handoff: %v", err)
	}
	if handoff.CompanyState.PrimarySurface != SurfaceCockpit {
		t.Fatalf("expected cockpit primary surface, got %q", handoff.CompanyState.PrimarySurface)
	}
	if handoff.Counts.WorkItems != 0 || handoff.Counts.Approvals != 0 || handoff.Counts.Releases != 0 || handoff.Counts.Deploys != 0 {
		t.Fatalf("expected empty counts, got %+v", handoff.Counts)
	}
	if handoff.SystemMemory.CurrentFocus != "handoff contract" {
		t.Fatalf("expected handoff memory focus to match, got %q", handoff.SystemMemory.CurrentFocus)
	}
	if handoff.Continuity == nil || handoff.Continuity.Record == nil {
		t.Fatalf("expected handoff continuity record, got %+v", handoff.Continuity)
	}
	if handoff.Continuity.Record.CurrentFocus != "handoff contract" {
		t.Fatalf("expected continuity focus to match, got %+v", handoff.Continuity.Record)
	}
	if len(handoff.Continuity.Suggestions) == 0 {
		t.Fatalf("expected continuity suggestions, got %+v", handoff.Continuity)
	}
}

func TestWriteAuthRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if service.AuthStatus().WriteAuthEnabled {
		t.Fatal("expected write auth disabled by default")
	}

	status, err := service.SetWriteToken("secret-token")
	if err != nil {
		t.Fatalf("set write token: %v", err)
	}
	if !status.WriteAuthEnabled {
		t.Fatal("expected write auth enabled after set")
	}
	if !service.AuthorizeWriteToken("secret-token") {
		t.Fatal("expected correct token to authorize")
	}
	if service.AuthorizeWriteToken("wrong-token") {
		t.Fatal("expected wrong token to fail authorization")
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if !reloaded.AuthStatus().WriteAuthEnabled {
		t.Fatal("expected write auth to persist after reload")
	}

	status, err = reloaded.ClearWriteToken()
	if err != nil {
		t.Fatalf("clear write token: %v", err)
	}
	if status.WriteAuthEnabled {
		t.Fatal("expected write auth disabled after clear")
	}
}

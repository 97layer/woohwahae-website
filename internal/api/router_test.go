package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func seedReadySecurityReview(t *testing.T, service *runtime.Service) {
	t.Helper()
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	if err := service.CreatePreflight(runtime.PreflightRecord{
		RecordID:    "security_review_001",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}
}

func TestStatusRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	seedReadySecurityReview(t, service)
	startedAt := time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC)
	router := NewRouterWithRuntime(service, DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: startedAt})

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Status       string                `json:"status"`
		CompanyState runtime.CompanyState  `json:"company_state"`
		Daemon       runtime.DaemonStatus  `json:"daemon"`
		Security     runtime.SecurityAudit `json:"security"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("expected ok status, got %q", response.Status)
	}
	if response.Daemon.Address != "127.0.0.1:17808" || response.Daemon.Service != "layer-osd" {
		t.Fatalf("unexpected daemon status: %+v", response.Daemon)
	}
	if response.Daemon.StartedAt.IsZero() {
		t.Fatalf("expected started_at, got %+v", response.Daemon)
	}
	if response.CompanyState.Progress == nil || len(response.CompanyState.Progress.Axes) != 7 {
		t.Fatalf("expected seven-axis progress dashboard, got %+v", response.CompanyState.Progress)
	}
	if response.Security.Status != "ok" {
		t.Fatalf("expected ok security posture, got %+v", response.Security)
	}
}

func TestHealthzRouteReturnsDaemonStatus(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	packet := service.Snapshot()
	failedAt := time.Now().UTC()
	packet.Deploys = []runtime.DeployRun{{DeployID: "deploy_001", ReleaseID: "release_001", Target: "vm", Status: "failed", Notes: []string{}, StartedAt: failedAt, FinishedAt: &failedAt}}
	if err := service.ImportSnapshot(packet); err != nil {
		t.Fatalf("seed degraded deploy state: %v", err)
	}
	seedReadySecurityReview(t, service)
	startedAt := time.Now().UTC().Add(-2 * time.Minute)
	router := NewRouterWithRuntime(service, DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: startedAt})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var item runtime.DaemonStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode healthz response: %v", err)
	}
	if item.Status != "degraded" {
		t.Fatalf("expected degraded health status, got %+v", item)
	}
	if len(item.DegradedReasons) == 0 || item.DegradedReasons[0] != "deploy_health=degraded" {
		t.Fatalf("expected degraded reasons, got %+v", item.DegradedReasons)
	}
}

func TestDaemonRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouterWithRuntime(service, DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: time.Now().UTC().Add(-time.Minute)})

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/daemon", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Daemon runtime.DaemonStatus `json:"daemon"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode daemon response: %v", err)
	}
	if err := response.Daemon.Validate(); err != nil {
		t.Fatalf("invalid daemon status payload: %v", err)
	}
}

func TestDaemonRouteDegradesWhenSourceIsNewerThanBoundDaemonStart(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, "internal", "api"), 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	source := filepath.Join(repoRoot, "internal", "api", "router.go")
	if err := os.WriteFile(source, []byte("package api"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	startedAt := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	baseline := startedAt.Add(-3 * time.Minute)
	if err := os.Chtimes(filepath.Join(repoRoot, "go.mod"), baseline, baseline); err != nil {
		t.Fatalf("chtimes baseline go.mod: %v", err)
	}
	newer := startedAt.Add(3 * time.Minute)
	if err := os.Chtimes(source, newer, newer); err != nil {
		t.Fatalf("chtimes source: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	seedReadySecurityReview(t, service)
	service.BindDaemonRuntime(startedAt)
	router := NewRouterWithRuntime(service, DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: startedAt})

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/daemon", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Daemon runtime.DaemonStatus `json:"daemon"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode daemon response: %v", err)
	}
	if response.Daemon.Status != "degraded" {
		t.Fatalf("expected degraded daemon status, got %+v", response.Daemon)
	}
	joined := strings.Join(response.Daemon.DegradedReasons, "\n")
	if !strings.Contains(joined, "daemon_source_freshness=degraded") {
		t.Fatalf("expected daemon freshness degraded reason, got %+v", response.Daemon.DegradedReasons)
	}
}

func TestDaemonRouteIncludesArchitectStatus(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	startedAt := time.Now().UTC().Add(-time.Minute)
	lastRunAt := startedAt.Add(30 * time.Second)
	lastError := "dispatch drift"
	router := NewRouterWithRuntime(service, DaemonRuntimeInfo{
		Address:   "127.0.0.1:17808",
		StartedAt: startedAt,
		ArchitectStatus: func() *runtime.DaemonArchitectStatus {
			return &runtime.DaemonArchitectStatus{
				Enabled:        true,
				AutoDispatch:   false,
				Interval:       "15s",
				PromoteLimit:   1,
				LastRunAt:      &lastRunAt,
				LastError:      &lastError,
				LastConsidered: 3,
				LastCreated:    1,
				LastExisting:   1,
				LastDispatched: 0,
				LastFailed:     1,
			}
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/daemon", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Daemon runtime.DaemonStatus `json:"daemon"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode daemon response: %v", err)
	}
	if err := response.Daemon.Validate(); err != nil {
		t.Fatalf("invalid daemon status payload: %v", err)
	}
	if response.Daemon.Status != "degraded" {
		t.Fatalf("expected degraded daemon status, got %+v", response.Daemon)
	}
	if response.Daemon.Architect == nil || !response.Daemon.Architect.Enabled {
		t.Fatalf("expected architect status, got %+v", response.Daemon.Architect)
	}
	if len(response.Daemon.DegradedReasons) == 0 || response.Daemon.DegradedReasons[len(response.Daemon.DegradedReasons)-1] != "architect_loop=degraded" {
		t.Fatalf("expected architect degraded reason, got %+v", response.Daemon.DegradedReasons)
	}
}

func TestCockpitRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/cockpit", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCockpitSnapshotRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	seedReadySecurityReview(t, service)
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "founder-chat")
	t.Setenv("GEMINI_API_KEY", "gemini")
	sourceObservation, err := service.CreateObservation(runtime.ObservationRecord{
		SourceChannel:     "crawler",
		CapturedAt:        time.Now().UTC(),
		Actor:             "rss_sensor",
		Topic:             "content_capture",
		Refs:              []string{"content_doc:feed_001", "route:woohwahae", "feed_source:https://example.com/feed.xml", "feed_kind:article"},
		Confidence:        "high",
		RawExcerpt:        "rss excerpt",
		NormalizedSummary: "slow note",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	sourceIntake, _, err := runtime.EnsureSourceIntakeObservation(service, sourceObservation, runtime.BuildSourceIntakeRecordFromContent(runtime.SourceIntakeContentInput{
		SourceObservation: sourceObservation,
		IntakeClass:       "public_collector",
		PolicyColor:       "green",
		Title:             "slow note",
		URL:               "https://example.com/slow-note",
		Excerpt:           "rss excerpt",
	}))
	if err != nil {
		t.Fatalf("ensure source intake observation: %v", err)
	}
	if _, err := runtime.OpenSourceIntakeDraftLane(service, sourceIntake, "woohwahae", "rss_sensor", "feed_sensor"); err != nil {
		t.Fatalf("open source intake draft lane: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/cockpit", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Status       string                               `json:"status"`
		Daemon       runtime.DaemonStatus                 `json:"daemon"`
		CompanyState runtime.CompanyState                 `json:"company_state"`
		Knowledge    runtime.KnowledgePacket              `json:"knowledge"`
		Conversation runtime.ConversationAutomationStatus `json:"conversation"`
		Capabilities runtime.CapabilityRegistry           `json:"capabilities"`
		Telegram     cockpitTelegramSnapshot              `json:"telegram"`
		Threads      cockpitThreadsSnapshot               `json:"threads"`
		Providers    []runtime.ProviderSummary            `json:"providers"`
		SourceIntake cockpitSourceIntakeSnapshot          `json:"source_intake"`
		Dashboard    cockpitDashboardSnapshot             `json:"dashboard"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode cockpit response: %v", err)
	}
	if response.Status != "ok" || response.Daemon.Status != "ok" {
		t.Fatalf("expected cockpit live status bundle, got status=%q daemon=%+v", response.Status, response.Daemon)
	}
	if response.CompanyState.Progress == nil || len(response.CompanyState.Progress.Axes) != 7 {
		t.Fatalf("expected cockpit company_state progress, got %+v", response.CompanyState.Progress)
	}
	if err := response.Capabilities.Validate(); err != nil {
		t.Fatalf("expected cockpit capabilities bundle, got %v", err)
	}
	if response.Telegram.Status.InboundMode != "assistant" || response.Telegram.Status.Adapter == "" {
		t.Fatalf("expected cockpit telegram bundle, got %+v", response.Telegram)
	}
	if response.Threads.Status.Adapter == "" {
		t.Fatalf("expected cockpit threads bundle, got %+v", response.Threads)
	}
	if response.Knowledge.ReviewTopOpen == nil {
		t.Fatalf("expected cockpit knowledge bundle, got %+v", response.Knowledge)
	}
	if response.SourceIntake.RecentCount == 0 || response.SourceIntake.Attention.Mode != "open_threads_prep" {
		t.Fatalf("expected cockpit source intake summary, got %+v", response.SourceIntake)
	}
	if strings.TrimSpace(response.SourceIntake.SummaryNote) == "" || strings.TrimSpace(response.SourceIntake.SummaryMeta) == "" {
		t.Fatalf("expected cockpit source intake summary strings, got %+v", response.SourceIntake)
	}
	if strings.TrimSpace(response.SourceIntake.QuietNote) == "" {
		t.Fatalf("expected cockpit source intake quiet note, got %+v", response.SourceIntake)
	}
	if strings.TrimSpace(response.Dashboard.JobCounts.SummaryNote) == "" || strings.TrimSpace(response.Dashboard.JobCounts.SummaryMeta) == "" {
		t.Fatalf("expected cockpit job summary strings, got %+v", response.Dashboard.JobCounts)
	}
	if strings.TrimSpace(response.Dashboard.ReviewNote) == "" || strings.TrimSpace(response.Dashboard.ReviewMeta) == "" {
		t.Fatalf("expected cockpit review summary strings, got review_note=%q review_meta=%q", response.Dashboard.ReviewNote, response.Dashboard.ReviewMeta)
	}
	if strings.TrimSpace(response.Dashboard.FounderAttention.Summary) == "" || strings.TrimSpace(response.Dashboard.PrimaryAttention.Summary) == "" {
		t.Fatalf("expected cockpit founder attention summaries, got founder=%+v primary=%+v", response.Dashboard.FounderAttention, response.Dashboard.PrimaryAttention)
	}
	if strings.TrimSpace(response.Dashboard.FounderFlow.Attention.Summary) == "" {
		t.Fatalf("expected cockpit founder flow defaults, got %+v", response.Dashboard.FounderFlow)
	}
	if response.Providers == nil {
		t.Fatalf("expected cockpit snapshot to include providers, got %+v", response)
	}
	for _, item := range response.Providers {
		if item.AuthReady || item.AuthSource != nil || len(item.AuthEnvKeys) != 0 {
			t.Fatalf("expected cockpit provider auth metadata to be redacted, got %+v", item)
		}
		for _, note := range item.Notes {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(note)), "provider credentials ") {
				t.Fatalf("expected cockpit provider notes to redact auth metadata, got %+v", item.Notes)
			}
		}
	}
	for _, item := range response.Dashboard.Providers {
		if item.AuthReady || item.AuthSource != nil || len(item.AuthEnvKeys) != 0 {
			t.Fatalf("expected dashboard provider auth metadata to be redacted, got %+v", item)
		}
	}
	if response.Dashboard.GeneratedAt.IsZero() || response.Dashboard.Daemon.Service != "layer-osd" {
		t.Fatalf("expected unified dashboard payload, got %+v", response.Dashboard)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode cockpit raw payload: %v", err)
	}
	for _, key := range []string{"gateway_calls", "events", "snapshot", "memory", "review_room", "branches", "work_items", "flows", "approvals"} {
		if _, exists := payload[key]; exists {
			t.Fatalf("expected slim cockpit payload to omit %q, got keys=%v", key, mapsKeys(payload))
		}
	}
	for _, key := range []string{"dashboard", "releases", "deploys", "rollbacks", "targets", "verifications"} {
		if _, exists := payload[key]; !exists {
			t.Fatalf("expected slim cockpit payload to include %q, got keys=%v", key, mapsKeys(payload))
		}
	}
	if _, exists := payload["source_intake"]; !exists {
		t.Fatalf("expected slim cockpit payload to include source_intake, got keys=%v", mapsKeys(payload))
	}
}

func TestCockpitSnapshotRoutePrefersPrepLaneAttention(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	seedReadySecurityReview(t, service)
	sourceObservation, err := service.CreateObservation(runtime.ObservationRecord{
		SourceChannel:     "crawler",
		CapturedAt:        time.Now().UTC(),
		Actor:             "rss_sensor",
		Topic:             "content_capture",
		Refs:              []string{"content_doc:feed_002", "route:woohwahae", "feed_source:https://example.com/feed.xml"},
		Confidence:        "high",
		RawExcerpt:        "rss excerpt",
		NormalizedSummary: "slow note",
	})
	if err != nil {
		t.Fatalf("create source observation: %v", err)
	}
	sourceIntake, _, err := runtime.EnsureSourceIntakeObservation(service, sourceObservation, runtime.BuildSourceIntakeRecordFromContent(runtime.SourceIntakeContentInput{
		SourceObservation: sourceObservation,
		IntakeClass:       "public_collector",
		PolicyColor:       "green",
		Title:             "slow note",
		URL:               "https://example.com/slow-note",
		Excerpt:           "rss excerpt",
	}))
	if err != nil {
		t.Fatalf("ensure source intake observation: %v", err)
	}
	lane, err := runtime.OpenSourceIntakeDraftLane(service, sourceIntake, "woohwahae", "rss_sensor", "feed_sensor")
	if err != nil {
		t.Fatalf("open source intake draft lane: %v", err)
	}
	if _, err := service.OpenSourceDraftSeedPublishPrep(lane.DraftSeed.ObservationID, "threads"); err != nil {
		t.Fatalf("open source draft seed publish prep: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/cockpit", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		SourceIntake cockpitSourceIntakeSnapshot `json:"source_intake"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode cockpit response: %v", err)
	}
	if response.SourceIntake.PrepCount != 1 || response.SourceIntake.Attention.Mode != "monitor_prep_lane" {
		t.Fatalf("expected prep-aware cockpit summary, got %+v", response.SourceIntake)
	}
}

func TestCockpitSnapshotRouteFullIncludesExpandedFields(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	seedReadySecurityReview(t, service)
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/cockpit?full=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode cockpit full payload: %v", err)
	}
	for _, key := range []string{"gateway_calls", "events", "snapshot", "memory", "review_room", "branches", "work_items", "flows", "approvals"} {
		if _, exists := payload[key]; !exists {
			t.Fatalf("expected full cockpit payload to include %q, got keys=%v", key, mapsKeys(payload))
		}
	}
}

func mapsKeys[K comparable, V any](items map[K]V) []K {
	keys := make([]K, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}

func TestHandoffRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CheckpointSession(runtime.SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Investigate flaky verification path",
		NextSteps:    []string{"resume route check"},
		OpenRisks:    []string{"single review item still open"},
		Refs:         []string{"review_room"},
	}); err != nil {
		t.Fatalf("seed continuity: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate flaky verification path."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/handoff", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Handoff runtime.HandoffPacket `json:"handoff"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode handoff response: %v", err)
	}
	if err := response.Handoff.Validate(); err != nil {
		t.Fatalf("invalid handoff payload: %v", err)
	}
	if len(response.Handoff.ParallelCandidates) != 0 {
		t.Fatalf("expected no parallel candidates with single review item, got %+v", response.Handoff.ParallelCandidates)
	}
	if response.Handoff.Continuity == nil || response.Handoff.Continuity.Record == nil {
		t.Fatalf("expected continuity in handoff route, got %+v", response.Handoff)
	}
	if response.Handoff.Continuity.Record.CurrentFocus != "Investigate flaky verification path" {
		t.Fatalf("expected continuity focus in handoff route, got %+v", response.Handoff.Continuity.Record)
	}
}

func TestCorpusEntriesRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	goal := "Keep official surfaces aligned"
	result, err := service.FinishSession(runtime.SessionFinishInput{CurrentFocus: "Close queue drift", CurrentGoal: &goal, NextSteps: []string{"restart daemon"}, OpenRisks: []string{"stale api"}})
	if err != nil {
		t.Fatalf("finish session: %v", err)
	}
	if result.Event.Kind != "session.finished" {
		t.Fatalf("unexpected event: %+v", result.Event)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/corpus/entries", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Entries []runtime.CapitalizationEntry `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode corpus response: %v", err)
	}
	if len(response.Entries) != 1 || response.Entries[0].SourceKind != "session.finished" {
		t.Fatalf("unexpected corpus entries: %+v", response.Entries)
	}
}

func TestProvidersRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Setenv("LAYER_OS_PROVIDERS", "openai,anthropic")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "openai=https://provider.example/v1/respond")
	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/providers", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Providers []runtime.ProviderSummary `json:"providers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode providers response: %v", err)
	}
	if len(response.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %+v", response.Providers)
	}
	if response.Providers[0].Provider != "anthropic" || response.Providers[1].Provider != "openai" {
		t.Fatalf("unexpected provider order: %+v", response.Providers)
	}
	for _, item := range response.Providers {
		if item.AuthReady || item.AuthSource != nil || len(item.AuthEnvKeys) != 0 {
			t.Fatalf("expected providers route to redact auth metadata, got %+v", item)
		}
		for _, note := range item.Notes {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(note)), "provider credentials ") {
				t.Fatalf("expected providers route notes to redact auth metadata, got %+v", item.Notes)
			}
		}
	}
}

func TestTelegramRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Founder queue still needs triage."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/telegram", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Telegram runtime.TelegramPacket `json:"telegram"`
		Enabled  bool                   `json:"enabled"`
		Adapter  string                 `json:"adapter"`
		Status   runtime.TelegramStatus `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode telegram response: %v", err)
	}
	if response.Telegram.ReviewOpenCount != 1 || response.Telegram.Headline == "" {
		t.Fatalf("unexpected telegram packet: %+v", response.Telegram)
	}
	if response.Enabled {
		t.Fatalf("expected default telegram adapter to be disabled")
	}
	if response.Adapter == "" {
		t.Fatalf("expected telegram adapter metadata in response")
	}
	if response.Status.Adapter == "" || response.Status.InboundMode == "" || response.Status.FounderDelivery == "" {
		t.Fatalf("expected telegram readiness status in response, got %+v", response.Status)
	}
	if response.Status.SendAdapter == "" || len(response.Status.Routes) != 3 {
		t.Fatalf("expected telegram route metadata in response, got %+v", response.Status)
	}
}

func TestThreadsRoute(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "")

	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/social/threads", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Enabled bool                  `json:"enabled"`
		Adapter string                `json:"adapter"`
		Status  runtime.ThreadsStatus `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode threads response: %v", err)
	}
	if response.Enabled || response.Adapter == "" || response.Status.Adapter == "" {
		t.Fatalf("unexpected threads status response: %+v", response)
	}
}

func TestKnowledgeRoute(t *testing.T) {
	t.Setenv("LAYER_OS_HOST_CLASS", "laptop")
	t.Setenv("LAYER_OS_POWER_MODE", "low")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate flaky verification path."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/knowledge", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Knowledge runtime.KnowledgePacket `json:"knowledge"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode knowledge response: %v", err)
	}
	if response.Knowledge.PrimaryAction != "review_room" {
		t.Fatalf("expected review_room primary action, got %+v", response.Knowledge)
	}
	if response.Knowledge.ReviewOpenCount != 1 {
		t.Fatalf("expected review open count 1, got %+v", response.Knowledge)
	}
	if response.Knowledge.CorpusLessons == nil || len(response.Knowledge.CorpusLessons) != 0 {
		t.Fatalf("expected empty corpus lessons in knowledge route, got %+v", response.Knowledge.CorpusLessons)
	}
	if response.Knowledge.ObservationLinks == nil || response.Knowledge.ProposalCandidates == nil {
		t.Fatalf("expected observation/proposal gate fields in knowledge route, got %+v", response.Knowledge)
	}
	joined := strings.ToLower(strings.Join(response.Knowledge.Surprising, "\n"))
	for _, needle := range []string{"open_risks", "next_steps"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected knowledge route surprising signal %q, got %+v", needle, response.Knowledge.Surprising)
		}
	}
	actionHints := strings.ToLower(strings.Join(response.Knowledge.ActionHints, "\n"))
	for _, needle := range []string{"define next step", "state explicit risk"} {
		if !strings.Contains(actionHints, needle) {
			t.Fatalf("expected knowledge route action hint %q, got %+v", needle, response.Knowledge.ActionHints)
		}
	}
	if len(response.Knowledge.ActionRoutes) < 2 {
		t.Fatalf("expected action routes in knowledge route, got %+v", response.Knowledge.ActionRoutes)
	}
	for _, route := range response.Knowledge.ActionRoutes[:2] {
		if route.TargetLane != "session_memory" {
			t.Fatalf("expected session_memory action route, got %+v", route)
		}
	}
	if response.Knowledge.EnvironmentAdvisory.AgentMode != "conserve" {
		t.Fatalf("unexpected knowledge advisory: %+v", response.Knowledge.EnvironmentAdvisory)
	}
}

func TestFounderViewRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/founder-view", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFounderSummaryRoute(t *testing.T) {
	t.Setenv("LAYER_OS_HOST_CLASS", "vm")
	t.Setenv("LAYER_OS_POWER_MODE", "normal")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Review blocked deploy before approval."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/founder-summary", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		FounderSummary runtime.FounderSummary `json:"founder_summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode founder summary response: %v", err)
	}
	if response.FounderSummary.PrimaryAction != "review_room" {
		t.Fatalf("expected review_room action, got %q", response.FounderSummary.PrimaryAction)
	}
	if response.FounderSummary.ReviewOpenCount != 1 {
		t.Fatalf("expected review_open_count 1, got %d", response.FounderSummary.ReviewOpenCount)
	}
	if response.FounderSummary.PriorityRationale == nil || response.FounderSummary.PriorityRationale.Rule != "founder_priority.review_open" {
		t.Fatalf("unexpected founder rationale: %+v", response.FounderSummary.PriorityRationale)
	}
	if response.FounderSummary.EnvironmentAdvisory.ContinuityRole != "continuity_host" {
		t.Fatalf("unexpected founder advisory: %+v", response.FounderSummary.EnvironmentAdvisory)
	}
}

func TestReviewRoomRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "issue_001"
	if _, err := service.AddStructuredReviewRoomItem("accepted", runtime.ReviewRoomItem{
		Text:     "Keep root budget tight.",
		Kind:     "agenda",
		Severity: "medium",
		Source:   "agent.codex",
		Ref:      &ref,
		Rationale: &runtime.ReviewRoomRationale{
			Trigger: "manual.add",
			Reason:  "operator explicitly promoted this housekeeping task",
			Rule:    "review_room.manual.add",
		},
		Evidence: []string{"note:root-budget", "owner:codex"},
	}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/review-room", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		ReviewRoom runtime.ReviewRoom `json:"review_room"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode review room response: %v", err)
	}
	if len(response.ReviewRoom.Accepted) != 1 {
		t.Fatalf("unexpected accepted items: %+v", response.ReviewRoom.Accepted)
	}
	item := response.ReviewRoom.Accepted[0]
	if item.Rationale == nil || item.Rationale.Rule != "review_room.manual.add" {
		t.Fatalf("unexpected review room rationale: %+v", item.Rationale)
	}
	if len(item.Evidence) != 2 || item.Evidence[0] != "note:root-budget" {
		t.Fatalf("unexpected review room evidence: %+v", item.Evidence)
	}
}

func TestReviewRoomTransitionRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Fix data race."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{"action": "accept", "item": "Fix data race.", "reason": "founder approved handling", "rule": "review_room.transition.accept.manual", "evidence": []string{"decision:founder", "owner:codex"}}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/review-room", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	room := service.ReviewRoom()
	if len(room.Open) != 0 || len(room.Accepted) != 1 || room.Accepted[0].Text != "Fix data race." {
		t.Fatalf("unexpected review room state: %+v", room)
	}
	if room.Accepted[0].Resolution == nil || room.Accepted[0].Resolution.Rule != "review_room.transition.accept.manual" {
		t.Fatalf("unexpected review room resolution: %+v", room.Accepted[0].Resolution)
	}
}

func TestReviewRoomAddRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{"section": "open", "item": "Fix data race.", "kind": "bug", "severity": "high", "source": "agent.codex", "ref": "issue_001", "why": "reduce deploy-risk before retry", "why_unresolved": "race root cause still unknown", "contradictions": []string{"accepted mitigation exists but failure persists", "keep root budget tight"}, "pattern_refs": []string{"issue_001", "session_003"}}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/review-room", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	room := service.ReviewRoom()
	if len(room.Open) != 1 || room.Open[0].Text != "Fix data race." {
		t.Fatalf("unexpected review room open items: %+v", room.Open)
	}
	if room.Open[0].Kind != "bug" || room.Open[0].Severity != "high" || room.Open[0].Source != "agent.codex" {
		t.Fatalf("unexpected review room metadata: %+v", room.Open[0])
	}
	if room.Open[0].Why != "reduce deploy-risk before retry" {
		t.Fatalf("unexpected review room why: %+v", room.Open[0])
	}
	if room.Open[0].WhyUnresolved == nil || *room.Open[0].WhyUnresolved != "race root cause still unknown" {
		t.Fatalf("unexpected review room why_unresolved: %+v", room.Open[0])
	}
	if len(room.Open[0].Contradictions) != 2 || room.Open[0].Contradictions[1] != "keep root budget tight" {
		t.Fatalf("unexpected review room contradictions: %+v", room.Open[0])
	}
	if len(room.Open[0].PatternRefs) != 2 || room.Open[0].PatternRefs[1] != "session_003" {
		t.Fatalf("unexpected review room pattern refs: %+v", room.Open[0])
	}
}

func TestAdaptersRoute(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/adapters", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Adapters runtime.AdapterSummary `json:"adapters"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode adapters response: %v", err)
	}
	if response.Adapters.GatewaySemantics != "record_only" || response.Adapters.GatewayDispatchEnabled || response.Adapters.GatewayRequiredMode != "single" {
		t.Fatalf("unexpected gateway adapter capabilities: %+v", response.Adapters)
	}
}

func TestCapabilitiesRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/capabilities", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Capabilities runtime.CapabilityRegistry `json:"capabilities"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode capabilities response: %v", err)
	}
	if response.Capabilities.DefaultActor != "system" {
		t.Fatalf("expected neutral default actor system, got %+v", response.Capabilities)
	}
	if len(response.Capabilities.Bindings) == 0 {
		t.Fatalf("expected capability bindings, got %+v", response.Capabilities)
	}
}

func TestWriterRoute(t *testing.T) {
	oldWriter := os.Getenv("LAYER_OS_WRITER_ID")
	defer os.Setenv("LAYER_OS_WRITER_ID", oldWriter)
	os.Setenv("LAYER_OS_WRITER_ID", "layer-osctl")

	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(runtime.WorkItem{
		ID:               "work_001",
		Title:            "Lease",
		Intent:           "writer",
		Stage:            runtime.StageDiscover,
		Surface:          runtime.SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/writer", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSnapshotRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate snapshot continuity."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/snapshot", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Snapshot runtime.SnapshotPacket `json:"snapshot"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode snapshot response: %v", err)
	}
	if len(response.Snapshot.ReviewRoom.Open) != 1 || response.Snapshot.ReviewRoom.Open[0].Text != "Investigate snapshot continuity." {
		t.Fatalf("expected review room in snapshot route, got %+v", response.Snapshot.ReviewRoom)
	}
}

func TestContractsAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/contracts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSnapshotImportRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"generated_at": time.Now().UTC(),
		"company_state": map[string]any{
			"shell_mode":        "founder",
			"approvals_pending": 0,
			"work_items_active": 0,
			"memory_health":     "ready",
			"deploy_health":     "ready",
			"primary_surface":   "cockpit",
			"active_surfaces":   []string{"cockpit", "telegram", "api"},
		},
		"system_memory": map[string]any{
			"current_focus": "snapshot import",
			"next_steps":    []string{},
			"open_risks":    []string{},
			"updated_at":    time.Now().UTC(),
		},
		"auth":       map[string]any{"write_auth_enabled": false},
		"work_items": []any{},
		"flows":      []any{},
		"approvals":  []any{},
		"releases":   []any{},
		"deploys":    []any{},
		"rollbacks":  []any{},
		"targets":    []any{},
		"events":     []any{},
		"review_room": map[string]any{
			"source":   ".layer-os/review_room.json",
			"accepted": []any{},
			"open": []any{map[string]any{
				"text":       "Imported review room item.",
				"kind":       "agenda",
				"severity":   "high",
				"source":     "import.test",
				"created_at": time.Now().UTC(),
				"updated_at": time.Now().UTC(),
			}},
			"deferred": []any{},
			"issues":   []any{},
		},
		"preflights":    []any{},
		"policies":      []any{},
		"gateway_calls": []any{},
		"executes":      []any{},
		"verifications": []any{},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/snapshot", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var response struct {
		Snapshot runtime.SnapshotPacket `json:"snapshot"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode snapshot import response: %v", err)
	}
	if len(response.Snapshot.ReviewRoom.Open) != 1 || response.Snapshot.ReviewRoom.Open[0].Text != "Imported review room item." {
		t.Fatalf("expected imported review room to persist, got %+v", response.Snapshot.ReviewRoom)
	}
}

func TestFounderActionRoutes(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	router := NewRouter(service)

	startBody := map[string]any{
		"flow_id":      "flow_001",
		"work_item_id": "work_001",
		"approval_id":  "approval_001",
		"title":        "Founder loop",
		"intent":       "close founder loop",
		"notes":        []string{"start"},
	}
	raw, _ := json.Marshal(startBody)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/founder-actions/start", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected founder start 201, got %d", rec.Code)
	}

	approveBody := map[string]any{
		"flow_id": "flow_001",
		"notes":   []string{"approve"},
	}
	raw, _ = json.Marshal(approveBody)
	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/founder-actions/approve", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected founder approve 200, got %d", rec.Code)
	}

	releaseBody := map[string]any{
		"flow_id":    "flow_001",
		"release_id": "release_001",
		"deploy_id":  "deploy_001",
		"target":     "vm",
		"channel":    "cockpit",
		"notes":      []string{"release"},
	}
	raw, _ = json.Marshal(releaseBody)
	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/founder-actions/release", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected founder release 201, got %d", rec.Code)
	}

	rollbackBody := map[string]any{
		"flow_id":     "flow_001",
		"rollback_id": "rollback_001",
		"notes":       []string{"rollback"},
	}
	raw, _ = json.Marshal(rollbackBody)
	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/founder-actions/rollback", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected founder rollback 201, got %d", rec.Code)
	}

	events := service.ListEvents()
	var releaseEvent *runtime.EventEnvelope
	var rollbackEvent *runtime.EventEnvelope
	for i := range events {
		switch events[i].Kind {
		case "founder.action.released":
			releaseEvent = &events[i]
		case "founder.action.rolled_back":
			rollbackEvent = &events[i]
		}
	}
	if releaseEvent == nil || rollbackEvent == nil {
		t.Fatalf("expected founder action events, got %+v", events)
	}
	if rule, _ := releaseEvent.Data["rule"].(string); rule != "founder_action.release" {
		t.Fatalf("unexpected founder release rule: %+v", releaseEvent.Data)
	}
	switch evidence := releaseEvent.Data["evidence"].(type) {
	case []string:
		if len(evidence) == 0 {
			t.Fatalf("unexpected founder release evidence: %+v", releaseEvent.Data["evidence"])
		}
	case []any:
		if len(evidence) == 0 {
			t.Fatalf("unexpected founder release evidence: %+v", releaseEvent.Data["evidence"])
		}
	default:
		t.Fatalf("unexpected founder release evidence: %+v", releaseEvent.Data["evidence"])
	}
	if rule, _ := rollbackEvent.Data["rule"].(string); rule != "founder_action.rollback" {
		t.Fatalf("unexpected founder rollback rule: %+v", rollbackEvent.Data)
	}
	switch evidence := rollbackEvent.Data["evidence"].(type) {
	case []string:
		if len(evidence) == 0 {
			t.Fatalf("unexpected founder rollback evidence: %+v", rollbackEvent.Data["evidence"])
		}
	case []any:
		if len(evidence) == 0 {
			t.Fatalf("unexpected founder rollback evidence: %+v", rollbackEvent.Data["evidence"])
		}
	default:
		t.Fatalf("unexpected founder rollback evidence: %+v", rollbackEvent.Data["evidence"])
	}
}

func TestAuditRoutes(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/structure", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected structure 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/residue", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected residue 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/surface", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected surface 200, got %d", rec.Code)
	}
}

func TestEventsRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(runtime.WorkItem{
		ID:               "work_001",
		Title:            "seed event",
		Intent:           "test route",
		Stage:            runtime.StageDiscover,
		Surface:          runtime.SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/events", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFlowRoutes(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(runtime.WorkItem{
		ID:               "work_001",
		Title:            "seed flow",
		Intent:           "test route",
		Stage:            runtime.StageDiscover,
		Surface:          runtime.SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"flow_id":      "flow_001",
		"work_item_id": "work_001",
		"status":       "active",
		"notes":        []string{"test"},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/flows", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/layer-os/flows", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/auth", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthWriteRoutes(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{"token": "secret-token"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/auth", bytes.NewReader(raw))
	req.RemoteAddr = "127.0.0.1:41001"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/layer-os/auth", nil)
	req.RemoteAddr = "127.0.0.1:41001"
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthBootstrapRejectsNonLoopbackRequest(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{"token": "secret-token"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/auth", bytes.NewReader(raw))
	req.RemoteAddr = "203.0.113.8:41001"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCreatePreflightRecord(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"record_id":   "preflight_001",
		"task":        "internalize planner",
		"mode":        "internal",
		"status":      "ready",
		"decision":    "go",
		"models_used": []string{},
		"steps":       []string{"load contracts"},
		"risks":       []string{"no live planner"},
		"checks":      []string{"run tests"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/preflights", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreatePreflightRecordUsesRequestModels(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"record_id":   "preflight_model_001",
		"task":        "internalize planner",
		"mode":        "internal",
		"status":      "ready",
		"decision":    "go",
		"models_used": []string{},
		"steps":       []string{"load contracts"},
		"risks":       []string{"no live planner"},
		"checks":      []string{"run tests"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/preflights", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Models", "gpt-5.4,claude-sonnet-4.5")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	items := service.ListPreflights()
	if len(items) != 1 {
		t.Fatalf("expected 1 preflight, got %d", len(items))
	}
	want := []string{"gpt-5.4", "claude-sonnet-4.5"}
	if !reflect.DeepEqual(items[0].ModelsUsed, want) {
		t.Fatalf("expected models %v, got %v", want, items[0].ModelsUsed)
	}
}

func TestVerificationRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/verifications", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestEvaluatePolicyRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"decision_id":       "policy_001",
		"intent":            "route kernel verify",
		"scope":             "kernel",
		"risk":              "medium",
		"novelty":           "low",
		"token_class":       "small",
		"requires_approval": true,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/policies/evaluate", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestEvaluatePolicyRouteUsesSingleReviewLaneForHighRisk(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"decision_id":       "policy_high",
		"intent":            "route kernel verify",
		"scope":             "kernel",
		"risk":              "high",
		"novelty":           "medium",
		"token_class":       "large",
		"requires_approval": true,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/policies/evaluate", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var item runtime.PolicyDecision
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if item.Mode != "single" {
		t.Fatalf("expected single mode, got %q", item.Mode)
	}
}

func TestCreateGatewayCallRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.EvaluatePolicy("policy_001", "route kernel verify", "kernel", "medium", "low", "small", true); err != nil {
		t.Fatalf("seed policy: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"call_id":      "gateway_001",
		"decision_id":  "policy_001",
		"provider":     "openai",
		"model":        "gpt-5.4",
		"request_kind": "plan",
		"status":       "recorded",
		"token_budget": 8000,
		"notes":        []string{"planned"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/gateway-calls", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestExecuteRollbackRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           runtime.StageVerify,
		Summary:         "review release",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	releasedAt := time.Now().UTC()
	if err := service.CreateRelease(runtime.ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/true"}}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"rollback_id": "rollback_001",
		"release_id":  "release_001",
		"notes":       []string{"test"},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/rollbacks/execute", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestRunExecuteRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(runtime.WorkItem{
		ID:               "work_001",
		Title:            "run founder loop",
		Intent:           "execute local work",
		Stage:            runtime.StageDiscover,
		Surface:          runtime.SurfaceCockpit,
		Pack:             "founder",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed work: %v", err)
	}
	if _, err := service.EvaluatePolicy("policy_001", "execute local work", "kernel", "low", "low", "small", false); err != nil {
		t.Fatalf("seed policy: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"execute_id":         "execute_001",
		"work_item_id":       "work_001",
		"policy_decision_id": "policy_001",
		"notes":              []string{"smoke"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/execute-runs/run", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestRunVerificationRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	root := seedVerificationRoot(t)
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", root)
	router := NewRouter(service)

	body := map[string]any{
		"record_id": "verify_001",
		"scope":     "kernel",
		"notes":     []string{"smoke"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/verifications/run", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func seedVerificationRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module verifytest\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "verify.go"), []byte("package verifytest\n\nfunc Ready() bool { return true }\n"), 0o644); err != nil {
		t.Fatalf("write verify.go: %v", err)
	}
	return root
}

func TestEventsRouteRespectsRetentionWindow(t *testing.T) {
	oldRetention := os.Getenv("LAYER_OS_EVENT_RETENTION")
	defer os.Setenv("LAYER_OS_EVENT_RETENTION", oldRetention)
	os.Setenv("LAYER_OS_EVENT_RETENTION", "2")

	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	for i := 1; i <= 3; i++ {
		if err := service.CreateWorkItem(runtime.WorkItem{
			ID:               fmt.Sprintf("work_events_%03d", i),
			Title:            "Seed runtime",
			Intent:           "initialize runtime",
			Stage:            runtime.StageDiscover,
			Surface:          runtime.SurfaceAPI,
			Pack:             "runtime",
			Priority:         "high",
			Risk:             "medium",
			RequiresApproval: true,
			Payload:          map[string]any{"source": "test"},
			CorrelationID:    fmt.Sprintf("corr_events_%03d", i),
			CreatedAt:        time.Now().UTC(),
		}); err != nil {
			t.Fatalf("seed event %d: %v", i, err)
		}
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/events", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload struct {
		Items []runtime.EventEnvelope `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode events payload: %v", err)
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 recent events, got %d", len(payload.Items))
	}
}

func TestCreateWorkItem(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateWorkItemUsesRequestActor(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_actor_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_actor_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	events := service.ListEvents()
	if len(events) == 0 {
		t.Fatal("expected events to be recorded")
	}
	if events[len(events)-1].Actor != "gemini" {
		t.Fatalf("expected actor gemini, got %q", events[len(events)-1].Actor)
	}
}

func TestCreateWorkItemRequiresBearerWhenEnabled(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 with valid token, got %d", rec.Code)
	}
}

func TestCreateApprovalItem(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"approval_id":      "approval_001",
		"work_item_id":     "work_001",
		"stage":            "verify",
		"summary":          "approve release",
		"risks":            []string{"rollback required"},
		"rollback_plan":    "revert release packet",
		"decision_surface": "cockpit",
		"status":           "pending",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/approval-inbox", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestResolveApprovalItem(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           runtime.StageVerify,
		Summary:         "approve release",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release packet",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"approval_id": "approval_001",
		"status":      "approved",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/approval-inbox/resolve", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateReleasePacket(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           runtime.StageVerify,
		Summary:         "approve release",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release packet",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"release_id":    "release_001",
		"work_item_id":  "work_001",
		"target":        "vm",
		"channel":       "website",
		"artifacts":     []string{"build.tar.gz"},
		"metrics":       map[string]any{"latency_ms": 120},
		"rollback_plan": "restore previous artifact",
		"approval_refs": []string{"approval_001"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/releases", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateDeployRun(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           runtime.StageVerify,
		Summary:         "approve release",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release packet",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := service.CreateRelease(runtime.ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   ptrTime(time.Now().UTC()),
	}); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"deploy_id":  "deploy_001",
		"release_id": "release_001",
		"target":     "vm",
		"status":     "succeeded",
		"notes":      []string{"initial deploy"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/deploys", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestExecuteDeployRun(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(runtime.DeployTarget{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
	}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := service.CreateApproval(runtime.ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           runtime.StageVerify,
		Summary:         "approve release",
		Risks:           []string{"rollback required"},
		RollbackPlan:    "revert release packet",
		DecisionSurface: runtime.SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	if _, err := service.ResolveApproval("approval_001", "approved"); err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if err := service.CreateRelease(runtime.ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "website",
		Artifacts:    []string{"build.tar.gz"},
		Metrics:      map[string]any{"latency_ms": 120},
		RollbackPlan: "restore previous artifact",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   ptrTime(time.Now().UTC()),
	}); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"deploy_id":  "deploy_002",
		"release_id": "release_001",
		"notes":      []string{"adapter run"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/deploys/execute", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateSystemMemory(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"current_focus": "approval state machine",
		"current_goal":  "stabilize runtime",
		"next_steps":    []string{"add release flow"},
		"open_risks":    []string{"release path missing"},
		"handoff_note":  "next agent continues here",
		"last_operator": "codex",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/memory", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateDeployTargetRejectsRelativeCommand(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"target_id": "vm",
		"command":   []string{"usr/bin/true"},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/deploy-targets", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateDeployTargetCleansWorkdirWithinRepoRoot(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	repoRoot := t.TempDir()
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	router := NewRouter(service)

	rawWorkdir := filepath.Join(repoRoot, "deploy") + string(os.PathSeparator) + "." + string(os.PathSeparator) + "release"
	cleanWorkdir := filepath.Join(repoRoot, "deploy", "release")
	body := map[string]any{
		"target_id": "vm",
		"command":   []string{"/usr/bin/true"},
		"workdir":   rawWorkdir,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/deploy-targets", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var item runtime.DeployTarget
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatalf("decode deploy target response: %v", err)
	}
	if item.Workdir == nil {
		t.Fatal("expected workdir to be returned")
	}
	if *item.Workdir != cleanWorkdir {
		t.Fatalf("expected clean workdir %q, got %q", cleanWorkdir, *item.Workdir)
	}
}

func TestCreateDeployTargetRejectsTraversalWorkdir(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	repoRoot := t.TempDir()
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	router := NewRouter(service)

	workdir := filepath.Join(repoRoot, "deploy") + string(os.PathSeparator) + ".." + string(os.PathSeparator) + ".."
	body := map[string]any{
		"target_id": "vm",
		"command":   []string{"/usr/bin/true"},
		"workdir":   workdir,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/deploy-targets", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSnapshotRouteRejectsTraversalTargetWorkdir(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	repoRoot := t.TempDir()
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	os.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	router := NewRouter(service)

	snapshot := service.Snapshot()
	snapshot.Targets = []runtime.DeployTarget{{
		TargetID: "vm",
		Command:  []string{"/usr/bin/true"},
		Workdir:  ptrString(filepath.Join(repoRoot, "deploy") + string(os.PathSeparator) + ".." + string(os.PathSeparator) + ".."),
	}}
	raw, _ := json.Marshal(snapshot)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/snapshot", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ptrString(value string) *string {
	return &value
}

func TestCorpusRecoverRoute(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	pathMD := filepath.Join(repoRoot, "knowledge", "corpus", "entries", "analysis.md")
	if err := os.MkdirAll(filepath.Dir(pathMD), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(pathMD, []byte("# Corpus Note\n\nPreserve this insight.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/corpus/recover", strings.NewReader(`{"cleanup":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result runtime.CorpusMarkdownRecoverResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode corpus recover response: %v", err)
	}
	if result.Created != 1 || result.Cleaned == 0 {
		t.Fatalf("unexpected corpus recover response: %+v", result)
	}
	if _, err := os.Stat(pathMD); !os.IsNotExist(err) {
		t.Fatalf("expected markdown cleanup, err=%v", err)
	}
}

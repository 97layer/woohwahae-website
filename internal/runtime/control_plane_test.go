package runtime

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type invalidJSONPayload struct {
	Message string
}

func TestPreflightRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	item := PreflightRecord{
		RecordID:    "preflight_001",
		Task:        "internalize planning",
		Mode:        "internal",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"load contracts", "run tests"},
		Risks:       []string{"no live planner"},
		Checks:      []string{"re-run tests externally"},
		GeneratedAt: time.Now().UTC(),
	}
	if err := service.CreatePreflight(item); err != nil {
		t.Fatalf("create preflight: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListPreflights()); got != 1 {
		t.Fatalf("expected 1 preflight, got %d", got)
	}
	if reloaded.Handoff().Counts.Preflights != 1 {
		t.Fatalf("expected handoff preflight count 1, got %d", reloaded.Handoff().Counts.Preflights)
	}
}

func TestPreflightModelsUsedFromEnv(t *testing.T) {
	oldModels := os.Getenv("LAYER_OS_MODELS")
	defer os.Setenv("LAYER_OS_MODELS", oldModels)
	os.Setenv("LAYER_OS_MODELS", "gpt-5.4,claude-sonnet-4.5")

	service, err := NewService(t.TempDir())
	service.gatewayAdapter = apiGatewayAdapter{}
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	item := PreflightRecord{
		RecordID:    "preflight_env_001",
		Task:        "internalize planning",
		Mode:        "internal",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"load contracts"},
		Risks:       []string{"no live planner"},
		Checks:      []string{"run tests"},
		GeneratedAt: time.Now().UTC(),
	}
	if err := service.CreatePreflight(item); err != nil {
		t.Fatalf("create preflight: %v", err)
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

func TestPolicyRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	item, err := service.EvaluatePolicy("policy_001", "route kernel verify", "kernel", "medium", "low", "small", true)
	if err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if item.Mode != "single" {
		t.Fatalf("expected single policy mode, got %q", item.Mode)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListPolicies()); got != 1 {
		t.Fatalf("expected 1 policy, got %d", got)
	}
	if reloaded.Handoff().Counts.Policies != 1 {
		t.Fatalf("expected handoff policy count 1, got %d", reloaded.Handoff().Counts.Policies)
	}
}

func TestHighRiskPolicyUsesSingleReviewLane(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	item, err := service.EvaluatePolicy("policy_high", "route high risk verify", "kernel", "high", "medium", "large", true)
	if err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if item.Mode != "single" {
		t.Fatalf("expected high-risk policy to stay single, got %q", item.Mode)
	}
	if len(item.Reasons) == 0 {
		t.Fatal("expected high-risk policy reasons")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type blockingGatewayAdapter struct {
	started chan struct{}
	release chan struct{}
}

func (a blockingGatewayAdapter) Name() string          { return "blocking" }
func (a blockingGatewayAdapter) Semantics() string     { return "dispatch" }
func (a blockingGatewayAdapter) DispatchEnabled() bool { return true }
func (a blockingGatewayAdapter) RequiredMode() string  { return "single" }

func (a blockingGatewayAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call, err := recordGatewayAdapter{}.Prepare(call, decision)
	if err != nil {
		return call, err
	}
	call.Notes = appendUniqueString(call.Notes, "adapter:blocking")
	return call, nil
}

func (a blockingGatewayAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	select {
	case a.started <- struct{}{}:
	default:
	}
	<-a.release
	now := time.Now().UTC()
	call.Status = "sent"
	call.AttemptCount = 1
	call.DispatchedAt = &now
	call.Notes = appendUniqueString(call.Notes, "dispatch:blocking-test")
	return call, nil
}

func TestGatewayAPIAdapterDispatchesToConfiguredEndpoint(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "openai=https://provider.example/v1/respond")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	calls := 0
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = apiGatewayAdapter{httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
	})}}
	if _, err := service.EvaluatePolicy("policy_001", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if err := service.CreateGatewayCall(GatewayCall{CallID: "gateway_api_001", DecisionID: "policy_001", Provider: "openai", Model: "gpt-5.4", RequestKind: "verify", Status: "recorded", TokenBudget: 12000, Notes: []string{"planned"}, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create gateway call: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 provider dispatch call, got %d", calls)
	}
	items := service.ListGatewayCalls()
	if len(items) != 1 || items[0].Status != "sent" {
		t.Fatalf("expected sent gateway call, got %+v", items)
	}
}

func TestGatewayAPIAdapterRetriesTransientFailure(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "openai=https://provider.example/v1/respond")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_PROVIDER_RETRIES", "1")
	calls := 0
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = apiGatewayAdapter{httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(`{"error":"upstream"}`)), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
	})}}
	if _, err := service.EvaluatePolicy("policy_retry", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if err := service.CreateGatewayCall(GatewayCall{CallID: "gateway_api_retry", DecisionID: "policy_retry", Provider: "openai", Model: "gpt-5.4", RequestKind: "verify", Status: "recorded", TokenBudget: 12000, Notes: []string{"planned"}, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create gateway call: %v", err)
	}
	items := service.ListGatewayCalls()
	if len(items) != 1 || items[0].Status != "sent" || items[0].AttemptCount != 2 {
		t.Fatalf("expected retried sent gateway call, got %+v", items)
	}
	if items[0].LastHTTPStatus == nil || *items[0].LastHTTPStatus != http.StatusAccepted {
		t.Fatalf("expected final http status accepted, got %+v", items[0])
	}
}

func TestGatewayAPIAdapterMarksMissingEndpointFailed(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = apiGatewayAdapter{}
	if _, err := service.EvaluatePolicy("policy_001", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if err := service.CreateGatewayCall(GatewayCall{CallID: "gateway_api_002", DecisionID: "policy_001", Provider: "openai", Model: "gpt-5.4", RequestKind: "verify", Status: "recorded", TokenBudget: 12000, Notes: []string{"planned"}, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create gateway call: %v", err)
	}
	items := service.ListGatewayCalls()
	if len(items) != 1 || items[0].Status != "failed" {
		t.Fatalf("expected failed gateway call, got %+v", items)
	}
}

func TestGatewayDispatchReleasesServiceLockWhileProviderWaits(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	adapter := blockingGatewayAdapter{started: make(chan struct{}, 1), release: make(chan struct{}, 1)}
	if _, err := service.EvaluatePolicy("policy_blocking", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}

	createDone := make(chan error, 1)
	go func() {
		createDone <- service.createGatewayCallWithAdapter(GatewayCall{CallID: "gateway_blocking_001", DecisionID: "policy_blocking", Provider: "openai", Model: "gpt-5.4", RequestKind: "verify", Status: "recorded", TokenBudget: 12000, Notes: []string{"planned"}, CreatedAt: time.Now().UTC()}, adapter)
	}()

	select {
	case <-adapter.started:
	case <-time.After(time.Second):
		t.Fatal("expected blocking dispatch to start")
	}

	handoffDone := make(chan struct{}, 1)
	go func() {
		_ = service.Handoff()
		handoffDone <- struct{}{}
	}()

	select {
	case <-handoffDone:
	case <-time.After(2 * time.Second):
		adapter.release <- struct{}{}
		t.Fatal("service lock stayed blocked during gateway dispatch")
	}

	adapter.release <- struct{}{}
	select {
	case err := <-createDone:
		if err != nil {
			t.Fatalf("create gateway call: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("gateway create did not finish after releasing dispatch")
	}

	items := service.ListGatewayCalls()
	if len(items) != 1 || items[0].Status != "sent" {
		t.Fatalf("expected sent gateway call after release, got %+v", items)
	}
}

func TestGatewayCallRequiresPolicy(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	err = service.CreateGatewayCall(GatewayCall{
		CallID:      "gateway_001",
		DecisionID:  "missing_policy",
		Provider:    "openai",
		Model:       "gpt-5.4",
		RequestKind: "plan",
		Status:      "recorded",
		TokenBudget: 8000,
		Notes:       []string{"smoke"},
		CreatedAt:   time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected missing policy error, got nil")
	}
}

func TestGatewayRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := service.EvaluatePolicy("policy_001", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if err := service.CreateGatewayCall(GatewayCall{
		CallID:      "gateway_001",
		DecisionID:  "policy_001",
		Provider:    "openai",
		Model:       "gpt-5.4",
		RequestKind: "verify",
		Status:      "recorded",
		TokenBudget: 12000,
		Notes:       []string{"planned"},
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create gateway call: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListGatewayCalls()); got != 1 {
		t.Fatalf("expected 1 gateway call, got %d", got)
	}
	if reloaded.Handoff().Counts.GatewayCalls != 1 {
		t.Fatalf("expected handoff gateway count 1, got %d", reloaded.Handoff().Counts.GatewayCalls)
	}
}

func TestGatewayCallRejectsNonDispatchablePolicy(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.EvaluatePolicy("policy_local", "local only", "kernel", "low", "low", "small", false); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	err = service.CreateGatewayCall(GatewayCall{
		CallID:      "gateway_local",
		DecisionID:  "policy_local",
		Provider:    "openai",
		Model:       "gpt-5.4",
		RequestKind: "verify",
		Status:      "recorded",
		TokenBudget: 8000,
		Notes:       []string{"planned"},
		CreatedAt:   time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("expected non-dispatchable policy error, got nil")
	}
}

func TestWorkItemRejectsUnsupportedPayloadValue(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	err = service.CreateWorkItem(WorkItem{
		ID:               "work_invalid_payload",
		Title:            "reject invalid payload",
		Intent:           "harden payload validation",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: false,
		Payload:          map[string]any{"nested": map[string]any{"invalid": invalidJSONPayload{Message: "no structs"}}},
		CorrelationID:    "corr_invalid_payload",
		CreatedAt:        time.Now().UTC(),
	})
	if err == nil || !strings.Contains(err.Error(), "work item payload") {
		t.Fatalf("expected work item payload validation error, got %v", err)
	}
}

func TestSnapshotImportRejectsUnsupportedNestedPayload(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	packet := SnapshotPacket{
		GeneratedAt: time.Now().UTC(),
		CompanyState: CompanyState{
			ShellMode:      "founder",
			MemoryHealth:   "ready",
			DeployHealth:   "ready",
			PrimarySurface: SurfaceCockpit,
			ActiveSurfaces: []Surface{SurfaceCockpit, SurfaceTelegram, SurfaceAPI},
		},
		SystemMemory: SystemMemory{
			CurrentFocus: "snapshot import",
			NextSteps:    []string{},
			OpenRisks:    []string{},
			UpdatedAt:    time.Now().UTC(),
		},
		Auth: AuthStatus{WriteAuthEnabled: false},
		WorkItems: []WorkItem{{
			ID:               "work_001",
			Title:            "reject nested payload",
			Intent:           "validate snapshot payload",
			Stage:            StageDiscover,
			Surface:          SurfaceAPI,
			Pack:             "core",
			Priority:         "high",
			Risk:             "medium",
			RequiresApproval: false,
			Payload:          map[string]any{"invalid": invalidJSONPayload{Message: "no structs"}},
			CorrelationID:    "corr_001",
			CreatedAt:        time.Now().UTC(),
		}},
		Flows:         []FlowRun{},
		Approvals:     []ApprovalItem{},
		Releases:      []ReleasePacket{},
		Deploys:       []DeployRun{},
		Rollbacks:     []RollbackRun{},
		Targets:       []DeployTarget{},
		Events:        []EventEnvelope{},
		Preflights:    []PreflightRecord{},
		Policies:      []PolicyDecision{},
		GatewayCalls:  []GatewayCall{},
		Executes:      []ExecuteRun{},
		Verifications: []VerificationRecord{},
	}

	err = service.ImportSnapshot(packet)
	if err == nil || !strings.Contains(err.Error(), "work item payload") {
		t.Fatalf("expected snapshot payload validation error, got %v", err)
	}
}

func TestExecuteRoundTripLocalMode(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.CreateWorkItem(WorkItem{
		ID:               "work_001",
		Title:            "run founder loop",
		Intent:           "execute local work",
		Stage:            StageDiscover,
		Surface:          SurfaceCockpit,
		Pack:             "founder",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create work: %v", err)
	}

	if _, err := service.EvaluatePolicy("policy_001", "execute local work", "kernel", "low", "low", "small", false); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}

	item, err := service.RunExecute("execute_001", "work_001", "policy_001", []string{"smoke"})
	if err != nil {
		t.Fatalf("run execute: %v", err)
	}
	if item.Status != "succeeded" {
		t.Fatalf("expected succeeded execute, got %q", item.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListExecutes()); got != 1 {
		t.Fatalf("expected 1 execute, got %d", got)
	}
	if reloaded.Handoff().Counts.Executes != 1 {
		t.Fatalf("expected handoff execute count 1, got %d", reloaded.Handoff().Counts.Executes)
	}
}

func TestExecuteBlockedPolicyStillRecordsFailure(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.CreateWorkItem(WorkItem{
		ID:               "work_001",
		Title:            "blocked work",
		Intent:           "blocked",
		Stage:            StageDiscover,
		Surface:          SurfaceCockpit,
		Pack:             "founder",
		Priority:         "high",
		Risk:             "high",
		RequiresApproval: true,
		Payload:          map[string]any{},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create work: %v", err)
	}

	if _, err := service.EvaluatePolicy("policy_001", "blocked", "kernel", "high", "high", "tiny", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	item, err := service.RunExecute("execute_001", "work_001", "policy_001", []string{"smoke"})
	if err == nil {
		t.Fatal("expected blocked execute error, got nil")
	}
	if item.Status != "failed" {
		t.Fatalf("expected failed execute, got %q", item.Status)
	}
	room := service.ReviewRoom()
	expected := "실행 `execute_001`이 작업 `work_001`에서 정책 `policy_001` 때문에 막혔어. 재시도 전에 founder 검토나 범위 축소가 필요해."
	if len(room.Open) != 1 || room.Open[0].Text != expected {
		t.Fatalf("unexpected review room open items: %+v", room.Open)
	}
	if room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.execute_blocked" {
		t.Fatalf("unexpected execute rationale: %+v", room.Open[0].Rationale)
	}
	if len(room.Open[0].Evidence) != 3 || room.Open[0].Evidence[0] != "execute:execute_001" {
		t.Fatalf("unexpected execute evidence: %+v", room.Open[0].Evidence)
	}
}

func TestExecuteRequiresReferences(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.RunExecute("execute_001", "missing_work", "missing_policy", []string{"smoke"}); err == nil {
		t.Fatal("expected missing reference error, got nil")
	}
}

func TestVerificationRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	item, err := service.RunVerification("verify_001", "kernel", dataDir, []string{"/usr/bin/true"}, []string{"smoke"})
	if err != nil {
		t.Fatalf("run verification: %v", err)
	}
	if item.Status != "passed" {
		t.Fatalf("expected passed verification, got %q", item.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListVerifications()); got != 1 {
		t.Fatalf("expected 1 verification, got %d", got)
	}
	if reloaded.Handoff().Counts.Verifications != 1 {
		t.Fatalf("expected handoff verification count 1, got %d", reloaded.Handoff().Counts.Verifications)
	}
}

func TestVerificationFailureStillRecordsRun(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	item, err := service.RunVerification("verify_002", "kernel", dataDir, []string{"/usr/bin/false"}, []string{"smoke"})
	if err == nil {
		t.Fatal("expected verification failure, got nil")
	}
	if item.Status != "failed" {
		t.Fatalf("expected failed verification, got %q", item.Status)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListVerifications()); got != 1 {
		t.Fatalf("expected 1 recorded verification, got %d", got)
	}
	expected := "검증 `verify_002`이 범위 `kernel`에서 실패했어. 릴리스 전에 명령 증거를 먼저 확인해야 해."
	if len(reloaded.ReviewRoom().Open) != 1 || reloaded.ReviewRoom().Open[0].Text != expected {
		t.Fatalf("unexpected review room open items: %+v", reloaded.ReviewRoom().Open)
	}
	if reloaded.ReviewRoom().Open[0].Rationale == nil || reloaded.ReviewRoom().Open[0].Rationale.Rule != "review_room.auto.verification_failed" {
		t.Fatalf("unexpected verification rationale: %+v", reloaded.ReviewRoom().Open[0].Rationale)
	}
	if len(reloaded.ReviewRoom().Open[0].Evidence) != 2 || reloaded.ReviewRoom().Open[0].Evidence[0] != "verification:verify_002" {
		t.Fatalf("unexpected verification evidence: %+v", reloaded.ReviewRoom().Open[0].Evidence)
	}
}

func TestVerificationPassClearsPriorScopeFailure(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.RunVerification("verify_fail_001", "kernel", dataDir, []string{"/usr/bin/false"}, []string{"smoke"}); err == nil {
		t.Fatal("expected verification failure, got nil")
	}
	if _, err := service.RunVerification("verify_pass_001", "kernel", dataDir, []string{"/usr/bin/true"}, []string{"smoke"}); err != nil {
		t.Fatalf("run passing verification: %v", err)
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListVerifications()); got != 2 {
		t.Fatalf("expected 2 recorded verifications, got %d", got)
	}
	if got := reloaded.ReviewRoomSummary().OpenCount; got != 0 {
		t.Fatalf("expected passed verification to clear prior scope failure, got %d", got)
	}
}

func TestGatewayFailedDispatchOpensReviewRoomItem(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = apiGatewayAdapter{}
	if _, err := service.EvaluatePolicy("policy_gateway_fail", "route kernel verify", "kernel", "high", "medium", "large", true); err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if err := service.CreateGatewayCall(GatewayCall{CallID: "gateway_api_fail_room", DecisionID: "policy_gateway_fail", Provider: "openai", Model: "gpt-5.4", RequestKind: "verify", Status: "recorded", TokenBudget: 12000, Notes: []string{"planned"}, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create gateway call: %v", err)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.gateway_failed" {
		t.Fatalf("expected gateway failure review-room item, got %+v", room.Open)
	}
}

func TestDispatchAgentJobRoutesThroughGateway(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "openai=https://provider.example/v1/respond")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = apiGatewayAdapter{httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		text := strings.TrimSpace(string(body))
		if !strings.Contains(text, "\"job\"") || !strings.Contains(text, "\"dispatch_transport\":\"http_push\"") {
			t.Fatalf("expected pushed agent run packet, got %s", text)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
	})}}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_dispatch_001", Kind: "plan", Role: "planner", Summary: "Plan provider lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create job: %v", err)
	}
	result, err := service.DispatchAgentJob("job_dispatch_001")
	if err != nil {
		t.Fatalf("dispatch job: %v", err)
	}
	if result.Gateway.Status != "sent" {
		t.Fatalf("unexpected gateway status: %+v", result)
	}
	if result.Job.Status != "running" {
		t.Fatalf("expected api-dispatched job to remain running pending explicit report, got %+v", result)
	}
	if result.Job.Result == nil || result.Job.Result["gateway_call_id"] == nil {
		t.Fatalf("expected gateway result in job result: %+v", result.Job)
	}
}

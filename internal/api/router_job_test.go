package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestJobsRouteCreateAndList(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	payload := runtime.AgentJob{JobID: "job_001", Kind: "plan", Role: "planner", Summary: "Plan founder lane", Status: "queued", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/api/layer-os/jobs", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Items []runtime.AgentJob `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode jobs response: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].JobID != "job_001" {
		t.Fatalf("unexpected jobs response: %+v", response.Items)
	}
}

func TestJobsRouteFiltersOpenAndLimit(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	seed := []runtime.AgentJob{
		{JobID: "job_001", Kind: "plan", Role: "planner", Summary: "First open", Status: "queued", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: []string{}, CreatedAt: now, UpdatedAt: now},
		{JobID: "job_002", Kind: "plan", Role: "planner", Summary: "Closed", Status: "succeeded", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: []string{}, CreatedAt: now, UpdatedAt: now},
		{JobID: "job_003", Kind: "plan", Role: "planner", Summary: "Second open", Status: "running", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: []string{}, CreatedAt: now, UpdatedAt: now},
	}
	for _, item := range seed {
		if err := service.CreateAgentJob(item); err != nil {
			t.Fatalf("seed agent job %s: %v", item.JobID, err)
		}
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/jobs?status=open&limit=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Items []runtime.AgentJob `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode jobs response: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].JobID != "job_001" {
		t.Fatalf("unexpected filtered jobs response: %+v", response.Items)
	}
}

func TestJobUpdateRouteEscalatesFailure(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(runtime.AgentJob{JobID: "job_001", Kind: "verify", Role: "verifier", Summary: "Verify lane", Status: "queued", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageVerify, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("seed agent job: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"job_id":"job_001","status":"failed","notes":["exit=1"],"result":{"reason":"test failure"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/update", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 1 {
		t.Fatalf("expected review room escalation, got %d", got)
	}
}

func TestDispatchAgentJobRoute(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(runtime.AgentJob{JobID: "job_dispatch_001", Kind: "plan", Role: "planner", Summary: "Plan founder lane", Status: "queued", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("seed job: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"job_id":"job_dispatch_001"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/dispatch", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result runtime.AgentDispatchResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode dispatch response: %v", err)
	}
	if result.Job.JobID != "job_dispatch_001" || result.Gateway.CallID == "" {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
	if result.Job.Status != "running" || result.Gateway.Status != "recorded" {
		t.Fatalf("expected packet-ready running dispatch, got %+v", result)
	}
	if result.Job.Result == nil || result.Job.Result["dispatch_state"] != "packet_ready" || result.Job.Result["dispatch_transport"] != "job_packet" {
		t.Fatalf("expected packet fallback result, got %+v", result.Job.Result)
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 0 {
		t.Fatalf("expected no review room failure for packet fallback, got %d", got)
	}
}

func TestAgentDispatchProfilesRoute(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/jobs/profiles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Items []runtime.AgentDispatchProfile `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode profiles: %v", err)
	}
	if len(response.Items) != 4 || response.Items[0].Role == "" {
		t.Fatalf("unexpected profiles: %+v", response.Items)
	}
	if response.Items[0].DispatchReady {
		t.Fatalf("expected planner dispatch not ready without endpoint, got %+v", response.Items[0])
	}
	foundDesigner := false
	for _, item := range response.Items {
		if item.Role == "designer" {
			foundDesigner = true
			break
		}
	}
	if !foundDesigner {
		t.Fatalf("expected designer dispatch profile, got %+v", response.Items)
	}
}

func TestReportAgentJobRouteWritesCapitalization(t *testing.T) {
	repoRoot, err := apiTestRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(runtime.AgentJob{JobID: "job_report_001", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: runtime.SurfaceAPI, Stage: runtime.StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("seed job: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"job_id":"job_report_001","status":"succeeded","notes":["patched"],"result":{"summary":"Patched runtime lane and captured the changed file.","artifacts":["internal/runtime/service_job_report.go"],"provider":"openai","gateway_call_id":"gateway_001"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/report", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response runtime.AgentJobReportResult
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode report response: %v", err)
	}
	if response.FollowUp.Mode != "continue_loop" {
		t.Fatalf("expected continue_loop follow-up, got %+v", response.FollowUp)
	}
	entries := service.ListCapitalizationEntries()
	if len(entries) != 1 || entries[0].SourceKind != "agent_job.succeeded" {
		t.Fatalf("unexpected capitalization entries: %+v", entries)
	}
	items := service.ListObservations(runtime.ObservationQuery{SourceChannel: "agent_job"})
	if len(items) != 1 || items[0].Topic != "agent_job.succeeded" {
		t.Fatalf("expected auto job observation, got %+v", items)
	}
}

func TestAgentRunPacketRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_001", Title: "Hot storage", Intent: "retrieve corpus", Summary: "Wire retrieval", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	ref := "proposal_001"
	if err := service.CreateAgentJob(runtime.AgentJob{JobID: "job_packet_001", Kind: "plan", Role: "planner", Summary: "Plan retrieval lane", Status: "queued", Source: "proposal", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Ref: &ref, Notes: []string{}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create job: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/jobs/packet?job_id=job_packet_001", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var packet runtime.AgentRunPacket
	if err := json.Unmarshal(rec.Body.Bytes(), &packet); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Job.JobID != "job_packet_001" || packet.Proposal == nil || packet.Proposal.ProposalID != "proposal_001" {
		t.Fatalf("unexpected packet: %+v", packet)
	}
	if packet.Runtime.SourceOfTruth != "daemon_api" || packet.Runtime.ReportPath != "/api/layer-os/jobs/report" {
		t.Fatalf("unexpected runtime contract: %+v", packet.Runtime)
	}
}

func TestDispatchPacketReportRouteFlow(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_flow_001", Title: "Corpus retrieval", Intent: "retrieve corpus", Summary: "Wire baton path", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	ref := "proposal_flow_001"
	if err := service.CreateAgentJob(runtime.AgentJob{JobID: "job_flow_001", Kind: "plan", Role: "planner", Summary: "Plan baton lane", Status: "queued", Source: "proposal", Surface: runtime.SurfaceAPI, Stage: runtime.StageDiscover, Ref: &ref, Notes: []string{}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create job: %v", err)
	}
	router := NewRouter(service)

	dispatchReq := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/dispatch", bytes.NewReader([]byte(`{"job_id":"job_flow_001"}`)))
	dispatchRec := httptest.NewRecorder()
	router.ServeHTTP(dispatchRec, dispatchReq)
	if dispatchRec.Code != http.StatusOK {
		t.Fatalf("expected dispatch 200, got %d body=%s", dispatchRec.Code, dispatchRec.Body.String())
	}

	packetReq := httptest.NewRequest(http.MethodGet, "/api/layer-os/jobs/packet?job_id=job_flow_001", nil)
	packetRec := httptest.NewRecorder()
	router.ServeHTTP(packetRec, packetReq)
	if packetRec.Code != http.StatusOK {
		t.Fatalf("expected packet 200, got %d body=%s", packetRec.Code, packetRec.Body.String())
	}
	var packet runtime.AgentRunPacket
	if err := json.Unmarshal(packetRec.Body.Bytes(), &packet); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Job.JobID != "job_flow_001" || packet.Job.Status != "running" || packet.Proposal == nil || packet.Proposal.ProposalID != "proposal_flow_001" {
		t.Fatalf("unexpected packet after dispatch: %+v", packet)
	}

	reportRaw := []byte(`{"job_id":"job_flow_001","status":"succeeded","notes":["delivered"],"result":{"summary":"Delivered planner packet for the next lane.","artifacts":[],"dispatch_transport":"job_packet"}}`)
	reportReq := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/report", bytes.NewReader(reportRaw))
	reportRec := httptest.NewRecorder()
	router.ServeHTTP(reportRec, reportReq)
	if reportRec.Code != http.StatusOK {
		t.Fatalf("expected report 200, got %d body=%s", reportRec.Code, reportRec.Body.String())
	}
	entries := service.ListCapitalizationEntries()
	if len(entries) != 1 || entries[0].SourceKind != "agent_job.succeeded" {
		t.Fatalf("unexpected capitalization entries: %+v", entries)
	}
	items := service.ListObservations(runtime.ObservationQuery{SourceChannel: "agent_job"})
	if len(items) != 1 || items[0].Topic != "agent_job.succeeded" {
		t.Fatalf("expected auto job observation, got %+v", items)
	}
}

func TestPromoteAgentJobsRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(runtime.ProposalItem{ProposalID: "proposal_promote_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	why := "Why is queue drift still unresolved?"
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "review.auto", WhyUnresolved: &why}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)
	raw := []byte(`{"limit":2,"dispatch":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/jobs/promote", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result runtime.AgentJobPromotionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode promote response: %v", err)
	}
	if result.Created != 2 || len(result.Items) != 2 {
		t.Fatalf("unexpected promote result: %+v", result)
	}
	if got := len(service.ListAgentJobs()); got != 2 {
		t.Fatalf("expected 2 promoted jobs, got %d", got)
	}
}

func apiTestRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(wd, "..", "..")), nil
}

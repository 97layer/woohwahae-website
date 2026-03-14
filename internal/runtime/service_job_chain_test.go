package runtime

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type captureTelegramAdapter struct {
	packets []TelegramPacket
	routes  []string
	fail    error
}

func (a *captureTelegramAdapter) Name() string  { return "capture" }
func (a *captureTelegramAdapter) Enabled() bool { return true }
func (a *captureTelegramAdapter) Send(packet TelegramPacket) error {
	a.routes = append(a.routes, TelegramRouteFounder)
	a.packets = append(a.packets, packet)
	return a.fail
}
func (a *captureTelegramAdapter) SendRoute(routeID string, packet TelegramPacket) error {
	a.routes = append(a.routes, routeID)
	a.packets = append(a.packets, packet)
	return a.fail
}
func (a *captureTelegramAdapter) RouteEnabled(_ string) bool { return true }

func TestReportAgentJobCreatesAndDispatchesChainedJob(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "claude,gemini")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "planner=claude,designer=gemini")
	t.Setenv("LAYER_OS_AGENT_ROLE_MODELS", "planner=claude-opus-4-6,designer=gemini-2.0-flash")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "gemini=https://designer.example/dispatch")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	telegram := &captureTelegramAdapter{}
	service.telegramAdapter = telegram
	service.gatewayAdapter = apiGatewayAdapter{httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		text := strings.TrimSpace(string(body))
		if !strings.Contains(text, "\"job\"") || !strings.Contains(text, "\"role\":\"designer\"") || !strings.Contains(text, "\"dispatch_transport\":\"http_push\"") || !strings.Contains(text, "chain_parent_job_id") {
			t.Fatalf("expected pushed agent run packet body, got %s", text)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
	})}}
	now := time.Now().UTC()
	payload := map[string]any{
		"chain_rules": map[string]any{
			"rules": []any{map[string]any{
				"rule_id":         "designer_followup",
				"on_status":       "succeeded",
				"next_kind":       "review",
				"next_role":       "designer",
				"next_stage":      string(StageExperience),
				"surface":         string(SurfaceAPI),
				"summary":         "Review the brand-aligned surface for the proposal.",
				"auto_dispatch":   true,
				"notify_telegram": true,
				"payload":         map[string]any{"target_surface": "brand"},
				"notes":           []any{"chain:auto_designer_review"},
			}},
		},
	}
	ref := "proposal_chain_001"
	if err := service.CreateProposal(ProposalItem{ProposalID: ref, Title: "Brand surface", Intent: "review surface", Summary: "Seed chain", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_chain_parent_001", Kind: "plan", Role: "planner", Summary: "Plan the brand surface lane.", Status: "running", Source: "proposal", Surface: SurfaceAPI, Stage: StageDiscover, Ref: &ref, Payload: payload, Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create parent job: %v", err)
	}

	report, err := service.ReportAgentJob("job_chain_parent_001", "succeeded", []string{"done"}, map[string]any{"summary": "Planner completed the surface plan.", "artifacts": []string{"plan.md"}})
	if err != nil {
		t.Fatalf("report agent job: %v", err)
	}
	if report.Job.Status != "succeeded" {
		t.Fatalf("expected parent succeeded, got %+v", report.Job)
	}
	if report.Chain == nil || report.Chain.ParentJobID != "job_chain_parent_001" {
		t.Fatalf("expected chain result on report, got %+v", report.Chain)
	}
	if report.FollowUp.Mode != "monitor_dispatched_jobs" {
		t.Fatalf("expected dispatched follow-up mode, got %+v", report.FollowUp)
	}
	if report.Chain.Considered != 1 || report.Chain.Created != 1 || report.Chain.Dispatched != 1 || report.Chain.TelegramRequested != 1 {
		t.Fatalf("unexpected chain result in report: %+v", report.Chain)
	}
	if len(report.FollowUp.JobIDs) != 1 {
		t.Fatalf("expected chained child job id in follow-up, got %+v", report.FollowUp)
	}
	storedFollowUp, ok := report.Job.Result["follow_up"].(map[string]any)
	if !ok || strings.TrimSpace(agentJobResultString(storedFollowUp["mode"])) != "monitor_dispatched_jobs" {
		t.Fatalf("expected persisted follow_up on chained job result, got %+v", report.Job.Result)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("expected clean report warnings, got %+v", report.Warnings)
	}
	jobs := service.ListAgentJobs()
	if len(jobs) != 2 {
		t.Fatalf("expected parent + child jobs, got %+v", jobs)
	}
	var child AgentJob
	found := false
	for _, item := range jobs {
		if item.JobID == "job_chain_parent_001" {
			continue
		}
		child = item
		found = true
	}
	if !found {
		t.Fatalf("expected chained job in %+v", jobs)
	}
	if child.Role != "designer" || child.Source != "agent_job.chain" {
		t.Fatalf("unexpected chained child job: %+v", child)
	}
	// auto-close: direct LLM adapter auto-reports job as succeeded when response received
	if child.Status != "succeeded" && child.Status != "running" {
		t.Fatalf("unexpected chained child status: %+v", child)
	}
	if child.Result == nil || child.Result["gateway_call_id"] == nil {
		t.Fatalf("expected gateway result in chained child, got %+v", child.Result)
	}
	if len(telegram.packets) < 1 || !strings.Contains(telegram.packets[0].Headline, "designer") {
		t.Fatalf("expected dispatch notice first, got %+v", telegram.packets)
	}
}

func TestReportAgentJobInvalidChainRulesOpenReviewItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateAgentJob(AgentJob{JobID: "job_chain_invalid_001", Kind: "plan", Role: "planner", Summary: "Plan the brand surface lane.", Status: "running", Source: "proposal", Surface: SurfaceAPI, Stage: StageDiscover, Payload: map[string]any{"chain_rules": map[string]any{"rules": []any{map[string]any{"rule_id": "broken"}}}}, Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create parent job: %v", err)
	}
	report, err := service.ReportAgentJob("job_chain_invalid_001", "succeeded", []string{"done"}, map[string]any{"summary": "Planner completed."})
	if err != nil {
		t.Fatalf("report agent job: %v", err)
	}
	if report.Chain != nil {
		t.Fatalf("expected no chain result when rule parsing fails, got %+v", report.Chain)
	}
	if report.FollowUp.Mode != "review_warnings" {
		t.Fatalf("expected warning follow-up mode, got %+v", report.FollowUp)
	}
	if len(report.Warnings) == 0 || !strings.Contains(report.Warnings[0], "agent job chain failed") {
		t.Fatalf("expected chain warning on report, got %+v", report.Warnings)
	}
	room := service.ReviewRoom()
	if len(room.Open) == 0 || room.Open[0].Rationale == nil || room.Open[0].Rationale.Rule != "review_room.auto.agent_job_chain_failed" {
		t.Fatalf("expected chain failure review-room item, got %+v", room.Open)
	}
}

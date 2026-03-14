package runtime

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateAgentJobPersistsAndShowsInFounderWaiting(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_001", Kind: "plan", Role: "planner", Summary: "Plan orchestration lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageDiscover, Notes: []string{"seed"}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListAgentJobs()); got != 1 {
		t.Fatalf("expected 1 agent job, got %d", got)
	}
	view := reloaded.FounderView()
	if len(view.Waiting) == 0 || view.Waiting[0].Kind != "agent_job" {
		t.Fatalf("expected agent job in founder waiting, got %+v", view.Waiting)
	}
	summary := reloaded.FounderSummary()
	if summary.PrimaryAction != "dispatch_job" {
		t.Fatalf("expected dispatch_job primary action, got %+v", summary)
	}
	if got := reloaded.Handoff().Counts.AgentJobs; got != 1 {
		t.Fatalf("expected agent job count 1, got %d", got)
	}
}

func TestUpdateAgentJobPromotesRiskAndResolvesReviewRoom(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_001", Kind: "verify", Role: "verifier", Summary: "Verify release lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageVerify, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	item, err := service.UpdateAgentJob("job_001", "running", []string{"started"}, nil)
	if err != nil {
		t.Fatalf("update running: %v", err)
	}
	if item.Status != "running" || len(service.FounderView().Now) == 0 || service.FounderView().Now[0].Kind != "agent_job" {
		t.Fatalf("expected running job in now lane, got item=%+v view=%+v", item, service.FounderView())
	}
	item, err = service.UpdateAgentJob("job_001", "failed", []string{"exit=1"}, map[string]any{"reason": "test failure"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if item.Status != "failed" || len(service.FounderView().Risk) == 0 || service.FounderView().Risk[0].Kind != "agent_job" {
		t.Fatalf("expected failed job in risk lane, got item=%+v view=%+v", item, service.FounderView())
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 1 {
		t.Fatalf("expected review room escalation, got %d", got)
	}
	if _, err := service.UpdateAgentJob("job_001", "succeeded", []string{"recovered"}, map[string]any{"reason": "retry passed"}); err != nil {
		t.Fatalf("update succeeded: %v", err)
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 0 {
		t.Fatalf("expected review room resolved after success, got %d", got)
	}
}

func TestRetryingAgentJobClearsFailedReviewRoomItem(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_retry_001", Kind: "plan", Role: "planner", Summary: "Retry stale lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.UpdateAgentJob("job_retry_001", "failed", []string{"dispatch_failed"}, map[string]any{"reason": "temporary"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 1 {
		t.Fatalf("expected failed job to open review-room item, got %d", got)
	}
	if _, err := service.UpdateAgentJob("job_retry_001", "running", []string{"dispatching"}, map[string]any{"reason": "retry"}); err != nil {
		t.Fatalf("update running: %v", err)
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 0 {
		t.Fatalf("expected retry to clear stale failed review item, got %d", got)
	}
}

func TestAgentJobSnapshotRoundTrip(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_001", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "review_room", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	packet := service.Snapshot()
	if got := len(packet.AgentJobs); got != 1 {
		t.Fatalf("expected snapshot agent jobs 1, got %d", got)
	}
	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	if got := len(target.ListAgentJobs()); got != 1 {
		t.Fatalf("expected imported agent jobs 1, got %d", got)
	}
}

func TestReportAgentJobCreatesCapitalizationEntry(t *testing.T) {
	repoRoot, err := runtimeTestRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_report_001", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	result, err := service.ReportAgentJob("job_report_001", "succeeded", []string{"patched"}, map[string]any{
		"summary":         "Patched runtime lane and captured the changed file.",
		"artifacts":       []string{"internal/runtime/service_job_report.go"},
		"provider":        "openai",
		"gateway_call_id": "gateway_001",
	})
	if err != nil {
		t.Fatalf("report agent job: %v", err)
	}
	if result.Job.Status != "succeeded" || result.Event.Kind != "agent_job.succeeded" {
		t.Fatalf("unexpected report result: %+v", result)
	}
	if result.FollowUp.Mode != "continue_loop" {
		t.Fatalf("expected continue_loop follow-up, got %+v", result.FollowUp)
	}
	storedFollowUp, ok := result.Job.Result["follow_up"].(map[string]any)
	if !ok || strings.TrimSpace(agentJobResultString(storedFollowUp["mode"])) != "continue_loop" {
		t.Fatalf("expected persisted follow_up on job result, got %+v", result.Job.Result)
	}
	if result.Capitalization == nil || result.Capitalization.SourceKind != "agent_job.succeeded" {
		t.Fatalf("unexpected capitalization result: %+v", result.Capitalization)
	}
	entries := service.ListCapitalizationEntries()
	if len(entries) != 1 || entries[0].EntryID != result.Capitalization.EntryID {
		t.Fatalf("unexpected capitalization entries: %+v", entries)
	}
	if result.Job.Result["reported_by_actor"] != "system" || result.Job.Result["report_transport"] != "job_report" {
		t.Fatalf("expected report provenance on job result, got %+v", result.Job.Result)
	}
	event := latestAgentJobEvent(t, service, "agent_job.succeeded")
	if event.Data["report_actor"] != "system" || event.Data["completion_mode"] != "explicit_job_report" {
		t.Fatalf("expected report provenance on event, got %+v", event.Data)
	}
}

func TestReportAgentJobCapturesAntigravityClassification(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateAgentJob(AgentJob{JobID: "job_report_antigravity_001", Kind: "verify", Role: "verifier", Summary: "Verify ingest lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageVerify, Notes: []string{}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	result, err := service.ReportAgentJob("job_report_antigravity_001", "succeeded", []string{"reported"}, map[string]any{
		"summary":        "Recovered antigravity artifact set for verification lane.",
		"artifacts":      []string{"task.md"},
		"report_source":  "antigravity",
		"source_channel": "antigravity",
		"source_family":  "agent_workspace",
		"source_run_id":  "run-001",
		"artifact_topic": "diagnosis",
		"artifact_type":  "ARTIFACT_TYPE_TASK",
		"artifact_stem":  "task.md",
		"severity":       "high",
	})
	if err != nil {
		t.Fatalf("report agent job: %v", err)
	}
	if result.Capitalization == nil {
		t.Fatalf("expected capitalization result, got %+v", result)
	}
	joined := strings.Join(result.Capitalization.Result.Items, " | ")
	for _, want := range []string{"report_source:antigravity", "source_channel:antigravity", "source_family:agent_workspace", "artifact_topic:diagnosis", "severity:high"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected capitalization result items to include %q, got %+v", want, result.Capitalization.Result.Items)
		}
	}
}

func TestAgentRunPacketIncludesLinkedProposal(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(ProposalItem{ProposalID: "proposal_001", Title: "Hot storage", Intent: "retrieve corpus", Summary: "Wire retrieval", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if _, err := service.CheckpointSession(SessionCheckpointInput{
		Source:       "terminal",
		CurrentFocus: "Plan retrieval lane",
		NextSteps:    []string{"dispatch planner"},
		OpenRisks:    []string{"retrieval scope unclear"},
		Refs:         []string{"proposal_001"},
	}); err != nil {
		t.Fatalf("checkpoint session: %v", err)
	}
	if _, err := service.SessionNote(SessionNoteInput{
		Source: "terminal",
		Kind:   continuityNoteKindTodo,
		Text:   "carry proposal context into the packet",
		Refs:   []string{"proposal_001"},
	}); err != nil {
		t.Fatalf("session note: %v", err)
	}
	ref := "proposal_001"
	if err := service.CreateAgentJob(AgentJob{JobID: "job_packet_001", Kind: "plan", Role: "planner", Summary: "Plan retrieval lane", Status: "queued", Source: "proposal", Surface: SurfaceAPI, Stage: StageDiscover, Ref: &ref, Notes: []string{}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	packet, err := service.AgentRunPacket("job_packet_001")
	if err != nil {
		t.Fatalf("agent run packet: %v", err)
	}
	if packet.Job.JobID != "job_packet_001" || packet.Proposal == nil || packet.Proposal.ProposalID != "proposal_001" {
		t.Fatalf("unexpected packet: %+v", packet)
	}
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate packet: %v", err)
	}
	if packet.Runtime.SourceOfTruth != "daemon_api" || packet.Runtime.DispatchTransport != "job_packet" {
		t.Fatalf("unexpected runtime contract: %+v", packet.Runtime)
	}
	if packet.Runtime.ReportPath != "/api/layer-os/jobs/report" || len(packet.Runtime.TerminalStatuses) != 3 {
		t.Fatalf("unexpected runtime report contract: %+v", packet.Runtime)
	}
	if packet.Prompting == nil || packet.Prompting.DecisionScope != "bounded" || packet.Prompting.AutonomyBudget != "multi_step" {
		t.Fatalf("expected prompting defaults in packet, got %+v", packet.Prompting)
	}
	if packet.Handoff == nil || packet.HandoffSummary != nil {
		t.Fatalf("expected planner packet to carry full handoff only, got handoff=%+v summary=%+v", packet.Handoff, packet.HandoffSummary)
	}
	if packet.Handoff.Continuity == nil || packet.Handoff.Continuity.Record == nil {
		t.Fatalf("expected handoff continuity in packet, got %+v", packet.Handoff)
	}
	if packet.Handoff.Continuity.Record.CurrentFocus != "Plan retrieval lane" {
		t.Fatalf("expected continuity focus in packet handoff, got %+v", packet.Handoff.Continuity.Record)
	}
	if len(packet.Handoff.Continuity.Suggestions) == 0 {
		t.Fatalf("expected continuity suggestions in packet handoff, got %+v", packet.Handoff.Continuity)
	}
}

func TestAgentRunPacketImplementerUsesHandoffSummary(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_impl_packet_001",
		Kind:      "implement",
		Role:      "implementer",
		Summary:   "Patch the worker loop",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   SurfaceAPI,
		Stage:     StageCompose,
		Payload:   map[string]any{"allowed_paths": []string{"cmd/layer-osctl/"}},
		Notes:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create implementer job: %v", err)
	}
	packet, err := service.AgentRunPacket("job_impl_packet_001")
	if err != nil {
		t.Fatalf("agent run packet: %v", err)
	}
	if packet.Handoff != nil || packet.HandoffSummary == nil {
		t.Fatalf("expected implementer packet summary only, got handoff=%+v summary=%+v", packet.Handoff, packet.HandoffSummary)
	}
	if packet.Prompting == nil || packet.Prompting.MutationPolicy != "scoped_write" {
		t.Fatalf("expected implementer scoped_write prompting, got %+v", packet.Prompting)
	}
	if packet.HandoffSummary.CurrentFocus != packet.Knowledge.CurrentFocus {
		t.Fatalf("expected summary focus to follow knowledge, got %+v", packet.HandoffSummary)
	}
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate implementer packet: %v", err)
	}
}

func TestDefaultJobMutationPolicyFollowsStageAndAllowedPaths(t *testing.T) {
	cases := []struct {
		name string
		job  AgentJob
		want string
	}{
		{
			name: "discover stage stays read only",
			job:  AgentJob{Role: "planner", Stage: StageDiscover},
			want: "read_only",
		},
		{
			name: "compose stage gets scoped write without role fallback",
			job:  AgentJob{Role: "verifier", Stage: StageCompose},
			want: "scoped_write",
		},
		{
			name: "experience stage gets scoped write",
			job:  AgentJob{Role: "planner", Stage: StageExperience},
			want: "scoped_write",
		},
		{
			name: "allowed paths force scoped write",
			job:  AgentJob{Role: "planner", Stage: StageVerify, Payload: map[string]any{"allowed_paths": []string{"docs/"}}},
			want: "scoped_write",
		},
		{
			name: "verify stage remains read only without explicit scope",
			job:  AgentJob{Role: "implementer", Stage: StageVerify},
			want: "read_only",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := defaultJobMutationPolicy(tc.job); got != tc.want {
				t.Fatalf("defaultJobMutationPolicy(%+v) = %q, want %q", tc.job, got, tc.want)
			}
		})
	}
}

func TestSessionBootstrapIncludesPrompting(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	packet := service.SessionBootstrap("daemon", false, false)
	if packet.Prompting == nil {
		t.Fatalf("expected bootstrap prompting, got %+v", packet)
	}
	if packet.Prompting.DecisionScope != "full" || packet.Prompting.MutationPolicy != "read_only" || packet.Prompting.AutonomyBudget != "single_step" {
		t.Fatalf("unexpected bootstrap prompting defaults: %+v", packet.Prompting)
	}
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate bootstrap packet: %v", err)
	}
}

func TestDispatchAgentJobFallsBackToPacketWhenProviderEndpointMissing(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_dispatch_001", Kind: "plan", Role: "planner", Summary: "Plan founder lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	result, err := service.DispatchAgentJob("job_dispatch_001")
	if err != nil {
		t.Fatalf("dispatch agent job: %v", err)
	}
	if result.Job.Status != "running" {
		t.Fatalf("expected running job, got %+v", result.Job)
	}
	if result.Gateway.Status != "recorded" {
		t.Fatalf("expected recorded gateway call, got %+v", result.Gateway)
	}
	if result.Job.Result == nil || result.Job.Result["dispatch_state"] != "packet_ready" || result.Job.Result["dispatch_transport"] != "job_packet" {
		t.Fatalf("expected packet-ready dispatch result, got %+v", result.Job.Result)
	}
	if result.Job.Result["dispatch_reason"] != "no_provider_endpoint" {
		t.Fatalf("expected no_provider_endpoint fallback reason, got %+v", result.Job.Result)
	}
	if got := service.ReviewRoomSummary().OpenCount; got != 0 {
		t.Fatalf("expected no review room drift for packet fallback, got %d", got)
	}
}

func TestDispatchAgentJobDirectLLMLeavesImplementerRunning(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Okay, I will inspect the failing lane and report back."},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("GOOGLE_API_KEY", "test_key")
	t.Setenv("LAYER_OS_PROVIDERS", "gemini")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = geminiDirectAdapter{httpClient: client}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_direct_impl_001", Kind: "implement", Role: "implementer", Summary: "Fix runtime lane", Status: "queued", Source: "architect.verification_failed", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_direct_impl_001")
	if err != nil {
		t.Fatalf("dispatch job: %v", err)
	}
	if result.Gateway.Status != "sent" {
		t.Fatalf("expected sent gateway, got %+v", result.Gateway)
	}
	if result.Job.Status != "running" {
		t.Fatalf("expected implementer to remain running pending explicit report, got %+v", result.Job)
	}
	if agentJobHasNote(result.Job, "auto_closed") {
		t.Fatalf("did not expect implementer auto_close note, got %+v", result.Job.Notes)
	}
	if result.Job.Result == nil || result.Job.Result["response_preview"] == nil {
		t.Fatalf("expected response preview to remain attached to running job, got %+v", result.Job.Result)
	}
}

func TestDispatchAgentJobDirectLLMAutoClosesDesigner(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Brand surface looks aligned with the current direction."},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("GOOGLE_API_KEY", "test_key")
	t.Setenv("LAYER_OS_PROVIDERS", "gemini")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = geminiDirectAdapter{httpClient: client}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_direct_design_001", Kind: "review", Role: "designer", Summary: "Review brand lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageExperience, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_direct_design_001")
	if err != nil {
		t.Fatalf("dispatch job: %v", err)
	}
	if result.Job.Status != "succeeded" {
		t.Fatalf("expected designer direct LLM lane to auto-close, got %+v", result.Job)
	}
	if !agentJobHasNote(result.Job, "auto_closed") {
		t.Fatalf("expected designer auto_close note, got %+v", result.Job.Notes)
	}
	if result.Job.Result == nil || result.Job.Result["response"] == nil {
		t.Fatalf("expected auto-closed response payload, got %+v", result.Job.Result)
	}
	if result.Job.Result["completion_mode"] != "direct_llm_auto_close" {
		t.Fatalf("expected direct_llm_auto_close provenance, got %+v", result.Job.Result)
	}
	artifacts, ok := result.Job.Result["artifacts"].([]any)
	if !ok || len(artifacts) != 1 {
		t.Fatalf("expected gateway call artifact ref on auto-close, got %+v", result.Job.Result)
	}
	event := latestAgentJobEvent(t, service, "agent_job.succeeded")
	if event.Data["dispatch_actor"] != "system" || event.Data["execution_origin"] != "direct_llm" || event.Data["completion_mode"] != "direct_llm_auto_close" {
		t.Fatalf("expected dispatch/report provenance on event, got %+v", event.Data)
	}
}

func TestDispatchAgentJobUsesImplicitProviderAndProviderDefaultModel(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "Planner lane acknowledged."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "claude")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_MODELS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = claudeDirectAdapter{httpClient: client}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_implicit_provider_001", Kind: "plan", Role: "planner", Summary: "Plan the founder lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageDiscover, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_implicit_provider_001")
	if err != nil {
		t.Fatalf("dispatch agent job: %v", err)
	}
	if result.Gateway.Provider != "claude" {
		t.Fatalf("expected implicit claude provider, got %+v", result.Gateway)
	}
	if result.Gateway.Model != "claude-opus-4-6" {
		t.Fatalf("expected provider default claude model, got %+v", result.Gateway)
	}
	if result.Job.Result == nil || result.Job.Result["provider"] != "claude" || result.Job.Result["model"] != "claude-opus-4-6" {
		t.Fatalf("expected runtime result to record implicit provider/model, got %+v", result.Job.Result)
	}
}

func TestDispatchAgentJobCouncilCallsMultipleProvidersSequentially(t *testing.T) {
	claudeCalls := 0
	openAICalls := 0
	claudeClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claudeCalls++
		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "Planner primary lane acknowledged from Claude."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	openAIClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAICalls++
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Planner secondary lane acknowledged from OpenAI."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("OPENAI_API_KEY", "test_key")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_MODELS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = multiProviderAdapter{
		providers: map[string]GatewayAdapter{
			"claude": claudeDirectAdapter{httpClient: claudeClient},
			"openai": openaiDirectAdapter{httpClient: openAIClient},
		},
		fallback: recordGatewayAdapter{},
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:   "job_council_001",
		Kind:    "plan",
		Role:    "planner",
		Summary: "Plan the next bounded founder action",
		Status:  "queued",
		Source:  "founder.manual",
		Surface: SurfaceAPI,
		Stage:   StageDiscover,
		Payload: map[string]any{
			"council": map[string]any{
				"providers":        []string{"claude", "openai"},
				"primary_provider": "claude",
			},
		},
		Notes:     []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create council job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_council_001")
	if err != nil {
		t.Fatalf("dispatch council job: %v", err)
	}
	if result.Gateway.Provider != "claude" || result.Gateway.Status != "sent" {
		t.Fatalf("expected claude primary gateway to be sent, got %+v", result.Gateway)
	}
	if result.Job.Status != "succeeded" {
		t.Fatalf("expected planner council lane to auto-close, got %+v", result.Job)
	}
	if claudeCalls != 1 || openAICalls != 1 {
		t.Fatalf("expected one sequential call per provider, got claude=%d openai=%d", claudeCalls, openAICalls)
	}
	calls := service.ListGatewayCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 gateway calls, got %+v", calls)
	}
	for _, call := range calls {
		if call.Status != "sent" {
			t.Fatalf("expected sent council gateway calls, got %+v", calls)
		}
	}
	if result.Job.Result["execution_origin"] != "direct_llm_council" {
		t.Fatalf("expected direct_llm_council execution origin, got %+v", result.Job.Result)
	}
	council, ok := result.Job.Result["council"].(map[string]any)
	if !ok {
		t.Fatalf("expected council result payload, got %+v", result.Job.Result)
	}
	if council["requested_count"] != float64(2) || council["succeeded_count"] != float64(2) {
		t.Fatalf("expected council counts in result, got %+v", council)
	}
	if council["primary_provider"] != "claude" || council["primary_status"] != "sent" {
		t.Fatalf("expected claude primary selection in council result, got %+v", council)
	}
	if result.Job.Result["provider"] != "claude" || result.Job.Result["completion_mode"] != "direct_llm_auto_close" {
		t.Fatalf("expected auto-closed primary provider result, got %+v", result.Job.Result)
	}
}

func TestDispatchAgentJobAutoCouncilForDerivedPlannerLane(t *testing.T) {
	claudeCalls := 0
	openAICalls := 0
	claudeClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claudeCalls++
		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "Planner primary lane acknowledged from Claude."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	openAIClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAICalls++
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Planner supporting lane acknowledged from OpenAI."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("OPENAI_API_KEY", "test_key")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "planner=claude,designer=openai")
	t.Setenv("LAYER_OS_AGENT_ROLE_MODELS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = multiProviderAdapter{
		providers: map[string]GatewayAdapter{
			"claude": claudeDirectAdapter{httpClient: claudeClient},
			"openai": openaiDirectAdapter{httpClient: openAIClient},
		},
		fallback: recordGatewayAdapter{},
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_auto_council_001",
		Kind:      "plan",
		Role:      "planner",
		Summary:   "Review the next bounded founder action",
		Status:    "queued",
		Source:    "review_room.open",
		Surface:   SurfaceAPI,
		Stage:     StageDiscover,
		Notes:     []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create auto council job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_auto_council_001")
	if err != nil {
		t.Fatalf("dispatch auto council job: %v", err)
	}
	if result.Gateway.Provider != "claude" || result.Gateway.Status != "sent" {
		t.Fatalf("expected claude primary gateway to be sent, got %+v", result.Gateway)
	}
	if claudeCalls != 1 || openAICalls != 1 {
		t.Fatalf("expected one auto council call per provider, got claude=%d openai=%d", claudeCalls, openAICalls)
	}
	if !agentJobHasNote(result.Job, "dispatch_council_auto") {
		t.Fatalf("expected auto council dispatch note, got %+v", result.Job.Notes)
	}
	if result.Job.Result["execution_origin"] != "direct_llm_council" || result.Job.Result["council_auto"] != true {
		t.Fatalf("expected auto council provenance, got %+v", result.Job.Result)
	}
	council, ok := result.Job.Result["council"].(map[string]any)
	if !ok {
		t.Fatalf("expected council result payload, got %+v", result.Job.Result)
	}
	if council["requested_count"] != float64(2) || council["primary_provider"] != "claude" {
		t.Fatalf("expected claude-led auto council payload, got %+v", council)
	}
}

func TestDispatchAgentJobDoesNotAutoCouncilFounderManualPlanner(t *testing.T) {
	claudeCalls := 0
	openAICalls := 0
	claudeClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claudeCalls++
		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "Planner primary lane acknowledged from Claude."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	openAIClient := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAICalls++
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Planner supporting lane acknowledged from OpenAI."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("OPENAI_API_KEY", "test_key")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "planner=claude,designer=openai")
	t.Setenv("LAYER_OS_AGENT_ROLE_MODELS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.gatewayAdapter = multiProviderAdapter{
		providers: map[string]GatewayAdapter{
			"claude": claudeDirectAdapter{httpClient: claudeClient},
			"openai": openaiDirectAdapter{httpClient: openAIClient},
		},
		fallback: recordGatewayAdapter{},
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_manual_planner_001",
		Kind:      "plan",
		Role:      "planner",
		Summary:   "Plan the next founder lane",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   SurfaceAPI,
		Stage:     StageDiscover,
		Notes:     []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create founder manual planner job: %v", err)
	}

	result, err := service.DispatchAgentJob("job_manual_planner_001")
	if err != nil {
		t.Fatalf("dispatch founder manual planner job: %v", err)
	}
	if claudeCalls != 1 || openAICalls != 0 {
		t.Fatalf("expected single-provider dispatch for founder.manual planner, got claude=%d openai=%d", claudeCalls, openAICalls)
	}
	if agentJobHasNote(result.Job, "dispatch_council_auto") {
		t.Fatalf("did not expect auto council note, got %+v", result.Job.Notes)
	}
	if _, ok := result.Job.Result["council"]; ok {
		t.Fatalf("did not expect council payload on founder.manual planner, got %+v", result.Job.Result)
	}
}

func TestReportAgentJobRejectsSucceededImplementerWithoutArtifacts(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_report_guard_001", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.ReportAgentJob("job_report_guard_001", "succeeded", []string{"reported"}, map[string]any{"summary": "Implemented the change."}); err == nil {
		t.Fatal("expected succeeded implementer report without artifacts to fail")
	}
}

func TestReportAgentJobRejectsSucceededImplementerOutsideAllowedPaths(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_report_guard_002", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Payload: map[string]any{"allowed_paths": []string{"internal/runtime/"}}, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.ReportAgentJob("job_report_guard_002", "succeeded", []string{"reported"}, map[string]any{
		"summary":       "Patched the wrong file.",
		"changed_paths": []string{"cmd/layer-osctl/main.go"},
		"artifacts":     []string{"cmd/layer-osctl/main.go"},
	}); err == nil {
		t.Fatal("expected succeeded implementer report outside allowed_paths to fail")
	}
}

func TestReportAgentJobRejectsMissingChangedPathEvidence(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{JobID: "job_report_guard_003", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Payload: map[string]any{"allowed_paths": []string{"internal/runtime/"}}, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.ReportAgentJob("job_report_guard_003", "succeeded", []string{"reported"}, map[string]any{
		"summary":       "Claimed a missing file changed.",
		"changed_paths": []string{"internal/runtime/does_not_exist.go"},
		"artifacts":     []string{"internal/runtime/does_not_exist.go"},
	}); err == nil {
		t.Fatal("expected succeeded implementer report with missing changed path to fail")
	}
}

func TestReportAgentJobAcceptsAbsoluteArtifactPathInsideAllowedRepoScope(t *testing.T) {
	repoRoot, err := runtimeTestRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateAgentJob(AgentJob{
		JobID:     "job_report_guard_004",
		Kind:      "implement",
		Role:      "implementer",
		Summary:   "Patch runtime lane",
		Status:    "queued",
		Source:    "founder.manual",
		Surface:   SurfaceAPI,
		Stage:     StageCompose,
		Payload:   map[string]any{"allowed_paths": []string{"internal/runtime/"}},
		Notes:     []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	artifactPath := filepath.Join(repoRoot, "internal", "runtime", "service_job_report.go")
	result, err := service.ReportAgentJob("job_report_guard_004", "succeeded", []string{"reported"}, map[string]any{
		"summary":   "Recorded the changed runtime file using an absolute path.",
		"artifacts": []string{artifactPath},
	})
	if err != nil {
		t.Fatalf("expected absolute repo path to pass validation, got %v", err)
	}
	if result.Job.Status != "succeeded" {
		t.Fatalf("expected succeeded job, got %+v", result.Job)
	}
}

func TestAgentDispatchProfilesRequireProviderEndpointForReady(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	profiles := service.AgentDispatchProfiles()
	if len(profiles) != 4 {
		t.Fatalf("expected 4 dispatch profiles, got %+v", profiles)
	}
	if profiles[0].DispatchReady {
		t.Fatalf("expected planner dispatch not ready without endpoint, got %+v", profiles[0])
	}
	found := false
	for _, note := range profiles[0].Notes {
		if note == "provider endpoint missing; dispatch falls back to job packet" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected packet fallback note, got %+v", profiles[0].Notes)
	}
	foundDesigner := false
	for _, profile := range profiles {
		if profile.Role != "designer" {
			continue
		}
		foundDesigner = true
		if profile.TokenBudget != 12000 || profile.TokenClass != "medium" || profile.Novelty != "medium" || profile.Risk != "medium" {
			t.Fatalf("unexpected designer dispatch profile: %+v", profile)
		}
	}
	if !foundDesigner {
		t.Fatalf("expected designer dispatch profile in %+v", profiles)
	}
}

func agentJobHasNote(job AgentJob, want string) bool {
	for _, note := range job.Notes {
		if strings.TrimSpace(note) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

func TestPromoteContextJobsCreatesIdempotentPlannerJobs(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(ProposalItem{ProposalID: "proposal_auto_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	why := "Why is queue drift still unresolved?"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "review.auto", WhyUnresolved: &why}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	result, err := service.PromoteContextJobs(0, false)
	if err != nil {
		t.Fatalf("promote context jobs: %v", err)
	}
	if result.Created != 2 || result.Existing != 0 || result.Dispatched != 0 {
		t.Fatalf("unexpected promotion result: %+v", result)
	}
	if got := len(service.ListAgentJobs()); got != 2 {
		t.Fatalf("expected 2 promoted jobs, got %d", got)
	}
	var sawParallel bool
	var sawThread bool
	for _, item := range result.Items {
		switch item.SourceKind {
		case "parallel_candidate":
			sawParallel = true
			if item.Status != "created" || item.Job == nil || item.Job.Kind != "plan" || item.Job.Role != "planner" {
				t.Fatalf("unexpected parallel promotion item: %+v", item)
			}
			if item.Job.Ref == nil || *item.Job.Ref != "proposal_auto_001" {
				t.Fatalf("expected proposal ref on parallel promotion, got %+v", item.Job)
			}
		case "open_thread":
			sawThread = true
			if item.Status != "created" || item.Job == nil || item.Job.Kind != "plan" || item.Job.Role != "planner" {
				t.Fatalf("unexpected open thread promotion item: %+v", item)
			}
			if item.Job.Payload == nil || item.Job.Payload["thread_id"] == nil {
				t.Fatalf("expected thread payload on open-thread promotion, got %+v", item.Job)
			}
		default:
			t.Fatalf("unexpected promotion source kind: %+v", item)
		}
	}
	if !sawParallel || !sawThread {
		t.Fatalf("expected both parallel and open-thread promotions, got %+v", result.Items)
	}

	second, err := service.PromoteContextJobs(0, false)
	if err != nil {
		t.Fatalf("re-promote context jobs: %v", err)
	}
	if second.Created != 0 || second.Existing != 2 {
		t.Fatalf("expected idempotent reuse on second promotion, got %+v", second)
	}
	if got := len(service.ListAgentJobs()); got != 2 {
		t.Fatalf("expected 2 jobs after idempotent rerun, got %d", got)
	}
}

func TestPromoteContextJobsDispatchesWhenRequested(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "api")
	t.Setenv("LAYER_OS_PROVIDERS", "openai")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateProposal(ProposalItem{ProposalID: "proposal_dispatch_001", Title: "Queue drift plan", Intent: "stabilize queue drift", Summary: "Plan queue drift lane", Surface: SurfaceAPI, Priority: "high", Risk: "medium", Status: "proposed", Notes: []string{"seed"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Queue drift", NextSteps: []string{"Triage queue drift"}, OpenRisks: []string{"Queue drift unresolved"}, UpdatedAt: now}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	why := "Why is queue drift still unresolved?"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "review.auto", WhyUnresolved: &why}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	result, err := service.PromoteContextJobs(1, true)
	if err != nil {
		t.Fatalf("promote+dispatch context jobs: %v", err)
	}
	if result.Created != 1 || result.Dispatched != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected dispatched promotion result: %+v", result)
	}
	item := result.Items[0]
	if item.Status != "dispatched" || item.Job == nil || item.Dispatch == nil {
		t.Fatalf("expected dispatched item with job+dispatch payload, got %+v", item)
	}
	if item.Job.Status != "running" {
		t.Fatalf("expected running job after dispatch, got %+v", item.Job)
	}
	if item.Dispatch.Gateway.Status != "recorded" || item.Dispatch.Job.Result == nil || item.Dispatch.Job.Result["dispatch_state"] != "packet_ready" {
		t.Fatalf("expected packet-ready dispatch fallback, got %+v", item.Dispatch)
	}
	if got := len(service.ListAgentJobs()); got != 1 {
		t.Fatalf("expected 1 promoted job after dispatch, got %d", got)
	}
}

func TestPromoteContextJobsSkipsOpenThreadAlreadyCoveredByProposalCandidate(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "intel_007"
	why := "We keep seeing the same signal without a managed proposal."
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{
		Text:          "Queue drift still unresolved.",
		Kind:          "agenda",
		Severity:      "high",
		Source:        "review.auto",
		Ref:           &ref,
		WhyUnresolved: &why,
	}); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "terminal",
		Topic:             "queue drift",
		Refs:              []string{ref},
		RawExcerpt:        "Terminal saw queue drift again.",
		NormalizedSummary: "Queue drift is still blocking the migration map.",
	}); err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "telegram",
		Topic:             "queue drift",
		Refs:              []string{ref},
		RawExcerpt:        "Telegram raised queue drift again.",
		NormalizedSummary: "Queue drift is still blocking the migration map.",
	}); err != nil {
		t.Fatalf("create second observation: %v", err)
	}

	packet := service.Knowledge()
	if len(packet.ProposalCandidates) == 0 || len(packet.OpenThreads) == 0 {
		t.Fatalf("expected proposal candidate and open thread, got %+v", packet)
	}
	coveredThreads := map[string]struct{}{}
	for _, threadID := range packet.ProposalCandidates[0].ThreadIDs {
		if strings.TrimSpace(threadID) == "" {
			continue
		}
		coveredThreads[strings.TrimSpace(threadID)] = struct{}{}
	}
	if len(coveredThreads) == 0 {
		t.Fatalf("expected proposal candidate to carry thread ids, got %+v", packet.ProposalCandidates[0])
	}

	result, err := service.PromoteContextJobs(0, false)
	if err != nil {
		t.Fatalf("promote context jobs: %v", err)
	}
	if result.Created != 0 || result.Existing != 0 {
		t.Fatalf("expected proposal-linked thread to stay out of created/existing context promotion, got %+v", result)
	}
	for _, item := range result.Items {
		if _, covered := coveredThreads[strings.TrimSpace(item.SourceID)]; covered {
			t.Fatalf("expected covered thread to be filtered out of promotion result, got %+v", item)
		}
		if item.Job == nil || item.Job.Payload == nil {
			continue
		}
		threadID, _ := item.Job.Payload["thread_id"].(string)
		if _, covered := coveredThreads[strings.TrimSpace(threadID)]; covered {
			t.Fatalf("expected covered thread to avoid promoted jobs, got %+v", item.Job)
		}
	}
	if got := len(service.ListAgentJobs()); got != 0 {
		t.Fatalf("expected no promoted jobs when proposal candidate already owns the thread, got %d", got)
	}
}

func latestAgentJobEvent(t *testing.T, service *Service, kind string) EventEnvelope {
	t.Helper()
	events := service.ListEvents()
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Kind == kind {
			return events[i]
		}
	}
	t.Fatalf("expected event kind %s", kind)
	return EventEnvelope{}
}

func runtimeTestRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	curr := wd
	for {
		if _, err := os.Stat(filepath.Join(curr, "AGENTS.md")); err == nil {
			return curr, nil
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return filepath.Clean(filepath.Join(wd, "..", "..")), nil
}

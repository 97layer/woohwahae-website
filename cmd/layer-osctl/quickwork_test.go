package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type quickWorkServiceStub struct {
	auth             runtime.AuthStatus
	adapters         runtime.AdapterSummary
	providers        []runtime.ProviderSummary
	bootstrap        runtime.SessionBootstrapPacket
	jobs             []runtime.AgentJob
	dispatched       []string
	finished         []runtime.SessionFinishInput
	promoteCalls     int
	promoteLimit     int
	promoteDispatch  bool
	promotionResults []runtime.AgentJobPromotionResult
}

func (s *quickWorkServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *quickWorkServiceStub) Adapters() runtime.AdapterSummary {
	return s.adapters
}
func (s *quickWorkServiceStub) Providers() []runtime.ProviderSummary {
	return append([]runtime.ProviderSummary{}, s.providers...)
}

func (s *quickWorkServiceStub) SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error) {
	return s.bootstrap, nil
}

func (s *quickWorkServiceStub) SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error) {
	s.finished = append(s.finished, input)
	return runtime.SessionFinishResult{Event: runtime.EventEnvelope{Kind: "session.finished"}}, nil
}

func (s *quickWorkServiceStub) CreateAgentJob(item runtime.AgentJob) error {
	s.jobs = append(s.jobs, item)
	return nil
}

func (s *quickWorkServiceStub) PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error) {
	s.promoteCalls++
	s.promoteLimit = limit
	s.promoteDispatch = dispatch
	if len(s.promotionResults) == 0 {
		return runtime.AgentJobPromotionResult{GeneratedAt: time.Now().UTC(), Items: []runtime.AgentJobPromotionItem{}}, nil
	}
	result := s.promotionResults[0]
	s.promotionResults = s.promotionResults[1:]
	for _, item := range result.Items {
		if item.Job == nil {
			continue
		}
		upserted := false
		for i := range s.jobs {
			if s.jobs[i].JobID == item.Job.JobID {
				s.jobs[i] = *item.Job
				upserted = true
				break
			}
		}
		if !upserted {
			s.jobs = append(s.jobs, *item.Job)
		}
	}
	return result, nil
}

func (s *quickWorkServiceStub) DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error) {
	s.dispatched = append(s.dispatched, jobID)
	for i := range s.jobs {
		if s.jobs[i].JobID == jobID {
			s.jobs[i].Status = "succeeded"
			s.jobs[i].Result = map[string]any{"dispatch_state": "packet_ready"}
			return runtime.AgentDispatchResult{Job: s.jobs[i]}, nil
		}
	}
	return runtime.AgentDispatchResult{Job: runtime.AgentJob{JobID: jobID, Status: "succeeded"}}, nil
}

func (s *quickWorkServiceStub) ListAgentJobs() []runtime.AgentJob {
	return append([]runtime.AgentJob{}, s.jobs...)
}

func TestRunQuickWorkCreatesAndDispatchesSingleJob(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{}
	var calls []string
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			calls = append(calls, path+" "+strings.Join(args, " "))
			return nil
		},
	}

	runQuickWork(service, []string{"--summary", "Stabilize backend worker lane", "--role", "implementer", "--allowed-paths", "cmd/layer-osctl/,scripts/", "--council", "claude,openai", "--council-primary", "claude"})

	if len(calls) != 1 || !strings.Contains(calls[0], "scripts/worker_orchestrator.sh up") {
		t.Fatalf("expected single up call, got %v", calls)
	}
	if len(service.jobs) != 1 || service.jobs[0].Summary != "Stabilize backend worker lane" || service.jobs[0].Role != "implementer" {
		t.Fatalf("unexpected created jobs: %+v", service.jobs)
	}
	if len(service.dispatched) != 1 || service.dispatched[0] != service.jobs[0].JobID {
		t.Fatalf("expected dispatch for created job, got %+v", service.dispatched)
	}
	allowed, ok := service.jobs[0].Payload["allowed_paths"].([]string)
	if !ok || len(allowed) != 2 || allowed[0] != "cmd/layer-osctl/" {
		t.Fatalf("expected allowed paths payload, got %#v", service.jobs[0].Payload)
	}
	council, ok := service.jobs[0].Payload["council"].(map[string]any)
	if !ok || council["primary_provider"] != "claude" {
		t.Fatalf("expected council payload, got %#v", service.jobs[0].Payload)
	}
}

func TestRunQuickWorkInteractiveDefaultsImplementerAllowedPaths(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{}
	var stderr bytes.Buffer
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader("백엔드 로직 개선\n\n\n"),
		stdout: io.Discard,
		stderr: &stderr,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, nil)

	if len(service.jobs) != 1 {
		t.Fatalf("expected one created job, got %+v", service.jobs)
	}
	allowed, ok := service.jobs[0].Payload["allowed_paths"].([]string)
	if !ok || len(allowed) != 3 {
		t.Fatalf("expected default implementer allowed paths, got %#v", service.jobs[0].Payload)
	}
	if !strings.Contains(stderr.String(), "Layer OS quick work") {
		t.Fatalf("expected interactive prompt heading, got %q", stderr.String())
	}
}

func TestRunQuickWorkUsesBootstrapRouteWhenNoSummary(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				ActionRoutes: []runtime.ActionRoute{{
					RouteID:    "route_001",
					Kind:       "job",
					Summary:    "Review the brand surface",
					TargetLane: "designer",
					Source:     "knowledge.route",
				}},
			},
		},
	}
	var calls []string
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			calls = append(calls, strings.Join(args, " "))
			return nil
		},
	}

	runQuickWork(service, nil)

	if len(service.jobs) != 1 || service.jobs[0].Role != "designer" || service.jobs[0].Summary != "Review the brand surface" {
		t.Fatalf("expected bootstrap-derived designer job, got %+v", service.jobs)
	}
	if len(calls) != 1 || calls[0] != "up --roles implementer,verifier,designer" {
		t.Fatalf("expected designer-aware up call, got %v", calls)
	}
}

func TestRunQuickWorkUsesBootstrapReviewAgendaWhenNoSummary(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				PrimaryAction:   "review_room",
				ReviewOpenCount: 1,
				ReviewTopOpen:   []string{"Queue drift still needs founder-facing rationale."},
			},
			ReviewRoom: &runtime.ReviewRoomSummary{
				Source:    "review_room.json",
				OpenCount: 1,
				TopOpen: []runtime.ReviewRoomItem{{
					Text:     "Queue drift still needs founder-facing rationale.",
					Kind:     "agenda",
					Severity: "high",
					Source:   "review.auto",
				}},
				TopAccepted: []runtime.ReviewRoomItem{},
				TopDeferred: []runtime.ReviewRoomItem{},
				Issues:      []string{},
			},
		},
	}
	var calls []string
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			calls = append(calls, strings.Join(args, " "))
			return nil
		},
	}

	runQuickWork(service, nil)

	if len(service.jobs) != 1 || service.jobs[0].Role != "planner" {
		t.Fatalf("expected planner job for top review agenda, got %+v", service.jobs)
	}
	if service.jobs[0].Summary != "Review agenda: Queue drift still needs founder-facing rationale." {
		t.Fatalf("expected review agenda summary, got %+v", service.jobs[0])
	}
	if len(calls) != 1 || calls[0] != "up --roles implementer,verifier,planner" {
		t.Fatalf("expected planner-aware up call, got %v", calls)
	}
}

func TestRunQuickWorkPromotesContextJobBeforePromptFallback(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		promotionResults: []runtime.AgentJobPromotionResult{{
			GeneratedAt: time.Now().UTC(),
			Created:     1,
			Dispatched:  1,
			Items: []runtime.AgentJobPromotionItem{{
				SourceKind: "open_thread",
				SourceID:   "thread_001",
				Summary:    "Investigate repeated founder drift before widening execution.",
				Status:     "dispatched",
				Job: &runtime.AgentJob{
					JobID:     "job_promoted_001",
					Kind:      "plan",
					Role:      "planner",
					Summary:   "Investigate repeated founder drift before widening execution.",
					Status:    "running",
					Source:    "knowledge.open_thread",
					Surface:   runtime.SurfaceAPI,
					Stage:     runtime.StageDiscover,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
			}},
		}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: &stdout,
		stderr: &stderr,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, nil)

	if service.promoteCalls != 1 || service.promoteLimit != 1 || !service.promoteDispatch {
		t.Fatalf("expected context promotion before prompt fallback, got calls=%d limit=%d dispatch=%v", service.promoteCalls, service.promoteLimit, service.promoteDispatch)
	}
	if strings.Contains(stderr.String(), "Layer OS quick work") {
		t.Fatalf("did not expect interactive prompt fallback, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "promoted[open_thread]: Investigate repeated founder drift before widening execution.") {
		t.Fatalf("expected promotion announcement, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "already running job_promoted_001 (planner)") {
		t.Fatalf("expected promoted planner lane to be reused, got %q", stdout.String())
	}
}

func TestRunQuickWorkAutoCouncilForDerivedPlannerLane(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("OPENAI_API_KEY", "test_key")
	service := &quickWorkServiceStub{
		adapters: runtime.AdapterSummary{
			Gateway:                "multi",
			GatewaySemantics:       "direct_llm",
			GatewayDispatchEnabled: true,
			GatewayRequiredMode:    "single",
		},
		providers: []runtime.ProviderSummary{
			{Provider: "claude", DispatchEnabled: true, GatewayAdapter: "multi", Semantics: "direct_llm"},
			{Provider: "openai", DispatchEnabled: true, GatewayAdapter: "multi", Semantics: "direct_llm"},
			{Provider: "gemini", DispatchEnabled: true, GatewayAdapter: "multi", Semantics: "direct_llm"},
		},
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				ReviewOpenCount: 1,
				ReviewTopOpen:   []string{"Provider drift still needs triage."},
			},
			ReviewRoom: &runtime.ReviewRoomSummary{
				Source:    "review_room.json",
				OpenCount: 1,
				TopOpen: []runtime.ReviewRoomItem{{
					Text:     "Provider drift still needs triage.",
					Kind:     "agenda",
					Severity: "high",
					Source:   "review.auto",
				}},
				TopAccepted: []runtime.ReviewRoomItem{},
				TopDeferred: []runtime.ReviewRoomItem{},
				Issues:      []string{},
			},
		},
	}
	var stdout bytes.Buffer
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: &stdout,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, nil)

	council, ok := service.jobs[0].Payload["council"].(map[string]any)
	if !ok {
		t.Fatalf("expected auto council payload, got %#v", service.jobs[0].Payload)
	}
	providers, ok := council["providers"].([]string)
	if !ok || len(providers) != 2 || providers[0] != "claude" || providers[1] != "openai" {
		t.Fatalf("unexpected auto council providers: %#v", council["providers"])
	}
	if council["primary_provider"] != "claude" {
		t.Fatalf("unexpected auto council primary: %#v", council)
	}
	if !strings.Contains(stdout.String(), "council auto[claude,openai] primary=claude") {
		t.Fatalf("expected council announcement, got %q", stdout.String())
	}
}

func TestRunQuickWorkDoesNotAutoCouncilFounderManualSummary(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	t.Setenv("ANTHROPIC_API_KEY", "test_key")
	t.Setenv("OPENAI_API_KEY", "test_key")
	service := &quickWorkServiceStub{
		adapters: runtime.AdapterSummary{
			Gateway:                "multi",
			GatewaySemantics:       "direct_llm",
			GatewayDispatchEnabled: true,
			GatewayRequiredMode:    "single",
		},
		providers: []runtime.ProviderSummary{
			{Provider: "claude", DispatchEnabled: true, GatewayAdapter: "multi", Semantics: "direct_llm"},
			{Provider: "openai", DispatchEnabled: true, GatewayAdapter: "multi", Semantics: "direct_llm"},
		},
	}
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, []string{"--summary", "Plan the next founder lane", "--role", "planner"})

	if _, ok := service.jobs[0].Payload["council"]; ok {
		t.Fatalf("expected founder.manual summary to stay single-provider by default, got %#v", service.jobs[0].Payload)
	}
}

func TestQuickWorkDirectProviderReadySupportsGeminiAPIKeyAlias(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "alias_key")
	if !quickWorkDirectProviderReady("gemini") {
		t.Fatal("expected quickwork gemini readiness to accept GEMINI_API_KEY")
	}
}

func TestRunQuickWorkLoopExitsAfterIdle(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{}
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, []string{"--summary", "Plan the next lane", "--role", "planner", "--loop", "--poll", "1ms", "--exit-on-idle", "1ms"})

	if len(service.jobs) != 1 || service.jobs[0].Role != "planner" {
		t.Fatalf("expected one planner job during loop, got %+v", service.jobs)
	}
}

func TestRunQuickWorkLoopPromotesContextJobBeforeIdleExit(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				OpenThreads: []runtime.OpenThread{{
					ThreadID: "thread_001",
					Question: "Investigate repeated founder drift before widening execution.",
					Status:   "open",
					Source:   "knowledge.open_thread",
				}},
			},
		},
		promotionResults: []runtime.AgentJobPromotionResult{
			{
				GeneratedAt: time.Now().UTC(),
				Created:     1,
				Dispatched:  1,
				Items: []runtime.AgentJobPromotionItem{{
					SourceKind: "open_thread",
					SourceID:   "thread_001",
					Summary:    "Investigate repeated founder drift before widening execution.",
					Status:     "dispatched",
					Job: &runtime.AgentJob{
						JobID:     "job_promoted_loop_001",
						Kind:      "plan",
						Role:      "planner",
						Summary:   "Investigate repeated founder drift before widening execution.",
						Status:    "succeeded",
						Source:    "knowledge.open_thread",
						Surface:   runtime.SurfaceAPI,
						Stage:     runtime.StageDiscover,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				}},
			},
			{
				GeneratedAt: time.Now().UTC(),
				Items:       []runtime.AgentJobPromotionItem{},
			},
		},
	}
	var stdout bytes.Buffer
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: &stdout,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, []string{"--loop", "--poll", "1ms", "--exit-on-idle", "1ms"})

	if service.promoteCalls == 0 {
		t.Fatal("expected loop mode to promote context before idling")
	}
	if !strings.Contains(stdout.String(), "promoted[open_thread]: Investigate repeated founder drift before widening execution.") {
		t.Fatalf("expected promotion announcement in loop mode, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "already completed job_promoted_loop_001 (succeeded)") {
		t.Fatalf("expected promoted loop job completion, got %q", stdout.String())
	}
	if len(service.finished) != 1 {
		t.Fatalf("expected session finish after loop idle exit, got %+v", service.finished)
	}
}

func TestQuickWorkTargetFromPromotionSkipsExistingCompletedJob(t *testing.T) {
	target, message, ok := quickWorkTargetFromPromotion(runtime.AgentJobPromotionResult{
		GeneratedAt: time.Now().UTC(),
		Items: []runtime.AgentJobPromotionItem{{
			SourceKind: "open_thread",
			SourceID:   "thread_001",
			Summary:    "Completed planner lane should not block idle exit.",
			Status:     "existing",
			Job: &runtime.AgentJob{
				JobID:   "job_done_001",
				Role:    "planner",
				Summary: "Completed planner lane should not block idle exit.",
				Status:  "succeeded",
			},
		}},
	})

	if ok {
		t.Fatalf("expected completed promoted job to be ignored, got target=%+v message=%q", target, message)
	}
}

func TestRunQuickWorkReusesQueuedOpenJob(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		jobs: []runtime.AgentJob{{
			JobID:     "job_existing_001",
			Kind:      "implement",
			Role:      "implementer",
			Summary:   "Finish the existing backend lane",
			Status:    "queued",
			Source:    "founder.manual",
			Surface:   runtime.SurfaceAPI,
			Stage:     runtime.StageCompose,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}},
	}
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, nil)

	if len(service.jobs) != 1 {
		t.Fatalf("expected existing job to be reused, got %+v", service.jobs)
	}
	if len(service.dispatched) != 1 || service.dispatched[0] != "job_existing_001" {
		t.Fatalf("expected existing queued job to be dispatched, got %+v", service.dispatched)
	}
}

func TestRunQuickWorkLoopFinishesSessionOnIdleExit(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				CurrentFocus: "Founder loop",
				NextSteps:    []string{"Wait for the next safe action"},
				OpenRisks:    []string{},
			},
		},
	}
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, []string{"--loop", "--poll", "1ms", "--exit-on-idle", "1ms"})

	if len(service.finished) != 1 {
		t.Fatalf("expected session finish on idle exit, got %+v", service.finished)
	}
	if service.finished[0].CurrentFocus != "Founder loop" {
		t.Fatalf("expected quickwork idle finish to reuse current focus, got %+v", service.finished[0])
	}
}

func TestRunQuickWorkLoopAnnouncesAgenda(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{
		bootstrap: runtime.SessionBootstrapPacket{
			Source: "daemon",
			Knowledge: runtime.KnowledgePacket{
				PrimaryAction:   "review_room",
				ReviewOpenCount: 1,
				ReviewTopOpen:   []string{"Dispatch/provider drift still needs review."},
			},
			ReviewRoom: &runtime.ReviewRoomSummary{
				Source:    "review_room.json",
				OpenCount: 1,
				TopOpen: []runtime.ReviewRoomItem{{
					Text:     "Dispatch/provider drift still needs review.",
					Kind:     "agenda",
					Severity: "high",
					Source:   "review.auto",
				}},
				TopAccepted: []runtime.ReviewRoomItem{},
				TopDeferred: []runtime.ReviewRoomItem{},
				Issues:      []string{},
			},
		},
	}
	var stdout bytes.Buffer
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: &stdout,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			return nil
		},
	}

	runQuickWork(service, []string{"--loop", "--poll", "1ms", "--exit-on-idle", "1ms"})

	if !strings.Contains(stdout.String(), "agenda[1]: Dispatch/provider drift still needs review.") {
		t.Fatalf("expected agenda announcement, got %q", stdout.String())
	}
}

func TestRunQuickWorkStatusRoutesToOrchestrator(t *testing.T) {
	original := quickWorkRunner
	defer func() { quickWorkRunner = original }()

	service := &quickWorkServiceStub{}
	var calls []string
	quickWorkRunner = quickWorkCommandRunner{
		stdin:  strings.NewReader(""),
		stdout: io.Discard,
		stderr: io.Discard,
		exec: func(path string, args []string, stdout io.Writer, stderr io.Writer) error {
			calls = append(calls, strings.Join(args, " "))
			return nil
		},
	}

	runQuickWork(service, []string{"--status"})

	if len(calls) != 1 || calls[0] != "status" {
		t.Fatalf("expected single status call, got %v", calls)
	}
}

func TestNormalizeQuickWorkRoleAcceptsDesigner(t *testing.T) {
	role, err := normalizeQuickWorkRole("designer")
	if err != nil {
		t.Fatalf("expected designer role, got %v", err)
	}
	if role != "designer" {
		t.Fatalf("expected designer, got %q", role)
	}
}

func TestNormalizeQuickWorkRoleRejectsUnknownRole(t *testing.T) {
	if _, err := normalizeQuickWorkRole("optimizer"); err == nil {
		t.Fatal("expected unknown quickwork role to fail")
	}
}

func TestQuickWorkSeedFromBootstrapFallsBackToPrimaryAction(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			PrimaryAction: "Close the active runtime seam",
		},
	}
	seed, ok := quickWorkSeedFromBootstrap(packet, "", []string{"internal/runtime/"})
	if !ok {
		t.Fatal("expected bootstrap seed")
	}
	if seed.Role != "implementer" || seed.Summary != "Close the active runtime seam" {
		t.Fatalf("unexpected seed: %+v", seed)
	}
	if len(seed.AllowedPaths) != 1 || seed.AllowedPaths[0] != "internal/runtime/" {
		t.Fatalf("unexpected allowed paths: %+v", seed.AllowedPaths)
	}
}

func TestQuickWorkSeedFromBootstrapPrefersReviewAgenda(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			PrimaryAction:   "review_room",
			ReviewOpenCount: 1,
			ReviewTopOpen:   []string{"Provider drift still needs triage."},
		},
		ReviewRoom: &runtime.ReviewRoomSummary{
			Source:    "review_room.json",
			OpenCount: 1,
			TopOpen: []runtime.ReviewRoomItem{{
				Text:     "Provider drift still needs triage.",
				Kind:     "agenda",
				Severity: "high",
				Source:   "review.auto",
			}},
			TopAccepted: []runtime.ReviewRoomItem{},
			TopDeferred: []runtime.ReviewRoomItem{},
			Issues:      []string{},
		},
	}
	seed, ok := quickWorkSeedFromBootstrap(packet, "", nil)
	if !ok {
		t.Fatal("expected bootstrap review agenda seed")
	}
	if seed.Role != "planner" || seed.Summary != "Review agenda: Provider drift still needs triage." {
		t.Fatalf("unexpected review agenda seed: %+v", seed)
	}
}

func TestQuickWorkSeedFromBootstrapFallsBackToProposalCandidatePlanner(t *testing.T) {
	packet := runtime.SessionBootstrapPacket{
		Source: "daemon",
		Knowledge: runtime.KnowledgePacket{
			ProposalCandidates: []runtime.ProposalCandidate{{
				Title:         "Queue drift plan",
				Summary:       "Turn repeated queue drift into a planner lane.",
				CreateCommand: "layer-osctl proposal create --title \"Queue drift plan\"",
				Source:        "knowledge.proposal_gate",
			}},
		},
	}
	seed, ok := quickWorkSeedFromBootstrap(packet, "", nil)
	if !ok {
		t.Fatal("expected proposal candidate fallback seed")
	}
	if seed.Role != "planner" || seed.Summary != "Turn repeated queue drift into a planner lane." {
		t.Fatalf("unexpected proposal fallback seed: %+v", seed)
	}
}

func TestWaitForQuickWorkTerminalJobReturnsSucceededJob(t *testing.T) {
	service := &quickWorkServiceStub{
		jobs: []runtime.AgentJob{{JobID: "job_001", Status: "succeeded", UpdatedAt: time.Now().UTC()}},
	}
	job, err := waitForQuickWorkTerminalJob(service, "job_001", time.Millisecond)
	if err != nil {
		t.Fatalf("wait terminal job: %v", err)
	}
	if job.JobID != "job_001" || job.Status != "succeeded" {
		t.Fatalf("unexpected terminal job: %+v", job)
	}
}

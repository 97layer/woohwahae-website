package runtime

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestKnowledgePacketCompressesRuntimeState(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	t.Setenv("LAYER_OS_ACTORS", "claude,gemini")
	t.Setenv("LAYER_OS_DEFAULT_ACTOR", "claude")
	t.Setenv("LAYER_OS_PROVIDERS", "openai,anthropic")
	t.Setenv("LAYER_OS_HOST_CLASS", "laptop")
	t.Setenv("LAYER_OS_POWER_MODE", "low")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	goal := "Ship compact knowledge plane"
	handoff := "carry compact state"
	operator := "system"
	if err := service.ReplaceMemory(SystemMemory{
		CurrentFocus: "Backend integrity",
		CurrentGoal:  &goal,
		NextSteps:    []string{"one", "two", "three", "four"},
		OpenRisks:    []string{"drift", "staleness", "coupling", "sprawl"},
		HandoffNote:  &handoff,
		LastOperator: &operator,
		UpdatedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "First review item.", Kind: "agenda", Severity: "high", Source: "agent.codex"}); err != nil {
		t.Fatalf("add first review item: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Second review item.", Kind: "agenda", Severity: "medium", Source: "agent.codex"}); err != nil {
		t.Fatalf("add second review item: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Third review item.", Kind: "agenda", Severity: "low", Source: "agent.codex"}); err != nil {
		t.Fatalf("add third review item: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Fourth review item.", Kind: "agenda", Severity: "low", Source: "agent.codex"}); err != nil {
		t.Fatalf("add fourth review item: %v", err)
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if packet.CurrentFocus != "Backend integrity" {
		t.Fatalf("unexpected current focus: %+v", packet)
	}
	if len(packet.NextSteps) != 3 || packet.NextSteps[2] != "three" {
		t.Fatalf("expected next_steps to be trimmed to 3, got %+v", packet.NextSteps)
	}
	if len(packet.OpenRisks) != 3 || packet.OpenRisks[2] != "coupling" {
		t.Fatalf("expected open_risks to be trimmed to 3, got %+v", packet.OpenRisks)
	}
	if packet.PrimaryAction != "review_room" || packet.ReviewOpenCount != 4 {
		t.Fatalf("expected review room priority in knowledge packet, got %+v", packet)
	}
	if len(packet.ReviewTopOpen) != 3 || packet.ReviewTopOpen[0] != "First review item." {
		t.Fatalf("expected top review open texts, got %+v", packet.ReviewTopOpen)
	}
	if packet.CorpusLessons == nil || len(packet.CorpusLessons) != 0 {
		t.Fatalf("expected empty corpus lessons, got %+v", packet.CorpusLessons)
	}
	if packet.Surprising == nil || len(packet.Surprising) != 0 {
		t.Fatalf("expected empty surprising signals, got %+v", packet.Surprising)
	}
	if packet.ObservationLinks == nil || len(packet.ObservationLinks) != 0 {
		t.Fatalf("expected empty observation links, got %+v", packet.ObservationLinks)
	}
	if packet.ProposalCandidates == nil || len(packet.ProposalCandidates) != 0 {
		t.Fatalf("expected empty proposal candidates, got %+v", packet.ProposalCandidates)
	}
	if packet.ActionHints == nil || len(packet.ActionHints) != 0 {
		t.Fatalf("expected empty action hints, got %+v", packet.ActionHints)
	}
	if packet.ActionRoutes == nil || len(packet.ActionRoutes) != 0 {
		t.Fatalf("expected empty action routes, got %+v", packet.ActionRoutes)
	}
	if packet.DefaultActor != "claude" {
		t.Fatalf("expected default actor claude, got %+v", packet)
	}
	if len(packet.Actors) < 2 || packet.Actors[0] != "claude" {
		t.Fatalf("expected actor registry in packet, got %+v", packet.Actors)
	}
	if len(packet.Providers) != 2 {
		t.Fatalf("expected provider registry in packet, got %+v", packet.Providers)
	}
	if packet.GatewaySemantics != "record_only" {
		t.Fatalf("expected record_only gateway semantics, got %+v", packet)
	}
	if packet.Authority.LegacyReferenceMode != authorityLegacyReferenceMode || packet.Authority.ReviewRoomWriteSurface != "/api/layer-os/review-room" {
		t.Fatalf("unexpected authority boundary: %+v", packet.Authority)
	}
	if packet.EnvironmentAdvisory.AgentMode != "conserve" || packet.EnvironmentAdvisory.Rule != "environment_advisory.low_power_conserve" {
		t.Fatalf("unexpected environment advisory: %+v", packet.EnvironmentAdvisory)
	}
}

func TestKnowledgePacketRetrievesRelatedCorpusLessons(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.FinishSession(SessionFinishInput{CurrentFocus: "Stabilize daemon path", NextSteps: []string{"verify daemon"}, OpenRisks: []string{}}); err != nil {
		t.Fatalf("seed daemon lesson: %v", err)
	}
	if _, err := service.FinishSession(SessionFinishInput{CurrentFocus: "Wire corpus retrieval", NextSteps: []string{"retrieve corpus", "shape lesson packet"}, OpenRisks: []string{}}); err != nil {
		t.Fatalf("seed retrieval lesson: %v", err)
	}
	if _, err := service.FinishSession(SessionFinishInput{CurrentFocus: "Review release checklist", NextSteps: []string{"close release checklist"}, OpenRisks: []string{}}); err != nil {
		t.Fatalf("seed release lesson: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Wire corpus retrieval", NextSteps: []string{"retrieve corpus"}, OpenRisks: []string{}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.CorpusLessons) == 0 {
		t.Fatalf("expected related corpus lessons, got %+v", packet)
	}
	if packet.CorpusLessons[0].SourceKind != "session.finished" {
		t.Fatalf("expected session lesson source kind, got %+v", packet.CorpusLessons)
	}
	if !strings.Contains(strings.ToLower(packet.CorpusLessons[0].Summary), "corpus") {
		t.Fatalf("expected retrieval lesson first, got %+v", packet.CorpusLessons)
	}
	if len(packet.CorpusLessons) > 3 {
		t.Fatalf("expected at most 3 corpus lessons, got %+v", packet.CorpusLessons)
	}
}

func TestKnowledgePacketDerivesSurprisingSignals(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := service.FinishSession(SessionFinishInput{CurrentFocus: "bootstrap runtime", NextSteps: []string{}, OpenRisks: []string{}}); err != nil {
			t.Fatalf("seed repeated session %d: %v", i, err)
		}
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate stalled bootstrap runtime."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.Surprising) == 0 {
		t.Fatalf("expected surprising signals, got %+v", packet)
	}
	joined := strings.ToLower(strings.Join(packet.Surprising, "\n"))
	for _, needle := range []string{"open_risks", "next_steps", "current_focus repeated"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected surprising signal %q, got %+v", needle, packet.Surprising)
		}
	}
}

func TestKnowledgePacketDerivesActionHints(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "bootstrap runtime", NextSteps: []string{}, OpenRisks: []string{}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Investigate stalled bootstrap runtime."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.NextSteps) != 0 || len(packet.OpenRisks) != 0 {
		t.Fatalf("expected packet to preserve empty memory lanes, got steps=%+v risks=%+v", packet.NextSteps, packet.OpenRisks)
	}
	if len(packet.ActionHints) == 0 {
		t.Fatalf("expected derived action hints, got %+v", packet)
	}
	if len(packet.ActionRoutes) < 2 {
		t.Fatalf("expected structured action routes, got %+v", packet.ActionRoutes)
	}
	joined := strings.ToLower(strings.Join(packet.ActionHints, "\n"))
	for _, needle := range []string{"define next step", "state explicit risk"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected action hint %q, got %+v", needle, packet.ActionHints)
		}
	}
	for _, route := range packet.ActionRoutes[:2] {
		if route.TargetLane != "session_memory" || route.Command != nil {
			t.Fatalf("expected session_memory action route without command, got %+v", route)
		}
	}
}

func TestKnowledgePacketDerivesCorpusFailureSignals(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Patch runtime lane", NextSteps: []string{"retry lane"}, OpenRisks: []string{"agent failure"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	statuses := []string{"failed", "failed", "canceled"}
	summaries := []string{"Patch runtime lane", "Patch runtime lane", "Stabilize review lane"}
	for i, status := range statuses {
		jobID := fmt.Sprintf("job_signal_%03d", i+1)
		if err := service.CreateAgentJob(AgentJob{JobID: jobID, Kind: "implement", Role: "implementer", Summary: summaries[i], Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
			t.Fatalf("create agent job %d: %v", i, err)
		}
		if _, err := service.ReportAgentJob(jobID, status, []string{"reported"}, map[string]any{"last_error": "boom"}); err != nil {
			t.Fatalf("report agent job %d: %v", i, err)
		}
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	joined := strings.ToLower(strings.Join(packet.Surprising, "\n"))
	for _, needle := range []string{"agent_job.failed repeated 2 times", "succeeded entries=0"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected corpus failure signal %q, got %+v", needle, packet.Surprising)
		}
	}
}

func TestKnowledgePacketDerivesBacklogAndDispatchRiskSignals(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "stabilize provider lane", NextSteps: []string{"inspect daemon status"}, OpenRisks: []string{"provider dispatch drift"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	for i := 0; i < 15; i++ {
		if _, err := service.AddReviewRoomItem("open", fmt.Sprintf("Review backlog item %02d", i+1)); err != nil {
			t.Fatalf("seed review room %d: %v", i, err)
		}
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	joined := strings.ToLower(strings.Join(packet.Surprising, "\n"))
	for _, needle := range []string{"review_backlog_high: 15 open items", "dispatch_risk_detected"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected surprising signal %q, got %+v", needle, packet.Surprising)
		}
	}
}

func TestKnowledgePacketDerivesSucceededAfterRepeatedFailureRolePattern(t *testing.T) {
	repoRoot, err := runtimeTestRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Patch runtime lane", NextSteps: []string{"retry lane"}, OpenRisks: []string{"provider queue risk"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	statuses := []string{"failed", "failed", "succeeded"}
	for i, status := range statuses {
		jobID := fmt.Sprintf("job_signal_role_%03d", i+1)
		if err := service.CreateAgentJob(AgentJob{JobID: jobID, Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
			t.Fatalf("create agent job %d: %v", i, err)
		}
		result := map[string]any{"last_error": "boom"}
		if status == "succeeded" {
			result["summary"] = "Patched runtime lane after retry."
			result["artifacts"] = []string{"internal/runtime/knowledge_test.go"}
		}
		if _, err := service.ReportAgentJob(jobID, status, []string{"reported"}, result); err != nil {
			t.Fatalf("report agent job %d: %v", i, err)
		}
	}

	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	joined := strings.ToLower(strings.Join(packet.Surprising, "\n"))
	for _, needle := range []string{"repeated_failure pattern", "role:implementer"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected surprising signal %q, got %+v", needle, packet.Surprising)
		}
	}
}

func TestKnowledgePacketDerivesObservationLinks(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	first, err := service.CreateObservation(ObservationRecord{SourceChannel: "terminal", Topic: "queue drift", Refs: []string{"proposal_001"}, RawExcerpt: "Queue drift surfaced in terminal review.", NormalizedSummary: "Queue drift is blocking the migration map."})
	if err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{SourceChannel: "telegram", Topic: "queue drift", Refs: []string{"proposal_001"}, RawExcerpt: "Founder repeated the queue drift concern.", NormalizedSummary: "Queue drift is blocking the migration map again.", CapturedAt: first.CapturedAt.Add(time.Second)}); err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.ObservationLinks) == 0 {
		t.Fatalf("expected observation links, got %+v", packet)
	}
	link := packet.ObservationLinks[0]
	if link.Count != 2 || len(link.Channels) != 2 {
		t.Fatalf("unexpected observation link: %+v", link)
	}
	if link.Kind != "ref" || len(link.Refs) == 0 || link.Refs[0] != "proposal_001" {
		t.Fatalf("expected ref-linked observation summary, got %+v", link)
	}
}

func TestKnowledgePacketDerivesReviewProposalCandidatesFromAgendaRef(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "[intel_009] Corpus pattern signal still lacks explicit proposal.", Kind: "agenda", Severity: "high", Source: "claude.arch.review.2026-03-08"}); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.ProposalCandidates) == 0 {
		t.Fatalf("expected review-driven proposal candidates, got %+v", packet)
	}
	candidate := packet.ProposalCandidates[0]
	if candidate.Source != "knowledge.review_gate" || len(candidate.Refs) == 0 || candidate.Refs[0] != "intel_009" {
		t.Fatalf("expected review-gated candidate around intel_009, got %+v", candidate)
	}
	if len(candidate.LinkIDs) == 0 || candidate.LinkIDs[0] != "review:intel_009" {
		t.Fatalf("expected synthetic review link id, got %+v", candidate)
	}
	if len(packet.ActionRoutes) == 0 {
		t.Fatalf("expected action routes, got %+v", packet)
	}
	route := packet.ActionRoutes[0]
	if route.TargetLane != "proposal" || route.TargetRef == nil || *route.TargetRef != candidate.ProposalID {
		t.Fatalf("expected proposal action route, got %+v", route)
	}
	if route.Command == nil || *route.Command != candidate.CreateCommand {
		t.Fatalf("expected proposal create command route, got %+v", route)
	}
}

func TestKnowledgePacketKeepsReviewRefConsistentAcrossThreadAndProposal(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	why := "No explicit rationale was recorded for this agenda."
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "[intel_011] Review why-layer still lacks consistent adoption.", Kind: "agenda", Severity: "high", Source: "claude.arch.review.2026-03-08", WhyUnresolved: &why}); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.OpenThreads) == 0 || len(packet.ProposalCandidates) == 0 {
		t.Fatalf("expected open thread and proposal candidate, got %+v", packet)
	}
	var threadID string
	for _, thread := range packet.OpenThreads {
		for _, ref := range thread.PatternRefs {
			if ref == "intel_011" {
				threadID = thread.ThreadID
				break
			}
		}
		if threadID != "" {
			break
		}
	}
	if threadID == "" {
		t.Fatalf("expected some open thread to carry intel_011 ref, got %+v", packet.OpenThreads)
	}
	candidate := packet.ProposalCandidates[0]
	if len(candidate.Refs) == 0 || candidate.Refs[0] != "intel_011" {
		t.Fatalf("expected proposal ref intel_011, got %+v", candidate)
	}
	if len(candidate.ThreadIDs) == 0 || candidate.ThreadIDs[0] != threadID {
		t.Fatalf("expected proposal to point at ref-linked thread, got candidate=%+v threads=%+v", candidate, packet.OpenThreads)
	}
}

func TestKnowledgePacketDerivesProposalCandidates(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "intel_007"
	why := "We keep seeing the same signal without a managed proposal."
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Queue drift still unresolved.", Kind: "agenda", Severity: "high", Source: "review.auto", Ref: &ref, WhyUnresolved: &why}); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{SourceChannel: "terminal", Topic: "queue drift", Refs: []string{"intel_007"}, RawExcerpt: "Terminal saw queue drift again.", NormalizedSummary: "Queue drift is still blocking the migration map."}); err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{SourceChannel: "telegram", Topic: "queue drift", Refs: []string{"intel_007"}, RawExcerpt: "Telegram raised queue drift again.", NormalizedSummary: "Queue drift is still blocking the migration map."}); err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	packet := service.Knowledge()
	if err := packet.Validate(); err != nil {
		t.Fatalf("validate knowledge packet: %v", err)
	}
	if len(packet.ProposalCandidates) == 0 {
		t.Fatalf("expected proposal candidates, got %+v", packet)
	}
	candidate := packet.ProposalCandidates[0]
	if candidate.Source != "knowledge.proposal_gate" || len(candidate.LinkIDs) == 0 || len(candidate.ThreadIDs) == 0 {
		t.Fatalf("unexpected proposal candidate: %+v", candidate)
	}
	if candidate.Priority != "high" || len(candidate.Refs) == 0 || candidate.Refs[0] != "intel_007" {
		t.Fatalf("expected gated candidate around intel_007, got %+v", candidate)
	}
	if candidate.ProposalID == "" || candidate.WorkItemID == "" || candidate.Surface != SurfaceAPI || candidate.CreatePath != "/api/layer-os/proposals" {
		t.Fatalf("expected explicit proposal helper, got %+v", candidate)
	}
	if !strings.Contains(candidate.CreateCommand, "layer-osctl proposal create") || !strings.Contains(candidate.CreateCommand, candidate.ProposalID) {
		t.Fatalf("expected create command helper, got %+v", candidate)
	}
	if candidate.PromotePath != "/api/layer-os/proposals/promote" {
		t.Fatalf("expected promote path helper, got %+v", candidate)
	}
	if !strings.Contains(candidate.PromoteCommand, "layer-osctl proposal promote") || !strings.Contains(candidate.PromoteCommand, candidate.WorkItemID) {
		t.Fatalf("expected promote command helper, got %+v", candidate)
	}
	joined := strings.Join(packet.ActionHints, "\n")
	if !strings.Contains(joined, candidate.ProposalID) {
		t.Fatalf("expected action hint to mention proposal helper, got %+v", packet.ActionHints)
	}
	if len(packet.ActionRoutes) == 0 {
		t.Fatalf("expected action routes, got %+v", packet)
	}
	route := packet.ActionRoutes[0]
	if route.TargetLane != "proposal" || route.TargetRef == nil || *route.TargetRef != candidate.ProposalID {
		t.Fatalf("expected proposal action route, got %+v", route)
	}
	if route.Command == nil || *route.Command != candidate.CreateCommand {
		t.Fatalf("expected action route create command, got %+v", route)
	}
}

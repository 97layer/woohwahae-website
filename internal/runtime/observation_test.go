package runtime

import (
	"testing"
	"time"
)

func TestCreateObservationDefaultsAndQuery(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	first, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "terminal",
		Topic:             "queue drift",
		Refs:              []string{"job_001", "job_001"},
		RawExcerpt:        "Queue drift surfaced in terminal thread.",
		NormalizedSummary: "Queue drift surfaced while reviewing terminal thread.",
	})
	if err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	if first.ObservationID == "" || first.Actor == "" {
		t.Fatalf("expected generated id and actor, got %+v", first)
	}
	if len(first.Refs) != 1 || first.Refs[0] != "job_001" {
		t.Fatalf("expected normalized refs, got %+v", first.Refs)
	}
	if first.Confidence != "medium" {
		t.Fatalf("expected default confidence medium, got %+v", first)
	}
	second, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "telegram",
		Topic:             "queue drift",
		Actor:             "founder",
		Refs:              []string{"thread_001"},
		Confidence:        "high",
		RawExcerpt:        "Founder raised same queue concern over Telegram.",
		NormalizedSummary: "Telegram repeated the same queue drift concern.",
		CapturedAt:        time.Now().UTC().Add(time.Second),
	})
	if err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	all := service.ListObservations(ObservationQuery{})
	if len(all) != 2 || all[0].ObservationID != second.ObservationID {
		t.Fatalf("expected newest-first observation list, got %+v", all)
	}
	telegram := service.ListObservations(ObservationQuery{SourceChannel: "telegram"})
	if len(telegram) != 1 || telegram[0].ObservationID != second.ObservationID {
		t.Fatalf("expected telegram query to match second observation, got %+v", telegram)
	}
	text := service.ListObservations(ObservationQuery{Text: "terminal thread"})
	if len(text) != 1 || text[0].ObservationID != first.ObservationID {
		t.Fatalf("expected text query to match first observation, got %+v", text)
	}
	byRef := service.ListObservations(ObservationQuery{Ref: "thread_001"})
	if len(byRef) != 1 || byRef[0].ObservationID != second.ObservationID {
		t.Fatalf("expected ref query to match second observation, got %+v", byRef)
	}
	if got := service.Handoff().Counts.Observations; got != 2 {
		t.Fatalf("expected handoff observation count 2, got %d", got)
	}
}

func TestObservationSnapshotRoundTrip(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	created, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "thread",
		Actor:             "founder",
		Topic:             "migration map",
		Refs:              []string{"proposal_001"},
		Confidence:        "high",
		RawExcerpt:        "Legacy migration map still has hidden coupling.",
		NormalizedSummary: "Legacy migration map still carries hidden coupling.",
	})
	if err != nil {
		t.Fatalf("create observation: %v", err)
	}
	packet := service.Snapshot()
	if got := len(packet.Observations); got != 1 {
		t.Fatalf("expected snapshot observations 1, got %d", got)
	}
	target, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new target service: %v", err)
	}
	if err := target.ImportSnapshot(packet); err != nil {
		t.Fatalf("import snapshot: %v", err)
	}
	items := target.ListObservations(ObservationQuery{})
	if len(items) != 1 || items[0].ObservationID != created.ObservationID {
		t.Fatalf("expected imported observations to match snapshot, got %+v", items)
	}
}

func TestFinishSessionAutoIngestsObservation(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.FinishSession(SessionFinishInput{CurrentFocus: "Lock queue", NextSteps: []string{"sync"}, OpenRisks: []string{"drift"}}); err != nil {
		t.Fatalf("finish session: %v", err)
	}
	items := service.ListObservations(ObservationQuery{SourceChannel: "session"})
	if len(items) != 1 {
		t.Fatalf("expected 1 auto session observation, got %+v", items)
	}
	if items[0].Topic != "session.finished" || items[0].Confidence != "high" {
		t.Fatalf("unexpected session observation: %+v", items[0])
	}
}

func TestReportAgentJobAutoIngestsObservation(t *testing.T) {
	repoRoot, err := runtimeTestRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreateAgentJob(AgentJob{JobID: "job_obs_001", Kind: "implement", Role: "implementer", Summary: "Patch runtime lane", Status: "queued", Source: "founder.manual", Surface: SurfaceAPI, Stage: StageCompose, Notes: []string{}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("create agent job: %v", err)
	}
	if _, err := service.ReportAgentJob("job_obs_001", "succeeded", []string{"patched"}, map[string]any{
		"summary":         "Patched runtime lane and recorded the output file.",
		"artifacts":       []string{"internal/runtime/observation_test.go"},
		"provider":        "openai",
		"gateway_call_id": "gateway_001",
	}); err != nil {
		t.Fatalf("report agent job: %v", err)
	}
	items := service.ListObservations(ObservationQuery{SourceChannel: "agent_job"})
	if len(items) != 1 {
		t.Fatalf("expected 1 auto job observation, got %+v", items)
	}
	if items[0].Topic != "agent_job.succeeded" || len(items[0].Refs) < 2 {
		t.Fatalf("unexpected job observation: %+v", items[0])
	}
}

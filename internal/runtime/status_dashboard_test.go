package runtime

import (
	"testing"
	"time"
)

func TestStatusIncludesProgressDashboard(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "openai,gemini")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_001",
		SourceChannel:     "telegram",
		CapturedAt:        now,
		Actor:             "founder",
		Topic:             "link",
		Refs:              []string{"youtube:abc123"},
		Confidence:        "high",
		RawExcerpt:        "Useful video note",
		NormalizedSummary: "Useful video note",
	}); err != nil {
		t.Fatalf("create observation: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		ObservationID:     "observation_002",
		SourceChannel:     "gemini",
		CapturedAt:        now,
		Actor:             "gemini",
		Topic:             "diagnosis",
		Refs:              []string{"gemini_sync:test"},
		Confidence:        "medium",
		RawExcerpt:        "Verifier artifact",
		NormalizedSummary: "Verifier artifact",
	}); err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	if err := service.DebugAppendCapitalizationEntry(CapitalizationEntry{
		EntryID:       "cap_event_001",
		CreatedAt:     now,
		Actor:         "founder",
		SourceEventID: "event_001",
		SourceKind:    "session.finished",
		Situation:     CapitalizationFacet{Summary: "Close queue drift"},
		Decision:      CapitalizationFacet{Summary: "Enable automation"},
		Cost:          CapitalizationFacet{Summary: "Carry-over cost"},
		Result:        CapitalizationFacet{Summary: "Corpus updated"},
	}); err != nil {
		t.Fatalf("append capitalization entry: %v", err)
	}

	status := service.Status()
	if status.Progress == nil {
		t.Fatal("expected progress dashboard on company state")
	}
	if len(status.Progress.Axes) != 7 {
		t.Fatalf("expected 7 progress axes, got %+v", status.Progress.Axes)
	}
	if status.Progress.OverallPercent <= 0 || status.Progress.OverallPercent > 100 {
		t.Fatalf("unexpected overall percent: %+v", status.Progress)
	}
	labels := map[string]int{}
	for _, axis := range status.Progress.Axes {
		labels[axis.Label] = axis.Percent
		if axis.Bar == "" || axis.Status == "" || len(axis.Signals) == 0 {
			t.Fatalf("expected axis formatting, got %+v", axis)
		}
	}
	for _, label := range []string{"핵심 인프라", "보안 자세", "지식 파이프라인", "에이전트 통제", "멀티에이전트 chain", "자동 트리거", "커밋 안정화"} {
		if _, ok := labels[label]; !ok {
			t.Fatalf("missing progress axis %q in %+v", label, status.Progress.Axes)
		}
	}
	if labels["지식 파이프라인"] < 80 {
		t.Fatalf("expected knowledge pipeline to reflect corpus+observation coverage, got %+v", status.Progress.Axes)
	}
	if labels["멀티에이전트 chain"] < 60 {
		t.Fatalf("expected multi-agent chain baseline score, got %+v", status.Progress.Axes)
	}
}

func TestCompanyStateValidateAcceptsOptionalProgress(t *testing.T) {
	state := CompanyState{
		ShellMode:      "founder",
		MemoryHealth:   "ready",
		DeployHealth:   "ready",
		PrimarySurface: SurfaceCockpit,
		ActiveSurfaces: []Surface{SurfaceCockpit, SurfaceTelegram, SurfaceAPI},
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("expected valid company state without progress, got %v", err)
	}
	state.Progress = &ProgressDashboard{OverallPercent: 55, OverallBar: "[#####-----]", OverallStatus: "building", Axes: []ProgressAxis{{AxisID: "infra_core", Label: "핵심 인프라", Percent: 55, Bar: "[#####-----]", Status: "building", Signals: []string{"ok"}}}}
	if err := state.Validate(); err != nil {
		t.Fatalf("expected valid company state with progress, got %v", err)
	}
}

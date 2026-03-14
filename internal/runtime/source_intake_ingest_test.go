package runtime

import (
	"strings"
	"testing"
)

func TestBuildSourceIntakeRecordFromContentUsesObservationHints(t *testing.T) {
	observation := BuildContentObservation(ContentObservationInput{
		SourceChannel: "crawler",
		Actor:         "rss_sensor",
		Title:         "Calm operating note",
		Summary:       "Calm operating note",
		Excerpt:       "A quieter operating cadence matters more than speed.",
		SourceURL:     "https://example.com/post-1",
		Kind:          "article",
		DocID:         "post_1",
		Tags:          []string{"signals", "beauty"},
		Refs:          []string{"route:woohwahae", "worldview:subtraction"},
	})

	record := BuildSourceIntakeRecordFromContent(SourceIntakeContentInput{
		SourceObservation: observation,
		IntakeClass:       "public_collector",
		Title:             "Calm operating note",
		URL:               "https://example.com/post-1",
		Excerpt:           "A quieter operating cadence matters more than speed.",
	})

	if record.IntakeClass != "public_collector" || record.PolicyColor != "green" {
		t.Fatalf("unexpected source intake record: %+v", record)
	}
	if record.PriorityScore < 60 || record.Disposition != "prep" {
		t.Fatalf("expected prep-worthy source intake priority, got %+v", record)
	}
	if !strings.Contains(record.DispositionNote, "single route cue") {
		t.Fatalf("expected prep disposition note, got %q", record.DispositionNote)
	}
	if !containsStringLocal(record.DomainTags, "signals") || !containsStringLocal(record.DomainTags, "beauty") {
		t.Fatalf("expected feed tags to flow into domain tags, got %+v", record.DomainTags)
	}
	if !containsStringLocal(record.WorldviewTags, "subtraction") {
		t.Fatalf("expected worldview tags from refs, got %+v", record.WorldviewTags)
	}
	if len(record.SuggestedRoutes) != 1 || record.SuggestedRoutes[0] != "woohwahae" {
		t.Fatalf("expected route hint from refs, got %+v", record.SuggestedRoutes)
	}
}

func TestBuildSourceIntakeRecordFromContentDefaultsYellowWhenRouteCueIsMissing(t *testing.T) {
	observation := BuildContentObservation(ContentObservationInput{
		SourceChannel: "crawler",
		Actor:         "rss_sensor",
		Title:         "Unsorted field note",
		Summary:       "Unsorted field note",
		Excerpt:       "A public signal arrived without a clear publishing lane.",
		SourceURL:     "https://example.com/post-2",
		Kind:          "article",
		DocID:         "post_2",
	})

	record := BuildSourceIntakeRecordFromContent(SourceIntakeContentInput{
		SourceObservation: observation,
		IntakeClass:       "public_collector",
		Title:             "Unsorted field note",
		URL:               "https://example.com/post-2",
		Excerpt:           "A public signal arrived without a clear publishing lane.",
	})

	if record.PolicyColor != "yellow" {
		t.Fatalf("expected missing route cue to default yellow, got %+v", record)
	}
	if record.Disposition != "review" {
		t.Fatalf("expected missing route cue to require review, got %+v", record)
	}
	if !strings.Contains(record.FounderNote, "route cue가 비어") {
		t.Fatalf("expected founder note to explain missing route cue, got %q", record.FounderNote)
	}
	if !strings.Contains(record.DispositionNote, "route cue가 비어") {
		t.Fatalf("expected disposition note to explain missing route cue, got %q", record.DispositionNote)
	}
}

func TestBuildSourceIntakeRecordFromContentDefaultsYellowWhenRouteCueConflicts(t *testing.T) {
	observation := BuildContentObservation(ContentObservationInput{
		SourceChannel: "crawler",
		Actor:         "rss_sensor",
		Title:         "Conflicting route note",
		Summary:       "Conflicting route note",
		Excerpt:       "A public signal arrived with multiple publishing lanes.",
		SourceURL:     "https://example.com/post-3",
		Kind:          "article",
		DocID:         "post_3",
		Refs:          []string{"route:97layer", "route:woohwahae"},
	})

	record := BuildSourceIntakeRecordFromContent(SourceIntakeContentInput{
		SourceObservation: observation,
		IntakeClass:       "public_collector",
		Title:             "Conflicting route note",
		URL:               "https://example.com/post-3",
		Excerpt:           "A public signal arrived with multiple publishing lanes.",
	})

	if record.PolicyColor != "yellow" {
		t.Fatalf("expected conflicting route cues to default yellow, got %+v", record)
	}
	if record.Disposition != "review" {
		t.Fatalf("expected conflicting route cues to require review, got %+v", record)
	}
	if !strings.Contains(record.FounderNote, "여러 route cue") {
		t.Fatalf("expected founder note to explain conflicting route cues, got %q", record.FounderNote)
	}
	if !strings.Contains(record.DispositionNote, "여러 개") {
		t.Fatalf("expected disposition note to explain conflicting route cues, got %q", record.DispositionNote)
	}
}

func TestEnsureSourceIntakeObservationCreatesAndDedupesByContentDoc(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	content, err := service.CreateObservation(BuildContentObservation(ContentObservationInput{
		SourceChannel: "crawler",
		Actor:         "rss_sensor",
		Title:         "Quiet beauty note",
		Summary:       "Quiet beauty note",
		Excerpt:       "Less noise, more signal.",
		SourceURL:     "https://example.com/post-1",
		Kind:          "article",
		DocID:         "post_1",
		Tags:          []string{"beauty"},
		Refs:          []string{"route:woosunhokr"},
	}))
	if err != nil {
		t.Fatalf("create content observation: %v", err)
	}

	record := BuildSourceIntakeRecordFromContent(SourceIntakeContentInput{
		SourceObservation: content,
		IntakeClass:       "public_collector",
		Title:             "Quiet beauty note",
		URL:               "https://example.com/post-1",
		Excerpt:           "Less noise, more signal.",
	})

	first, created, err := EnsureSourceIntakeObservation(service, content, record)
	if err != nil {
		t.Fatalf("ensure source intake: %v", err)
	}
	if !created {
		t.Fatalf("expected source intake observation to be created")
	}
	if strings.TrimSpace(first.Topic) != SourceIntakeTopic {
		t.Fatalf("unexpected source intake topic: %+v", first)
	}
	if !containsStringLocal(first.Refs, "source_observation:"+content.ObservationID) || !containsStringLocal(first.Refs, "content_doc:post_1") {
		t.Fatalf("unexpected source intake refs: %+v", first.Refs)
	}

	second, createdAgain, err := EnsureSourceIntakeObservation(service, content, record)
	if err != nil {
		t.Fatalf("ensure source intake again: %v", err)
	}
	if createdAgain {
		t.Fatalf("expected second ensure to reuse source intake observation")
	}
	if first.ObservationID != second.ObservationID {
		t.Fatalf("expected same source intake observation, got %s vs %s", first.ObservationID, second.ObservationID)
	}
}

func TestSourceIntakeAutoRouteTargetRequiresPrepDisposition(t *testing.T) {
	if target, ok := SourceIntakeAutoRouteTarget(SourceIntakeRecord{
		PolicyColor:     "green",
		Disposition:     "observe",
		SuggestedRoutes: []string{"woohwahae"},
	}, []string{"route:woohwahae"}); ok || target != "" {
		t.Fatalf("expected observe disposition to stay quiet, got target=%q ok=%v", target, ok)
	}

	if target, ok := SourceIntakeAutoRouteTarget(SourceIntakeRecord{
		PolicyColor:     "green",
		Disposition:     "prep",
		SuggestedRoutes: []string{"woohwahae"},
	}, []string{"route:woohwahae"}); !ok || target != "woohwahae" {
		t.Fatalf("expected prep disposition to auto-route, got target=%q ok=%v", target, ok)
	}
}

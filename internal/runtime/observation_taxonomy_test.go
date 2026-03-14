package runtime

import "testing"

func TestCanonicalObservationChannelSupportsFutureAdapters(t *testing.T) {
	cases := map[string]string{
		"tg":              "telegram",
		"antigravity-ide": "antigravity",
		"gemini-cli":      "gemini",
		"crawl":           "crawler",
		"archive":         "personal_db",
		"notebook-lm":     "notebook_lm",
		"terminal":        "terminal",
	}
	for raw, want := range cases {
		if got := canonicalObservationChannel(raw); got != want {
			t.Fatalf("canonicalObservationChannel(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestObservationMetadataRefsBuildsStableTaxonomyRefs(t *testing.T) {
	refs := ObservationMetadataRefs("antigravity-ide", "Implementation Plan", map[string]string{
		"artifact_type":  "ARTIFACT_TYPE_PLAN",
		"artifact_stem":  "implementation_plan.md",
		"severity":       "High",
		"ingest_adapter": "Antigravity",
	})
	wants := []string{
		"source:antigravity",
		"source_family:agent_workspace",
		"topic:implementation_plan",
		"artifact_type:artifact_type_plan",
		"artifact_stem:implementation_plan_md",
		"severity:high",
		"ingest_adapter:antigravity",
	}
	for _, want := range wants {
		if !containsString(refs, want) {
			t.Fatalf("expected ref %q in %+v", want, refs)
		}
	}
}

func TestObservationLinkIgnoresTaxonomyRefs(t *testing.T) {
	items := []ObservationRecord{
		{
			ObservationID:     "observation_001",
			SourceChannel:     "antigravity",
			Actor:             "antigravity",
			Topic:             "task",
			Refs:              []string{"source:antigravity", "source_family:agent_workspace", "topic:task", "antigravity_run:run-001", "artifact_stem:task_md"},
			Confidence:        "high",
			RawExcerpt:        "task excerpt",
			NormalizedSummary: "Implement ingest lane",
		},
		{
			ObservationID:     "observation_002",
			SourceChannel:     "antigravity",
			Actor:             "antigravity",
			Topic:             "walkthrough",
			Refs:              []string{"source:antigravity", "source_family:agent_workspace", "topic:walkthrough", "antigravity_run:run-001", "artifact_stem:walkthrough_md"},
			Confidence:        "high",
			RawExcerpt:        "walkthrough excerpt",
			NormalizedSummary: "Inspect review backlog",
		},
	}
	links := deriveObservationLinks(items)
	if len(links) != 0 {
		t.Fatalf("expected taxonomy refs to be ignored, got %+v", links)
	}
}

func TestObservationLinkDoesNotLinkGenericContentTopics(t *testing.T) {
	items := []ObservationRecord{
		{ObservationID: "observation_010", SourceChannel: "telegram", Actor: "founder", Topic: "message", Confidence: "high", RawExcerpt: "one", NormalizedSummary: "An unrelated founder message."},
		{ObservationID: "observation_011", SourceChannel: "crawler", Actor: "crawler", Topic: "article", Confidence: "high", RawExcerpt: "two", NormalizedSummary: "A separate crawled article."},
	}
	if links := deriveObservationLinks(items); len(links) != 0 {
		t.Fatalf("expected generic content topics not to self-link, got %+v", links)
	}
}

func TestObservationLinkDoesNotLinkGenericGeminiArtifactTopics(t *testing.T) {
	items := []ObservationRecord{
		{ObservationID: "observation_020", SourceChannel: "gemini", Actor: "gemini", Topic: "task", Confidence: "high", RawExcerpt: "one", NormalizedSummary: "A Gemini task artifact was recovered."},
		{ObservationID: "observation_021", SourceChannel: "gemini", Actor: "gemini", Topic: "walkthrough", Confidence: "high", RawExcerpt: "two", NormalizedSummary: "A Gemini walkthrough artifact was recovered."},
	}
	if links := deriveObservationLinks(items); len(links) != 0 {
		t.Fatalf("expected generic gemini artifact topics not to self-link, got %+v", links)
	}
}

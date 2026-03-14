package runtime

import "testing"

func TestBuildTelegramObservationClassifiesDirectionAndLinkMetadata(t *testing.T) {
	obs := BuildTelegramObservation(TelegramObservationInput{
		Topic:     "",
		Title:     "Founder shared a YouTube strategy clip",
		Excerpt:   "https://youtube.com/watch?v=abc123\nThis should influence recommendation strategy.",
		SourceURL: "https://youtube.com/watch?v=abc123",
		Direction: "inbound",
		Kind:      "link",
		MessageID: "msg-001",
		ChatID:    "founder-room",
		Username:  "founder",
		Tags:      []string{"recommendation", "media"},
	})
	if obs.SourceChannel != "telegram" || obs.Actor != "founder" {
		t.Fatalf("unexpected telegram observation: %+v", obs)
	}
	wants := []string{
		"source:telegram",
		"source_family:founder_inbox",
		"content_kind:link",
		"interaction_direction:inbound",
		"telegram_message:msg_001",
		"telegram_chat:founder_room",
		"telegram_user:founder",
		"content_host:youtube_com",
		"tag:recommendation",
	}
	for _, want := range wants {
		if !containsString(obs.Refs, want) {
			t.Fatalf("expected ref %q in %+v", want, obs.Refs)
		}
	}
}

func TestBuildContentObservationClassifiesNotebookSource(t *testing.T) {
	obs := BuildContentObservation(ContentObservationInput{
		SourceChannel: "notebook-lm",
		Title:         "Second brain clustering note",
		Excerpt:       "A founder note about clustering content signals into reusable themes.",
		SourceURL:     "https://example.com/research/clustering",
		Author:        "Founder",
		Kind:          "note",
		DocID:         "note-001",
		Tags:          []string{"research", "second-brain"},
	})
	if obs.SourceChannel != "notebook_lm" {
		t.Fatalf("expected notebook_lm source channel, got %+v", obs)
	}
	wants := []string{
		"source:notebook_lm",
		"source_family:founder_archive",
		"content_kind:note",
		"content_author:founder",
		"content_doc:note_001",
		"content_host:example_com",
		"tag:research",
		"tag:second_brain",
	}
	for _, want := range wants {
		if !containsString(obs.Refs, want) {
			t.Fatalf("expected ref %q in %+v", want, obs.Refs)
		}
	}
}

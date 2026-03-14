package runtime

import (
	"strings"
	"testing"
)

func TestTelegramSemanticsProfilesCoverCanonicalInteractions(t *testing.T) {
	profiles := TelegramSemanticsProfiles()
	if len(profiles) != 5 {
		t.Fatalf("expected 5 canonical profiles, got %d", len(profiles))
	}
	names := map[string]bool{}
	for _, p := range profiles {
		if names[p.Interaction] {
			t.Fatalf("duplicate interaction %q", p.Interaction)
		}
		names[p.Interaction] = true
		if strings.TrimSpace(p.Topic) == "" || strings.TrimSpace(p.ContentKind) == "" || strings.TrimSpace(p.Direction) == "" {
			t.Fatalf("profile %q missing topic/kind/direction: %+v", p.Interaction, p)
		}
	}
	for _, want := range []string{"inbound_conversation", "outbound_publication", "feedback_reply", "youtube_link_injection", "bookmark_share"} {
		if !names[want] {
			t.Fatalf("missing canonical interaction %q", want)
		}
	}
}

func TestFindTelegramProfileNormalizesInteraction(t *testing.T) {
	profile, ok := FindTelegramProfile("Inbound-Conversation")
	if !ok {
		t.Fatalf("expected to find inbound_conversation profile")
	}
	if profile.Interaction != "inbound_conversation" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestApplyTelegramProfileFillsGapsButKeepsExplicitFields(t *testing.T) {
	base := TelegramObservationInput{
		Topic:     "custom_topic",
		Kind:      "",
		Direction: "",
		Tags:      []string{"existing"},
		Refs:      []string{"custom_ref"},
	}
	profile := TelegramSemanticsProfiles()[0] // inbound_conversation

	updated := ApplyTelegramProfile(profile, base)

	if updated.Topic != "custom_topic" { // caller wins
		t.Fatalf("expected topic to stay caller-provided, got %q", updated.Topic)
	}
	if updated.Kind != profile.ContentKind {
		t.Fatalf("expected kind to default from profile, got %q", updated.Kind)
	}
	if updated.Direction != profile.Direction {
		t.Fatalf("expected direction to default from profile, got %q", updated.Direction)
	}
	if !containsExactString(updated.Tags, "existing") || !containsExactString(updated.Tags, "founder_inbox") {
		t.Fatalf("expected merged tags, got %+v", updated.Tags)
	}
	if !containsExactString(updated.Refs, "custom_ref") || !containsExactString(updated.Refs, "interaction_direction:inbound") {
		t.Fatalf("expected merged refs, got %+v", updated.Refs)
	}
}

func TestTelegramProfileRefHintsAreNonLinkable(t *testing.T) {
	for _, profile := range TelegramSemanticsProfiles() {
		for _, ref := range profile.RefHints {
			if !isNonLinkableObservationRef(ref) {
				t.Fatalf("expected ref %q to be marked non-linkable", ref)
			}
		}
	}
}

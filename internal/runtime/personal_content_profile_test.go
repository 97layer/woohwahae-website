package runtime

import (
	"strings"
	"testing"
)

func TestPersonalContentProfilesFamilyAndChannels(t *testing.T) {
	profiles := PersonalContentProfiles()
	if len(profiles) == 0 {
		t.Fatalf("expected profiles")
	}
	foundNotebook := false
	for _, p := range profiles {
		if canonicalObservationChannel(p.SourceChannel) == "" {
			t.Fatalf("profile missing source channel: %+v", p)
		}
		if strings.TrimSpace(p.Kind) == "" || strings.TrimSpace(p.Topic) == "" {
			t.Fatalf("profile missing kind/topic: %+v", p)
		}
		if p.SourceChannel == "notebook_lm" {
			foundNotebook = true
		}
	}
	if !foundNotebook {
		t.Fatalf("expected notebook_lm profile present")
	}
}

func TestFindPersonalContentProfileMatchesChannelAndKind(t *testing.T) {
	profile, ok := FindPersonalContentProfile("personal_db", "Note")
	if !ok {
		t.Fatalf("expected to find personal_db note")
	}
	if profile.Topic != "note" || profile.Kind != "note" {
		t.Fatalf("unexpected profile %+v", profile)
	}
}

func TestApplyPersonalContentProfileDefaultsAndDocHint(t *testing.T) {
	base := ContentObservationInput{
		Title:  "Founder note",
		DocID:  "",
		Tags:   []string{"existing"},
		Refs:   []string{"custom_ref"},
		Author: "Founder",
	}
	profile := PersonalContentProfiles()[0] // personal_db note

	updated := ApplyPersonalContentProfile(profile, base)
	if updated.SourceChannel != "personal_db" {
		t.Fatalf("expected source channel personal_db, got %q", updated.SourceChannel)
	}
	if updated.Kind != "note" || updated.Topic != "note" {
		t.Fatalf("expected note defaults, got kind=%q topic=%q", updated.Kind, updated.Topic)
	}
	if !containsExactString(updated.Tags, "existing") || !containsExactString(updated.Tags, "founder_archive") {
		t.Fatalf("expected merged tags, got %+v", updated.Tags)
	}
	if !containsExactString(updated.Refs, "custom_ref") || !containsExactString(updated.Refs, "source_family:founder_archive") {
		t.Fatalf("expected merged refs, got %+v", updated.Refs)
	}
	if updated.DocID != "" {
		t.Fatalf("did not expect doc hint for personal_db note, got %q", updated.DocID)
	}
}

func TestApplyPersonalContentProfileAddsNotebookDocHint(t *testing.T) {
	profile := PersonalContentProfiles()[len(PersonalContentProfiles())-1] // notebook_lm qa
	base := ContentObservationInput{
		SourceURL: "https://notebooklm.google/some/doc",
		Tags:      nil,
	}
	updated := ApplyPersonalContentProfile(profile, base)
	if updated.SourceChannel != "notebook_lm" {
		t.Fatalf("expected notebook_lm source, got %q", updated.SourceChannel)
	}
	if updated.DocID == "" {
		t.Fatalf("expected doc hint to be applied for notebook_lm")
	}
}

func TestApplyPersonalContentProfileKeepsCallerDocID(t *testing.T) {
	profile := PersonalContentProfiles()[len(PersonalContentProfiles())-1] // notebook_lm qa
	base := ContentObservationInput{
		DocID: "caller_doc",
	}
	updated := ApplyPersonalContentProfile(profile, base)
	if updated.DocID != "caller_doc" {
		t.Fatalf("expected caller docID to remain, got %q", updated.DocID)
	}
}

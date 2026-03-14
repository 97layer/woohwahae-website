package runtime

import "strings"

// PersonalContentProfile captures canonical personal_db / notebook_lm shapes so
// ingest helpers can stay deterministic without changing the core ingest path.
type PersonalContentProfile struct {
	SourceChannel string   // personal_db or notebook_lm
	Kind          string   // note, summary, qa, transcript, bookmark, outline
	Topic         string   // topic attached to the observation
	DefaultTags   []string // tag:<value> refs for later retrieval filtering
	RefHints      []string // non-linkable refs describing the family
	DocHint       string   // optional doc-id hint when caller lacks one
	Summary       string   // human rationale for the mapping
}

// PersonalContentProfiles lists the Phase 1.5 canonical profiles. notebook_lm
// and personal_db stay in the same founder_archive family but keep distinct
// source_channel values so later retrieval can preserve provenance.
func PersonalContentProfiles() []PersonalContentProfile {
	return []PersonalContentProfile{
		{
			SourceChannel: "personal_db",
			Kind:          "note",
			Topic:         "note",
			DefaultTags:   []string{"founder_archive", "note"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Free-form personal note captured in the founder archive.",
		},
		{
			SourceChannel: "personal_db",
			Kind:          "summary",
			Topic:         "summary",
			DefaultTags:   []string{"founder_archive", "summary"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Condensed personal summary intended for later retrieval.",
		},
		{
			SourceChannel: "personal_db",
			Kind:          "qa",
			Topic:         "qa",
			DefaultTags:   []string{"founder_archive", "qa"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Question/answer pair with explicit prompt/answer fields.",
		},
		{
			SourceChannel: "personal_db",
			Kind:          "transcript",
			Topic:         "transcript",
			DefaultTags:   []string{"founder_archive", "transcript"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Transcribed meeting or call; keep identity for redaction and reuse.",
		},
		{
			SourceChannel: "personal_db",
			Kind:          "bookmark",
			Topic:         "bookmark",
			DefaultTags:   []string{"founder_archive", "bookmark"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Saved link with doc-id/title/author so crawler dedupe can merge later.",
		},
		{
			SourceChannel: "personal_db",
			Kind:          "outline",
			Topic:         "outline",
			DefaultTags:   []string{"founder_archive", "outline"},
			RefHints:      []string{"source_family:founder_archive"},
			Summary:       "Structured outline for future drafts; keep headings intact.",
		},
		{
			SourceChannel: "notebook_lm",
			Kind:          "summary",
			Topic:         "summary",
			DefaultTags:   []string{"founder_archive", "notebook_lm", "summary"},
			RefHints:      []string{"source_family:founder_archive"},
			DocHint:       "notebooklm_auto",
			Summary:       "NotebookLM auto-summary; provenance stays explicit via source_channel.",
		},
		{
			SourceChannel: "notebook_lm",
			Kind:          "qa",
			Topic:         "qa",
			DefaultTags:   []string{"founder_archive", "notebook_lm", "qa"},
			RefHints:      []string{"source_family:founder_archive"},
			DocHint:       "notebooklm_auto",
			Summary:       "NotebookLM Q&A result; keep doc hint to group regenerated answers.",
		},
	}
}

// FindPersonalContentProfile resolves a profile by source channel + kind.
func FindPersonalContentProfile(sourceChannel string, kind string) (PersonalContentProfile, bool) {
	channel := canonicalObservationChannel(sourceChannel)
	kind = normalizeObservationMetaToken(kind)
	for _, profile := range PersonalContentProfiles() {
		if canonicalObservationChannel(profile.SourceChannel) == channel && normalizeObservationMetaToken(profile.Kind) == kind {
			return profile, true
		}
	}
	return PersonalContentProfile{}, false
}

// ApplyPersonalContentProfile stamps stable defaults; caller-provided fields
// always win. This keeps identity/dedupe hints consistent across adapters.
func ApplyPersonalContentProfile(profile PersonalContentProfile, input ContentObservationInput) ContentObservationInput {
	if strings.TrimSpace(input.SourceChannel) == "" {
		input.SourceChannel = profile.SourceChannel
	}
	if strings.TrimSpace(input.Kind) == "" {
		input.Kind = profile.Kind
	}
	if strings.TrimSpace(input.Topic) == "" {
		input.Topic = profile.Topic
	}
	for _, tag := range profile.DefaultTags {
		input.Tags = appendUniqueString(input.Tags, tag)
	}
	for _, ref := range profile.RefHints {
		input.Refs = appendUniqueString(input.Refs, ref)
	}
	if strings.TrimSpace(input.DocID) == "" && strings.TrimSpace(profile.DocHint) != "" {
		input.DocID = profile.DocHint
	}
	return input
}

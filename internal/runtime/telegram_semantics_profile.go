package runtime

import "strings"

// TelegramInteractionProfile captures the canonical mapping for a Telegram
// interaction so ingest callers can stamp stable topics, kinds, directions,
// and tagging without rewriting the core ingest path.
type TelegramInteractionProfile struct {
	Interaction string   // stable identifier, e.g. inbound_conversation
	Topic       string   // normalized topic to attach to the observation
	ContentKind string   // maps to content_kind metadata
	Direction   string   // maps to interaction_direction metadata
	DefaultTags []string // tag:<value> refs appended for quick filtering
	RefHints    []string // non-linkable refs that describe the interaction
	Summary     string   // human-readable explanation
}

// TelegramSemanticsProfiles returns the Phase 1.5 canonical Telegram profiles
// used by ingest-to-surface so helpers and fixtures stay in one place.
func TelegramSemanticsProfiles() []TelegramInteractionProfile {
	return []TelegramInteractionProfile{
		{
			Interaction: "inbound_conversation",
			Topic:       "conversation",
			ContentKind: "message",
			Direction:   "inbound",
			DefaultTags: []string{"founder_inbox", "conversation"},
			RefHints:    []string{"interaction_direction:inbound"},
			Summary:     "Founder inbox or partner reaches in; treat as a high-signal conversation message, not a broadcast.",
		},
		{
			Interaction: "outbound_publication",
			Topic:       "broadcast",
			ContentKind: "publication",
			Direction:   "outbound",
			DefaultTags: []string{"publication", "broadcast"},
			RefHints:    []string{"interaction_direction:outbound"},
			Summary:     "Founder sends a one-to-many update; keep it distinct from inbox chatter.",
		},
		{
			Interaction: "feedback_reply",
			Topic:       "feedback",
			ContentKind: "feedback",
			Direction:   "inbound",
			DefaultTags: []string{"feedback", "reply"},
			RefHints:    []string{"interaction_direction:inbound"},
			Summary:     "Structured reply or critique that should land in review or follow-up queues.",
		},
		{
			Interaction: "youtube_link_injection",
			Topic:       "youtube_link",
			ContentKind: "link",
			Direction:   "inbound",
			DefaultTags: []string{"youtube", "link"},
			RefHints:    []string{"content_kind:link"},
			Summary:     "Inbound share of a YouTube link; keep host + link tags for later ranking and dedupe.",
		},
		{
			Interaction: "bookmark_share",
			Topic:       "bookmark",
			ContentKind: "bookmark",
			Direction:   "outbound",
			DefaultTags: []string{"bookmark", "share"},
			RefHints:    []string{"interaction_direction:outbound"},
			Summary:     "Founder or agent pushes a saved link/bookmark to the channel; preserve doc identity for reuse.",
		},
	}
}

// FindTelegramProfile returns a profile by interaction id (normalized).
func FindTelegramProfile(interaction string) (TelegramInteractionProfile, bool) {
	needle := normalizeObservationMetaToken(interaction)
	if needle == "" {
		return TelegramInteractionProfile{}, false
	}
	for _, profile := range TelegramSemanticsProfiles() {
		if normalizeObservationMetaToken(profile.Interaction) == needle {
			return profile, true
		}
	}
	return TelegramInteractionProfile{}, false
}

// ApplyTelegramProfile merges a canonical interaction profile onto a raw
// TelegramObservationInput. Caller-provided fields win; profile fills gaps and
// appends tagging hints so the downstream ingest surface stays deterministic.
func ApplyTelegramProfile(profile TelegramInteractionProfile, input TelegramObservationInput) TelegramObservationInput {
	if strings.TrimSpace(input.Topic) == "" {
		input.Topic = profile.Topic
	}
	if strings.TrimSpace(input.Kind) == "" {
		input.Kind = profile.ContentKind
	}
	if strings.TrimSpace(input.Direction) == "" {
		input.Direction = profile.Direction
	}
	for _, tag := range profile.DefaultTags {
		input.Tags = appendUniqueString(input.Tags, tag)
	}
	for _, ref := range profile.RefHints {
		input.Refs = appendUniqueString(input.Refs, ref)
	}
	return input
}

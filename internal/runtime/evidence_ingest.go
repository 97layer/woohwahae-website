package runtime

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"strings"
	"time"
)

type ExternalObservationInput struct {
	ObservationID string
	SourceChannel string
	Topic         string
	Actor         string
	Confidence    string
	Summary       string
	Excerpt       string
	Refs          []string
	Attributes    map[string]string
	CapturedAt    time.Time
}

type TelegramObservationInput struct {
	ObservationID string
	Actor         string
	Topic         string
	Confidence    string
	Summary       string
	Excerpt       string
	Title         string
	SourceURL     string
	Direction     string
	Kind          string
	MessageID     string
	ChatID        string
	Username      string
	Tags          []string
	Refs          []string
	CapturedAt    time.Time
}

type ContentObservationInput struct {
	ObservationID string
	SourceChannel string
	Actor         string
	Topic         string
	Confidence    string
	Summary       string
	Excerpt       string
	Title         string
	SourceURL     string
	Author        string
	Kind          string
	DocID         string
	Tags          []string
	Refs          []string
	CapturedAt    time.Time
}

func BuildExternalObservation(input ExternalObservationInput) ObservationRecord {
	channel := canonicalObservationChannel(input.SourceChannel)
	topic := strings.TrimSpace(input.Topic)
	if topic == "" {
		topic = "external_capture"
	}
	summary := strings.TrimSpace(input.Summary)
	excerpt := strings.TrimSpace(input.Excerpt)
	if summary == "" {
		summary = externalObservationSummary(excerpt, topic)
	}
	if excerpt == "" {
		excerpt = summary
	}
	refs := append([]string{}, input.Refs...)
	for _, ref := range ObservationMetadataRefs(channel, topic, input.Attributes) {
		refs = appendUniqueString(refs, ref)
	}
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = channel
	}
	return normalizeObservationRecord(ObservationRecord{
		ObservationID:     strings.TrimSpace(input.ObservationID),
		SourceChannel:     channel,
		CapturedAt:        input.CapturedAt,
		Actor:             actor,
		Topic:             topic,
		Refs:              refs,
		Confidence:        strings.TrimSpace(input.Confidence),
		RawExcerpt:        excerpt,
		NormalizedSummary: summary,
	}, actor)
}

func BuildTelegramObservation(input TelegramObservationInput) ObservationRecord {
	kind := normalizeObservationMetaToken(input.Kind)
	if kind == "" {
		kind = "message"
	}
	topic := strings.TrimSpace(input.Topic)
	if topic == "" {
		topic = kind
	}
	attrs := map[string]string{
		"ingest_adapter": "telegram",
		"content_kind":   kind,
	}
	if direction := normalizeObservationMetaToken(input.Direction); direction != "" {
		attrs["interaction_direction"] = direction
	}
	if messageID := normalizeObservationMetaToken(input.MessageID); messageID != "" {
		attrs["telegram_message"] = messageID
	}
	if chatID := normalizeObservationMetaToken(input.ChatID); chatID != "" {
		attrs["telegram_chat"] = chatID
	}
	if username := normalizeObservationMetaToken(input.Username); username != "" {
		attrs["telegram_user"] = username
	}
	if host := evidenceURLHost(input.SourceURL); host != "" {
		attrs["content_host"] = host
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = strings.TrimSpace(input.Title)
	}
	refs := append([]string{}, input.Refs...)
	for _, tag := range input.Tags {
		if normalized := normalizeObservationMetaToken(tag); normalized != "" {
			refs = appendUniqueString(refs, "tag:"+normalized)
		}
	}
	confidence := strings.TrimSpace(input.Confidence)
	if confidence == "" {
		if strings.TrimSpace(input.MessageID) != "" {
			confidence = "high"
		} else {
			confidence = "medium"
		}
	}
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		if username := strings.TrimSpace(input.Username); username != "" {
			actor = username
		} else {
			actor = "telegram"
		}
	}
	return BuildExternalObservation(ExternalObservationInput{
		ObservationID: input.ObservationID,
		SourceChannel: "telegram",
		Topic:         topic,
		Actor:         actor,
		Confidence:    confidence,
		Summary:       summary,
		Excerpt:       strings.TrimSpace(input.Excerpt),
		Refs:          refs,
		Attributes:    attrs,
		CapturedAt:    input.CapturedAt,
	})
}

func BuildContentObservation(input ContentObservationInput) ObservationRecord {
	channel := canonicalObservationChannel(input.SourceChannel)
	if channel == "" {
		channel = "crawler"
	}
	kind := normalizeObservationMetaToken(input.Kind)
	if kind == "" {
		kind = "content_capture"
	}
	topic := strings.TrimSpace(input.Topic)
	if topic == "" {
		topic = kind
	}
	host := evidenceURLHost(input.SourceURL)
	docID := stableContentDocID(strings.TrimSpace(input.DocID), strings.TrimSpace(input.SourceURL))
	attrs := map[string]string{
		"ingest_adapter": "content",
		"content_kind":   kind,
	}
	if host != "" {
		attrs["content_host"] = host
	}
	if author := normalizeObservationMetaToken(input.Author); author != "" {
		attrs["content_author"] = author
	}
	if docID != "" {
		attrs["content_doc"] = docID
	}
	refs := append([]string{}, input.Refs...)
	for _, tag := range input.Tags {
		if normalized := normalizeObservationMetaToken(tag); normalized != "" {
			refs = appendUniqueString(refs, "tag:"+normalized)
		}
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = strings.TrimSpace(input.Title)
	}
	confidence := strings.TrimSpace(input.Confidence)
	if confidence == "" {
		if docID != "" || host != "" {
			confidence = "high"
		} else {
			confidence = "medium"
		}
	}
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = channel
	}
	return BuildExternalObservation(ExternalObservationInput{
		ObservationID: input.ObservationID,
		SourceChannel: channel,
		Topic:         topic,
		Actor:         actor,
		Confidence:    confidence,
		Summary:       summary,
		Excerpt:       strings.TrimSpace(input.Excerpt),
		Refs:          refs,
		Attributes:    attrs,
		CapturedAt:    input.CapturedAt,
	})
}

func TelegramDedupeRef(messageID string) string {
	messageID = normalizeObservationMetaToken(messageID)
	if messageID == "" {
		return ""
	}
	return "telegram_message:" + messageID
}

func ContentDedupeRef(docID string, sourceURL string) string {
	docID = stableContentDocID(docID, sourceURL)
	if docID == "" {
		return ""
	}
	return "content_doc:" + docID
}

func externalObservationSummary(excerpt string, fallback string) string {
	for _, line := range strings.Split(strings.TrimSpace(excerpt), "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		if trimmed == "" {
			continue
		}
		return limitText(trimmed, 180)
	}
	return limitText(strings.TrimSpace(fallback), 180)
}

func evidenceURLHost(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		parsed, err = url.Parse("https://" + strings.TrimPrefix(rawURL, "//"))
		if err != nil {
			return ""
		}
	}
	return normalizeObservationMetaToken(parsed.Hostname())
}

func stableContentDocID(docID string, sourceURL string) string {
	if normalized := normalizeObservationMetaToken(docID); normalized != "" {
		return normalized
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return ""
	}
	return "url_" + externalEvidenceHash(sourceURL)
}

func externalEvidenceHash(parts ...string) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(strings.TrimSpace(part)))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

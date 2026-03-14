package runtime

import (
	"sort"
	"strings"
	"unicode"
)

var nonLinkableObservationRefPrefixes = []string{
	"event:",
	"source:",
	"source_family:",
	"topic:",
	"severity:",
	"artifact_type:",
	"artifact_stem:",
	"ingest_adapter:",
	"interaction_direction:",
	"content_kind:",
	"content_host:",
	"content_author:",
	"content_doc:",
	"telegram_message:",
	"telegram_chat:",
	"telegram_route:",
	"telegram_chat_type:",
	"telegram_user:",
	"threads_creation:",
	"threads_thread:",
	"antigravity_sync:",
	"antigravity_run:",
	"antigravity_artifact:",
	"antigravity_path:",
	"tag:",
	"gemini_sync:",
	"gemini_artifact:",
	"gemini_path:",
}

func canonicalObservationChannel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "":
		return ""
	case "tg":
		return "telegram"
	case "ag", "antigravity-ide", "antigravity_ide":
		return "antigravity"
	case "gem", "gemini", "gemini-cli", "gemini_cli":
		return "gemini"
	case "crawl", "content-crawl", "content_crawl", "web-crawl", "web_crawl":
		return "crawler"
	case "archive", "personal-archive", "personal_archive", "personal-db":
		return "personal_db"
	case "notebook", "notebooklm", "notebook-lm", "notebook_lm":
		return "notebook_lm"
	default:
		return value
	}
}

func ObservationChannelFamily(channel string) string {
	switch canonicalObservationChannel(channel) {
	case "antigravity", "gemini":
		return "agent_workspace"
	case "telegram":
		return "founder_inbox"
	case "personal_db", "notebook_lm":
		return "founder_archive"
	case "crawler":
		return "external_content"
	case "session", "agent_job":
		return "runtime"
	case "terminal", "cockpit", "api":
		return "operator_surface"
	default:
		return "external"
	}
}

func ObservationMetadataRefs(channel string, topic string, attrs map[string]string) []string {
	refs := []string{}
	channel = canonicalObservationChannel(channel)
	if channel != "" {
		refs = appendUniqueString(refs, "source:"+channel)
		family := ObservationChannelFamily(channel)
		if family != "" {
			refs = appendUniqueString(refs, "source_family:"+family)
		}
	}
	if normalizedTopic := normalizeObservationMetaToken(topic); normalizedTopic != "" {
		refs = appendUniqueString(refs, "topic:"+normalizedTopic)
	}
	if len(attrs) == 0 {
		return refs
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		normalizedKey := normalizeObservationMetaToken(key)
		normalizedValue := normalizeObservationMetaToken(attrs[key])
		if normalizedKey == "" || normalizedValue == "" {
			continue
		}
		refs = appendUniqueString(refs, normalizedKey+":"+normalizedValue)
	}
	return refs
}

func isNonLinkableObservationRef(ref string) bool {
	ref = strings.ToLower(strings.TrimSpace(ref))
	if ref == "" {
		return true
	}
	for _, prefix := range nonLinkableObservationRefPrefixes {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}

func normalizeObservationMetaToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			builder.WriteRune(r)
			lastUnderscore = false
		case r == '.' || r == '-' || r == '_' || unicode.IsSpace(r) || r == '/' || r == ':':
			if !lastUnderscore {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(builder.String(), "_")
}

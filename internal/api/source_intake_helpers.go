package api

import (
	"strings"

	"layer-os/internal/runtime"
)

func firstObservationRefValue(refs []string, prefix string) string {
	for _, ref := range refs {
		trimmed := strings.TrimSpace(ref)
		if trimmed == "" {
			continue
		}
		if prefix == "" {
			if !strings.Contains(trimmed, ":") {
				return trimmed
			}
			continue
		}
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func routeDecisionRawValue(raw string, prefix string) string {
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func sourceIntakeRouteLabel(route string) string {
	switch strings.TrimSpace(route) {
	case "97layer":
		return "97layer"
	case "woosunhokr":
		return "우순호"
	case "woohwahae":
		return "우화해"
	default:
		return strings.TrimSpace(route)
	}
}

func sourceIntakeOriginLabel(intakeClass string, feedSource string) string {
	if strings.TrimSpace(feedSource) != "" {
		return "feed sensor"
	}
	switch strings.TrimSpace(intakeClass) {
	case "manual_drop":
		return "founder drop"
	case "authorized_connector":
		return "authorized connector"
	case "public_collector":
		return "public collector"
	default:
		return strings.TrimSpace(intakeClass)
	}
}

func sourceIntakeFeedLabel(feedSource string) string {
	trimmed := strings.TrimSpace(feedSource)
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	return trimmed
}

func firstPrepObservationSourceID(item runtime.ObservationRecord) string {
	if value := prepObservationValue(item.RawExcerpt, "sources="); value != "" {
		for _, sourceID := range splitCSVValues(value) {
			if trimmed := strings.TrimSpace(sourceID); trimmed != "" && !strings.Contains(trimmed, ":") {
				return trimmed
			}
		}
	}
	for _, ref := range item.Refs {
		trimmed := strings.TrimSpace(ref)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "source_draft_seed:") || strings.Contains(trimmed, ":") {
			continue
		}
		return trimmed
	}
	return ""
}

func prepObservationValue(raw string, prefix string) string {
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func splitCSVValues(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "none") {
		return []string{}
	}
	parts := strings.Split(trimmed, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func prepChannelLabel(channel string) string {
	switch strings.TrimSpace(strings.ToLower(channel)) {
	case "threads":
		return "Threads"
	default:
		return strings.TrimSpace(channel)
	}
}

func nonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

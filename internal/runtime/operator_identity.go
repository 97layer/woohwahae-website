package runtime

import (
	"os"
	"strings"
)

const neutralDefaultActor = "system"

func normalizeActor(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "layer-osctl", "layer-osd", "":
		return ""
	default:
		return value
	}
}

func DefaultActor() string {
	if actor := normalizeActor(os.Getenv("LAYER_OS_DEFAULT_ACTOR")); actor != "" {
		return actor
	}
	return neutralDefaultActor
}

func ResolveActor(candidates ...string) string {
	for _, candidate := range candidates {
		if actor := normalizeActor(candidate); actor != "" {
			return actor
		}
	}
	return DefaultActor()
}

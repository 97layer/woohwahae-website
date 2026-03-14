package runtime

import (
	"strings"
	"unicode"
)

func reviewItemRefs(item ReviewRoomItem) []string {
	item = normalizeReviewRoomItem(item)
	refs := []string{}
	appendRef := func(value string) {
		trimmed := strings.TrimSpace(value)
		if !validReviewGateRef(trimmed) {
			return
		}
		for _, existing := range refs {
			if existing == trimmed {
				return
			}
		}
		refs = append(refs, trimmed)
	}
	if item.Ref != nil {
		appendRef(*item.Ref)
	}
	for _, ref := range item.PatternRefs {
		appendRef(ref)
	}
	if ref, ok := leadingReviewTextRef(item.Text); ok {
		appendRef(ref)
	}
	return refs
}

func validReviewGateRef(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !strings.Contains(trimmed, "_") {
		return false
	}
	for idx, r := range trimmed {
		if idx == 0 && !unicode.IsLetter(r) {
			return false
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			continue
		}
		return false
	}
	return true
}

func leadingReviewTextRef(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "[") {
		return "", false
	}
	end := strings.Index(trimmed, "]")
	if end <= 1 {
		return "", false
	}
	candidate := strings.TrimSpace(trimmed[1:end])
	if !validReviewGateRef(candidate) {
		return "", false
	}
	return candidate, true
}

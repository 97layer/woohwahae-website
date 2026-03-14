package runtime

import (
	"sort"
	"strings"
	"unicode"
)

var reviewRoomConflictStopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "before": {}, "for": {}, "from": {}, "into": {}, "item": {}, "its": {}, "later": {},
	"manual": {}, "not": {}, "open": {}, "operator": {}, "review": {}, "room": {}, "still": {}, "that": {}, "the": {}, "this": {},
	"with": {}, "without": {},
}

func annotateReviewRoomContradictions(room ReviewRoom, section string, item ReviewRoomItem) ReviewRoomItem {
	if normalizeReviewSection(section) != "open" {
		return normalizeReviewRoomItem(item)
	}
	item = normalizeReviewRoomItem(item)
	matches := reviewRoomDecisionContradictions(room.Accepted, item)
	if len(matches) == 0 {
		return item
	}
	item.Contradictions = normalizeReviewRoomEvidence(append(item.Contradictions, matches...))
	if item.Contradiction == nil && len(item.Contradictions) == 1 {
		item.Contradiction = normalizeReviewRoomOptionalString(&item.Contradictions[0])
	}
	return normalizeReviewRoomItem(item)
}

func reviewRoomDecisionContradictions(decisions []ReviewRoomItem, item ReviewRoomItem) []string {
	item = normalizeReviewRoomItem(item)
	agendaKeywords := reviewRoomConflictKeywords(item)
	if len(agendaKeywords) == 0 {
		return nil
	}
	matches := []string{}
	for _, decision := range decisions {
		decision = normalizeReviewRoomItem(decision)
		if strings.TrimSpace(decision.Text) == "" {
			continue
		}
		if len(intersectStrings(agendaKeywords, reviewRoomConflictKeywords(decision))) == 0 {
			continue
		}
		matches = append(matches, decision.Text)
	}
	return normalizeReviewRoomEvidence(matches)
}

func reviewRoomConflictKeywords(item ReviewRoomItem) []string {
	parts := []string{item.Text, item.Why}
	if item.WhyUnresolved != nil {
		parts = append(parts, strings.TrimSpace(*item.WhyUnresolved))
	}
	if item.Resolution != nil {
		parts = append(parts, strings.TrimSpace(item.Resolution.Reason))
	}
	return reviewRoomKeywordSet(strings.Join(normalizeReviewRoomEvidence(parts), " "))
}

func reviewRoomKeywordSet(text string) []string {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return nil
	}
	seen := map[string]struct{}{}
	items := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, skip := reviewRoomConflictStopwords[item]; skip {
			continue
		}
		if utf8Len(item) < 2 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func utf8Len(value string) int {
	count := 0
	for range value {
		count++
	}
	return count
}

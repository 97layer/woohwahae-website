package runtime

import (
	"strings"
	"time"
)

type reviewRoomDiskLegacy struct {
	Source    string     `json:"source"`
	Accepted  []string   `json:"accepted"`
	Open      []string   `json:"open"`
	Deferred  []string   `json:"deferred"`
	Issues    []string   `json:"issues"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func normalizeReviewRoomItem(item ReviewRoomItem) ReviewRoomItem {
	item.Text = strings.TrimSpace(item.Text)
	item.Kind = strings.TrimSpace(item.Kind)
	item.Severity = strings.TrimSpace(item.Severity)
	item.Source = strings.TrimSpace(item.Source)
	item.Why = strings.TrimSpace(item.Why)
	item.Ref = normalizeReviewRoomOptionalString(item.Ref)
	item.WhyUnresolved = normalizeReviewRoomOptionalString(item.WhyUnresolved)
	item.Contradiction = normalizeReviewRoomOptionalString(item.Contradiction)
	item.Contradictions = normalizeReviewRoomEvidence(item.Contradictions)
	if len(item.Contradictions) == 0 && item.Contradiction != nil {
		item.Contradictions = []string{strings.TrimSpace(*item.Contradiction)}
	}
	if item.Contradiction == nil && len(item.Contradictions) == 1 {
		item.Contradiction = normalizeReviewRoomOptionalString(&item.Contradictions[0])
	}
	item.PatternRefs = normalizeReviewRoomEvidence(item.PatternRefs)
	item.Rationale = normalizeReviewRoomRationale(item.Rationale)
	item.Resolution = normalizeReviewRoomResolution(item.Resolution)
	item.Evidence = normalizeReviewRoomEvidence(item.Evidence)
	if item.Kind == "" {
		item.Kind = "agenda"
	}
	if item.Severity == "" {
		item.Severity = "medium"
	}
	if item.Source == "" {
		item.Source = "manual"
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = zeroSafeNow()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	return item
}

func normalizeReviewRoomOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeReviewRoomRationale(rationale *ReviewRoomRationale) *ReviewRoomRationale {
	if rationale == nil {
		return nil
	}
	out := &ReviewRoomRationale{
		Trigger: strings.TrimSpace(rationale.Trigger),
		Reason:  strings.TrimSpace(rationale.Reason),
		Rule:    strings.TrimSpace(rationale.Rule),
	}
	if out.Trigger == "" && out.Reason == "" && out.Rule == "" {
		return nil
	}
	return out
}

func normalizeReviewRoomResolution(resolution *ReviewRoomResolution) *ReviewRoomResolution {
	if resolution == nil {
		return nil
	}
	action := normalizeReviewAction(resolution.Action)
	if action == "" {
		action = strings.TrimSpace(strings.ToLower(resolution.Action))
	}
	out := &ReviewRoomResolution{
		Action:     action,
		Reason:     strings.TrimSpace(resolution.Reason),
		Rule:       strings.TrimSpace(resolution.Rule),
		Evidence:   normalizeReviewRoomEvidence(resolution.Evidence),
		ResolvedAt: resolution.ResolvedAt,
	}
	if out.Action == "" && out.Reason == "" && out.Rule == "" && len(out.Evidence) == 0 {
		return nil
	}
	return out
}

func normalizeReviewRoomEvidence(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeReviewRoomItems(items []ReviewRoomItem) []ReviewRoomItem {
	if items == nil {
		return []ReviewRoomItem{}
	}
	out := make([]ReviewRoomItem, len(items))
	for i, item := range items {
		out[i] = cloneReviewRoomItem(item)
	}
	return out
}

func cloneReviewRoomItem(item ReviewRoomItem) ReviewRoomItem {
	out := normalizeReviewRoomItem(item)
	if out.Ref != nil {
		value := *out.Ref
		out.Ref = &value
	}
	if out.WhyUnresolved != nil {
		value := *out.WhyUnresolved
		out.WhyUnresolved = &value
	}
	if out.Contradiction != nil {
		value := *out.Contradiction
		out.Contradiction = &value
	}
	if out.Contradictions != nil {
		out.Contradictions = append([]string{}, out.Contradictions...)
	}
	if out.PatternRefs != nil {
		out.PatternRefs = append([]string{}, out.PatternRefs...)
	}
	if out.Rationale != nil {
		rationale := *out.Rationale
		out.Rationale = &rationale
	}
	if out.Resolution != nil {
		resolution := *out.Resolution
		if resolution.Evidence != nil {
			resolution.Evidence = append([]string{}, resolution.Evidence...)
		}
		out.Resolution = &resolution
	}
	if out.Evidence != nil {
		out.Evidence = append([]string{}, out.Evidence...)
	}
	return out
}

func reviewRoomLegacyItems(items []string, source string, updatedAt *time.Time) []ReviewRoomItem {
	out := make([]ReviewRoomItem, 0, len(items))
	for _, item := range items {
		entry := ReviewRoomItem{
			Text:     item,
			Kind:     "legacy_note",
			Severity: "medium",
			Source:   source,
		}
		if updatedAt != nil && !updatedAt.IsZero() {
			entry.CreatedAt = *updatedAt
			entry.UpdatedAt = *updatedAt
		}
		out = append(out, normalizeReviewRoomItem(entry))
	}
	return out
}

func newStructuredReviewRoomItem(text string, kind string, severity string, source string, ref *string, rationale *ReviewRoomRationale, evidence []string) ReviewRoomItem {
	item := ReviewRoomItem{
		Text:      text,
		Kind:      kind,
		Severity:  severity,
		Source:    source,
		Ref:       ref,
		Rationale: rationale,
		Evidence:  evidence,
	}
	return normalizeReviewRoomItem(item)
}

func stampDeferredWhyUnresolved(item ReviewRoomItem) ReviewRoomItem {
	item = normalizeReviewRoomItem(item)
	if item.WhyUnresolved != nil {
		return item
	}
	if item.Resolution != nil {
		item.WhyUnresolved = normalizeReviewRoomOptionalString(&item.Resolution.Reason)
	}
	return normalizeReviewRoomItem(item)
}

func newManualReviewRoomItem(text string, kind string, severity string, source string, ref *string) ReviewRoomItem {
	return newStructuredReviewRoomItem(text, kind, severity, source, ref, nil, nil)
}

func newSignalReviewRoomItem(text string, trigger string, ref *string, reason string, rule string, evidence []string) ReviewRoomItem {
	severity := "high"
	if strings.Contains(trigger, "verification") {
		severity = "medium"
	}
	return newStructuredReviewRoomItem(text, "runtime_signal", severity, trigger, ref, &ReviewRoomRationale{
		Trigger: trigger,
		Reason:  reason,
		Rule:    rule,
	}, evidence)
}

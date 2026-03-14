package runtime

import (
	"fmt"
	"strings"
)

type SourceIntakeObservationService interface {
	ListObservations(query ObservationQuery) []ObservationRecord
	CreateObservation(item ObservationRecord) (ObservationRecord, error)
}

type SourceIntakeContentInput struct {
	SourceObservation ObservationRecord
	IntakeClass       string
	PolicyColor       string
	Title             string
	URL               string
	Excerpt           string
	FounderNote       string
	DomainTags        []string
	WorldviewTags     []string
	SuggestedRoutes   []string
	PriorityScore     int
	Disposition       string
	DispositionNote   string
}

func BuildSourceIntakeRecordFromContent(input SourceIntakeContentInput) SourceIntakeRecord {
	observation := normalizeObservationRecord(input.SourceObservation, input.SourceObservation.Actor)
	record := SourceIntakeRecord{
		IntakeClass:     strings.TrimSpace(input.IntakeClass),
		PolicyColor:     strings.TrimSpace(input.PolicyColor),
		Title:           strings.TrimSpace(input.Title),
		URL:             strings.TrimSpace(input.URL),
		Excerpt:         strings.TrimSpace(input.Excerpt),
		FounderNote:     strings.TrimSpace(input.FounderNote),
		PriorityScore:   input.PriorityScore,
		Disposition:     strings.TrimSpace(input.Disposition),
		DispositionNote: strings.TrimSpace(input.DispositionNote),
	}
	if record.IntakeClass == "" {
		record.IntakeClass = sourceIntakeClassForObservation(observation)
	}
	record.DomainTags = sourceIntakeMergedTags(input.DomainTags, observation.Refs, "domain:", true)
	record.WorldviewTags = sourceIntakeMergedTags(input.WorldviewTags, observation.Refs, "worldview:", false)
	record.SuggestedRoutes = sourceIntakeMergedRoutes(input.SuggestedRoutes, observation.Refs)
	if record.PolicyColor == "" {
		record.PolicyColor = defaultSourceIntakePolicyForContent(observation)
	}
	if record.FounderNote == "" {
		record.FounderNote = defaultSourceIntakeFounderNoteForContent(observation)
	}
	if record.Title == "" {
		record.Title = observation.NormalizedSummary
	}
	if record.Excerpt == "" {
		record.Excerpt = observation.RawExcerpt
	}
	return FinalizeSourceIntakeRecord(record)
}

func defaultSourceIntakePolicyForContent(observation ObservationRecord) string {
	if sourceIntakeClassForObservation(observation) != "public_collector" {
		return "green"
	}
	hints := explicitSourceIntakeRouteHints(observation.Refs)
	if len(hints) == 1 {
		return "green"
	}
	return "yellow"
}

func defaultSourceIntakeFounderNoteForContent(observation ObservationRecord) string {
	if sourceIntakeClassForObservation(observation) != "public_collector" {
		return ""
	}
	hints := explicitSourceIntakeRouteHints(observation.Refs)
	switch len(hints) {
	case 0:
		return "sensor가 source를 주웠지만 route cue가 비어 있어 founder 분류가 필요합니다."
	case 1:
		return ""
	default:
		return "sensor가 여러 route cue를 함께 가져와 founder route 결정이 필요합니다."
	}
}

func EnsureSourceIntakeObservation(service SourceIntakeObservationService, sourceObservation ObservationRecord, record SourceIntakeRecord) (ObservationRecord, bool, error) {
	sourceObservation = normalizeObservationRecord(sourceObservation, sourceObservation.Actor)
	sourceObservationID := strings.TrimSpace(sourceObservation.ObservationID)
	if sourceObservationID == "" {
		return ObservationRecord{}, false, fmt.Errorf("source observation is required")
	}

	if existing, ok := findExistingSourceIntakeObservation(service, sourceObservationID, sourceObservation.Refs); ok {
		return existing, false, nil
	}

	record = NormalizeSourceIntakeRecord(record)
	refs := []string{"source_observation:" + sourceObservationID, sourceObservationID}
	for _, ref := range sourceObservation.Refs {
		refs = appendUniqueString(refs, ref)
	}
	for _, ref := range SourceIntakeRefs(record) {
		refs = appendUniqueString(refs, ref)
	}

	created, err := service.CreateObservation(ObservationRecord{
		Topic:             SourceIntakeTopic,
		SourceChannel:     sourceObservation.SourceChannel,
		Actor:             sourceObservation.Actor,
		Refs:              refs,
		Confidence:        sourceObservation.Confidence,
		RawExcerpt:        BuildSourceIntakeRawExcerpt(record),
		NormalizedSummary: SourceIntakeSummary(record),
		CapturedAt:        sourceObservation.CapturedAt,
	})
	if err != nil {
		return ObservationRecord{}, false, err
	}
	return created, true, nil
}

func sourceIntakeClassForObservation(item ObservationRecord) string {
	switch normalizeObservationChannel(item.SourceChannel) {
	case "crawler", "web", "feed":
		return "public_collector"
	default:
		return "authorized_connector"
	}
}

func sourceIntakeMergedTags(explicit []string, refs []string, prefix string, includeTagFallback bool) []string {
	items := normalizeSourceIntakeTags(explicit)
	for _, ref := range refs {
		if value, ok := sourceIntakeRefValue(ref, prefix); ok {
			items = appendUniqueString(items, value)
		}
	}
	if includeTagFallback && len(items) == 0 {
		for _, ref := range refs {
			if value, ok := sourceIntakeRefValue(ref, "tag:"); ok {
				items = appendUniqueString(items, value)
			}
		}
	}
	return normalizeSourceIntakeTags(items)
}

func sourceIntakeMergedRoutes(explicit []string, refs []string) []string {
	items := []string{}
	for _, item := range explicit {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" {
			continue
		}
		items = appendUniqueString(items, value)
	}
	for _, ref := range refs {
		if value, ok := sourceIntakeRefValue(ref, "route:"); ok {
			items = appendUniqueString(items, value)
		}
	}
	return normalizeSourceIntakeRoutes(items)
}

func sourceIntakeRefValue(ref string, prefix string) (string, bool) {
	ref = strings.TrimSpace(ref)
	prefix = strings.TrimSpace(prefix)
	if ref == "" || prefix == "" || !strings.HasPrefix(ref, prefix) {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(ref, prefix))
	if value == "" {
		return "", false
	}
	return value, true
}

func findExistingSourceIntakeObservation(service SourceIntakeObservationService, sourceObservationID string, sourceRefs []string) (ObservationRecord, bool) {
	for _, ref := range sourceIntakeLookupRefs(sourceObservationID, sourceRefs) {
		items := service.ListObservations(ObservationQuery{Ref: ref, Limit: 16})
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Topic)) != SourceIntakeTopic {
				continue
			}
			return item, true
		}
	}
	return ObservationRecord{}, false
}

func sourceIntakeLookupRefs(sourceObservationID string, sourceRefs []string) []string {
	out := []string{}
	if trimmed := strings.TrimSpace(sourceObservationID); trimmed != "" {
		out = appendUniqueString(out, "source_observation:"+trimmed)
		out = appendUniqueString(out, trimmed)
	}
	for _, ref := range sourceRefs {
		trimmed := strings.TrimSpace(ref)
		if strings.HasPrefix(trimmed, "content_doc:") {
			out = appendUniqueString(out, trimmed)
		}
	}
	return out
}

package api

import (
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

func buildSourceIntakeView(service *runtime.Service) sourceIntakeView {
	items := service.ListObservations(runtime.ObservationQuery{Topic: runtime.SourceIntakeTopic, Limit: 16})
	draftObservations := service.ListObservations(runtime.ObservationQuery{Topic: runtime.SourceDraftSeedTopic, Limit: 16})
	routeObservations := service.ListObservations(runtime.ObservationQuery{Topic: runtime.IntakeRouteDecisionTopic, Limit: 16})
	prepObservations := service.ListObservations(runtime.ObservationQuery{Topic: cockpitBrandPrepTopic, Limit: 16})

	drafts := normalizeSourceIntakeDraftSeedViews(draftObservations)
	routeDecisions := normalizeSourceIntakeRouteDecisionViews(routeObservations)
	prepLanes := normalizeSourceIntakePrepLaneViews(prepObservations)

	draftsBySource := groupSourceIntakeDraftsBySource(drafts)
	routeBySource := groupSourceIntakeRouteDecisionsBySource(routeDecisions)
	prepBySource := groupSourceIntakePrepLanesBySource(prepLanes)

	normalizedItems := make([]sourceIntakeItemView, 0, len(items))
	feedCount := 0
	draftSeedCount := 0
	prepCount := 0
	reviewCount := 0
	prepReadyCount := 0
	autoRoutedCount := 0
	for _, item := range items {
		record := runtime.ParseSourceIntakeRawExcerpt(item.RawExcerpt)
		feedSource := firstObservationRefValue(item.Refs, "feed_source:")
		if feedSource != "" {
			feedCount++
		}
		normalized := sourceIntakeItemView{
			ObservationID:    strings.TrimSpace(item.ObservationID),
			CapturedAt:       item.CapturedAt,
			Summary:          item.NormalizedSummary,
			IntakeClass:      strings.TrimSpace(record.IntakeClass),
			IntakeClassLabel: sourceIntakeIntakeClassLabel(record.IntakeClass),
			PolicyColor:      strings.TrimSpace(record.PolicyColor),
			PolicyColorLabel: sourceIntakePolicyColorLabel(record.PolicyColor),
			Title:            strings.TrimSpace(record.Title),
			URL:              strings.TrimSpace(record.URL),
			Excerpt:          strings.TrimSpace(record.Excerpt),
			FounderNote:      strings.TrimSpace(record.FounderNote),
			PriorityScore:    record.PriorityScore,
			Disposition:      strings.TrimSpace(record.Disposition),
			DispositionLabel: sourceIntakeDispositionLabel(record.Disposition),
			DispositionNote:  strings.TrimSpace(record.DispositionNote),
			DomainTags:       filterNonEmptyValues(record.DomainTags, "none"),
			WorldviewTags:    filterNonEmptyValues(record.WorldviewTags, "none"),
			SuggestedRoutes:  append([]string{}, record.SuggestedRoutes...),
			Refs:             append([]string{}, item.Refs...),
			FeedSource:       feedSource,
			FeedKind:         firstObservationRefValue(item.Refs, "feed_kind:"),
			OriginLabel:      sourceIntakeOriginLabel(record.IntakeClass, feedSource),
			FeedSourceLabel:  sourceIntakeFeedLabel(feedSource),
		}
		if seeds := draftsBySource[normalized.ObservationID]; len(seeds) > 0 {
			normalized.DraftSeeds = seeds
			seed := seeds[0]
			normalized.DraftSeed = &seed
			draftSeedCount += len(seeds)
		}
		if decisions := routeBySource[normalized.ObservationID]; len(decisions) > 0 {
			normalized.RouteDecisions = decisions
			decision := decisions[0]
			normalized.RouteDecision = &decision
			if decision.RouteSource == "feed_sensor" {
				autoRoutedCount++
			}
		}
		if preps := prepBySource[normalized.ObservationID]; len(preps) > 0 {
			normalized.PrepLanes = preps
			prep := preps[0]
			normalized.PrepLane = &prep
			prepCount++
		}
		if normalized.Disposition == "review" {
			reviewCount++
		}
		if normalized.Disposition == "prep" && normalized.PrepLane == nil {
			prepReadyCount++
		}
		normalizedItems = append(normalizedItems, normalized)
	}

	prioritizedItems := sortSourceIntakeViewItems(normalizedItems)
	actionItems := make([]sourceIntakeItemView, 0, len(prioritizedItems))
	quietItems := make([]sourceIntakeItemView, 0, len(prioritizedItems))
	for _, item := range prioritizedItems {
		if sourceIntakeViewAttentionRank(item) > 600 {
			actionItems = append(actionItems, item)
		} else {
			quietItems = append(quietItems, item)
		}
	}

	return sourceIntakeView{
		GeneratedAt:      time.Now().UTC(),
		RuntimeAvailable: true,
		SummaryNote:      summarizeSourceIntakeViewSummaryNote(actionItems, quietItems),
		SummaryMeta:      summarizeSourceIntakeViewSummaryMeta(actionItems, quietItems, drafts, prepLanes),
		QuietNote:        summarizeSourceIntakeQuietNote(quietItems),
		Defaults: sourceIntakeFormDefaults{
			IntakeClass:     "manual_drop",
			PolicyColor:     "green",
			Title:           "",
			URL:             "",
			Excerpt:         "",
			FounderNote:     "",
			PriorityScore:   0,
			Disposition:     "observe",
			DispositionNote: "",
			DomainTags:      []string{},
			WorldviewTags:   []string{},
			SuggestedRoutes: []string{"97layer"},
		},
		RouteOptions: []sourceIntakeRouteOption{
			{RouteID: "97layer", Label: "97layer", Cue: "raw intake / broad source hub"},
			{RouteID: "woosunhokr", Label: "우순호", Cue: "beauty craft / refined translation"},
			{RouteID: "woohwahae", Label: "우화해", Cue: "slow magazine / polished public shell"},
		},
		IntakeClasses: []sourceIntakeOption{
			{Value: "manual_drop", Label: "manual drop"},
			{Value: "authorized_connector", Label: "authorized connector"},
			{Value: "public_collector", Label: "public collector"},
		},
		PolicyColors: []sourceIntakeOption{
			{Value: "green", Label: "green"},
			{Value: "yellow", Label: "yellow"},
			{Value: "red", Label: "red"},
		},
		Items:           sourceIntakeSliceLimit(actionItems, 8),
		ActionCount:     len(actionItems),
		QuietCount:      len(quietItems),
		QuietItems:      sourceIntakeSliceLimit(quietItems, 4),
		Drafts:          sourceIntakeSliceLimit(drafts, 8),
		RouteDecisions:  sourceIntakeSliceLimit(routeDecisions, 8),
		PrepLanes:       sourceIntakeSliceLimit(prepLanes, 8),
		Attention:       summarizeSourceIntakeViewAttention(prioritizedItems),
		RecentCount:     len(prioritizedItems),
		FeedCount:       feedCount,
		DraftSeedCount:  draftSeedCount,
		PrepCount:       prepCount,
		ReviewCount:     reviewCount,
		PrepReadyCount:  prepReadyCount,
		AutoRoutedCount: autoRoutedCount,
	}
}

func summarizeSourceIntakeViewSummaryNote(actionItems []sourceIntakeItemView, quietItems []sourceIntakeItemView) string {
	if len(actionItems) > 0 {
		return "review, prep, draft/prep lane이 열린 intake만 먼저 보여줍니다."
	}
	if len(quietItems) > 0 {
		return "급한 intake는 없고 quiet candidate만 축적 중입니다."
	}
	return "링크나 텍스트를 하나 저장하면 source intake가 여기부터 쌓입니다."
}

func summarizeSourceIntakeViewSummaryMeta(actionItems []sourceIntakeItemView, quietItems []sourceIntakeItemView, drafts []sourceIntakeDraftSeedView, prepLanes []sourceIntakePrepLaneView) string {
	return "action " + strconv.Itoa(len(actionItems)) +
		" · quiet " + strconv.Itoa(len(quietItems)) +
		" · drafts " + strconv.Itoa(len(drafts)) +
		" · prep " + strconv.Itoa(len(prepLanes))
}

func summarizeSourceIntakeQuietNote(quietItems []sourceIntakeItemView) string {
	if len(quietItems) == 0 {
		return "quiet candidate는 아직 없습니다."
	}
	return "자동 수집됐지만 아직 founder attention을 바로 요구하지 않는 후보 " + strconv.Itoa(len(quietItems)) + "개입니다."
}

func buildSourceIntakeCreateResult(service *runtime.Service, created runtime.ObservationRecord) sourceIntakeCreateResult {
	view := buildSourceIntakeView(service)
	item, ok := findSourceIntakeItemView(view, created.ObservationID)
	if !ok {
		item = normalizeSourceIntakeItemView(created)
	}
	return sourceIntakeCreateResult{
		Created:    item,
		NextAction: summarizeSourceIntakeViewAttention([]sourceIntakeItemView{item}),
	}
}

package api

import (
	"sort"
	"strings"

	"layer-os/internal/runtime"
)

func normalizeSourceIntakeDraftSeedViews(items []runtime.ObservationRecord) []sourceIntakeDraftSeedView {
	out := make([]sourceIntakeDraftSeedView, 0, len(items))
	for _, item := range items {
		record := runtime.ParseSourceDraftSeedRawExcerpt(item.RawExcerpt)
		out = append(out, sourceIntakeDraftSeedView{
			ObservationID:            item.ObservationID,
			CapturedAt:               item.CapturedAt,
			Summary:                  item.NormalizedSummary,
			TargetAccount:            record.TargetAccount,
			TargetAccountLabel:       sourceIntakeRouteLabel(record.TargetAccount),
			TargetTone:               record.TargetToneLevel,
			Title:                    record.Title,
			SourceObservationID:      record.SourceObservationID,
			RouteDecisionID:          record.RouteDecisionObservationID,
			ParentDraftObservationID: record.ParentDraftObservationID,
			RevisionNote:             record.RevisionNote,
			SourceTitle:              record.SourceTitle,
			SourceURL:                record.SourceURL,
			FounderNote:              record.FounderNote,
			DomainTags:               append([]string{}, record.DomainTags...),
			WorldviewTags:            append([]string{}, record.WorldviewTags...),
			Draft:                    record.Draft,
			Preview:                  sourceIntakeLimitText(strings.Join(strings.Fields(record.Draft), " "), 180),
			Refs:                     append([]string{}, item.Refs...),
		})
	}
	sort.SliceStable(out, func(left, right int) bool {
		if !out[left].CapturedAt.Equal(out[right].CapturedAt) {
			return out[left].CapturedAt.After(out[right].CapturedAt)
		}
		return out[left].ObservationID > out[right].ObservationID
	})
	return out
}

func normalizeSourceIntakeItemView(item runtime.ObservationRecord) sourceIntakeItemView {
	record := runtime.ParseSourceIntakeRawExcerpt(item.RawExcerpt)
	feedSource := firstObservationRefValue(item.Refs, "feed_source:")
	return sourceIntakeItemView{
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
}

func findSourceIntakeItemView(view sourceIntakeView, observationID string) (sourceIntakeItemView, bool) {
	for _, item := range view.Items {
		if item.ObservationID == observationID {
			return item, true
		}
	}
	for _, item := range view.QuietItems {
		if item.ObservationID == observationID {
			return item, true
		}
	}
	return sourceIntakeItemView{}, false
}

func normalizeSourceIntakeRouteDecisionViews(items []runtime.ObservationRecord) []sourceIntakeRouteDecisionView {
	out := make([]sourceIntakeRouteDecisionView, 0, len(items))
	for _, item := range items {
		sourceObservationID := firstObservationRefValue(item.Refs, "")
		if sourceObservationID == "" {
			sourceObservationID = routeDecisionRawValue(item.RawExcerpt, "source_observation_id=")
		}
		decision := routeDecisionRawValue(item.RawExcerpt, "decision=")
		title := routeDecisionRawValue(item.RawExcerpt, "title=")
		routeSource := routeDecisionRawValue(item.RawExcerpt, "route_source=")
		out = append(out, sourceIntakeRouteDecisionView{
			ObservationID:       item.ObservationID,
			CapturedAt:          item.CapturedAt,
			Summary:             item.NormalizedSummary,
			SourceObservationID: sourceObservationID,
			Decision:            decision,
			DecisionLabel:       sourceIntakeRouteLabel(decision),
			Title:               title,
			RouteSource:         routeSource,
			RouteSourceLabel:    sourceIntakeRouteSourceLabel(routeSource),
			Refs:                append([]string{}, item.Refs...),
		})
	}
	sort.SliceStable(out, func(left, right int) bool {
		if !out[left].CapturedAt.Equal(out[right].CapturedAt) {
			return out[left].CapturedAt.After(out[right].CapturedAt)
		}
		return out[left].ObservationID > out[right].ObservationID
	})
	return out
}

func normalizeSourceIntakePrepLaneViews(items []runtime.ObservationRecord) []sourceIntakePrepLaneView {
	out := make([]sourceIntakePrepLaneView, 0, len(items))
	for _, item := range items {
		channel := prepObservationValue(item.RawExcerpt, "channel=")
		targetAccount := prepObservationValue(item.RawExcerpt, "target_account=")
		title := prepObservationValue(item.RawExcerpt, "title=")
		bodyPreview := prepObservationBodyPreview(item.RawExcerpt)
		out = append(out, sourceIntakePrepLaneView{
			ObservationID:       item.ObservationID,
			CapturedAt:          item.CapturedAt,
			Summary:             item.NormalizedSummary,
			Channel:             strings.TrimSpace(channel),
			ChannelLabel:        prepChannelLabel(channel),
			TargetAccount:       strings.TrimSpace(targetAccount),
			TargetAccountLabel:  sourceIntakeRouteLabel(targetAccount),
			Title:               nonEmptyString(title, item.NormalizedSummary),
			BodyPreview:         bodyPreview,
			SourceObservationID: firstPrepObservationSourceID(item),
			DraftObservationID:  firstObservationRefValue(item.Refs, "source_draft_seed:"),
			ApprovalID:          firstApprovalID(item.Refs),
			FlowID:              firstFlowID(item.Refs),
			Refs:                append([]string{}, item.Refs...),
		})
	}
	sort.SliceStable(out, func(left, right int) bool {
		if !out[left].CapturedAt.Equal(out[right].CapturedAt) {
			return out[left].CapturedAt.After(out[right].CapturedAt)
		}
		return out[left].ObservationID > out[right].ObservationID
	})
	return out
}

func groupSourceIntakeDraftsBySource(items []sourceIntakeDraftSeedView) map[string][]sourceIntakeDraftSeedView {
	grouped := map[string][]sourceIntakeDraftSeedView{}
	seen := map[string]bool{}
	for _, item := range items {
		sourceObservationID := strings.TrimSpace(item.SourceObservationID)
		targetAccount := strings.TrimSpace(item.TargetAccount)
		if sourceObservationID == "" || targetAccount == "" {
			continue
		}
		key := sourceObservationID + ":" + targetAccount
		if seen[key] {
			continue
		}
		seen[key] = true
		grouped[sourceObservationID] = append(grouped[sourceObservationID], item)
	}
	for sourceID := range grouped {
		sort.SliceStable(grouped[sourceID], func(left, right int) bool {
			leftRank := sourceIntakeDraftAccountRank(grouped[sourceID][left].TargetAccount)
			rightRank := sourceIntakeDraftAccountRank(grouped[sourceID][right].TargetAccount)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
			if !grouped[sourceID][left].CapturedAt.Equal(grouped[sourceID][right].CapturedAt) {
				return grouped[sourceID][left].CapturedAt.After(grouped[sourceID][right].CapturedAt)
			}
			return grouped[sourceID][left].ObservationID > grouped[sourceID][right].ObservationID
		})
	}
	return grouped
}

func groupSourceIntakeRouteDecisionsBySource(items []sourceIntakeRouteDecisionView) map[string][]sourceIntakeRouteDecisionView {
	grouped := map[string][]sourceIntakeRouteDecisionView{}
	for _, item := range items {
		sourceObservationID := strings.TrimSpace(item.SourceObservationID)
		if sourceObservationID == "" {
			continue
		}
		grouped[sourceObservationID] = append(grouped[sourceObservationID], item)
	}
	return grouped
}

func groupSourceIntakePrepLanesBySource(items []sourceIntakePrepLaneView) map[string][]sourceIntakePrepLaneView {
	grouped := map[string][]sourceIntakePrepLaneView{}
	seen := map[string]bool{}
	for _, item := range items {
		sourceObservationID := strings.TrimSpace(item.SourceObservationID)
		targetAccount := strings.TrimSpace(item.TargetAccount)
		channel := strings.TrimSpace(item.Channel)
		if sourceObservationID == "" || targetAccount == "" || channel == "" {
			continue
		}
		key := sourceObservationID + ":" + targetAccount + ":" + channel
		if seen[key] {
			continue
		}
		seen[key] = true
		grouped[sourceObservationID] = append(grouped[sourceObservationID], item)
	}
	for sourceID := range grouped {
		sort.SliceStable(grouped[sourceID], func(left, right int) bool {
			leftRank := sourceIntakeDraftAccountRank(grouped[sourceID][left].TargetAccount)
			rightRank := sourceIntakeDraftAccountRank(grouped[sourceID][right].TargetAccount)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
			if !grouped[sourceID][left].CapturedAt.Equal(grouped[sourceID][right].CapturedAt) {
				return grouped[sourceID][left].CapturedAt.After(grouped[sourceID][right].CapturedAt)
			}
			return grouped[sourceID][left].ObservationID > grouped[sourceID][right].ObservationID
		})
	}
	return grouped
}

func sourceIntakeDraftAccountRank(account string) int {
	switch strings.TrimSpace(account) {
	case "97layer":
		return 0
	case "woosunhokr":
		return 1
	case "woohwahae":
		return 2
	default:
		return 9
	}
}

func filterNonEmptyValues(items []string, skip string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || trimmed == skip {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func firstSourceID(items []string) string {
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func prepObservationBodyPreview(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	bodyStart := -1
	for index, rawLine := range lines {
		if strings.TrimSpace(rawLine) == "draft:" {
			bodyStart = index + 1
			break
		}
	}
	if bodyStart < 0 || bodyStart > len(lines) {
		return ""
	}
	return sourceIntakeLimitText(strings.Join(strings.Fields(strings.Join(lines[bodyStart:], "\n")), " "), 180)
}

func firstApprovalID(refs []string) string {
	for _, ref := range refs {
		trimmed := strings.TrimSpace(ref)
		if strings.HasPrefix(trimmed, "approval_") {
			return trimmed
		}
	}
	return ""
}

func firstFlowID(refs []string) string {
	for _, ref := range refs {
		trimmed := strings.TrimSpace(ref)
		if strings.HasPrefix(trimmed, "flow_") {
			return trimmed
		}
	}
	return ""
}

func sourceIntakeLimitText(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 || len([]rune(trimmed)) <= max {
		return trimmed
	}
	runes := []rune(trimmed)
	return strings.TrimSpace(string(runes[:max-1])) + "..."
}

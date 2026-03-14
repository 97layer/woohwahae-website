package api

import (
	"sort"
	"strconv"
	"strings"
)

func sortSourceIntakeViewItems(items []sourceIntakeItemView) []sourceIntakeItemView {
	out := append([]sourceIntakeItemView{}, items...)
	sort.SliceStable(out, func(left, right int) bool {
		leftRank := sourceIntakeViewAttentionRank(out[left])
		rightRank := sourceIntakeViewAttentionRank(out[right])
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if out[left].PriorityScore != out[right].PriorityScore {
			return out[left].PriorityScore > out[right].PriorityScore
		}
		if !out[left].CapturedAt.Equal(out[right].CapturedAt) {
			return out[left].CapturedAt.After(out[right].CapturedAt)
		}
		return out[left].ObservationID > out[right].ObservationID
	})
	return out
}

func sourceIntakeViewAttentionRank(item sourceIntakeItemView) int {
	switch {
	case item.PolicyColor == "red":
		return 900
	case item.PrepLane != nil:
		return 860
	case len(item.DraftSeeds) > 0:
		return 820
	case item.Disposition == "review":
		return 780
	case item.Disposition == "prep":
		return 720
	case item.RouteDecision != nil && item.RouteDecision.RouteSource == "feed_sensor":
		return 680
	case item.RouteDecision != nil && strings.TrimSpace(item.RouteDecision.Decision) != "":
		return 640
	default:
		return 600
	}
}

func summarizeSourceIntakeViewAttention(items []sourceIntakeItemView) sourceIntakeAttentionView {
	if len(items) == 0 {
		return sourceIntakeAttentionView{
			Mode:        "drop_source",
			Summary:     "링크나 메모를 하나 저장해 source intake를 시작하세요.",
			Detail:      "지금은 founder가 던진 raw source를 먼저 정규화해 쌓는 단계입니다.",
			ActionLabel: "source intake 저장",
		}
	}
	for _, item := range items {
		if item.PolicyColor == "red" {
			return sourceIntakeAttentionView{
				Mode:                "review_policy",
				Summary:             nonEmptyString(item.Title, item.Summary, "최근 source") + "는 red policy라 founder 판단이 먼저 필요합니다.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.ObservationID,
				SourceObservationID: item.ObservationID,
			}
		}
	}
	for _, item := range items {
		if item.PrepLane != nil && item.RouteDecision != nil && item.RouteDecision.RouteSource == "feed_sensor" {
			return sourceIntakeAttentionView{
				Mode:                "monitor_prep_lane",
				Summary:             nonEmptyString(item.PrepLane.TargetAccountLabel, "target") + " route와 " + nonEmptyString(item.PrepLane.ChannelLabel, item.PrepLane.Channel, "prep") + " 준비를 시스템이 먼저 열었습니다. founder review를 이어가세요.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "prep:" + item.PrepLane.Channel,
				SourceObservationID: item.ObservationID,
				DraftObservationID:  item.PrepLane.DraftObservationID,
			}
		}
	}
	for _, item := range items {
		if item.PrepLane != nil {
			return sourceIntakeAttentionView{
				Mode:                "monitor_prep_lane",
				Summary:             nonEmptyString(item.PrepLane.TargetAccountLabel, "target") + " " + nonEmptyString(item.PrepLane.ChannelLabel, item.PrepLane.Channel, "prep") + " 준비 lane을 이어가세요.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "prep:" + item.PrepLane.Channel,
				SourceObservationID: item.ObservationID,
				DraftObservationID:  item.PrepLane.DraftObservationID,
			}
		}
	}
	for _, item := range items {
		if len(item.DraftSeeds) > 0 && item.RouteDecision != nil && item.RouteDecision.RouteSource == "feed_sensor" {
			return sourceIntakeAttentionView{
				Mode:                "open_threads_prep",
				Summary:             nonEmptyString(item.DraftSeeds[0].TargetAccountLabel, "target") + " route를 시스템이 먼저 열어뒀습니다. Threads 준비만 이어가면 됩니다.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.DraftSeeds[0].ObservationID,
				SourceObservationID: item.ObservationID,
				DraftObservationID:  item.DraftSeeds[0].ObservationID,
				ActionLabel:         "Threads 준비 열기",
			}
		}
	}
	for _, item := range items {
		if len(item.DraftSeeds) > 0 {
			return sourceIntakeAttentionView{
				Mode:                "open_threads_prep",
				Summary:             nonEmptyString(item.DraftSeeds[0].TargetAccountLabel, "target") + " Threads 준비를 먼저 여세요.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.DraftSeeds[0].ObservationID,
				SourceObservationID: item.ObservationID,
				DraftObservationID:  item.DraftSeeds[0].ObservationID,
				ActionLabel:         "Threads 준비 열기",
			}
		}
	}
	for _, item := range items {
		if item.Disposition == "review" {
			return sourceIntakeAttentionView{
				Mode:                "review_policy",
				Summary:             nonEmptyString(item.Title, item.Summary, "최근 source") + "는 founder review가 필요한 intake입니다.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.ObservationID,
				SourceObservationID: item.ObservationID,
			}
		}
	}
	for _, item := range items {
		if item.RouteDecision != nil && item.RouteDecision.RouteSource == "feed_sensor" {
			return sourceIntakeAttentionView{
				Mode:                "review_route",
				Summary:             nonEmptyString(item.RouteDecision.DecisionLabel, "target") + " route를 시스템이 먼저 잡았습니다. 방금 열린 source를 확인하세요.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.ObservationID,
				SourceObservationID: item.ObservationID,
			}
		}
	}
	for _, item := range items {
		if item.Disposition == "prep" {
			return sourceIntakeAttentionView{
				Mode:                "prep_candidate",
				Summary:             nonEmptyString(item.Title, item.Summary, "최근 source") + "는 draft/prep 후보입니다.",
				Detail:              sourceIntakeViewDetail(item),
				Ref:                 "observation:" + item.ObservationID,
				SourceObservationID: item.ObservationID,
			}
		}
	}
	latest := items[0]
	return sourceIntakeAttentionView{
		Mode:                "review_route",
		Summary:             "최근 source intake를 확인하고 다음 draft seed 여부를 살피세요.",
		Detail:              sourceIntakeViewDetail(latest),
		Ref:                 "observation:" + latest.ObservationID,
		SourceObservationID: latest.ObservationID,
	}
}

func sourceIntakeViewDetail(item sourceIntakeItemView) string {
	subject := nonEmptyString(item.Title, item.Summary, "최근 source")
	switch {
	case item.PrepLane != nil:
		return subject + " · " + nonEmptyString(item.PrepLane.ChannelLabel, item.PrepLane.Channel, "prep") + " lane"
	case item.DispositionNote != "":
		return subject + " · priority " + strconv.Itoa(item.PriorityScore) + " · " + item.DispositionNote
	case item.FeedSource != "":
		return subject + " · " + item.OriginLabel + " · " + sourceIntakeFeedLabel(item.FeedSource)
	case item.FounderNote != "":
		return subject + " · founder note " + item.FounderNote
	default:
		return subject + " · " + item.OriginLabel
	}
}

func sourceIntakeIntakeClassLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "manual_drop":
		return "manual drop"
	case "authorized_connector":
		return "authorized connector"
	case "public_collector":
		return "public collector"
	default:
		return strings.TrimSpace(value)
	}
}

func sourceIntakePolicyColorLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "green":
		return "green"
	case "yellow":
		return "yellow"
	case "red":
		return "red"
	default:
		return strings.TrimSpace(value)
	}
}

func sourceIntakeDispositionLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "observe":
		return "observe"
	case "review":
		return "review"
	case "prep":
		return "prep"
	default:
		return "observe"
	}
}

func sourceIntakeRouteSourceLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "feed_sensor":
		return "sensor auto-route"
	case "telegram":
		return "founder route"
	default:
		return nonEmptyString(strings.TrimSpace(value), "route")
	}
}

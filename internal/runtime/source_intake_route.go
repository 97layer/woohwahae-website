package runtime

import (
	"fmt"
	"strings"
)

type SourceIntakeDraftLaneResult struct {
	RouteDecision ObservationRecord `json:"route_decision"`
	DraftSeed     ObservationRecord `json:"draft_seed"`
	Proposal      ProposalItem      `json:"proposal"`
	Reused        bool              `json:"reused"`
}

func normalizeSourceIntakeRouteChoice(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "97layer":
		return "97layer"
	case "woosunhokr", "woosunho", "우순호":
		return "woosunhokr"
	case "woohwahae", "우화해":
		return "woohwahae"
	case "hold", "보류":
		return "hold"
	default:
		return ""
	}
}

func SourceIntakeRouteChoiceLabel(value string) string {
	switch normalizeSourceIntakeRouteChoice(value) {
	case "97layer":
		return "97layer"
	case "woosunhokr":
		return "우순호"
	case "woohwahae":
		return "우화해"
	case "hold":
		return "hold"
	default:
		return strings.TrimSpace(value)
	}
}

func EnsureSourceIntakeRouteDecisionObservation(service SourceIntakeObservationService, sourceObservation ObservationRecord, decision string, actor string, routeSource string) (ObservationRecord, bool, error) {
	decision = normalizeSourceIntakeRouteChoice(decision)
	if decision == "" {
		return ObservationRecord{}, false, fmt.Errorf("route decision is required")
	}
	sourceObservationID := strings.TrimSpace(sourceObservation.ObservationID)
	if sourceObservationID == "" {
		return ObservationRecord{}, false, fmt.Errorf("source observation is required")
	}
	if existing, ok := findExistingSourceIntakeRouteDecisionObservation(service, sourceObservationID, decision); ok {
		if fullService, ok := service.(*Service); ok {
			if err := fullService.resolveSourceIntakeRouteReview(sourceObservationID, "route decision opened or reused a downstream lane for this source intake", "review_room.auto.source_intake_sensor_route_cleared_by_route", []string{"route_decision:" + existing.ObservationID, "decision:" + decision}); err != nil {
				return ObservationRecord{}, false, err
			}
		}
		return existing, false, nil
	}

	record := ParseSourceIntakeRawExcerpt(sourceObservation.RawExcerpt)
	subject := sourceIntakeReviewSubject(record, sourceObservation)
	rawExcerpt := strings.Join([]string{
		"source_observation_id=" + sourceObservationID,
		"decision=" + decision,
		"title=" + subject,
		"summary=" + strings.TrimSpace(sourceObservation.NormalizedSummary),
		"route_source=" + strings.TrimSpace(routeSource),
	}, "\n")
	created, err := service.CreateObservation(ObservationRecord{
		Topic:         IntakeRouteDecisionTopic,
		SourceChannel: sourceObservation.SourceChannel,
		Actor:         strings.TrimSpace(actor),
		Refs: []string{
			sourceObservationID,
			"decision_route:" + decision,
			"route_source:" + strings.TrimSpace(routeSource),
		},
		Confidence:        "high",
		RawExcerpt:        rawExcerpt,
		NormalizedSummary: fmt.Sprintf("Intake route decided · %s -> %s", subject, SourceIntakeRouteChoiceLabel(decision)),
		CapturedAt:        zeroSafeNow(),
	})
	if err != nil {
		return ObservationRecord{}, false, err
	}
	if fullService, ok := service.(*Service); ok {
		if err := fullService.resolveSourceIntakeRouteReview(sourceObservationID, "route decision opened or reused a downstream lane for this source intake", "review_room.auto.source_intake_sensor_route_cleared_by_route", []string{"route_decision:" + created.ObservationID, "decision:" + decision, "route_source:" + strings.TrimSpace(routeSource)}); err != nil {
			return ObservationRecord{}, false, err
		}
	}
	return created, true, nil
}

func OpenSourceIntakeDraftLane(service *Service, sourceObservation ObservationRecord, decision string, actor string, routeSource string) (SourceIntakeDraftLaneResult, error) {
	decision = normalizeSourceIntakeRouteChoice(decision)
	if decision == "" || decision == "hold" {
		return SourceIntakeDraftLaneResult{}, fmt.Errorf("route decision must target a canonical account")
	}
	routeDecision, routeOpened, err := EnsureSourceIntakeRouteDecisionObservation(service, sourceObservation, decision, actor, routeSource)
	if err != nil {
		return SourceIntakeDraftLaneResult{}, err
	}
	draftSeed, draftOpened, err := EnsureSourceDraftSeedObservation(service, sourceObservation, routeDecision.ObservationID, decision)
	if err != nil {
		return SourceIntakeDraftLaneResult{}, err
	}
	proposal, proposalOpened, err := EnsureSourceDraftSeedProposal(service, sourceObservation, routeDecision.ObservationID, decision)
	if err != nil {
		return SourceIntakeDraftLaneResult{}, err
	}
	return SourceIntakeDraftLaneResult{
		RouteDecision: routeDecision,
		DraftSeed:     draftSeed,
		Proposal:      proposal,
		Reused:        !routeOpened && !draftOpened && !proposalOpened,
	}, nil
}

func SourceIntakeAutoRouteTarget(record SourceIntakeRecord, refs []string) (string, bool) {
	record = NormalizeSourceIntakeRecord(record)
	if record.PolicyColor != "green" {
		return "", false
	}
	if record.Disposition != "" && record.Disposition != "prep" {
		return "", false
	}
	hints := explicitSourceIntakeRouteHints(refs)
	if len(hints) != 1 {
		return "", false
	}
	target := normalizeSourceIntakeRouteChoice(hints[0])
	if target == "" || target == "hold" {
		return "", false
	}
	return target, true
}

func explicitSourceIntakeRouteHints(refs []string) []string {
	out := []string{}
	for _, ref := range refs {
		if value, ok := sourceIntakeRefValue(ref, "route:"); ok {
			out = appendUniqueString(out, normalizeSourceIntakeRouteChoice(value))
		}
	}
	filtered := []string{}
	for _, item := range out {
		if item == "" || item == "hold" {
			continue
		}
		filtered = appendUniqueString(filtered, item)
	}
	return filtered
}

func findExistingSourceIntakeRouteDecisionObservation(service SourceIntakeObservationService, sourceObservationID string, decision string) (ObservationRecord, bool) {
	decision = normalizeSourceIntakeRouteChoice(decision)
	for _, ref := range []string{strings.TrimSpace(sourceObservationID), "decision_route:" + decision} {
		items := service.ListObservations(ObservationQuery{Ref: ref, Limit: 32})
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Topic)) != IntakeRouteDecisionTopic {
				continue
			}
			if !observationHasRef(item, strings.TrimSpace(sourceObservationID)) || !observationHasRef(item, "decision_route:"+decision) {
				continue
			}
			return item, true
		}
	}
	return ObservationRecord{}, false
}

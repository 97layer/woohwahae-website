package runtime

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SourceIntakeTopic        = "source_intake"
	IntakeRouteDecisionTopic = "intake_route_decision"
	SourceDraftSeedTopic     = "source_draft_seed"
	SourceDraftSeedLaneNote  = "lane:source_draft_seed"
)

type SourceIntakeRecord struct {
	IntakeClass     string
	PolicyColor     string
	Title           string
	URL             string
	Excerpt         string
	FounderNote     string
	DomainTags      []string
	WorldviewTags   []string
	SuggestedRoutes []string
	PriorityScore   int
	Disposition     string
	DispositionNote string
}

type SourceDraftSeedRecord struct {
	TargetAccount              string
	TargetToneLevel            string
	Title                      string
	SourceObservationID        string
	RouteDecisionObservationID string
	ParentDraftObservationID   string
	RevisionNote               string
	SourceTitle                string
	SourceURL                  string
	FounderNote                string
	DomainTags                 []string
	WorldviewTags              []string
	Draft                      string
}

type SourceIntakeObservationInput struct {
	SourceChannel string
	Actor         string
	Confidence    string
	Refs          []string
	CapturedAt    time.Time
	Record        SourceIntakeRecord
}

func NormalizeSourceIntakeRecord(input SourceIntakeRecord) SourceIntakeRecord {
	record := SourceIntakeRecord{
		IntakeClass:     normalizeSourceIntakeClass(input.IntakeClass),
		PolicyColor:     normalizeSourceIntakePolicyColor(input.PolicyColor),
		Title:           strings.TrimSpace(input.Title),
		URL:             normalizeSourceIntakeURL(input.URL),
		Excerpt:         strings.TrimSpace(input.Excerpt),
		FounderNote:     strings.TrimSpace(input.FounderNote),
		DomainTags:      normalizeSourceIntakeTags(input.DomainTags),
		WorldviewTags:   normalizeSourceIntakeTags(input.WorldviewTags),
		SuggestedRoutes: normalizeSourceIntakeRoutes(input.SuggestedRoutes),
		PriorityScore:   normalizeSourceIntakePriorityScore(input.PriorityScore),
		Disposition:     normalizeSourceIntakeDisposition(input.Disposition),
		DispositionNote: strings.TrimSpace(input.DispositionNote),
	}
	if record.Excerpt == "" {
		record.Excerpt = record.Title
	}
	return record
}

func BuildSourceIntakeRawExcerpt(input SourceIntakeRecord) string {
	record := NormalizeSourceIntakeRecord(input)
	return strings.Join([]string{
		"intake_class=" + record.IntakeClass,
		"policy_color=" + record.PolicyColor,
		"title=" + record.Title,
		"url=" + record.URL,
		"founder_note=" + record.FounderNote,
		"priority_score=" + strconv.Itoa(record.PriorityScore),
		"disposition=" + record.Disposition,
		"disposition_note=" + record.DispositionNote,
		"domain_tags=" + joinSourceIntakeTags(record.DomainTags),
		"worldview_tags=" + joinSourceIntakeTags(record.WorldviewTags),
		"suggested_routes=" + joinSourceIntakeRoutes(record.SuggestedRoutes),
		"excerpt:",
		record.Excerpt,
	}, "\n")
}

func ParseSourceIntakeRawExcerpt(raw string) SourceIntakeRecord {
	record := SourceIntakeRecord{
		IntakeClass:     "manual_drop",
		PolicyColor:     "green",
		SuggestedRoutes: []string{"97layer"},
	}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	excerptStart := -1
	for index, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		switch {
		case line == "excerpt:":
			excerptStart = index + 1
		case strings.HasPrefix(line, "intake_class="):
			record.IntakeClass = normalizeSourceIntakeClass(strings.TrimSpace(strings.TrimPrefix(line, "intake_class=")))
		case strings.HasPrefix(line, "policy_color="):
			record.PolicyColor = normalizeSourceIntakePolicyColor(strings.TrimSpace(strings.TrimPrefix(line, "policy_color=")))
		case strings.HasPrefix(line, "title="):
			record.Title = strings.TrimSpace(strings.TrimPrefix(line, "title="))
		case strings.HasPrefix(line, "url="):
			record.URL = normalizeSourceIntakeURL(strings.TrimSpace(strings.TrimPrefix(line, "url=")))
		case strings.HasPrefix(line, "founder_note="):
			record.FounderNote = strings.TrimSpace(strings.TrimPrefix(line, "founder_note="))
		case strings.HasPrefix(line, "priority_score="):
			record.PriorityScore = parseSourceIntakePriorityScore(strings.TrimSpace(strings.TrimPrefix(line, "priority_score=")))
		case strings.HasPrefix(line, "disposition="):
			record.Disposition = normalizeSourceIntakeDisposition(strings.TrimSpace(strings.TrimPrefix(line, "disposition=")))
		case strings.HasPrefix(line, "disposition_note="):
			record.DispositionNote = strings.TrimSpace(strings.TrimPrefix(line, "disposition_note="))
		case strings.HasPrefix(line, "domain_tags="):
			record.DomainTags = normalizeSourceIntakeTags(strings.Split(strings.TrimPrefix(line, "domain_tags="), ","))
		case strings.HasPrefix(line, "worldview_tags="):
			record.WorldviewTags = normalizeSourceIntakeTags(strings.Split(strings.TrimPrefix(line, "worldview_tags="), ","))
		case strings.HasPrefix(line, "suggested_routes="):
			record.SuggestedRoutes = normalizeSourceIntakeRoutes(strings.Split(strings.TrimPrefix(line, "suggested_routes="), ","))
		}
	}
	if excerptStart >= 0 && excerptStart <= len(lines) {
		record.Excerpt = strings.TrimSpace(strings.Join(lines[excerptStart:], "\n"))
	}
	return NormalizeSourceIntakeRecord(record)
}

func NormalizeSourceDraftSeedRecord(input SourceDraftSeedRecord) SourceDraftSeedRecord {
	record := SourceDraftSeedRecord{
		TargetAccount:              normalizeSourceDraftTarget(input.TargetAccount),
		TargetToneLevel:            normalizeSourceDraftToneLevel(input.TargetToneLevel),
		Title:                      strings.TrimSpace(input.Title),
		SourceObservationID:        strings.TrimSpace(input.SourceObservationID),
		RouteDecisionObservationID: strings.TrimSpace(input.RouteDecisionObservationID),
		ParentDraftObservationID:   strings.TrimSpace(input.ParentDraftObservationID),
		RevisionNote:               strings.TrimSpace(input.RevisionNote),
		SourceTitle:                strings.TrimSpace(input.SourceTitle),
		SourceURL:                  normalizeSourceIntakeURL(input.SourceURL),
		FounderNote:                strings.TrimSpace(input.FounderNote),
		DomainTags:                 normalizeSourceIntakeTags(input.DomainTags),
		WorldviewTags:              normalizeSourceIntakeTags(input.WorldviewTags),
		Draft:                      strings.TrimSpace(input.Draft),
	}
	if record.TargetToneLevel == "" {
		record.TargetToneLevel = sourceDraftSeedToneLevel(record.TargetAccount)
	}
	return record
}

func BuildSourceDraftSeedRawExcerpt(input SourceDraftSeedRecord) string {
	record := NormalizeSourceDraftSeedRecord(input)
	return strings.Join([]string{
		"target_account=" + record.TargetAccount,
		"target_tone=" + record.TargetToneLevel,
		"title=" + record.Title,
		"source_observation_id=" + record.SourceObservationID,
		"route_decision_id=" + record.RouteDecisionObservationID,
		"parent_draft_observation_id=" + record.ParentDraftObservationID,
		"revision_note=" + record.RevisionNote,
		"source_title=" + record.SourceTitle,
		"source_url=" + record.SourceURL,
		"founder_note=" + record.FounderNote,
		"domain_tags=" + joinSourceIntakeTags(record.DomainTags),
		"worldview_tags=" + joinSourceIntakeTags(record.WorldviewTags),
		"draft:",
		record.Draft,
	}, "\n")
}

func ParseSourceDraftSeedRawExcerpt(raw string) SourceDraftSeedRecord {
	record := SourceDraftSeedRecord{}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	draftStart := -1
	for index, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		switch {
		case line == "draft:":
			draftStart = index + 1
		case strings.HasPrefix(line, "target_account="):
			record.TargetAccount = normalizeSourceDraftTarget(strings.TrimSpace(strings.TrimPrefix(line, "target_account=")))
		case strings.HasPrefix(line, "target_tone="):
			record.TargetToneLevel = normalizeSourceDraftToneLevel(strings.TrimSpace(strings.TrimPrefix(line, "target_tone=")))
		case strings.HasPrefix(line, "title="):
			record.Title = strings.TrimSpace(strings.TrimPrefix(line, "title="))
		case strings.HasPrefix(line, "source_observation_id="):
			record.SourceObservationID = strings.TrimSpace(strings.TrimPrefix(line, "source_observation_id="))
		case strings.HasPrefix(line, "route_decision_id="):
			record.RouteDecisionObservationID = strings.TrimSpace(strings.TrimPrefix(line, "route_decision_id="))
		case strings.HasPrefix(line, "parent_draft_observation_id="):
			record.ParentDraftObservationID = strings.TrimSpace(strings.TrimPrefix(line, "parent_draft_observation_id="))
		case strings.HasPrefix(line, "revision_note="):
			record.RevisionNote = strings.TrimSpace(strings.TrimPrefix(line, "revision_note="))
		case strings.HasPrefix(line, "source_title="):
			record.SourceTitle = strings.TrimSpace(strings.TrimPrefix(line, "source_title="))
		case strings.HasPrefix(line, "source_url="):
			record.SourceURL = normalizeSourceIntakeURL(strings.TrimSpace(strings.TrimPrefix(line, "source_url=")))
		case strings.HasPrefix(line, "founder_note="):
			record.FounderNote = strings.TrimSpace(strings.TrimPrefix(line, "founder_note="))
		case strings.HasPrefix(line, "domain_tags="):
			record.DomainTags = normalizeSourceIntakeTags(strings.Split(strings.TrimPrefix(line, "domain_tags="), ","))
		case strings.HasPrefix(line, "worldview_tags="):
			record.WorldviewTags = normalizeSourceIntakeTags(strings.Split(strings.TrimPrefix(line, "worldview_tags="), ","))
		}
	}
	if draftStart >= 0 && draftStart <= len(lines) {
		record.Draft = strings.TrimSpace(strings.Join(lines[draftStart:], "\n"))
	}
	return NormalizeSourceDraftSeedRecord(record)
}

func SourceIntakeSummary(input SourceIntakeRecord) string {
	record := NormalizeSourceIntakeRecord(input)
	subject := record.Title
	if subject == "" {
		subject = record.URL
	}
	if subject == "" {
		subject = sourceIntakeLimitText(strings.Join(strings.Fields(record.Excerpt), " "), 72)
	}
	routes := sourceIntakeRouteLabels(record.SuggestedRoutes)
	if routes == "" {
		routes = "97layer"
	}
	return sourceIntakeLimitText("Source intake · "+subject+" -> "+routes, 180)
}

func SourceIntakePriority(input SourceIntakeRecord) (int, string, string) {
	record := NormalizeSourceIntakeRecord(input)
	if record.PriorityScore > 0 && record.Disposition != "" {
		return record.PriorityScore, record.Disposition, strings.TrimSpace(record.DispositionNote)
	}

	score := 28
	if record.Title != "" {
		score += 8
	}
	if record.URL != "" {
		score += 5
	}
	switch excerptLength := len([]rune(record.Excerpt)); {
	case excerptLength >= 120:
		score += 12
	case excerptLength >= 48:
		score += 7
	}
	score += minSourceIntakeScoreBonus(len(record.DomainTags)*4, 12)
	score += minSourceIntakeScoreBonus(len(record.WorldviewTags)*5, 15)
	hintCount := len(record.SuggestedRoutes)
	switch record.PolicyColor {
	case "red":
		score += 50
	case "yellow":
		score += 22
	case "green":
		score += 12
	}
	switch hintCount {
	case 0:
		score += 8
	case 1:
		score += 18
	default:
		score += 12
	}
	score = normalizeSourceIntakePriorityScore(score)

	disposition := "observe"
	dispositionNote := "source를 corpus 후보로만 유지하고 founder가 당겨올 때까지 조용히 둡니다."
	switch {
	case record.PolicyColor == "red":
		disposition = "review"
		dispositionNote = "red policy source라 founder 검토가 먼저 필요합니다."
	case record.PolicyColor == "yellow":
		disposition = "review"
		switch {
		case strings.Contains(record.FounderNote, "route cue가 비어"):
			dispositionNote = "route cue가 비어 있어 founder route 분류가 필요합니다."
		case strings.Contains(record.FounderNote, "여러 route cue"):
			dispositionNote = "route cue가 여러 개라 founder route 결정을 먼저 받아야 합니다."
		case hintCount == 0:
			dispositionNote = "route cue가 비어 있어 founder route 분류가 필요합니다."
		case hintCount > 1:
			dispositionNote = "route cue가 여러 개라 founder route 결정을 먼저 받아야 합니다."
		default:
			dispositionNote = "yellow policy source라 founder가 먼저 의미와 쓰임을 확인해야 합니다."
		}
	case hintCount == 1 && score >= 60:
		disposition = "prep"
		dispositionNote = "single route cue와 충분한 signal이 있어 draft/prep 후보로 바로 올릴 수 있습니다."
	case score >= 48:
		disposition = "review"
		dispositionNote = "signal은 충분하지만 founder가 route와 쓰임을 한 번 확인하는 편이 안전합니다."
	}
	return score, disposition, dispositionNote
}

func FinalizeSourceIntakeRecord(input SourceIntakeRecord) SourceIntakeRecord {
	record := NormalizeSourceIntakeRecord(input)
	score, disposition, note := SourceIntakePriority(record)
	if record.PriorityScore == 0 {
		record.PriorityScore = score
	}
	if record.Disposition == "" {
		record.Disposition = disposition
	}
	if record.DispositionNote == "" {
		record.DispositionNote = note
	}
	return NormalizeSourceIntakeRecord(record)
}

func BuildSourceIntakeObservation(input SourceIntakeObservationInput) (ObservationRecord, SourceIntakeRecord, error) {
	record := FinalizeSourceIntakeRecord(input.Record)
	if strings.TrimSpace(record.Title) == "" && strings.TrimSpace(record.Excerpt) == "" {
		return ObservationRecord{}, SourceIntakeRecord{}, fmt.Errorf("source intake title or excerpt is required")
	}

	sourceChannel := strings.TrimSpace(input.SourceChannel)
	if sourceChannel == "" {
		sourceChannel = "cockpit"
	}
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = "founder"
	}
	confidence := strings.TrimSpace(input.Confidence)
	if confidence == "" {
		confidence = "high"
	}
	refs := []string{}
	for _, ref := range SourceIntakeRefs(record) {
		refs = appendUniqueString(refs, ref)
	}
	for _, ref := range input.Refs {
		refs = appendUniqueString(refs, strings.TrimSpace(ref))
	}

	observation := ObservationRecord{
		Topic:             SourceIntakeTopic,
		SourceChannel:     sourceChannel,
		Actor:             actor,
		Refs:              refs,
		Confidence:        confidence,
		RawExcerpt:        BuildSourceIntakeRawExcerpt(record),
		NormalizedSummary: SourceIntakeSummary(record),
		CapturedAt:        input.CapturedAt,
	}
	return observation, record, nil
}

func SourceIntakeRefs(input SourceIntakeRecord) []string {
	record := NormalizeSourceIntakeRecord(input)
	refs := make([]string, 0, len(record.SuggestedRoutes)+len(record.DomainTags)+len(record.WorldviewTags))
	for _, route := range record.SuggestedRoutes {
		refs = appendUniqueString(refs, "route:"+route)
	}
	for _, tag := range record.DomainTags {
		refs = appendUniqueString(refs, "domain:"+tag)
	}
	for _, tag := range record.WorldviewTags {
		refs = appendUniqueString(refs, "worldview:"+tag)
	}
	return refs
}

func normalizeSourceIntakeClass(value string) string {
	switch strings.TrimSpace(value) {
	case "manual_drop", "authorized_connector", "public_collector":
		return strings.TrimSpace(value)
	default:
		return "manual_drop"
	}
}

func normalizeSourceIntakePolicyColor(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "green", "yellow", "red":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "green"
	}
}

func normalizeSourceIntakePriorityScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func parseSourceIntakePriorityScore(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return normalizeSourceIntakePriorityScore(parsed)
}

func normalizeSourceIntakeDisposition(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "observe", "review", "prep":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func minSourceIntakeScoreBonus(value int, max int) int {
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}

func normalizeSourceIntakeTags(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || value == "none" {
			continue
		}
		out = appendUniqueString(out, value)
	}
	return out
}

func normalizeSourceIntakeRoutes(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		switch strings.ToLower(strings.TrimSpace(item)) {
		case "97layer":
			out = appendUniqueString(out, "97layer")
		case "woosunhokr", "woosunho", "우순호":
			out = appendUniqueString(out, "woosunhokr")
		case "woohwahae", "우화해":
			out = appendUniqueString(out, "woohwahae")
		}
	}
	if len(out) == 0 {
		return []string{"97layer"}
	}
	return out
}

func normalizeSourceDraftToneLevel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "raw", "refined", "polished":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func joinSourceIntakeTags(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ",")
}

func joinSourceIntakeRoutes(items []string) string {
	routes := normalizeSourceIntakeRoutes(items)
	if len(routes) == 0 {
		return "97layer"
	}
	return strings.Join(routes, ",")
}

func sourceIntakeRouteLabels(routes []string) string {
	labels := make([]string, 0, len(routes))
	for _, route := range normalizeSourceIntakeRoutes(routes) {
		switch route {
		case "97layer":
			labels = append(labels, "97layer")
		case "woosunhokr":
			labels = append(labels, "우순호")
		case "woohwahae":
			labels = append(labels, "우화해")
		}
	}
	return strings.Join(labels, ", ")
}

func sourceIntakeLimitText(value string, max int) string {
	text := strings.TrimSpace(value)
	if max <= 0 || len([]rune(text)) <= max {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:max-1])) + "…"
}

func normalizeSourceIntakeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.String()
}

func ExtractFirstSourceURL(text string) string {
	for _, part := range strings.Fields(strings.TrimSpace(text)) {
		if value := normalizeSourceIntakeURL(part); value != "" {
			return value
		}
	}
	return ""
}

func EnsureSourceDraftSeedProposal(service *Service, sourceObservation ObservationRecord, routeDecisionObservationID string, targetAccount string) (ProposalItem, bool, error) {
	targetAccount = normalizeSourceDraftTarget(targetAccount)
	if targetAccount == "" {
		return ProposalItem{}, false, fmt.Errorf("target account is required")
	}
	sourceObservationID := strings.TrimSpace(sourceObservation.ObservationID)
	if sourceObservationID == "" {
		return ProposalItem{}, false, fmt.Errorf("source observation is required")
	}

	for _, item := range service.ListProposals() {
		if sourceDraftSeedMatches(item, sourceObservationID, targetAccount) {
			return item, false, nil
		}
	}

	record := ParseSourceIntakeRawExcerpt(sourceObservation.RawExcerpt)
	draftSeed, _, err := EnsureSourceDraftSeedObservation(service, sourceObservation, routeDecisionObservationID, targetAccount)
	if err != nil {
		return ProposalItem{}, false, err
	}
	draftRecord := ParseSourceDraftSeedRawExcerpt(draftSeed.RawExcerpt)
	now := zeroSafeNow()
	proposal := ProposalItem{
		ProposalID: nextSourceDraftSeedProposalID(now, targetAccount, sourceObservationID),
		Title:      sourceDraftSeedTitle(targetAccount, record, sourceObservation),
		Intent:     sourceDraftSeedIntent(targetAccount, record),
		Summary:    sourceDraftSeedSummary(targetAccount, record, sourceObservation, draftRecord.Draft),
		Surface:    SurfaceTelegram,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes: []string{
			SourceDraftSeedLaneNote,
			"account:" + targetAccount,
			"source_observation:" + sourceObservationID,
			"source_topic:" + SourceIntakeTopic,
			"source_channel:" + strings.TrimSpace(sourceObservation.SourceChannel),
			"route_decision:" + strings.TrimSpace(routeDecisionObservationID),
			"draft_seed_observation:" + strings.TrimSpace(draftSeed.ObservationID),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.CreateProposal(proposal); err != nil {
		return ProposalItem{}, false, err
	}
	return proposal, true, nil
}

func EnsureSourceDraftSeedObservation(service *Service, sourceObservation ObservationRecord, routeDecisionObservationID string, targetAccount string) (ObservationRecord, bool, error) {
	targetAccount = normalizeSourceDraftTarget(targetAccount)
	if targetAccount == "" {
		return ObservationRecord{}, false, fmt.Errorf("target account is required")
	}
	sourceObservationID := strings.TrimSpace(sourceObservation.ObservationID)
	if sourceObservationID == "" {
		return ObservationRecord{}, false, fmt.Errorf("source observation is required")
	}
	for _, item := range service.ListObservations(ObservationQuery{Topic: SourceDraftSeedTopic, Limit: 128}) {
		if sourceDraftSeedObservationMatches(item, sourceObservationID, targetAccount) {
			if err := service.resolveSourceIntakeRedPolicyReview(sourceObservationID, "route decision opened or reused a draft seed for this source intake", "review_room.auto.source_intake_red_policy_cleared_by_draft_seed", []string{"draft_seed:" + item.ObservationID, "route_decision:" + strings.TrimSpace(routeDecisionObservationID), "account:" + targetAccount}); err != nil {
				return ObservationRecord{}, false, err
			}
			return item, false, nil
		}
	}

	sourceRecord := ParseSourceIntakeRawExcerpt(sourceObservation.RawExcerpt)
	now := zeroSafeNow()
	draftRecord := buildSourceDraftSeedRecord(sourceRecord, sourceObservation, routeDecisionObservationID, targetAccount)
	refs := []string{}
	for _, value := range []string{
		sourceObservationID,
		routeDecisionObservationID,
		"account:" + targetAccount,
		"source_topic:" + SourceIntakeTopic,
		"route_decision:" + strings.TrimSpace(routeDecisionObservationID),
	} {
		refs = appendUniqueString(refs, value)
	}
	observation, err := service.CreateObservation(ObservationRecord{
		Topic:             SourceDraftSeedTopic,
		SourceChannel:     sourceObservation.SourceChannel,
		Actor:             sourceObservation.Actor,
		Refs:              refs,
		Confidence:        "medium",
		RawExcerpt:        BuildSourceDraftSeedRawExcerpt(draftRecord),
		NormalizedSummary: sourceDraftSeedSummary(targetAccount, sourceRecord, sourceObservation, draftRecord.Draft),
		CapturedAt:        now,
	})
	if err != nil {
		return ObservationRecord{}, false, err
	}
	if err := service.resolveSourceIntakeRedPolicyReview(sourceObservationID, "route decision opened or reused a draft seed for this source intake", "review_room.auto.source_intake_red_policy_cleared_by_draft_seed", []string{"draft_seed:" + observation.ObservationID, "route_decision:" + strings.TrimSpace(routeDecisionObservationID), "account:" + targetAccount}); err != nil {
		return ObservationRecord{}, false, err
	}
	return observation, true, nil
}

func ReviseSourceDraftSeedObservation(service *Service, draftObservation ObservationRecord, instruction string) (ObservationRecord, error) {
	if strings.TrimSpace(strings.ToLower(draftObservation.Topic)) != SourceDraftSeedTopic {
		return ObservationRecord{}, fmt.Errorf("draft observation is required")
	}
	instruction = strings.TrimSpace(instruction)
	if instruction == "" {
		return ObservationRecord{}, fmt.Errorf("revision instruction is required")
	}

	draftRecord := ParseSourceDraftSeedRawExcerpt(draftObservation.RawExcerpt)
	if draftRecord.TargetAccount == "" || draftRecord.SourceObservationID == "" {
		return ObservationRecord{}, fmt.Errorf("draft seed is missing canonical refs")
	}
	sourceObservation, ok := findObservationByID(service, draftRecord.SourceObservationID)
	if !ok {
		return ObservationRecord{}, fmt.Errorf("source intake not found")
	}
	sourceRecord := ParseSourceIntakeRawExcerpt(sourceObservation.RawExcerpt)
	revisedRecord := buildSourceDraftSeedRecord(sourceRecord, sourceObservation, draftRecord.RouteDecisionObservationID, draftRecord.TargetAccount)
	revisedRecord.ParentDraftObservationID = draftObservation.ObservationID
	revisedRecord.RevisionNote = instruction
	revisedRecord.Draft = reviseSourceDraftSeedBody(revisedRecord.Draft, draftRecord.TargetAccount, instruction)

	refs := []string{}
	for _, value := range append([]string{}, draftObservation.Refs...) {
		refs = appendUniqueString(refs, value)
	}
	for _, value := range []string{
		draftRecord.SourceObservationID,
		draftRecord.RouteDecisionObservationID,
		"account:" + draftRecord.TargetAccount,
		"draft_seed_parent:" + strings.TrimSpace(draftObservation.ObservationID),
	} {
		refs = appendUniqueString(refs, value)
	}

	return service.CreateObservation(ObservationRecord{
		Topic:             SourceDraftSeedTopic,
		SourceChannel:     draftObservation.SourceChannel,
		Actor:             draftObservation.Actor,
		Refs:              refs,
		Confidence:        "medium",
		RawExcerpt:        BuildSourceDraftSeedRawExcerpt(revisedRecord),
		NormalizedSummary: sourceDraftSeedSummary(draftRecord.TargetAccount, sourceRecord, sourceObservation, revisedRecord.Draft),
		CapturedAt:        zeroSafeNow(),
	})
}

func findObservationByID(service *Service, observationID string) (ObservationRecord, bool) {
	observationID = strings.TrimSpace(observationID)
	if observationID == "" {
		return ObservationRecord{}, false
	}
	for _, item := range service.ListObservations(ObservationQuery{Limit: 256}) {
		if item.ObservationID == observationID {
			return item, true
		}
	}
	return ObservationRecord{}, false
}

func (s *Service) resolveSourceIntakeRedPolicyReview(sourceObservationID string, reason string, rule string, evidence []string) error {
	sourceObservationID = strings.TrimSpace(sourceObservationID)
	if sourceObservationID == "" {
		return nil
	}
	target, ok := func() (string, bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		ref := sourceObservationID
		return findOpenReviewRoomItem(s.reviewRoom.current(), "source_intake.red_policy", &ref)
	}()
	if !ok {
		return nil
	}
	resolutionEvidence := append([]string{"observation:" + sourceObservationID}, evidence...)
	_, err := s.TransitionStructuredReviewRoomItem("resolve", target, &ReviewRoomResolution{
		Action:   "resolve",
		Reason:   strings.TrimSpace(reason),
		Rule:     strings.TrimSpace(rule),
		Evidence: normalizeReviewRoomEvidence(resolutionEvidence),
	})
	if err == ErrReviewRoomItemNotFound {
		return nil
	}
	return err
}

func (s *Service) resolveSourceIntakeRouteReview(sourceObservationID string, reason string, rule string, evidence []string) error {
	sourceObservationID = strings.TrimSpace(sourceObservationID)
	if sourceObservationID == "" {
		return nil
	}
	target, ok := func() (string, bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		ref := sourceObservationID
		return findOpenReviewRoomItem(s.reviewRoom.current(), "source_intake.sensor_route", &ref)
	}()
	if !ok {
		return nil
	}
	resolutionEvidence := append([]string{"observation:" + sourceObservationID}, evidence...)
	_, err := s.TransitionStructuredReviewRoomItem("resolve", target, &ReviewRoomResolution{
		Action:   "resolve",
		Reason:   strings.TrimSpace(reason),
		Rule:     strings.TrimSpace(rule),
		Evidence: normalizeReviewRoomEvidence(resolutionEvidence),
	})
	if err == ErrReviewRoomItemNotFound {
		return nil
	}
	return err
}

func sourceDraftSeedObservationMatches(item ObservationRecord, sourceObservationID string, targetAccount string) bool {
	if strings.TrimSpace(strings.ToLower(item.Topic)) != SourceDraftSeedTopic {
		return false
	}
	refs := strings.Join(item.Refs, " ")
	return strings.Contains(refs, sourceObservationID) && strings.Contains(refs, "account:"+targetAccount)
}

func reviseSourceDraftSeedBody(currentDraft string, targetAccount string, instruction string) string {
	instruction = strings.TrimSpace(instruction)
	currentDraft = strings.TrimSpace(currentDraft)
	if instruction == "" {
		return currentDraft
	}
	switch normalizeSourceDraftTarget(targetAccount) {
	case "97layer":
		return joinSourceDraftParagraphs(
			currentDraft,
			sourceIntakeLimitText("이번엔 "+instruction+" 쪽으로 다시 눌러본다. 아직 완성보다 기록에 가깝게 둔다.", 220),
		)
	case "woosunhokr":
		return joinSourceDraftParagraphs(
			currentDraft,
			sourceIntakeLimitText("이 버전은 "+instruction+" 쪽으로 더 정리해서, 미용사의 단상과 실무 감각이 같이 보이게 잡는다.", 220),
		)
	case "woohwahae":
		return joinSourceDraftParagraphs(
			currentDraft,
			sourceIntakeLimitText("이 버전은 "+instruction+" 쪽으로 더 덜어내서, 조용한 공적 문장으로 정리한다.", 220),
		)
	default:
		return joinSourceDraftParagraphs(currentDraft, sourceIntakeLimitText("이번엔 "+instruction+" 쪽으로 다시 잡아본다.", 180))
	}
}

func sourceDraftSeedMatches(item ProposalItem, sourceObservationID string, targetAccount string) bool {
	if !proposalHasNote(item.Notes, SourceDraftSeedLaneNote) {
		return false
	}
	return proposalNoteValue(item.Notes, "source_observation:") == sourceObservationID &&
		proposalNoteValue(item.Notes, "account:") == targetAccount
}

func proposalHasNote(notes []string, exact string) bool {
	for _, note := range notes {
		if strings.TrimSpace(note) == exact {
			return true
		}
	}
	return false
}

func proposalNoteValue(notes []string, prefix string) string {
	for _, note := range notes {
		value := strings.TrimSpace(note)
		if strings.HasPrefix(value, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(value, prefix))
		}
	}
	return ""
}

func normalizeSourceDraftTarget(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "97layer":
		return "97layer"
	case "woosunhokr":
		return "woosunhokr"
	case "woohwahae":
		return "woohwahae"
	default:
		return ""
	}
}

func sourceDraftSeedToneLevel(targetAccount string) string {
	switch normalizeSourceDraftTarget(targetAccount) {
	case "97layer":
		return "raw"
	case "woosunhokr":
		return "refined"
	case "woohwahae":
		return "polished"
	default:
		return ""
	}
}

func sourceDraftSeedTitle(targetAccount string, record SourceIntakeRecord, sourceObservation ObservationRecord) string {
	subject := sourceDraftSeedSubject(record, sourceObservation)
	switch targetAccount {
	case "97layer":
		return sourceIntakeLimitText("97layer raw draft · "+subject, 120)
	case "woosunhokr":
		return sourceIntakeLimitText("우순호 draft · "+subject, 120)
	case "woohwahae":
		return sourceIntakeLimitText("우화해 draft · "+subject, 120)
	default:
		return sourceIntakeLimitText("draft · "+subject, 120)
	}
}

func sourceDraftSeedIntent(targetAccount string, record SourceIntakeRecord) string {
	subject := sourceDraftSeedSubject(record, ObservationRecord{})
	switch targetAccount {
	case "97layer":
		return sourceIntakeLimitText("Open a raw 97layer note from source intake: "+subject, 180)
	case "woosunhokr":
		return sourceIntakeLimitText("Translate this source into a refined woosunhokr beauty-practice note: "+subject, 180)
	case "woohwahae":
		return sourceIntakeLimitText("Distill this source into a polished woohwahae public note: "+subject, 180)
	default:
		return sourceIntakeLimitText("Open a draft from source intake: "+subject, 180)
	}
}

func sourceDraftSeedSummary(targetAccount string, record SourceIntakeRecord, sourceObservation ObservationRecord, draft string) string {
	subject := sourceDraftSeedSubject(record, sourceObservation)
	return sourceDraftSeedSummaryWithPreview(targetAccount, subject, draft)
}

func sourceDraftSeedSummaryWithPreview(targetAccount string, subject string, draft string) string {
	preview := sourceDraftSeedPreview(draft)
	if preview == "" {
		return sourceIntakeLimitText(
			fmt.Sprintf("Draft seed opened · %s -> %s", subject, sourceDraftSeedTargetLabel(targetAccount)),
			180,
		)
	}
	return sourceIntakeLimitText(
		fmt.Sprintf("Draft seed opened · %s -> %s · %s", subject, sourceDraftSeedTargetLabel(targetAccount), preview),
		180,
	)
}

func sourceDraftSeedSubject(record SourceIntakeRecord, sourceObservation ObservationRecord) string {
	if strings.TrimSpace(record.Title) != "" {
		return strings.TrimSpace(record.Title)
	}
	if strings.TrimSpace(record.URL) != "" {
		return strings.TrimSpace(record.URL)
	}
	text := strings.TrimSpace(record.Excerpt)
	if text == "" {
		text = strings.TrimSpace(sourceObservation.NormalizedSummary)
	}
	return sourceIntakeLimitText(strings.Join(strings.Fields(text), " "), 72)
}

func sourceDraftSeedTargetLabel(value string) string {
	switch normalizeSourceDraftTarget(value) {
	case "97layer":
		return "97layer"
	case "woosunhokr":
		return "우순호"
	case "woohwahae":
		return "우화해"
	default:
		return strings.TrimSpace(value)
	}
}

func sourceDraftSeedPreview(draft string) string {
	text := strings.Join(strings.Fields(strings.TrimSpace(draft)), " ")
	return sourceIntakeLimitText(text, 96)
}

func buildSourceDraftSeedRecord(record SourceIntakeRecord, sourceObservation ObservationRecord, routeDecisionObservationID string, targetAccount string) SourceDraftSeedRecord {
	subject := sourceDraftSeedSubject(record, sourceObservation)
	return NormalizeSourceDraftSeedRecord(SourceDraftSeedRecord{
		TargetAccount:              targetAccount,
		TargetToneLevel:            sourceDraftSeedToneLevel(targetAccount),
		Title:                      sourceDraftSeedTitle(targetAccount, record, sourceObservation),
		SourceObservationID:        strings.TrimSpace(sourceObservation.ObservationID),
		RouteDecisionObservationID: strings.TrimSpace(routeDecisionObservationID),
		SourceTitle:                subject,
		SourceURL:                  record.URL,
		FounderNote:                record.FounderNote,
		DomainTags:                 record.DomainTags,
		WorldviewTags:              record.WorldviewTags,
		Draft:                      buildSourceDraftSeedBody(targetAccount, subject, record),
	})
}

func buildSourceDraftSeedBody(targetAccount string, subject string, record SourceIntakeRecord) string {
	excerpt := sourceDraftSeedExcerpt(record)
	founderNote := sourceIntakeLimitText(strings.Join(strings.Fields(strings.TrimSpace(record.FounderNote)), " "), 180)
	tagCue := sourceDraftSeedTagCue(record)
	focusCue := sourceDraftSeedFocusCue(record)
	subjectCue := sourceDraftSeedSubjectCue(subject, record)
	switch normalizeSourceDraftTarget(targetAccount) {
	case "97layer":
		return joinSourceDraftParagraphs(
			sourceIntakeLimitText("요즘 붙들고 있는 건 "+subjectCue+"에서 다시 걸려 나온 "+focusCue+"이다.", 220),
			excerpt,
			founderNote,
			sourceIntakeLimitText("97layer에서는 이걸 아직 정리보다 기록에 가까운 메모로 둔다. "+tagCue+" 같은 넓은 원천이 한 생각으로 묶이는지부터 본다.", 240),
		)
	case "woosunhokr":
		return joinSourceDraftParagraphs(
			sourceIntakeLimitText(subjectCue+"를 우순호 쪽으로 옮기면 결국 "+focusCue+"가 사람에게 어떻게 닿는지가 먼저 보인다.", 220),
			excerpt,
			founderNote,
			sourceIntakeLimitText("우순호에서는 이 소재를 미용사의 단상, 태도, 실무 감각이 같이 보이는 문장으로 더 정리해본다.", 220),
		)
	case "woohwahae":
		return joinSourceDraftParagraphs(
			sourceIntakeLimitText(subjectCue+"를 우화해 쪽으로 옮기면 결국 "+focusCue+"를 덜어낸 뒤 남는 생활의 리듬이 더 중요해진다.", 220),
			excerpt,
			founderNote,
			sourceIntakeLimitText("우화해에서는 이 소재를 슬로우 라이프, 태도, 조용한 공적 문장 쪽으로 더 다듬는다.", 220),
		)
	default:
		return joinSourceDraftParagraphs(subject, excerpt, founderNote)
	}
}

func sourceDraftSeedExcerpt(record SourceIntakeRecord) string {
	text := strings.Join(strings.Fields(strings.TrimSpace(record.Excerpt)), " ")
	if text == "" {
		return ""
	}
	return sourceIntakeLimitText(text, 220)
}

func sourceDraftSeedTagCue(record SourceIntakeRecord) string {
	tags := append([]string{}, record.DomainTags...)
	tags = append(tags, record.WorldviewTags...)
	if len(tags) == 0 {
		return "지금 생각"
	}
	if len(tags) == 1 {
		return tags[0]
	}
	return strings.Join(tags[:minSourceIntakeInt(len(tags), 2)], " / ")
}

func sourceDraftSeedFocusCue(record SourceIntakeRecord) string {
	combined := strings.ToLower(strings.Join(append(append([]string{}, record.DomainTags...), record.WorldviewTags...), " "))
	switch {
	case containsAnySourceDraftCue(combined, "beauty", "hair", "aesthetic", "style", "practice", "craft", "미용"):
		return "기준과 손기술 사이의 감각"
	case containsAnySourceDraftCue(combined, "system", "dev", "development", "build", "product", "automation", "operator", "개발", "시스템"):
		return "기능보다 구조와 순서"
	case containsAnySourceDraftCue(combined, "finance", "stock", "market", "money", "investment", "주식", "금융"):
		return "숫자보다 기준과 리듬"
	case containsAnySourceDraftCue(combined, "brand", "identity", "subtraction", "worldview", "브랜드", "정체성", "소거"):
		return "기준을 덜어내며 선명하게 만드는 감각"
	default:
		return "기준과 태도"
	}
}

func sourceDraftSeedSubjectCue(subject string, record SourceIntakeRecord) string {
	switch {
	case strings.TrimSpace(record.Title) != "":
		return strings.TrimSpace(record.Title)
	case strings.TrimSpace(record.URL) != "":
		return "이 링크"
	case strings.TrimSpace(subject) != "":
		return strings.TrimSpace(subject)
	default:
		return "이 소재"
	}
}

func containsAnySourceDraftCue(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}

func joinSourceDraftParagraphs(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return strings.Join(out, "\n\n")
}

func minSourceIntakeInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

var sourceDraftSeedSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func nextSourceDraftSeedProposalID(now time.Time, targetAccount string, sourceObservationID string) string {
	stamp := now.Format("20060102150405")
	slug := strings.ToLower(strings.TrimSpace(targetAccount)) + "-" + strings.ToLower(strings.TrimSpace(sourceObservationID))
	slug = sourceDraftSeedSanitizer.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 40 {
		slug = slug[:40]
		slug = strings.Trim(slug, "-")
	}
	if slug == "" {
		slug = "draft-seed"
	}
	return "proposal_source_draft_" + stamp + "_" + slug
}

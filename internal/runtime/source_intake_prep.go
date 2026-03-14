package runtime

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	brandPublishLaneNote        = "lane:brand_publish"
	sourceDraftThreadsPrepShape = "source_draft_threads_v1"
	defaultDraftPrepChannel     = "threads"
)

var sourceDraftPrepSlugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

type SourceDraftPublishPrepOpenedFromDraft struct {
	DraftObservationID string `json:"draft_observation_id"`
	TargetAccount      string `json:"target_account"`
	TargetAccountLabel string `json:"target_account_label"`
	TargetTone         string `json:"target_tone"`
	Title              string `json:"title"`
	BodyPreview        string `json:"body_preview"`
	RevisionNote       string `json:"revision_note,omitempty"`
}

type SourceDraftPublishPrepLane struct {
	Channel            string   `json:"channel"`
	Label              string   `json:"label"`
	TargetAccount      string   `json:"target_account"`
	TargetAccountLabel string   `json:"target_account_label"`
	TargetToneLevel    string   `json:"target_tone_level"`
	Title              string   `json:"title"`
	BodyPreview        string   `json:"body_preview"`
	SourceIDs          []string `json:"source_ids"`
	ExternalRefs       []string `json:"external_refs"`
	PrepShapeID        string   `json:"prep_shape_id"`
}

type SourceDraftPublishPrepFollowUp struct {
	Mode                string `json:"mode"`
	Summary             string `json:"summary"`
	Detail              string `json:"detail"`
	Ref                 string `json:"ref,omitempty"`
	SourceObservationID string `json:"source_observation_id,omitempty"`
	DraftObservationID  string `json:"draft_observation_id,omitempty"`
	ActionLabel         string `json:"action_label,omitempty"`
}

type SourceDraftPublishPrepResult struct {
	OpenedFromDraft SourceDraftPublishPrepOpenedFromDraft `json:"opened_from_draft"`
	Lane            SourceDraftPublishPrepLane            `json:"lane"`
	Proposal        ProposalItem                          `json:"proposal"`
	WorkItem        WorkItem                              `json:"work_item"`
	Approval        ApprovalItem                          `json:"approval"`
	Flow            FlowRun                               `json:"flow"`
	Observation     ObservationRecord                     `json:"observation"`
	Reused          bool                                  `json:"reused"`
	FollowUp        SourceDraftPublishPrepFollowUp        `json:"follow_up"`
}

type sourceDraftPrepShape struct {
	Channel     string
	Title       string
	Body        string
	BodyPreview string
	PrepShapeID string
}

type sourceDraftPrepIDs struct {
	ProposalID string
	WorkItemID string
	ApprovalID string
	FlowID     string
}

func (s *Service) OpenSourceDraftSeedPublishPrep(draftObservationID string, channel string) (SourceDraftPublishPrepResult, error) {
	draftObservationID = strings.TrimSpace(draftObservationID)
	channel = normalizeSourceDraftPrepChannel(channel)
	if draftObservationID == "" {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("draft_observation_id is required")
	}
	if channel == "" {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("publish prep channel is required")
	}

	draftObservation, ok := findObservationByID(s, draftObservationID)
	if !ok {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("draft seed not found")
	}
	if strings.TrimSpace(strings.ToLower(draftObservation.Topic)) != SourceDraftSeedTopic {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("draft seed not found")
	}

	draft := ParseSourceDraftSeedRawExcerpt(draftObservation.RawExcerpt)
	if draft.TargetAccount == "" {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("draft seed is missing target account")
	}
	shape, err := buildSourceDraftPrepShape(draft, channel)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}

	if existingObservation, ok := findSourceDraftPrepObservation(s, draftObservationID, channel); ok {
		if err := s.resolveSourceIntakeRedPolicyReview(draft.SourceObservationID, "publish prep opened or reused the downstream lane for this source intake", "review_room.auto.source_intake_red_policy_cleared_by_prep", []string{"draft_seed:" + draftObservationID, "prep_channel:" + channel}); err != nil {
			return SourceDraftPublishPrepResult{}, err
		}
		return sourceDraftPublishPrepResultFromObservation(s, draftObservationID, draft, shape, existingObservation, true)
	}

	ids := buildSourceDraftPrepIDs(channel, draftObservationID)
	notes := buildSourceDraftPrepLaneNotes(draft, draftObservationID, shape)

	proposal, err := ensureSourceDraftPrepProposal(s, ids.ProposalID, channel, draftObservationID, draft, shape, notes)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	workItem, err := ensureSourceDraftPrepWorkItem(s, proposal, ids.WorkItemID)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	approval, err := ensureSourceDraftPrepApproval(s, workItem, ids.ApprovalID, proposal.Title)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	flow, err := s.SyncFlow(ids.FlowID, workItem.ID, &approval.ApprovalID, nil, nil, nil, nil, nil, nil, notes)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	observation, err := ensureSourceDraftPrepObservation(s, draftObservation, draftObservationID, draft, shape, proposal, workItem, approval, flow, notes)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	if err := s.resolveSourceIntakeRedPolicyReview(draft.SourceObservationID, "publish prep opened or reused the downstream lane for this source intake", "review_room.auto.source_intake_red_policy_cleared_by_prep", []string{"draft_seed:" + draftObservationID, "prep_channel:" + channel, "approval:" + approval.ApprovalID}); err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	return buildSourceDraftPublishPrepResult(draftObservationID, draft, shape, proposal, workItem, approval, flow, observation, false), nil
}

func normalizeSourceDraftPrepChannel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", defaultDraftPrepChannel:
		return defaultDraftPrepChannel
	default:
		return strings.TrimSpace(strings.ToLower(value))
	}
}

func buildSourceDraftPrepShape(draft SourceDraftSeedRecord, channel string) (sourceDraftPrepShape, error) {
	channel = normalizeSourceDraftPrepChannel(channel)
	if channel != defaultDraftPrepChannel {
		return sourceDraftPrepShape{}, fmt.Errorf("unsupported publish prep channel: %s", channel)
	}
	subject := sourceDraftPrepSubject(draft)
	body := joinSourceDraftParagraphs(
		sourceDraftPrepLead(draft.TargetAccount, subject, sourceDraftSeedFocusCue(SourceIntakeRecord{
			DomainTags:    draft.DomainTags,
			WorldviewTags: draft.WorldviewTags,
		})),
		sourceDraftPrepExcerpt(draft),
		sourceDraftPrepFounderCue(draft),
		sourceDraftPrepRevisionCue(draft),
	)
	return sourceDraftPrepShape{
		Channel:     channel,
		Title:       subject,
		Body:        body,
		BodyPreview: sourceIntakeLimitText(strings.Join(strings.Fields(body), " "), 180),
		PrepShapeID: sourceDraftThreadsPrepShape,
	}, nil
}

func sourceDraftPrepSubject(draft SourceDraftSeedRecord) string {
	if value := strings.Join(strings.Fields(strings.TrimSpace(draft.SourceTitle)), " "); value != "" {
		return sourceIntakeLimitText(value, 92)
	}
	title := strings.Join(strings.Fields(strings.TrimSpace(draft.Title)), " ")
	for _, prefix := range []string{"97layer raw draft · ", "우순호 draft · ", "우화해 draft · ", "draft · "} {
		if strings.HasPrefix(strings.ToLower(title), strings.ToLower(prefix)) {
			title = strings.TrimSpace(title[len(prefix):])
			break
		}
	}
	if title != "" {
		return sourceIntakeLimitText(title, 92)
	}
	if value := strings.Join(strings.Fields(strings.TrimSpace(draft.SourceURL)), " "); value != "" {
		return sourceIntakeLimitText(value, 92)
	}
	return "source note"
}

func sourceDraftPrepExcerpt(draft SourceDraftSeedRecord) string {
	founderNote := strings.Join(strings.Fields(strings.TrimSpace(draft.FounderNote)), " ")
	revisionNote := strings.Join(strings.Fields(strings.TrimSpace(draft.RevisionNote)), " ")
	paragraphs := splitSourceDraftParagraphs(draft.Draft)
	longest := ""
	for _, paragraph := range paragraphs {
		collapsed := strings.Join(strings.Fields(paragraph), " ")
		if collapsed == "" {
			continue
		}
		if sourceDraftPrepIsMetaParagraph(collapsed) {
			continue
		}
		if founderNote != "" && collapsed == founderNote {
			continue
		}
		if revisionNote != "" && strings.Contains(collapsed, revisionNote) {
			continue
		}
		if len(collapsed) > len(longest) {
			longest = collapsed
		}
	}
	return longest
}

func sourceDraftPrepIsMetaParagraph(paragraph string) bool {
	paragraph = strings.Join(strings.Fields(strings.TrimSpace(paragraph)), " ")
	return strings.HasPrefix(paragraph, "요즘 붙들고 있는 건") ||
		strings.Contains(paragraph, "우순호 쪽으로 옮기면 결국") ||
		strings.Contains(paragraph, "우화해 쪽으로 옮기면 결국") ||
		strings.HasPrefix(paragraph, "97layer에서는") ||
		strings.HasPrefix(paragraph, "우순호에서는") ||
		strings.HasPrefix(paragraph, "우화해에서는") ||
		strings.Contains(paragraph, "기록에 가까운 메모로 둔다") ||
		strings.Contains(paragraph, "더 정리해본다") ||
		strings.Contains(paragraph, "더 다듬는다")
}

func sourceDraftPrepLead(targetAccount string, subject string, focusCue string) string {
	switch normalizeSourceDraftTarget(targetAccount) {
	case "97layer":
		return fmt.Sprintf("요즘 자꾸 다시 보게 되는 건 %s 안에 들어 있는 %s 쪽이다.", subject, focusCue)
	case "woosunhokr":
		return fmt.Sprintf("%s를 보다 보면 결국 %s에 가까운 순간이 사람에게 어떻게 닿는지가 먼저 남는다.", subject, focusCue)
	case "woohwahae":
		return fmt.Sprintf("%s에서 덜어내고 남는 건 결국 %s에 가까운 리듬이 생활 안으로 번지는 순간이다.", subject, focusCue)
	default:
		return fmt.Sprintf("%s를 다시 보면 결국 %s가 먼저 남는다.", subject, focusCue)
	}
}

func sourceDraftPrepFounderCue(draft SourceDraftSeedRecord) string {
	note := strings.Join(strings.Fields(strings.TrimSpace(draft.FounderNote)), " ")
	if note == "" {
		return ""
	}
	switch normalizeSourceDraftTarget(draft.TargetAccount) {
	case "97layer":
		return sourceIntakeLimitText("이번엔 "+note+"라는 축을 더 또렷하게 적어둔다.", 160)
	case "woosunhokr":
		return sourceIntakeLimitText("이 소재는 "+note+"라는 결로 더 눌러볼 만하다.", 160)
	case "woohwahae":
		return sourceIntakeLimitText("이 소재는 "+note+"라는 결로 더 덜어낼 만하다.", 160)
	default:
		return sourceIntakeLimitText(note, 160)
	}
}

func sourceDraftPrepRevisionCue(draft SourceDraftSeedRecord) string {
	note := strings.Join(strings.Fields(strings.TrimSpace(draft.RevisionNote)), " ")
	if note == "" {
		return ""
	}
	switch normalizeSourceDraftTarget(draft.TargetAccount) {
	case "97layer":
		return sourceIntakeLimitText("이번 메모는 "+note+"라는 요청을 따라 다시 눌러본 버전이다.", 170)
	case "woosunhokr":
		return sourceIntakeLimitText("이번 버전은 "+note+"라는 요청을 따라 더 정제해본다.", 170)
	case "woohwahae":
		return sourceIntakeLimitText("이번 버전은 "+note+"라는 요청을 따라 한 번 더 덜어낸다.", 170)
	default:
		return sourceIntakeLimitText("이번 버전은 "+note+"라는 요청을 따라 다시 잡는다.", 170)
	}
}

func splitSourceDraftParagraphs(value string) []string {
	raw := strings.ReplaceAll(value, "\r\n", "\n")
	parts := strings.Split(raw, "\n\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func buildSourceDraftPrepIDs(channel string, draftObservationID string) sourceDraftPrepIDs {
	slug := sourceDraftPrepSlugSanitizer.ReplaceAllString(strings.ToLower(strings.TrimSpace(channel+"-"+draftObservationID)), "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 56 {
		slug = strings.Trim(slug[:56], "-")
	}
	if slug == "" {
		slug = "source-draft-prep"
	}
	return sourceDraftPrepIDs{
		ProposalID: "proposal_brand_prep_" + slug,
		WorkItemID: "work_brand_prep_" + slug,
		ApprovalID: "approval_brand_prep_" + slug,
		FlowID:     "flow_brand_prep_" + slug,
	}
}

func buildSourceDraftPrepLaneNotes(draft SourceDraftSeedRecord, draftObservationID string, shape sourceDraftPrepShape) []string {
	notes := []string{
		brandPublishLaneNote,
		"channel:" + shape.Channel,
		"account:" + draft.TargetAccount,
		"tone:" + sourceDraftSeedToneLevel(draft.TargetAccount),
		"source_draft_seed:" + draftObservationID,
		"prep_shape:" + shape.PrepShapeID,
	}
	if draft.SourceObservationID != "" {
		notes = append(notes, "source_observation:"+draft.SourceObservationID)
	}
	if draft.RouteDecisionObservationID != "" {
		notes = append(notes, "route_decision:"+draft.RouteDecisionObservationID)
	}
	if draft.RevisionNote != "" {
		notes = append(notes, "revision:"+sourceIntakeLimitText(draft.RevisionNote, 80))
	}
	return uniqueSourceDraftPrepStrings(notes)
}

func ensureSourceDraftPrepProposal(s *Service, expectedID string, channel string, draftObservationID string, draft SourceDraftSeedRecord, shape sourceDraftPrepShape, notes []string) (ProposalItem, error) {
	if proposal, ok := findSourceDraftPrepProposal(s, draftObservationID, channel); ok {
		return proposal, nil
	}
	accountLabel := sourceDraftSeedTargetLabel(draft.TargetAccount)
	now := zeroSafeNow()
	proposal := ProposalItem{
		ProposalID: expectedID,
		Title:      sourceDraftPrepProposalTitle(shape.Channel, accountLabel, shape.Title),
		Intent:     sourceDraftPrepIntent(shape.Channel),
		Summary:    sourceDraftPrepProposalSummary(shape.Channel, accountLabel, shape.Title, shape.Body),
		Surface:    SurfaceCockpit,
		Priority:   "high",
		Risk:       "medium",
		Status:     "proposed",
		Notes:      append([]string{}, notes...),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.CreateProposal(proposal); err != nil {
		return ProposalItem{}, err
	}
	return proposal, nil
}

func ensureSourceDraftPrepWorkItem(s *Service, proposal ProposalItem, expectedID string) (WorkItem, error) {
	if work, ok := findSourceDraftPrepWorkItem(s, proposal, expectedID); ok {
		return work, nil
	}
	if proposal.Status != "proposed" {
		return WorkItem{}, fmt.Errorf("brand prep proposal is not open")
	}
	updatedProposal, workItem, err := s.PromoteProposal(proposal.ProposalID, expectedID)
	if err != nil {
		return WorkItem{}, err
	}
	proposal = updatedProposal
	return workItem, nil
}

func ensureSourceDraftPrepApproval(s *Service, workItem WorkItem, expectedID string, summary string) (ApprovalItem, error) {
	if approval, ok := findSourceDraftPrepApproval(s, workItem.ID, expectedID); ok {
		return approval, nil
	}
	item := ApprovalItem{
		ApprovalID:      expectedID,
		WorkItemID:      workItem.ID,
		Stage:           StageVerify,
		Summary:         summary,
		Risks:           []string{"Threads wording needs founder approval", "publish only after the draft matches tone and proof"},
		RollbackPlan:    "do not publish; revise the draft and reopen approval",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}
	if err := s.CreateApproval(item); err != nil {
		return ApprovalItem{}, err
	}
	return item, nil
}

func ensureSourceDraftPrepObservation(s *Service, draftObservation ObservationRecord, draftObservationID string, draft SourceDraftSeedRecord, shape sourceDraftPrepShape, proposal ProposalItem, workItem WorkItem, approval ApprovalItem, flow FlowRun, notes []string) (ObservationRecord, error) {
	if observation, ok := findSourceDraftPrepObservation(s, draftObservationID, shape.Channel); ok {
		return observation, nil
	}
	refs := mergeObservationRefs(
		approval.ApprovalID,
		proposal.ProposalID,
		workItem.ID,
		flow.FlowID,
		"",
		"",
		sourceDraftPrepExternalRefs(draftObservationID, draft),
	)
	refs = appendUniqueString(refs, "source_draft_seed:"+draftObservationID)
	observation, err := s.CreateObservation(ObservationRecord{
		SourceChannel:     draftObservation.SourceChannel,
		Actor:             draftObservation.Actor,
		Topic:             threadsBrandPrepTopic,
		Refs:              refs,
		Confidence:        "high",
		RawExcerpt:        buildSourceDraftPrepObservationRawExcerpt(shape, draft, notes),
		NormalizedSummary: sourceDraftPrepProposalSummary(shape.Channel, sourceDraftSeedTargetLabel(draft.TargetAccount), shape.Title, shape.Body),
		CapturedAt:        zeroSafeNow(),
	})
	if err != nil {
		return ObservationRecord{}, err
	}
	return observation, nil
}

func sourceDraftPrepExternalRefs(draftObservationID string, draft SourceDraftSeedRecord) []string {
	refs := []string{draftObservationID}
	if draft.SourceObservationID != "" {
		refs = append(refs, draft.SourceObservationID)
	}
	if draft.RouteDecisionObservationID != "" {
		refs = append(refs, draft.RouteDecisionObservationID)
	}
	if draft.ParentDraftObservationID != "" {
		refs = append(refs, draft.ParentDraftObservationID)
	}
	return uniqueSourceDraftPrepStrings(refs)
}

func buildSourceDraftPrepObservationRawExcerpt(shape sourceDraftPrepShape, draft SourceDraftSeedRecord, notes []string) string {
	lines := []string{
		"channel=" + shape.Channel,
		"target_account=" + draft.TargetAccount,
		"title=" + shape.Title,
		"sources=" + sourceDraftPrepSourcesValue(draft),
	}
	if len(notes) > 0 {
		lines = append(lines, "notes="+strings.Join(uniqueSourceDraftPrepStrings(notes), ","))
	}
	lines = append(lines, "draft:", shape.Body)
	return strings.Join(lines, "\n")
}

func sourceDraftPrepSourcesValue(draft SourceDraftSeedRecord) string {
	sourceIDs := []string{}
	if draft.SourceObservationID != "" {
		sourceIDs = append(sourceIDs, draft.SourceObservationID)
	}
	sourceIDs = uniqueSourceDraftPrepStrings(sourceIDs)
	if len(sourceIDs) == 0 {
		return "none"
	}
	return strings.Join(sourceIDs, ",")
}

func sourceDraftPrepProposalTitle(channel string, accountLabel string, title string) string {
	return sourceIntakeLimitText(fmt.Sprintf("Brand publish · %s · %s · %s", accountLabel, sourceDraftPrepChannelLabel(channel), strings.TrimSpace(title)), 160)
}

func sourceDraftPrepIntent(channel string) string {
	return sourceIntakeLimitText(fmt.Sprintf("prepare %s publication from a source draft seed", normalizeSourceDraftPrepChannel(channel)), 180)
}

func sourceDraftPrepProposalSummary(channel string, accountLabel string, title string, body string) string {
	return sourceIntakeLimitText(fmt.Sprintf("%s draft for %s is ready for founder review: %s. %s", sourceDraftPrepChannelLabel(channel), strings.TrimSpace(accountLabel), strings.TrimSpace(title), strings.Join(strings.Fields(strings.TrimSpace(body)), " ")), 180)
}

func sourceDraftPrepChannelLabel(channel string) string {
	switch normalizeSourceDraftPrepChannel(channel) {
	case "threads":
		return "Threads"
	default:
		return strings.TrimSpace(channel)
	}
}

func findSourceDraftPrepObservation(s *Service, draftObservationID string, channel string) (ObservationRecord, bool) {
	items := s.ListObservations(ObservationQuery{
		Topic: threadsBrandPrepTopic,
		Text:  "source_draft_seed:" + strings.TrimSpace(draftObservationID),
		Limit: 32,
	})
	for _, item := range items {
		draft, err := parseBrandThreadsDraft(item)
		if err != nil {
			continue
		}
		if normalizeSourceDraftPrepChannel(draft.Channel) == normalizeSourceDraftPrepChannel(channel) {
			return item, true
		}
	}
	return ObservationRecord{}, false
}

func findSourceDraftPrepProposal(s *Service, draftObservationID string, channel string) (ProposalItem, bool) {
	for _, item := range s.ListProposals() {
		if !proposalHasNote(item.Notes, brandPublishLaneNote) {
			continue
		}
		if proposalNoteValue(item.Notes, "source_draft_seed:") != strings.TrimSpace(draftObservationID) {
			continue
		}
		if proposalNoteValue(item.Notes, "channel:") != normalizeSourceDraftPrepChannel(channel) {
			continue
		}
		return item, true
	}
	return ProposalItem{}, false
}

func findSourceDraftPrepWorkItem(s *Service, proposal ProposalItem, expectedID string) (WorkItem, bool) {
	expectedID = strings.TrimSpace(expectedID)
	if proposal.PromotedWorkItemID != nil {
		if item, ok := findWorkItemByID(s, *proposal.PromotedWorkItemID); ok {
			return item, true
		}
	}
	if expectedID != "" {
		if item, ok := findWorkItemByID(s, expectedID); ok {
			return item, true
		}
	}
	for _, item := range s.ListWorkItems() {
		if strings.TrimSpace(item.CorrelationID) == proposal.ProposalID {
			return item, true
		}
	}
	return WorkItem{}, false
}

func findSourceDraftPrepApproval(s *Service, workItemID string, expectedID string) (ApprovalItem, bool) {
	expectedID = strings.TrimSpace(expectedID)
	for _, item := range s.ListApprovals() {
		if expectedID != "" && item.ApprovalID == expectedID {
			return item, true
		}
		if strings.TrimSpace(item.WorkItemID) == strings.TrimSpace(workItemID) {
			return item, true
		}
	}
	return ApprovalItem{}, false
}

func findSourceDraftPrepFlow(s *Service, workItemID string, expectedID string) (FlowRun, bool) {
	expectedID = strings.TrimSpace(expectedID)
	for _, item := range s.ListFlows() {
		if expectedID != "" && item.FlowID == expectedID {
			return item, true
		}
		if strings.TrimSpace(item.WorkItemID) == strings.TrimSpace(workItemID) {
			return item, true
		}
	}
	return FlowRun{}, false
}

func findWorkItemByID(s *Service, workItemID string) (WorkItem, bool) {
	for _, item := range s.ListWorkItems() {
		if item.ID == strings.TrimSpace(workItemID) {
			return item, true
		}
	}
	return WorkItem{}, false
}

func buildSourceDraftPublishPrepResult(draftObservationID string, draft SourceDraftSeedRecord, shape sourceDraftPrepShape, proposal ProposalItem, workItem WorkItem, approval ApprovalItem, flow FlowRun, observation ObservationRecord, reused bool) SourceDraftPublishPrepResult {
	targetAccountLabel := firstNonEmptyString(sourceDraftSeedTargetLabel(draft.TargetAccount), draft.TargetAccount, "account")
	channelLabel := firstNonEmptyString(sourceDraftPrepChannelLabel(shape.Channel), shape.Channel, "prep")
	laneTitle := firstNonEmptyString(shape.Title, draft.Title, "draft")
	stateLabel := "새로 연"
	if reused {
		stateLabel = "이미 열려 있던"
	}
	return SourceDraftPublishPrepResult{
		OpenedFromDraft: SourceDraftPublishPrepOpenedFromDraft{
			DraftObservationID: draftObservationID,
			TargetAccount:      draft.TargetAccount,
			TargetAccountLabel: sourceDraftSeedTargetLabel(draft.TargetAccount),
			TargetTone:         sourceDraftSeedToneLevel(draft.TargetAccount),
			Title:              shape.Title,
			BodyPreview:        shape.BodyPreview,
			RevisionNote:       draft.RevisionNote,
		},
		Lane: SourceDraftPublishPrepLane{
			Channel:            shape.Channel,
			Label:              sourceDraftPrepChannelLabel(shape.Channel),
			TargetAccount:      draft.TargetAccount,
			TargetAccountLabel: sourceDraftSeedTargetLabel(draft.TargetAccount),
			TargetToneLevel:    sourceDraftSeedToneLevel(draft.TargetAccount),
			Title:              shape.Title,
			BodyPreview:        shape.BodyPreview,
			SourceIDs:          splitCSV(sourceDraftPrepSourcesValue(draft)),
			ExternalRefs:       sourceDraftPrepExternalRefs(draftObservationID, draft),
			PrepShapeID:        shape.PrepShapeID,
		},
		Proposal:    proposal,
		WorkItem:    workItem,
		Approval:    approval,
		Flow:        flow,
		Observation: observation,
		Reused:      reused,
		FollowUp: SourceDraftPublishPrepFollowUp{
			Mode:                "monitor_prep_lane",
			Summary:             targetAccountLabel + " " + channelLabel + " 준비가 열렸습니다. 초안 다듬기와 founder review를 이어가세요.",
			Detail:              stateLabel + " prep lane · " + laneTitle,
			Ref:                 "prep:" + shape.Channel,
			SourceObservationID: draft.SourceObservationID,
			DraftObservationID:  draftObservationID,
			ActionLabel:         channelLabel + " 준비 lane 확인",
		},
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func sourceDraftPublishPrepResultFromObservation(s *Service, draftObservationID string, draft SourceDraftSeedRecord, fallbackShape sourceDraftPrepShape, observation ObservationRecord, reused bool) (SourceDraftPublishPrepResult, error) {
	parsed, err := parseBrandThreadsDraft(observation)
	if err != nil {
		return SourceDraftPublishPrepResult{}, err
	}
	shape := fallbackShape
	shape.Channel = normalizeSourceDraftPrepChannel(parsed.Channel)
	shape.Title = parsed.Title
	shape.Body = parsed.Body
	shape.BodyPreview = sourceIntakeLimitText(strings.Join(strings.Fields(parsed.Body), " "), 180)
	if shape.PrepShapeID == "" {
		shape.PrepShapeID = sourceDraftThreadsPrepShape
	}
	proposal, ok := findSourceDraftPrepProposal(s, draftObservationID, parsed.Channel)
	if !ok && parsed.ProposalID != "" {
		for _, item := range s.ListProposals() {
			if item.ProposalID == parsed.ProposalID {
				proposal = item
				ok = true
				break
			}
		}
	}
	if !ok {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("prep proposal not found")
	}
	workItem, ok := findSourceDraftPrepWorkItem(s, proposal, parsed.WorkItemID)
	if !ok {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("prep work item not found")
	}
	approval, ok := findSourceDraftPrepApproval(s, workItem.ID, parsed.ApprovalID)
	if !ok {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("prep approval not found")
	}
	flow, ok := findSourceDraftPrepFlow(s, workItem.ID, parsed.FlowID)
	if !ok {
		return SourceDraftPublishPrepResult{}, fmt.Errorf("prep flow not found")
	}
	return buildSourceDraftPublishPrepResult(draftObservationID, draft, shape, proposal, workItem, approval, flow, observation, reused), nil
}

func uniqueSourceDraftPrepStrings(items []string) []string {
	out := []string{}
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = appendUniqueString(out, trimmed)
		}
	}
	return out
}

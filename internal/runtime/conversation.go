package runtime

import (
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	conversationEventKind            = "conversation.note"
	conversationAutomationFailedKind = "conversation.automation_failed"
	conversationObservationTopic     = "conversation.note"
	conversationRiskSource           = "auto.conversation"
	conversationRiskKind             = "conversation_risk"
	conversationDedupeNotePrefix     = "conversation_dedupe:"
)

var (
	conversationEmailPattern    = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	conversationPhonePattern    = regexp.MustCompile(`(?m)(?:\+?\d[\d\-() ]{7,}\d)`)
	conversationResidentPattern = regexp.MustCompile(`\b\d{6}-?\d{7}\b`)
	conversationSecretPattern   = regexp.MustCompile(`(?i)\b(bearer|token|secret|password|passwd|api[_-]?key)\b\s*[:=]\s*[^\s,;]+`)
)

type ConversationNoteInput struct {
	ConversationID string     `json:"conversation_id,omitempty"`
	Text           string     `json:"text"`
	Actor          string     `json:"actor,omitempty"`
	SourceChannel  string     `json:"source_channel,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	Summary        string     `json:"summary,omitempty"`
	Refs           []string   `json:"refs,omitempty"`
	Confidence     string     `json:"confidence,omitempty"`
	CapturedAt     *time.Time `json:"captured_at,omitempty"`
	AutoPlan       *bool      `json:"auto_plan,omitempty"`
	AutoRisk       *bool      `json:"auto_risk,omitempty"`
	AutoDispatch   *bool      `json:"auto_dispatch,omitempty"`
	LLMTag         *bool      `json:"llm_tag,omitempty"`
}

type ConversationNoteResult struct {
	ConversationID string                     `json:"conversation_id"`
	Event          EventEnvelope              `json:"event"`
	Observation    ObservationRecord          `json:"observation"`
	Profile        ConversationProfile        `json:"profile"`
	Proposal       *ProposalItem              `json:"proposal,omitempty"`
	Job            *AgentJob                  `json:"job,omitempty"`
	Dispatch       *AgentDispatchResult       `json:"dispatch,omitempty"`
	ReviewItem     *ReviewRoomItem            `json:"review_item,omitempty"`
	Warnings       []string                   `json:"warnings,omitempty"`
	Toggles        ConversationRuntimeToggles `json:"toggles"`
}

type ConversationProfile struct {
	Actor               string   `json:"actor"`
	SourceChannel       string   `json:"source_channel"`
	FounderLike         bool     `json:"founder_like"`
	AutoRisk            bool     `json:"auto_risk"`
	AutoPlan            bool     `json:"auto_plan"`
	AutoDispatchPlanner bool     `json:"auto_dispatch_planner"`
	LLMTag              bool     `json:"llm_tag"`
	TrustMode           string   `json:"trust_mode"`
	Notes               []string `json:"notes"`
}

type ConversationRuntimeToggles struct {
	AutoPlanEnabled     bool `json:"auto_plan_enabled"`
	AutoRiskEnabled     bool `json:"auto_risk_enabled"`
	AutoDispatchEnabled bool `json:"auto_dispatch_enabled"`
	LLMTagEnabled       bool `json:"llm_tag_enabled"`
}

type ConversationAutomationStatus struct {
	Toggles              ConversationRuntimeToggles `json:"toggles"`
	ActiveProfile        ConversationProfile        `json:"active_profile"`
	RecentProcessedCount int                        `json:"recent_processed_count"`
	RecentFailureCount   int                        `json:"recent_failure_count"`
	LastFailureReasons   []string                   `json:"last_failure_reasons"`
	RecentRisks          []ReviewRoomItem           `json:"recent_risks"`
}

type AgentBehaviorProfile struct {
	Role                 string   `json:"role"`
	Mission              string   `json:"mission"`
	CriticalRules        []string `json:"critical_rules"`
	Workflow             []string `json:"workflow"`
	RequiredOutputs      []string `json:"required_outputs"`
	EvidenceRequirements []string `json:"evidence_requirements"`
}

func (i ConversationNoteInput) Validate() error {
	if strings.TrimSpace(i.Text) == "" {
		return errors.New("conversation note text is required")
	}
	if strings.TrimSpace(normalizeConversationSourceChannel(i.SourceChannel)) == "" {
		return errors.New("conversation note source_channel is required")
	}
	if confidence := normalizeObservationConfidence(i.Confidence); confidence == "" {
		return errors.New("conversation note confidence is invalid")
	}
	return nil
}

func (p AgentBehaviorProfile) Validate() error {
	if strings.TrimSpace(p.Role) == "" {
		return errors.New("agent behavior profile role is required")
	}
	if strings.TrimSpace(p.Mission) == "" {
		return errors.New("agent behavior profile mission is required")
	}
	if p.CriticalRules == nil || p.Workflow == nil || p.RequiredOutputs == nil || p.EvidenceRequirements == nil {
		return errors.New("agent behavior profile arrays are required")
	}
	return nil
}

func (s *Service) ConversationAutomationStatus() ConversationAutomationStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conversationAutomationStatusLocked("api")
}

func (s *Service) conversationAutomationStatusLocked(sourceChannel string) ConversationAutomationStatus {
	toggles := conversationRuntimeToggles()
	profile := conversationProfileForActor(s.currentActor(), sourceChannel, toggles)
	events := s.event.list()
	room := s.currentReviewRoomLocked()
	status := ConversationAutomationStatus{
		Toggles:            toggles,
		ActiveProfile:      profile,
		LastFailureReasons: []string{},
		RecentRisks:        []ReviewRoomItem{},
	}
	for i := len(events) - 1; i >= 0; i-- {
		item := events[i]
		switch strings.TrimSpace(item.Kind) {
		case conversationEventKind:
			status.RecentProcessedCount++
		case conversationAutomationFailedKind:
			status.RecentFailureCount++
			if len(status.LastFailureReasons) < 3 {
				if message, ok := item.Data["error"].(string); ok && strings.TrimSpace(message) != "" {
					status.LastFailureReasons = append(status.LastFailureReasons, strings.TrimSpace(message))
				}
			}
		}
	}
	for _, item := range room.Open {
		if strings.TrimSpace(item.Source) != conversationRiskSource {
			continue
		}
		status.RecentRisks = append(status.RecentRisks, cloneReviewRoomItem(item))
		if len(status.RecentRisks) == 3 {
			break
		}
	}
	return status
}

func (s *Service) CreateConversationNote(input ConversationNoteInput) (ConversationNoteResult, error) {
	override := normalizeActor(input.Actor)
	if override != "" && override != s.currentActorOverride() && override != s.currentActor() {
		var result ConversationNoteResult
		err := s.WithActor(override, func(inner *Service) error {
			var err error
			result, err = inner.createConversationNote(input)
			return err
		})
		return result, err
	}
	return s.createConversationNote(input)
}

func (s *Service) createConversationNote(input ConversationNoteInput) (ConversationNoteResult, error) {
	s.mu.Lock()
	result, dispatchJobID, shouldDispatch, err := s.createConversationNoteLocked(input)
	s.mu.Unlock()
	if err != nil {
		return ConversationNoteResult{}, err
	}
	if shouldDispatch && strings.TrimSpace(dispatchJobID) != "" {
		dispatch, dispatchErr := s.DispatchAgentJob(dispatchJobID)
		if dispatchErr != nil {
			result.Warnings = append(result.Warnings, "planner dispatch failed: "+strings.TrimSpace(dispatchErr.Error()))
			s.recordConversationAutomationFailure(dispatchJobID, dispatchErr)
			s.maybeNotifyFounderAttention()
			return result, nil
		}
		result.Dispatch = &dispatch
		result.Job = &dispatch.Job
	}
	s.maybeNotifyFounderAttention()
	return result, nil
}

func (s *Service) createConversationNoteLocked(input ConversationNoteInput) (ConversationNoteResult, string, bool, error) {
	input = normalizeConversationNoteInput(input, s.currentActor())
	if err := input.Validate(); err != nil {
		return ConversationNoteResult{}, "", false, err
	}

	toggles := conversationRuntimeToggles()
	effective := applyConversationOverrides(toggles, input)
	profile := conversationProfileForActor(s.currentActor(), input.SourceChannel, effective)
	maskedText := maskConversationText(input.Text)
	maskedSummary := input.Summary
	if strings.TrimSpace(maskedSummary) == "" {
		maskedSummary = deriveConversationSummary(maskedText)
	} else {
		maskedSummary = limitText(maskConversationText(maskedSummary), 180)
	}
	tags := classifyConversationTags(input.Tags, maskedText, maskedSummary)
	refs := conversationRefs(input.Refs, input.SourceChannel, tags)
	dedupeKey := conversationDedupeKey(profile.Actor, input.SourceChannel, maskedSummary, refs)
	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		conversationID = nextConversationID(zeroSafeNow())
	}
	capturedAt := zeroSafeNow()
	if input.CapturedAt != nil && !input.CapturedAt.IsZero() {
		capturedAt = input.CapturedAt.UTC()
	}

	oldObservations := s.observation.list()
	oldProposals := s.proposal.list()
	oldJobs := s.agentJob.list()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()

	event := newEvent(conversationEventKind, s.currentActor(), surfaceForConversationSource(input.SourceChannel), conversationID, StageDiscover, map[string]any{
		"text":              maskedText,
		"summary":           maskedSummary,
		"tags":              append([]string{}, tags...),
		"refs":              append([]string{}, refs...),
		"confidence":        input.Confidence,
		"source_channel":    input.SourceChannel,
		"dedupe_key":        dedupeKey,
		"auto_generated":    true,
		"llm_tag_requested": effective.LLMTagEnabled,
		"llm_tag_applied":   false,
	})
	if err := s.event.append(event); err != nil {
		return ConversationNoteResult{}, "", false, err
	}

	observation, err := s.recordObservationLocked(ObservationRecord{
		SourceChannel:     conversationObservationChannel(input.SourceChannel),
		Actor:             profile.Actor,
		Topic:             conversationObservationTopic,
		Refs:              append(refs, "event:"+event.EventID, conversationDedupeNotePrefix+dedupeKey),
		Confidence:        input.Confidence,
		RawExcerpt:        maskedText,
		NormalizedSummary: maskedSummary,
		CapturedAt:        capturedAt,
	}, false)
	if err != nil {
		s.observation = newObservationStore(oldObservations)
		s.proposal = newProposalStore(oldProposals)
		s.agentJob = newAgentJobStore(oldJobs)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return ConversationNoteResult{}, "", false, err
	}

	result := ConversationNoteResult{
		ConversationID: conversationID,
		Event:          event,
		Observation:    observation,
		Profile:        profile,
		Warnings:       []string{},
		Toggles:        effective,
	}
	if effective.LLMTagEnabled {
		result.Warnings = append(result.Warnings, "LLM tagging requested; v1 keeps rule-based tagging only")
	}

	if shouldAutoConversationRisk(tags, maskedText, maskedSummary, effective) {
		reviewItem, reviewErr := s.upsertConversationRiskReviewItemLocked(dedupeKey, maskedSummary, event, profile, refs, len(tags) > 0)
		if reviewErr != nil {
			s.observation = newObservationStore(oldObservations)
			s.proposal = newProposalStore(oldProposals)
			s.agentJob = newAgentJobStore(oldJobs)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return ConversationNoteResult{}, "", false, reviewErr
		}
		result.ReviewItem = &reviewItem
	}

	dispatchJobID := ""
	shouldDispatch := false
	if shouldAutoConversationPlan(tags, maskedText, maskedSummary, refs, effective) {
		proposal, proposalCreated, proposalErr := s.upsertConversationProposalLocked(dedupeKey, maskedSummary, refs, profile, event, tags)
		if proposalErr != nil {
			s.observation = newObservationStore(oldObservations)
			s.proposal = newProposalStore(oldProposals)
			s.agentJob = newAgentJobStore(oldJobs)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return ConversationNoteResult{}, "", false, proposalErr
		}
		result.Proposal = &proposal
		job, jobCreated, jobErr := s.upsertConversationPlannerJobLocked(dedupeKey, proposal, maskedSummary, refs, profile, event, tags)
		if jobErr != nil {
			s.observation = newObservationStore(oldObservations)
			s.proposal = newProposalStore(oldProposals)
			s.agentJob = newAgentJobStore(oldJobs)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return ConversationNoteResult{}, "", false, jobErr
		}
		result.Job = &job
		shouldDispatch = proposalCreated || jobCreated
		if effective.AutoDispatchEnabled && profile.AutoDispatchPlanner && strings.TrimSpace(job.Status) == "queued" {
			dispatchJobID = job.JobID
			shouldDispatch = true
		}
	}

	if err := s.persistLocked(); err != nil {
		s.observation = newObservationStore(oldObservations)
		s.proposal = newProposalStore(oldProposals)
		s.agentJob = newAgentJobStore(oldJobs)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return ConversationNoteResult{}, "", false, err
	}
	return result, dispatchJobID, shouldDispatch, nil
}

func (s *Service) recordConversationAutomationFailure(jobID string, err error) {
	if strings.TrimSpace(jobID) == "" || err == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if appendErr := s.event.append(s.newSystemEvent(conversationAutomationFailedKind, map[string]any{
		"job_id": jobID,
		"error":  strings.TrimSpace(err.Error()),
	})); appendErr == nil {
		_ = s.persistLocked()
	}
}

func normalizeConversationNoteInput(input ConversationNoteInput, fallbackActor string) ConversationNoteInput {
	input.Text = limitText(strings.TrimSpace(maskConversationText(input.Text)), 500)
	input.Actor = ResolveActor(input.Actor, fallbackActor)
	input.SourceChannel = normalizeConversationSourceChannel(input.SourceChannel)
	input.Tags = normalizeConversationTags(input.Tags)
	input.Summary = limitText(strings.TrimSpace(maskConversationText(input.Summary)), 180)
	input.Refs = normalizeObservationRefs(input.Refs)
	input.Confidence = normalizeObservationConfidence(input.Confidence)
	if input.Confidence == "" {
		input.Confidence = "medium"
	}
	return input
}

func normalizeConversationSourceChannel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "conversation:")
	switch value {
	case "", "api":
		return "api"
	case "cli", "shell":
		return "terminal"
	case "tg":
		return "telegram"
	default:
		return value
	}
}

func normalizeConversationTags(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(strings.ToLower(item))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func conversationRuntimeToggles() ConversationRuntimeToggles {
	return ConversationRuntimeToggles{
		AutoPlanEnabled:     conversationEnvEnabled("LAYER_OS_CONVERSATION_AUTOPLAN", true),
		AutoRiskEnabled:     conversationEnvEnabled("LAYER_OS_CONVERSATION_AUTORISK", true),
		AutoDispatchEnabled: conversationEnvEnabled("LAYER_OS_CONVERSATION_AUTODISPATCH", false),
		LLMTagEnabled:       conversationEnvEnabled("LAYER_OS_CONVERSATION_LLM_TAG", false),
	}
}

func conversationEnvEnabled(key string, defaultValue bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "", "default":
		return defaultValue
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func applyConversationOverrides(base ConversationRuntimeToggles, input ConversationNoteInput) ConversationRuntimeToggles {
	if input.AutoPlan != nil {
		base.AutoPlanEnabled = *input.AutoPlan
	}
	if input.AutoRisk != nil {
		base.AutoRiskEnabled = *input.AutoRisk
	}
	if input.AutoDispatch != nil {
		base.AutoDispatchEnabled = *input.AutoDispatch
	}
	if input.LLMTag != nil {
		base.LLMTagEnabled = *input.LLMTag
	}
	return base
}

func ConversationNoteInputFromObservation(item ObservationRecord, tags []string) ConversationNoteInput {
	item = normalizeObservationRecord(item, item.Actor)
	text := strings.TrimSpace(item.RawExcerpt)
	if text == "" {
		text = strings.TrimSpace(item.NormalizedSummary)
	}
	input := ConversationNoteInput{
		Text:          text,
		Actor:         item.Actor,
		SourceChannel: item.SourceChannel,
		Tags:          append([]string{}, tags...),
		Summary:       strings.TrimSpace(item.NormalizedSummary),
		Refs:          append([]string{}, item.Refs...),
		Confidence:    item.Confidence,
	}
	if !item.CapturedAt.IsZero() {
		capturedAt := item.CapturedAt.UTC()
		input.CapturedAt = &capturedAt
	}
	return normalizeConversationNoteInput(input, item.Actor)
}

func conversationProfileForActor(actor string, sourceChannel string, toggles ConversationRuntimeToggles) ConversationProfile {
	actor = ResolveActor(actor)
	sourceChannel = normalizeConversationSourceChannel(sourceChannel)
	founderLike := !isRuntimeAgentActor(actor)
	notes := []string{"masked-before-persist", "summary-only knowledge propagation"}
	if founderLike {
		notes = append(notes, "founder-like channel keeps planner auto-dispatch eligible")
	} else {
		notes = append(notes, "non-founder actor stays approval-gated beyond planner draft")
	}
	if !toggles.LLMTagEnabled {
		notes = append(notes, "rule-based tagging only")
	}
	return ConversationProfile{
		Actor:               actor,
		SourceChannel:       sourceChannel,
		FounderLike:         founderLike,
		AutoRisk:            toggles.AutoRiskEnabled,
		AutoPlan:            toggles.AutoPlanEnabled,
		AutoDispatchPlanner: toggles.AutoDispatchEnabled && founderLike,
		LLMTag:              toggles.LLMTagEnabled,
		TrustMode:           conversationTrustMode(founderLike, sourceChannel),
		Notes:               notes,
	}
}

func conversationTrustMode(founderLike bool, sourceChannel string) string {
	if !founderLike {
		return "scoped_agent"
	}
	switch normalizeConversationSourceChannel(sourceChannel) {
	case "telegram":
		return "founder_inbox"
	case "session":
		return "session_handoff"
	default:
		return "founder_operator"
	}
}

func isRuntimeAgentActor(actor string) bool {
	switch normalizeActor(actor) {
	case "codex", "claude", "gemini", "designer", "planner", "implementer", "verifier":
		return true
	default:
		return false
	}
}

func maskConversationText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	masked := conversationEmailPattern.ReplaceAllString(text, "[redacted-email]")
	masked = conversationResidentPattern.ReplaceAllString(masked, "[redacted-sensitive-id]")
	masked = conversationPhonePattern.ReplaceAllString(masked, "[redacted-phone]")
	masked = conversationSecretPattern.ReplaceAllStringFunc(masked, func(raw string) string {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) + ": [redacted-secret]"
		}
		parts = strings.SplitN(raw, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) + "=[redacted-secret]"
		}
		return "[redacted-secret]"
	})
	return limitText(masked, 500)
}

func deriveConversationSummary(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "conversation note"
	}
	separators := []string{"\n", ". ", "! ", "? ", " | "}
	for _, separator := range separators {
		if idx := strings.Index(text, separator); idx > 0 {
			return limitText(strings.TrimSpace(text[:idx]), 180)
		}
	}
	return limitText(text, 180)
}

func classifyConversationTags(tags []string, text string, summary string) []string {
	items := normalizeConversationTags(tags)
	add := func(tag string) {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" || containsExactString(items, tag) {
			return
		}
		items = append(items, tag)
	}
	combined := strings.ToLower(strings.TrimSpace(text + "\n" + summary))
	if containsAnyFragment(combined, []string{"risk", "issue", "blocker", "critical", "security", "vuln", "failure", "취약", "보안", "리스크", "위험", "깨", "막힘", "오류"}) {
		add("risk")
	}
	if containsAnyFragment(combined, []string{"idea", "proposal", "should", "could", "consider", "아이디어", "제안", "생각", "하면 좋"}) {
		add("idea")
	}
	if containsAnyFragment(combined, []string{"todo", "next", "action", "implement", "build", "add", "wire", "fix", "해야", "할 일", "구현", "추가", "수정", "진행"}) {
		add("todo")
	}
	if containsAnyFragment(combined, []string{"concern", "wrong", "mismatch", "review", "critique", "problem", "문제", "비판", "우려", "불일치", "검토"}) {
		add("critique")
	}
	sort.Strings(items)
	return items
}

func conversationRefs(refs []string, sourceChannel string, tags []string) []string {
	items := normalizeObservationRefs(refs)
	items = append(items, ObservationMetadataRefs(sourceChannel, conversationObservationTopic, map[string]string{"source_channel": sourceChannel})...)
	for _, tag := range tags {
		items = append(items, "tag:"+strings.TrimSpace(tag))
	}
	return normalizeObservationRefs(items)
}

func conversationDedupeKey(actor string, sourceChannel string, summary string, refs []string) string {
	sortedRefs := append([]string{}, refs...)
	sort.Strings(sortedRefs)
	payload := strings.Join([]string{ResolveActor(actor), normalizeConversationSourceChannel(sourceChannel), strings.ToLower(strings.TrimSpace(summary)), strings.Join(sortedRefs, "|")}, "||")
	return fmt.Sprintf("conv_%08x", crc32.ChecksumIEEE([]byte(payload)))
}

func nextConversationID(now interface{ UnixNano() int64 }) string {
	return "conversation_" + fmt.Sprintf("%d", now.UnixNano())
}

func conversationObservationChannel(sourceChannel string) string {
	return "conversation:" + normalizeConversationSourceChannel(sourceChannel)
}

func surfaceForConversationSource(sourceChannel string) Surface {
	return surfaceForObservationChannel(normalizeConversationSourceChannel(sourceChannel))
}

func shouldAutoConversationRisk(tags []string, text string, summary string, toggles ConversationRuntimeToggles) bool {
	if !toggles.AutoRiskEnabled {
		return false
	}
	for _, tag := range tags {
		if tag == "risk" || tag == "critique" {
			return true
		}
	}
	combined := strings.ToLower(strings.TrimSpace(text + "\n" + summary))
	return containsAnyFragment(combined, []string{"risk", "security", "vuln", "critical", "blocker", "취약", "보안", "위험", "리스크", "막힘"})
}

func shouldAutoConversationPlan(tags []string, text string, summary string, refs []string, toggles ConversationRuntimeToggles) bool {
	if !toggles.AutoPlanEnabled {
		return false
	}
	_ = refs
	for _, tag := range tags {
		if tag == "idea" || tag == "todo" {
			return true
		}
	}
	combined := strings.ToLower(strings.TrimSpace(text + "\n" + summary))
	if containsAnyFragment(combined, []string{"todo", "next", "action", "implement", "build", "add", "wire", "fix", "해야", "할 일", "구현", "추가", "수정", "진행"}) {
		return true
	}
	return false
}

func (s *Service) upsertConversationProposalLocked(dedupeKey string, summary string, refs []string, profile ConversationProfile, event EventEnvelope, tags []string) (ProposalItem, bool, error) {
	for _, item := range s.proposal.list() {
		if proposalHasConversationDedupe(item, dedupeKey) {
			return item, false, nil
		}
	}
	proposalID := deriveProposalID(dedupeKey)
	intent := "turn a captured conversation signal into an explicit founder-visible proposal with assumptions and next actions"
	notes := []string{
		"auto_generated:true",
		conversationDedupeNotePrefix + dedupeKey,
		"source_event:" + event.EventID,
		"source_channel:" + profile.SourceChannel,
	}
	for _, ref := range refs {
		notes = append(notes, "ref:"+ref)
	}
	proposal := ProposalItem{
		ProposalID: proposalID,
		Title:      limitText("Conversation follow-up: "+summary, 80),
		Intent:     intent,
		Summary:    limitText(summary, 180),
		Surface:    surfaceForConversationSource(profile.SourceChannel),
		Priority:   conversationProposalPriority(tags),
		Risk:       conversationProposalRisk(tags),
		Status:     "proposed",
		Notes:      normalizeReviewRoomEvidence(notes),
	}
	created, err := s.createProposalLocked(proposal)
	if err != nil {
		return ProposalItem{}, false, err
	}
	return created, true, nil
}

func (s *Service) upsertConversationPlannerJobLocked(dedupeKey string, proposal ProposalItem, summary string, refs []string, profile ConversationProfile, event EventEnvelope, tags []string) (AgentJob, bool, error) {
	for _, item := range s.agentJob.list() {
		if jobHasConversationDedupe(item, dedupeKey) {
			return item, false, nil
		}
	}
	ref := proposal.ProposalID
	roleProfile := agentBehaviorProfileForRole("planner")
	payload := map[string]any{
		"auto_generated":       true,
		"conversation_summary": limitText(summary, 180),
		"conversation_tags":    append([]string{}, tags...),
		"conversation_refs":    append([]string{}, refs...),
		"source_event_id":      event.EventID,
		"source_channel":       profile.SourceChannel,
		"role_profile":         agentBehaviorProfilePayload(roleProfile),
		"assumptions": []string{
			"conversation note is masked before persist",
			"planner must not assume implementation approval",
		},
		"options": []string{
			"propose the smallest safe implementation slice",
			"surface explicit risks and dependencies",
			"defer unresolved execution blockers to review-room when needed",
		},
		"instructions": conversationPlannerInstructions(summary, refs, roleProfile),
	}
	job := AgentJob{
		JobID:   "job_" + strings.TrimPrefix(dedupeKey, "conv_"),
		Kind:    "plan",
		Role:    "planner",
		Summary: limitText("Plan conversation follow-up: "+summary, 120),
		Status:  "queued",
		Source:  "conversation.auto",
		Surface: surfaceForConversationSource(profile.SourceChannel),
		Stage:   StageDiscover,
		Ref:     &ref,
		Payload: payload,
		Notes: normalizeReviewRoomEvidence([]string{
			"auto_generated:true",
			conversationDedupeNotePrefix + dedupeKey,
			"source_event:" + event.EventID,
			"source_channel:" + profile.SourceChannel,
		}),
	}
	created, err := s.createAgentJobLocked(job)
	if err != nil {
		return AgentJob{}, false, err
	}
	return created, true, nil
}

func (s *Service) upsertConversationRiskReviewItemLocked(dedupeKey string, summary string, event EventEnvelope, profile ConversationProfile, refs []string, escalate bool) (ReviewRoomItem, error) {
	room := s.currentReviewRoomLocked()
	evidence := normalizeReviewRoomEvidence(append([]string{
		conversationDedupeNotePrefix + dedupeKey,
		"event:" + event.EventID,
		"actor:" + profile.Actor,
		"source_channel:" + profile.SourceChannel,
	}, refs...))
	for index, item := range room.Open {
		if strings.TrimSpace(item.Source) != conversationRiskSource || !reviewRoomItemHasEvidence(item, conversationDedupeNotePrefix+dedupeKey) {
			continue
		}
		updated := cloneReviewRoomItem(item)
		updated.Evidence = normalizeReviewRoomEvidence(append(updated.Evidence, evidence...))
		updated.UpdatedAt = zeroSafeNow()
		if escalate {
			updated.Severity = escalateConversationRiskSeverity(updated.Severity)
		}
		room.Open[index] = normalizeReviewRoomItem(updated)
		room.UpdatedAt = &updated.UpdatedAt
		s.reviewRoom = newReviewRoomStore(room)
		_ = s.event.append(s.newSystemEvent("conversation.risk_updated", map[string]any{
			"item":       updated.Text,
			"severity":   updated.Severity,
			"dedupe_key": dedupeKey,
			"evidence":   append([]string{}, updated.Evidence...),
		}))
		return reviewRoomItemFromSection(room, "open", updated.Text), nil
	}
	ref := event.EventID
	why := "conversation risk requires founder review before downstream automation depends on the note"
	item := ReviewRoomItem{
		Text:          limitText("Conversation risk: "+summary, 180),
		Kind:          conversationRiskKind,
		Severity:      "high",
		Source:        conversationRiskSource,
		Ref:           &ref,
		WhyUnresolved: &why,
		Rationale: &ReviewRoomRationale{
			Trigger: conversationEventKind,
			Reason:  why,
			Rule:    "review_room.auto.conversation_risk",
		},
		Evidence: evidence,
	}
	if err := s.autoOpenReviewRoomItemLocked(item); err != nil {
		return ReviewRoomItem{}, err
	}
	return reviewRoomItemFromSection(s.reviewRoom.current(), "open", item.Text), nil
}

func escalateConversationRiskSeverity(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "low", "medium":
		return "high"
	case "high":
		return "critical"
	default:
		return value
	}
}

func proposalHasConversationDedupe(item ProposalItem, dedupeKey string) bool {
	for _, note := range item.Notes {
		if strings.TrimSpace(note) == conversationDedupeNotePrefix+dedupeKey {
			return true
		}
	}
	return false
}

func jobHasConversationDedupe(item AgentJob, dedupeKey string) bool {
	for _, note := range item.Notes {
		if strings.TrimSpace(note) == conversationDedupeNotePrefix+dedupeKey {
			return true
		}
	}
	return false
}

func conversationPlannerInstructions(summary string, refs []string, profile *AgentBehaviorProfile) string {
	parts := []string{}
	if profile != nil && strings.TrimSpace(profile.Mission) != "" {
		parts = append(parts, profile.Mission)
	}
	if profile != nil && len(profile.CriticalRules) > 0 {
		parts = append(parts, "Critical rules: "+strings.Join(profile.CriticalRules, "; "))
	}
	if len(refs) > 0 {
		parts = append(parts, "Refs: "+strings.Join(refs, ", "))
	}
	parts = append(parts,
		"Conversation signal: "+limitText(summary, 180),
		"Return a concise plan with assumptions, options, risks, and the next founder-visible step.",
	)
	return strings.Join(parts, "\n")
}

func agentBehaviorProfilePayload(profile *AgentBehaviorProfile) map[string]any {
	if profile == nil {
		return map[string]any{}
	}
	return map[string]any{
		"role":                  profile.Role,
		"mission":               profile.Mission,
		"critical_rules":        append([]string{}, profile.CriticalRules...),
		"workflow":              append([]string{}, profile.Workflow...),
		"required_outputs":      append([]string{}, profile.RequiredOutputs...),
		"evidence_requirements": append([]string{}, profile.EvidenceRequirements...),
	}
}

func conversationProposalPriority(tags []string) string {
	for _, tag := range tags {
		if tag == "risk" || tag == "todo" {
			return "high"
		}
	}
	return "medium"
}

func conversationProposalRisk(tags []string) string {
	for _, tag := range tags {
		if tag == "risk" || tag == "critique" {
			return "high"
		}
	}
	return "medium"
}

func containsAnyFragment(text string, fragments []string) bool {
	for _, fragment := range fragments {
		if fragment != "" && strings.Contains(text, strings.ToLower(strings.TrimSpace(fragment))) {
			return true
		}
	}
	return false
}

func agentBehaviorProfileForRole(role string) *AgentBehaviorProfile {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "planner":
		return &AgentBehaviorProfile{
			Role:    "planner",
			Mission: "Turn an approved conversation signal into the smallest viable plan with explicit assumptions, risks, and next actions.",
			CriticalRules: []string{
				"Stay within canonical runtime contracts",
				"Do not assume implementation approval",
				"Escalate unresolved blockers to review-room rather than inventing side channels",
			},
			Workflow:             []string{"triage signal", "state assumptions", "propose options", "surface risks", "recommend next step"},
			RequiredOutputs:      []string{"plan", "assumptions", "options", "risks", "recommended_next_step"},
			EvidenceRequirements: []string{"conversation refs", "proposal ref when present", "review-room link when risk exists"},
		}
	case "implementer":
		return &AgentBehaviorProfile{
			Role:                 "implementer",
			Mission:              "Implement the approved change with minimal surface area and preserve canonical Layer OS contracts.",
			CriticalRules:        []string{"no new side-channel state", "respect no-touch lanes", "keep changes minimal and test-backed"},
			Workflow:             []string{"inspect scope", "patch minimal slice", "run focused validation", "prepare handoff evidence"},
			RequiredOutputs:      []string{"patch", "focused_tests", "known_gaps"},
			EvidenceRequirements: []string{"changed files", "test output", "follow-up risk if unresolved"},
		}
	case "verifier":
		return &AgentBehaviorProfile{
			Role:                 "verifier",
			Mission:              "Verify the claimed implementation against contracts, tests, and observable runtime evidence.",
			CriticalRules:        []string{"prefer primary evidence", "report drift plainly", "never fake success"},
			Workflow:             []string{"read claim", "check code", "run validation", "report pass/fail evidence"},
			RequiredOutputs:      []string{"verdict", "evidence", "open_risks"},
			EvidenceRequirements: []string{"failing or passing commands", "file refs", "review-room escalation when needed"},
		}
	case "designer":
		return &AgentBehaviorProfile{
			Role:                 "designer",
			Mission:              "Review brand and experience changes while keeping runtime nouns and approval boundaries intact.",
			CriticalRules:        []string{"designer is a lane, not a new kingdom", "keep founder-facing language clear", "respect existing no-touch boundaries"},
			Workflow:             []string{"review IA", "review visual/system fit", "report recommendations", "flag founder decisions"},
			RequiredOutputs:      []string{"design_review", "recommendations", "open_questions"},
			EvidenceRequirements: []string{"screens or file refs", "consistency rationale", "founder decision points"},
		}
	default:
		return nil
	}
}

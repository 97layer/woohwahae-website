package runtime

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

func (s *Service) ListObservations(query ObservationQuery) []ObservationRecord {
	s.mu.Lock()
	items := s.observation.list()
	s.mu.Unlock()
	return filterObservations(items, query)
}

func (s *Service) CreateObservation(item ObservationRecord) (ObservationRecord, error) {
	s.mu.Lock()

	oldItems := s.observation.list()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()
	item, err := s.createObservationLocked(item)
	if err != nil {
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.mu.Unlock()
		return ObservationRecord{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.observation = newObservationStore(oldItems)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		s.mu.Unlock()
		return ObservationRecord{}, err
	}
	s.mu.Unlock()
	s.maybeNotifyFounderAttention()
	return item, nil
}

func normalizeObservationRecord(item ObservationRecord, fallbackActor string) ObservationRecord {
	now := zeroSafeNow()
	if strings.TrimSpace(item.ObservationID) == "" {
		item.ObservationID = nextObservationID(now)
	}
	item.SourceChannel = normalizeObservationChannel(item.SourceChannel)
	if item.CapturedAt.IsZero() {
		item.CapturedAt = now
	}
	item.Actor = strings.TrimSpace(item.Actor)
	if item.Actor == "" {
		item.Actor = strings.TrimSpace(fallbackActor)
	}
	item.Topic = strings.TrimSpace(item.Topic)
	item.Refs = normalizeObservationRefs(item.Refs)
	item.Confidence = normalizeObservationConfidence(item.Confidence)
	item.RawExcerpt = strings.TrimSpace(item.RawExcerpt)
	item.NormalizedSummary = strings.TrimSpace(item.NormalizedSummary)
	if item.RawExcerpt == "" {
		item.RawExcerpt = item.NormalizedSummary
	}
	if item.NormalizedSummary == "" {
		item.NormalizedSummary = item.RawExcerpt
	}
	return item
}

func nextObservationID(now interface{ UnixNano() int64 }) string {
	return "observation_" + strconv.FormatInt(now.UnixNano(), 10)
}

func normalizeObservationChannel(value string) string {
	return canonicalObservationChannel(value)
}

func normalizeObservationConfidence(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if validObservationConfidence(value) {
		return value
	}
	return "medium"
}

func normalizeObservationRefs(items []string) []string {
	if items == nil {
		return []string{}
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func filterObservations(items []ObservationRecord, query ObservationQuery) []ObservationRecord {
	sourceChannel := normalizeObservationChannel(query.SourceChannel)
	topic := strings.ToLower(strings.TrimSpace(query.Topic))
	actor := strings.ToLower(strings.TrimSpace(query.Actor))
	ref := strings.TrimSpace(query.Ref)
	text := strings.ToLower(strings.TrimSpace(query.Text))
	filtered := make([]ObservationRecord, 0, len(items))
	for _, item := range items {
		if sourceChannel != "" && normalizeObservationChannel(item.SourceChannel) != sourceChannel {
			continue
		}
		if topic != "" && !strings.Contains(strings.ToLower(item.Topic), topic) {
			continue
		}
		if actor != "" && !strings.Contains(strings.ToLower(item.Actor), actor) {
			continue
		}
		if ref != "" && !observationHasRef(item, ref) {
			continue
		}
		if text != "" && !observationMatchesText(item, text) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].CapturedAt.Equal(filtered[j].CapturedAt) {
			return filtered[i].ObservationID > filtered[j].ObservationID
		}
		return filtered[i].CapturedAt.After(filtered[j].CapturedAt)
	})
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
	}
	return filtered
}

func observationHasRef(item ObservationRecord, ref string) bool {
	for _, candidate := range item.Refs {
		if strings.TrimSpace(candidate) == ref {
			return true
		}
	}
	return false
}

func observationMatchesText(item ObservationRecord, text string) bool {
	if text == "" {
		return true
	}
	fields := []string{
		item.Topic,
		item.Actor,
		item.SourceChannel,
		item.RawExcerpt,
		item.NormalizedSummary,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), text) {
			return true
		}
	}
	for _, ref := range item.Refs {
		if strings.Contains(strings.ToLower(ref), text) {
			return true
		}
	}
	return false
}

func surfaceForObservationChannel(channel string) Surface {
	switch normalizeObservationChannel(channel) {
	case string(SurfaceTelegram):
		return SurfaceTelegram
	case string(SurfaceCockpit):
		return SurfaceCockpit
	default:
		return SurfaceAPI
	}
}

func (q ObservationQuery) Validate() error {
	if q.Limit < 0 {
		return errors.New("observation query limit must not be negative")
	}
	return nil
}

func (s *Service) recordObservationLocked(item ObservationRecord, emitEvent bool) (ObservationRecord, error) {
	item = normalizeObservationRecord(item, s.currentActor())
	if err := s.observation.create(item); err != nil {
		return ObservationRecord{}, err
	}
	if !emitEvent {
		return item, nil
	}
	event := newEvent("observation.created", s.currentActor(), surfaceForObservationChannel(item.SourceChannel), item.ObservationID, StageDiscover, map[string]any{
		"source_channel": item.SourceChannel,
		"topic":          item.Topic,
		"confidence":     item.Confidence,
		"summary":        item.NormalizedSummary,
		"refs":           append([]string{}, item.Refs...),
	})
	if err := s.event.append(event); err != nil {
		return ObservationRecord{}, err
	}
	return item, nil
}

func (s *Service) createObservationLocked(item ObservationRecord) (ObservationRecord, error) {
	recorded, err := s.recordObservationLocked(item, true)
	if err != nil {
		return ObservationRecord{}, err
	}
	if suggestion, ok := reviewRoomSuggestionForObservation(recorded); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			return ObservationRecord{}, err
		}
	}
	return recorded, nil
}

func (s *Service) appendAutoObservationLocked(item ObservationRecord) error {
	oldItems := s.observation.list()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()
	if _, err := s.createObservationLocked(item); err != nil {
		s.observation = newObservationStore(oldItems)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.observation = newObservationStore(oldItems)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func sessionFinishObservation(input SessionFinishInput, event EventEnvelope, actor string) ObservationRecord {
	refs := []string{"event:" + event.EventID}
	parts := []string{"focus=" + strings.TrimSpace(input.CurrentFocus)}
	if goal := normalizeOptionalString(input.CurrentGoal); goal != nil {
		parts = append(parts, "goal="+*goal)
	}
	if len(input.NextSteps) > 0 {
		parts = append(parts, "steps="+strings.Join(input.NextSteps, "; "))
	}
	if len(input.OpenRisks) > 0 {
		parts = append(parts, "risks="+strings.Join(input.OpenRisks, "; "))
	}
	if handoff := normalizeOptionalString(input.HandoffNote); handoff != nil {
		parts = append(parts, "handoff="+*handoff)
	}
	if note := normalizeOptionalString(input.Note); note != nil {
		parts = append(parts, "note="+*note)
	}
	summary := "Session finished on " + strings.TrimSpace(input.CurrentFocus)
	if goal := normalizeOptionalString(input.CurrentGoal); goal != nil {
		summary += " toward " + *goal
	}
	return ObservationRecord{
		SourceChannel:     "session",
		Actor:             strings.TrimSpace(actor),
		Topic:             event.Kind,
		Refs:              refs,
		Confidence:        "high",
		RawExcerpt:        strings.Join(parts, " | "),
		NormalizedSummary: summary,
		CapturedAt:        event.Timestamp,
	}
}

func agentJobReportObservation(job AgentJob, event EventEnvelope, actor string) ObservationRecord {
	refs := []string{"job:" + job.JobID, "event:" + event.EventID}
	if job.Ref != nil && strings.TrimSpace(*job.Ref) != "" {
		refs = append(refs, "ref:"+strings.TrimSpace(*job.Ref))
	}
	if gateway := observationResultString(job.Result, "gateway_call_id"); gateway != "" {
		refs = append(refs, "gateway:"+gateway)
	}
	parts := []string{"job=" + job.JobID, "status=" + job.Status, "summary=" + strings.TrimSpace(job.Summary)}
	if provider := observationResultString(job.Result, "provider"); provider != "" {
		parts = append(parts, "provider="+provider)
	}
	if model := observationResultString(job.Result, "model"); model != "" {
		parts = append(parts, "model="+model)
	}
	if len(job.Notes) > 0 {
		parts = append(parts, "notes="+strings.Join(job.Notes, "; "))
	}
	return ObservationRecord{
		SourceChannel:     "agent_job",
		Actor:             strings.TrimSpace(actor),
		Topic:             event.Kind,
		Refs:              refs,
		Confidence:        "high",
		RawExcerpt:        strings.Join(parts, " | "),
		NormalizedSummary: "Agent job " + job.JobID + " " + job.Status + ": " + strings.TrimSpace(job.Summary),
		CapturedAt:        event.Timestamp,
	}
}

func observationResultString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	value, ok := data[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

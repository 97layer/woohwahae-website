package runtime

import "strings"

func (s *Service) CreateEvent(input EventCreateInput) (EventEnvelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(input.WorkItemID) == "" {
		input.WorkItemID = "system"
	}
	if input.Surface == "" {
		input.Surface = SurfaceAPI
	}
	if input.Stage == "" {
		input.Stage = StageDiscover
	}
	if input.Data == nil {
		input.Data = map[string]any{}
	}
	if err := input.Validate(); err != nil {
		return EventEnvelope{}, err
	}

	oldEventState := s.event.state()
	item := newEvent(input.Kind, s.currentActor(), input.Surface, input.WorkItemID, input.Stage, input.Data)
	if err := s.event.append(item); err != nil {
		return EventEnvelope{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.event = newEventStoreFromState(oldEventState)
		return EventEnvelope{}, err
	}
	return item, nil
}

func (s *Service) SessionBootstrap(source string, degraded bool, full bool) SessionBootstrapPacket {
	knowledge := s.Knowledge()
	s.mu.Lock()
	continuityRecords := s.continuity.list()
	s.mu.Unlock()
	continuity := continuityViewForBootstrap(continuityRecords, continuityRecordCompatibilityAny, continuitySourceTerminal, knowledge)
	var resume *SessionCheckpoint
	if continuity.Record != nil {
		item := sessionCheckpointFromContinuityRecord(*continuity.Record)
		resume = &item
	}
	packet := SessionBootstrapPacket{
		GeneratedAt: zeroSafeNow(),
		Source:      strings.TrimSpace(source),
		ReadOnly:    true,
		Degraded:    degraded,
		Tooling:     toolingHealthForBootstrap(),
		Knowledge:   knowledge,
		Prompting:   func() *PromptingContract { item := defaultSessionPrompting(knowledge); return &item }(),
		Continuity:  continuity,
		Resume:      resume,
	}
	if packet.Source == "" {
		packet.Source = "daemon"
	}
	if full {
		handoff := s.Handoff()
		reviewRoom := s.ReviewRoomSummary()
		capabilities := s.Capabilities()
		packet.Handoff = &handoff
		packet.ReviewRoom = &reviewRoom
		packet.Capabilities = &capabilities
	}
	return packet
}

func toolingHealthForBootstrap() *ToolingHealth {
	audit := AuditMCP("")
	missingRequired := []string{}
	for _, item := range audit.Servers {
		if item.Required && len(item.Issues) > 0 {
			missingRequired = append(missingRequired, item.Name)
		}
	}
	var note *string
	if audit.Status != "ok" || len(missingRequired) > 0 {
		value := "restart Codex session if newly added MCP tools are still not visible in this session"
		note = &value
	}
	return &ToolingHealth{
		Status:           audit.Status,
		RequiredMCPReady: len(missingRequired) == 0,
		SessionNote:      note,
		MissingRequired:  missingRequired,
	}
}

func (s *Service) CheckpointSession(input SessionCheckpointInput) (SessionCheckpoint, error) {
	input = normalizeSessionCheckpointInput(input)
	if err := input.Validate(); err != nil {
		return SessionCheckpoint{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldContinuity := s.continuity.list()
	oldMemory := s.memory.current()
	actor := s.currentActor()
	now := zeroSafeNow()
	records := s.continuity.list()

	var record ContinuityRecord
	if existing, _, ok := latestContinuityRecord(records, actor, input.Source, continuityStatusActive); ok {
		record = applySessionCheckpointToContinuityRecord(existing, input, now)
	} else {
		record = applySessionCheckpointToContinuityRecord(continuityRecordFromMemory(oldMemory, actor, input.Source, now), input, now)
	}
	records = upsertContinuityRecord(records, record)
	if err := s.continuity.replace(records); err != nil {
		return SessionCheckpoint{}, err
	}

	memory := continuityMemoryProjection(oldMemory, record, actor)
	if err := s.memory.replace(memory); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		return SessionCheckpoint{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		s.memory = newMemoryStore(oldMemory)
		return SessionCheckpoint{}, err
	}
	return sessionCheckpointFromContinuityRecord(record), nil
}

func (s *Service) SessionNote(input SessionNoteInput) (SessionNoteResult, error) {
	input = normalizeSessionNoteInput(input)
	if err := input.Validate(); err != nil {
		return SessionNoteResult{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldContinuity := s.continuity.list()
	oldMemory := s.memory.current()
	oldObservations := s.observation.list()
	oldEventState := s.event.state()
	actor := s.currentActor()
	now := zeroSafeNow()
	records := s.continuity.list()

	var base ContinuityRecord
	if existing, _, ok := latestOpenContinuityRecord(records, actor, input.Source); ok {
		base = existing
	} else {
		base = continuityRecordFromMemory(oldMemory, actor, input.Source, now)
	}
	record, note := applySessionNoteToContinuityRecord(base, input, now)
	records = upsertContinuityRecord(records, record)
	if err := s.continuity.replace(records); err != nil {
		return SessionNoteResult{}, err
	}

	memory := continuityMemoryProjection(oldMemory, record, actor)
	if err := s.memory.replace(memory); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		return SessionNoteResult{}, err
	}

	observation, err := s.recordObservationLocked(continuityObservationForNote(record, note), true)
	if err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		s.memory = newMemoryStore(oldMemory)
		s.observation = newObservationStore(oldObservations)
		s.event = newEventStoreFromState(oldEventState)
		return SessionNoteResult{}, err
	}
	event := newEvent("session.noted", actor, SurfaceAPI, "system", StageDiscover, map[string]any{
		"kind":          note.Kind,
		"current_focus": record.CurrentFocus,
		"note":          note.Text,
	})
	if err := s.event.append(event); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		s.memory = newMemoryStore(oldMemory)
		s.observation = newObservationStore(oldObservations)
		s.event = newEventStoreFromState(oldEventState)
		return SessionNoteResult{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		s.memory = newMemoryStore(oldMemory)
		s.observation = newObservationStore(oldObservations)
		s.event = newEventStoreFromState(oldEventState)
		return SessionNoteResult{}, err
	}
	return SessionNoteResult{
		Record:      record,
		Note:        note,
		Observation: observation,
		Event:       event,
	}, nil
}

func (s *Service) FinishSession(input SessionFinishInput) (SessionFinishResult, error) {
	if input.NextSteps == nil {
		input.NextSteps = []string{}
	}
	if input.OpenRisks == nil {
		input.OpenRisks = []string{}
	}
	if err := input.Validate(); err != nil {
		return SessionFinishResult{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldContinuity := s.continuity.list()
	oldMemory := s.memory.current()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()

	memory := s.memory.current()
	memory.CurrentFocus = strings.TrimSpace(input.CurrentFocus)
	memory.CurrentGoal = normalizeOptionalString(input.CurrentGoal)
	memory.NextSteps = append([]string{}, input.NextSteps...)
	memory.OpenRisks = append([]string{}, input.OpenRisks...)
	memory.HandoffNote = normalizeOptionalString(input.HandoffNote)
	operator := s.currentActor()
	memory.LastOperator = &operator
	memory.UpdatedAt = zeroSafeNow()
	if err := s.memory.replace(memory); err != nil {
		return SessionFinishResult{}, err
	}

	data := map[string]any{
		"current_focus": memory.CurrentFocus,
		"next_steps":    append([]string{}, memory.NextSteps...),
		"open_risks":    append([]string{}, memory.OpenRisks...),
	}
	if memory.CurrentGoal != nil {
		data["current_goal"] = *memory.CurrentGoal
	}
	if memory.HandoffNote != nil {
		data["handoff_note"] = *memory.HandoffNote
	}
	if note := normalizeOptionalString(input.Note); note != nil {
		data["note"] = *note
	}
	event := newEvent("session.finished", s.currentActor(), SurfaceAPI, "system", StageDiscover, data)
	if err := s.event.append(event); err != nil {
		s.memory = newMemoryStore(oldMemory)
		return SessionFinishResult{}, err
	}
	if suggestion, ok := reviewRoomSuggestionForSessionFinish(input, s.currentActor()); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.memory = newMemoryStore(oldMemory)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return SessionFinishResult{}, err
		}
	}
	records := s.continuity.list()
	if existing, _, ok := latestOpenContinuityRecord(records, s.currentActor(), continuitySourceTerminal); ok {
		records = upsertContinuityRecord(records, applySessionFinishToContinuityRecord(existing, input, memory.UpdatedAt))
		if err := s.continuity.replace(records); err != nil {
			s.memory = newMemoryStore(oldMemory)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return SessionFinishResult{}, err
		}
	}
	if err := s.persistLocked(); err != nil {
		s.continuity = newContinuityStore(oldContinuity)
		s.memory = newMemoryStore(oldMemory)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return SessionFinishResult{}, err
	}
	entry := newCapitalizationEntry(input, event, SummarizeReviewRoom(s.currentReviewRoomLocked()), s.currentActor())
	if err := s.disk.appendCapitalizationEntry(entry); err != nil {
		warning := newSignalReviewRoomItem("`session.finish` 이후 세션 자본화 기록 추가가 실패했어. 이후 자동화를 믿기 전에 지식 코퍼스 경로를 확인해야 해.", "capitalization.failed", nil, "capitalization append failure requires founder review before automation depends on the corpus", "review_room.auto.capitalization_failed", []string{"entry:" + entry.EntryID, "error:" + err.Error()})
		if openErr := s.autoOpenReviewRoomItemLocked(warning); openErr == nil {
			_ = s.persistLocked()
		}
	}
	if err := s.appendAutoObservationLocked(sessionFinishObservation(input, event, s.currentActor())); err != nil {
		warning := newSignalReviewRoomItem("`session.finish` 이후 observation 자동 적재가 실패했어. 크로스채널 합성을 믿기 전에 observation 레지스트리를 확인해야 해.", "observation.failed", nil, "observation append failure requires founder review before automation depends on the observation registry", "review_room.auto.session_observation_failed", []string{"event:" + event.EventID, "error:" + err.Error()})
		if openErr := s.autoOpenReviewRoomItemLocked(warning); openErr == nil {
			_ = s.persistLocked()
		}
	}
	return SessionFinishResult{Memory: memory, Event: event}, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeSessionCheckpointInput(input SessionCheckpointInput) SessionCheckpointInput {
	input.Source = normalizeSessionCheckpointSource(input.Source)
	input.CurrentFocus = strings.TrimSpace(input.CurrentFocus)
	input.CurrentGoal = normalizeOptionalString(input.CurrentGoal)
	input.NextSteps = normalizeSessionCheckpointList(input.NextSteps)
	input.OpenRisks = normalizeSessionCheckpointList(input.OpenRisks)
	input.HandoffNote = normalizeOptionalString(input.HandoffNote)
	input.Note = normalizeOptionalString(input.Note)
	input.Refs = normalizeObservationRefs(input.Refs)
	return input
}

func normalizeSessionNoteInput(input SessionNoteInput) SessionNoteInput {
	input.Source = normalizeContinuitySource(input.Source)
	input.Kind = normalizeContinuityNoteKind(input.Kind)
	input.Text = limitText(strings.TrimSpace(maskConversationText(input.Text)), 500)
	input.Refs = normalizeObservationRefs(input.Refs)
	return input
}

func normalizeSessionCheckpoint(item SessionCheckpoint) SessionCheckpoint {
	item.CheckpointID = strings.TrimSpace(item.CheckpointID)
	item.Source = normalizeSessionCheckpointSource(item.Source)
	item.Actor = ResolveActor(item.Actor)
	item.CurrentFocus = strings.TrimSpace(item.CurrentFocus)
	item.CurrentGoal = normalizeOptionalString(item.CurrentGoal)
	item.NextSteps = normalizeSessionCheckpointList(item.NextSteps)
	item.OpenRisks = normalizeSessionCheckpointList(item.OpenRisks)
	item.HandoffNote = normalizeOptionalString(item.HandoffNote)
	item.Note = normalizeOptionalString(item.Note)
	item.Refs = normalizeObservationRefs(item.Refs)
	return item
}

func normalizeSessionCheckpointSource(value string) string {
	return normalizeContinuitySource(value)
}

func normalizeSessionCheckpointList(items []string) []string {
	if items == nil {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
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

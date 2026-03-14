package runtime

import (
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	continuitySourceTerminal         = "terminal"
	continuityStatusActive           = "active"
	continuityStatusResolved         = "resolved"
	continuityStatusStale            = "stale"
	continuityNoteKindNote           = "note"
	continuityNoteKindTodo           = "todo"
	continuityNoteKindDecision       = "decision"
	continuityNoteKindRisk           = "risk"
	continuitySuggestionKindRisk     = "risk"
	continuitySuggestionKindTodo     = "todo"
	continuitySuggestionKindRoute    = "route"
	continuitySuggestionKindThread   = "thread"
	continuityObservationTopic       = "continuity.note"
	continuitySuggestionLimit        = 3
	continuityNoteHistoryLimit       = 50
	continuityStaleAfter             = 24 * time.Hour
	continuityRecordCompatibilityAny = "*"
)

type continuityStore struct {
	mu    sync.RWMutex
	items []ContinuityRecord
}

func newContinuityStore(items []ContinuityRecord) *continuityStore {
	if items == nil {
		items = []ContinuityRecord{}
	}
	return &continuityStore{items: normalizeContinuityRecordsAt(items, zeroSafeNow())}
}

func (s *continuityStore) list() []ContinuityRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ContinuityRecord, len(s.items))
	copy(out, s.items)
	return normalizeContinuityRecordsAt(out, zeroSafeNow())
}

func (s *continuityStore) replace(items []ContinuityRecord) error {
	items = normalizeContinuityRecordsAt(items, zeroSafeNow())
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = items
	return nil
}

func normalizeContinuityRecordsAt(items []ContinuityRecord, now time.Time) []ContinuityRecord {
	if items == nil {
		return []ContinuityRecord{}
	}
	out := make([]ContinuityRecord, 0, len(items))
	for _, item := range items {
		normalized := normalizeContinuityRecordAt(item, now)
		if normalized.RecordID == "" {
			continue
		}
		out = append(out, normalized)
	}
	sortContinuityRecords(out)
	return out
}

func sortContinuityRecords(items []ContinuityRecord) {
	sort.SliceStable(items, func(i, j int) bool {
		if continuityStatusRank(items[i].Status) != continuityStatusRank(items[j].Status) {
			return continuityStatusRank(items[i].Status) < continuityStatusRank(items[j].Status)
		}
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].RecordID > items[j].RecordID
	})
}

func continuityStatusRank(value string) int {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case continuityStatusActive:
		return 0
	case continuityStatusStale:
		return 1
	case continuityStatusResolved:
		return 2
	default:
		return 3
	}
}

func normalizeContinuitySource(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return continuitySourceTerminal
	}
	return value
}

func normalizeContinuityNoteKind(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return continuityNoteKindNote
	}
	return value
}

func normalizeContinuitySuggestionKind(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func validContinuityStatus(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case continuityStatusActive, continuityStatusResolved, continuityStatusStale:
		return true
	default:
		return false
	}
}

func validContinuityNoteKind(value string) bool {
	switch normalizeContinuityNoteKind(value) {
	case continuityNoteKindNote, continuityNoteKindTodo, continuityNoteKindDecision, continuityNoteKindRisk:
		return true
	default:
		return false
	}
}

func validContinuitySuggestionKind(value string) bool {
	switch normalizeContinuitySuggestionKind(value) {
	case continuitySuggestionKindRisk, continuitySuggestionKindTodo, continuitySuggestionKindRoute, continuitySuggestionKindThread:
		return true
	default:
		return false
	}
}

func nextContinuityRecordID(now interface{ UnixNano() int64 }) string {
	return "continuity_record_" + strconv.FormatInt(now.UnixNano(), 10)
}

func nextContinuityNoteID(now interface{ UnixNano() int64 }) string {
	return "continuity_note_" + strconv.FormatInt(now.UnixNano(), 10)
}

func normalizeContinuityNoteAt(item ContinuityNote, now time.Time) ContinuityNote {
	item.NoteID = strings.TrimSpace(item.NoteID)
	if item.NoteID == "" {
		item.NoteID = nextContinuityNoteID(now)
	}
	item.Kind = normalizeContinuityNoteKind(item.Kind)
	item.Text = limitText(strings.TrimSpace(maskConversationText(item.Text)), 500)
	item.Refs = normalizeObservationRefs(item.Refs)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	return item
}

func normalizeContinuityRecordAt(item ContinuityRecord, now time.Time) ContinuityRecord {
	item.RecordID = strings.TrimSpace(item.RecordID)
	if item.RecordID == "" {
		item.RecordID = nextContinuityRecordID(now)
	}
	item.Source = normalizeContinuitySource(item.Source)
	item.Actor = ResolveActor(item.Actor)
	item.CurrentFocus = strings.TrimSpace(item.CurrentFocus)
	item.CurrentGoal = normalizeOptionalString(item.CurrentGoal)
	item.NextSteps = normalizeSessionCheckpointList(item.NextSteps)
	item.OpenRisks = normalizeSessionCheckpointList(item.OpenRisks)
	item.HandoffNote = normalizeOptionalString(item.HandoffNote)
	item.Refs = normalizeObservationRefs(item.Refs)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	if item.Notes == nil {
		item.Notes = []ContinuityNote{}
	}
	notes := make([]ContinuityNote, 0, len(item.Notes))
	for _, note := range item.Notes {
		normalized := normalizeContinuityNoteAt(note, item.UpdatedAt)
		if normalized.Text == "" {
			continue
		}
		notes = append(notes, normalized)
		item.Refs = normalizeObservationRefs(append(item.Refs, normalized.Refs...))
	}
	item.Notes = trimContinuityNotes(notes)
	item.Status = continuityStatusAt(item, now)
	if item.Status == continuityStatusResolved {
		if item.ResolvedAt == nil || item.ResolvedAt.IsZero() {
			resolvedAt := item.UpdatedAt
			item.ResolvedAt = &resolvedAt
		}
	} else {
		item.ResolvedAt = nil
	}
	return item
}

func trimContinuityNotes(items []ContinuityNote) []ContinuityNote {
	if items == nil {
		return []ContinuityNote{}
	}
	if len(items) <= continuityNoteHistoryLimit {
		return append([]ContinuityNote{}, items...)
	}
	return append([]ContinuityNote{}, items[len(items)-continuityNoteHistoryLimit:]...)
}

func continuityStatusAt(item ContinuityRecord, now time.Time) string {
	status := strings.TrimSpace(strings.ToLower(item.Status))
	if status == continuityStatusResolved {
		return continuityStatusResolved
	}
	if status == "" {
		status = continuityStatusActive
	}
	if item.UpdatedAt.IsZero() {
		return status
	}
	if now.Sub(item.UpdatedAt) >= continuityStaleAfter {
		return continuityStatusStale
	}
	return continuityStatusActive
}

func continuityRecordFromSessionCheckpoint(item SessionCheckpoint) ContinuityRecord {
	item = normalizeSessionCheckpoint(item)
	now := item.UpdatedAt
	if now.IsZero() {
		now = zeroSafeNow()
	}
	recordID := strings.TrimSpace(strings.Replace(item.CheckpointID, "session_checkpoint", "continuity_record", 1))
	if recordID == "" {
		recordID = nextContinuityRecordID(now)
	}
	record := ContinuityRecord{
		RecordID:     recordID,
		Source:       item.Source,
		Actor:        item.Actor,
		Status:       continuityStatusActive,
		CurrentFocus: item.CurrentFocus,
		CurrentGoal:  item.CurrentGoal,
		NextSteps:    append([]string{}, item.NextSteps...),
		OpenRisks:    append([]string{}, item.OpenRisks...),
		HandoffNote:  item.HandoffNote,
		Refs:         append([]string{}, item.Refs...),
		CreatedAt:    now,
		UpdatedAt:    now,
		Notes:        []ContinuityNote{},
	}
	if item.Note != nil {
		record.Notes = append(record.Notes, ContinuityNote{
			NoteID:    strings.Replace(recordID, "continuity_record", "continuity_note", 1),
			Kind:      continuityNoteKindNote,
			Text:      *item.Note,
			Refs:      append([]string{}, item.Refs...),
			CreatedAt: now,
		})
	}
	return normalizeContinuityRecordAt(record, zeroSafeNow())
}

func sessionCheckpointFromContinuityRecord(item ContinuityRecord) SessionCheckpoint {
	item = normalizeContinuityRecordAt(item, zeroSafeNow())
	var latestNote *string
	if len(item.Notes) > 0 {
		value := strings.TrimSpace(item.Notes[len(item.Notes)-1].Text)
		if value != "" {
			latestNote = &value
		}
	}
	return normalizeSessionCheckpoint(SessionCheckpoint{
		CheckpointID: strings.Replace(item.RecordID, "continuity_record", "session_checkpoint", 1),
		Source:       item.Source,
		Actor:        item.Actor,
		CurrentFocus: item.CurrentFocus,
		CurrentGoal:  item.CurrentGoal,
		NextSteps:    append([]string{}, item.NextSteps...),
		OpenRisks:    append([]string{}, item.OpenRisks...),
		HandoffNote:  item.HandoffNote,
		Note:         latestNote,
		Refs:         append([]string{}, item.Refs...),
		UpdatedAt:    item.UpdatedAt,
	})
}

func continuityRecordFromMemory(memory SystemMemory, actor string, source string, now time.Time) ContinuityRecord {
	memory = normalizeContinuityMemoryBase(memory, now)
	record := ContinuityRecord{
		RecordID:     nextContinuityRecordID(now),
		Source:       normalizeContinuitySource(source),
		Actor:        ResolveActor(actor),
		Status:       continuityStatusActive,
		CurrentFocus: memory.CurrentFocus,
		CurrentGoal:  memory.CurrentGoal,
		NextSteps:    append([]string{}, memory.NextSteps...),
		OpenRisks:    append([]string{}, memory.OpenRisks...),
		HandoffNote:  memory.HandoffNote,
		Refs:         []string{},
		Notes:        []ContinuityNote{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return normalizeContinuityRecordAt(record, now)
}

func applySessionCheckpointToContinuityRecord(item ContinuityRecord, input SessionCheckpointInput, now time.Time) ContinuityRecord {
	item = normalizeContinuityRecordAt(item, now)
	item.Source = input.Source
	item.Status = continuityStatusActive
	item.CurrentFocus = input.CurrentFocus
	item.CurrentGoal = input.CurrentGoal
	item.NextSteps = append([]string{}, input.NextSteps...)
	item.OpenRisks = append([]string{}, input.OpenRisks...)
	item.HandoffNote = input.HandoffNote
	item.Refs = normalizeObservationRefs(append(item.Refs, input.Refs...))
	item.UpdatedAt = now
	item.ResolvedAt = nil
	if input.Note != nil {
		item.Notes = append(item.Notes, normalizeContinuityNoteAt(ContinuityNote{
			Kind:      continuityNoteKindNote,
			Text:      *input.Note,
			Refs:      append([]string{}, input.Refs...),
			CreatedAt: now,
		}, now))
	}
	return normalizeContinuityRecordAt(item, now)
}

func applySessionFinishToContinuityRecord(item ContinuityRecord, input SessionFinishInput, now time.Time) ContinuityRecord {
	item = normalizeContinuityRecordAt(item, now)
	item.Status = continuityStatusResolved
	item.CurrentFocus = strings.TrimSpace(input.CurrentFocus)
	item.CurrentGoal = normalizeOptionalString(input.CurrentGoal)
	item.NextSteps = normalizeSessionCheckpointList(input.NextSteps)
	item.OpenRisks = normalizeSessionCheckpointList(input.OpenRisks)
	item.HandoffNote = normalizeOptionalString(input.HandoffNote)
	item.UpdatedAt = now
	item.ResolvedAt = &now
	if input.Note != nil {
		item.Notes = append(item.Notes, normalizeContinuityNoteAt(ContinuityNote{
			Kind:      continuityNoteKindDecision,
			Text:      *input.Note,
			Refs:      append([]string{}, item.Refs...),
			CreatedAt: now,
		}, now))
	}
	return normalizeContinuityRecordAt(item, now)
}

func applySessionNoteToContinuityRecord(item ContinuityRecord, input SessionNoteInput, now time.Time) (ContinuityRecord, ContinuityNote) {
	item = normalizeContinuityRecordAt(item, now)
	note := normalizeContinuityNoteAt(ContinuityNote{
		Kind:      input.Kind,
		Text:      input.Text,
		Refs:      append([]string{}, input.Refs...),
		CreatedAt: now,
	}, now)
	item.Status = continuityStatusActive
	item.Source = input.Source
	item.Refs = normalizeObservationRefs(append(item.Refs, note.Refs...))
	item.Notes = append(item.Notes, note)
	item.UpdatedAt = now
	item.ResolvedAt = nil
	switch note.Kind {
	case continuityNoteKindTodo:
		item.NextSteps = normalizeSessionCheckpointList(append(item.NextSteps, note.Text))
	case continuityNoteKindRisk:
		item.OpenRisks = normalizeSessionCheckpointList(append(item.OpenRisks, note.Text))
	}
	return normalizeContinuityRecordAt(item, now), note
}

func continuityMemoryProjection(base SystemMemory, item ContinuityRecord, actor string) SystemMemory {
	base = normalizeContinuityMemoryBase(base, item.UpdatedAt)
	item = normalizeContinuityRecordAt(item, zeroSafeNow())
	base.CurrentFocus = item.CurrentFocus
	base.CurrentGoal = item.CurrentGoal
	base.NextSteps = append([]string{}, item.NextSteps...)
	base.OpenRisks = append([]string{}, item.OpenRisks...)
	base.HandoffNote = item.HandoffNote
	if resolvedActor := ResolveActor(actor); resolvedActor != "" {
		base.LastOperator = &resolvedActor
	}
	base.UpdatedAt = item.UpdatedAt
	return base
}

func latestContinuityRecord(items []ContinuityRecord, actor string, source string, statuses ...string) (ContinuityRecord, int, bool) {
	source = normalizeContinuitySource(source)
	actor = ResolveActor(actor)
	allowed := map[string]bool{}
	for _, status := range statuses {
		allowed[strings.TrimSpace(strings.ToLower(status))] = true
	}
	for idx, item := range normalizeContinuityRecordsAt(items, zeroSafeNow()) {
		if actor != "" && actor != continuityRecordCompatibilityAny && item.Actor != actor {
			continue
		}
		if source != "" && source != continuityRecordCompatibilityAny && item.Source != source {
			continue
		}
		if len(allowed) > 0 && !allowed[item.Status] {
			continue
		}
		return item, idx, true
	}
	return ContinuityRecord{}, -1, false
}

func latestOpenContinuityRecord(items []ContinuityRecord, actor string, source string) (ContinuityRecord, int, bool) {
	if item, idx, ok := latestContinuityRecord(items, actor, source, continuityStatusActive); ok {
		return item, idx, true
	}
	return latestContinuityRecord(items, actor, source, continuityStatusStale)
}

func upsertContinuityRecord(items []ContinuityRecord, next ContinuityRecord) []ContinuityRecord {
	next = normalizeContinuityRecordAt(next, zeroSafeNow())
	updated := false
	out := make([]ContinuityRecord, 0, len(items)+1)
	for _, item := range items {
		if item.RecordID != next.RecordID {
			out = append(out, normalizeContinuityRecordAt(item, zeroSafeNow()))
			continue
		}
		out = append(out, next)
		updated = true
	}
	if !updated {
		out = append(out, next)
	}
	return normalizeContinuityRecordsAt(out, zeroSafeNow())
}

func continuityObservationForNote(item ContinuityRecord, note ContinuityNote) ObservationRecord {
	item = normalizeContinuityRecordAt(item, zeroSafeNow())
	note = normalizeContinuityNoteAt(note, item.UpdatedAt)
	refs := append([]string{"continuity:" + item.RecordID, "continuity_note:" + note.NoteID}, note.Refs...)
	parts := []string{
		"kind=" + note.Kind,
		"focus=" + item.CurrentFocus,
		"text=" + note.Text,
	}
	if len(note.Refs) > 0 {
		parts = append(parts, "refs="+strings.Join(note.Refs, "; "))
	}
	return ObservationRecord{
		SourceChannel:     continuitySourceTerminal,
		Actor:             item.Actor,
		Topic:             continuityObservationTopic,
		Refs:              normalizeObservationRefs(refs),
		Confidence:        "high",
		RawExcerpt:        strings.Join(parts, " | "),
		NormalizedSummary: limitText("Continuity "+note.Kind+": "+note.Text, 180),
		CapturedAt:        note.CreatedAt,
	}
}

func continuityViewForBootstrap(items []ContinuityRecord, actor string, source string, knowledge KnowledgePacket) *ContinuityView {
	record, _, ok := latestOpenContinuityRecord(items, actor, source)
	view := &ContinuityView{
		RelatedThreads: []OpenThread{},
		Suggestions:    []ContinuitySuggestion{},
	}
	if !ok {
		return view
	}
	copy := record
	view.Record = &copy
	view.RelatedThreads = continuityRelatedThreads(record, knowledge.OpenThreads)
	view.Suggestions = continuitySuggestions(record, knowledge, view.RelatedThreads)
	return view
}

func continuityRelatedThreads(item ContinuityRecord, threads []OpenThread) []OpenThread {
	if len(threads) == 0 {
		return []OpenThread{}
	}
	matched := []OpenThread{}
	for _, thread := range threads {
		if continuityThreadMatchesRefs(thread, item.Refs) {
			matched = append(matched, thread)
		}
	}
	if len(matched) == 0 {
		matched = append(matched, threads...)
	}
	return limitOpenThreads(matched, continuitySuggestionLimit)
}

func continuityThreadMatchesRefs(item OpenThread, refs []string) bool {
	if len(refs) == 0 {
		return false
	}
	candidates := append(append([]string{}, item.PatternRefs...), item.Evidence...)
	candidates = append(candidates, item.ThreadID, item.Question)
	for _, candidate := range candidates {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		for _, ref := range refs {
			ref = strings.ToLower(strings.TrimSpace(ref))
			if ref == "" {
				continue
			}
			if strings.Contains(candidate, ref) || strings.Contains(ref, candidate) {
				return true
			}
		}
	}
	return false
}

func continuitySuggestions(item ContinuityRecord, knowledge KnowledgePacket, relatedThreads []OpenThread) []ContinuitySuggestion {
	suggestions := []ContinuitySuggestion{}
	seen := map[string]bool{}
	appendSuggestion := func(suggestion ContinuitySuggestion) {
		suggestion = normalizeContinuitySuggestion(suggestion)
		if suggestion.SuggestionID == "" || suggestion.Summary == "" {
			return
		}
		key := suggestion.Kind + "|" + strings.ToLower(strings.TrimSpace(suggestion.Summary))
		if seen[key] {
			return
		}
		seen[key] = true
		suggestions = append(suggestions, suggestion)
	}
	for _, risk := range item.OpenRisks {
		appendSuggestion(newContinuitySuggestion(continuitySuggestionKindRisk, risk, nil, nil, "continuity.open_risk"))
		if len(suggestions) == continuitySuggestionLimit {
			return suggestions
		}
	}
	for idx := len(item.Notes) - 1; idx >= 0; idx-- {
		note := item.Notes[idx]
		if note.Kind != continuityNoteKindTodo {
			continue
		}
		appendSuggestion(newContinuitySuggestion(
			continuitySuggestionKindTodo,
			note.Text,
			firstContinuityRef(note.Refs),
			nil,
			"continuity.note",
		))
		if len(suggestions) == continuitySuggestionLimit {
			return suggestions
		}
	}
	for _, route := range knowledge.ActionRoutes {
		appendSuggestion(newContinuitySuggestion(continuitySuggestionKindRoute, route.Summary, route.TargetRef, route.Command, route.Source))
		if len(suggestions) == continuitySuggestionLimit {
			return suggestions
		}
	}
	if len(relatedThreads) == 0 {
		relatedThreads = knowledge.OpenThreads
	}
	for _, thread := range relatedThreads {
		appendSuggestion(newContinuitySuggestion(
			continuitySuggestionKindThread,
			thread.Question,
			optionalKnowledgeString(thread.ThreadID),
			nil,
			thread.Source,
		))
		if len(suggestions) == continuitySuggestionLimit {
			return suggestions
		}
	}
	return suggestions
}

func firstContinuityRef(items []string) *string {
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		return &value
	}
	return nil
}

func newContinuitySuggestion(kind string, summary string, ref *string, command *string, source string) ContinuitySuggestion {
	kind = normalizeContinuitySuggestionKind(kind)
	summary = limitText(strings.TrimSpace(summary), 180)
	source = strings.TrimSpace(strings.ToLower(source))
	payload := kind + "|" + summary + "|" + source
	if ref != nil {
		payload += "|" + strings.TrimSpace(*ref)
	}
	return ContinuitySuggestion{
		SuggestionID: "continuity_suggestion_" + strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(payload))), 16),
		Kind:         kind,
		Summary:      summary,
		Ref:          normalizeOptionalString(ref),
		Command:      normalizeOptionalString(command),
		Source:       source,
	}
}

func normalizeContinuitySuggestion(item ContinuitySuggestion) ContinuitySuggestion {
	item.SuggestionID = strings.TrimSpace(item.SuggestionID)
	item.Kind = normalizeContinuitySuggestionKind(item.Kind)
	item.Summary = limitText(strings.TrimSpace(item.Summary), 180)
	item.Ref = normalizeOptionalString(item.Ref)
	item.Command = normalizeOptionalString(item.Command)
	item.Source = strings.TrimSpace(strings.ToLower(item.Source))
	if item.SuggestionID == "" && item.Summary != "" {
		item = newContinuitySuggestion(item.Kind, item.Summary, item.Ref, item.Command, item.Source)
	}
	return item
}

func continuityRecordForCompatibility(items []ContinuityRecord) *SessionCheckpoint {
	record, _, ok := latestOpenContinuityRecord(items, continuityRecordCompatibilityAny, continuityRecordCompatibilityAny)
	if !ok {
		return nil
	}
	checkpoint := sessionCheckpointFromContinuityRecord(record)
	return &checkpoint
}

func normalizeContinuityMemoryBase(memory SystemMemory, now time.Time) SystemMemory {
	if strings.TrimSpace(memory.CurrentFocus) == "" {
		memory.CurrentFocus = defaultSystemMemory().CurrentFocus
	}
	if memory.NextSteps == nil {
		memory.NextSteps = []string{}
	}
	if memory.OpenRisks == nil {
		memory.OpenRisks = []string{}
	}
	memory.CurrentGoal = normalizeOptionalString(memory.CurrentGoal)
	memory.HandoffNote = normalizeOptionalString(memory.HandoffNote)
	memory.LastOperator = normalizeOptionalString(memory.LastOperator)
	if memory.UpdatedAt.IsZero() {
		memory.UpdatedAt = now
	}
	return memory
}

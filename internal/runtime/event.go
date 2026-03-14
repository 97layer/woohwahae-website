package runtime

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultEventRetentionLimit = 256

func zeroSafeNow() time.Time {
	return time.Now().UTC()
}

type eventStore struct {
	mu             sync.RWMutex
	items          []EventEnvelope
	pendingArchive []EventEnvelope
	archivedCount  int
	limit          int
}

type eventStoreState struct {
	items          []EventEnvelope
	pendingArchive []EventEnvelope
	archivedCount  int
	limit          int
}

func newEventStoreWithArchiveCount(items []EventEnvelope, archivedCount int) *eventStore {
	limit := eventRetentionLimitFromEnv()
	retained, _ := splitRetainedEvents(dedupeEvents(items), limit)
	if archivedCount < 0 {
		archivedCount = 0
	}
	return &eventStore{
		items:          retained,
		pendingArchive: []EventEnvelope{},
		archivedCount:  archivedCount,
		limit:          limit,
	}
}

func newEventStoreFromEvidence(evidence []EventEnvelope) *eventStore {
	limit := eventRetentionLimitFromEnv()
	retained, archived := splitRetainedEvents(dedupeEvents(evidence), limit)
	return &eventStore{
		items:          retained,
		pendingArchive: archived,
		archivedCount:  0,
		limit:          limit,
	}
}

func newEventStoreFromState(state eventStoreState) *eventStore {
	items := append([]EventEnvelope{}, state.items...)
	pending := append([]EventEnvelope{}, state.pendingArchive...)
	limit := state.limit
	if limit <= 0 {
		limit = eventRetentionLimitFromEnv()
	}
	return &eventStore{
		items:          items,
		pendingArchive: pending,
		archivedCount:  state.archivedCount,
		limit:          limit,
	}
}

func (s *eventStore) state() eventStoreState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return eventStoreState{
		items:          append([]EventEnvelope{}, s.items...),
		pendingArchive: append([]EventEnvelope{}, s.pendingArchive...),
		archivedCount:  s.archivedCount,
		limit:          s.limit,
	}
}

func (s *eventStore) recent() []EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventEnvelope, len(s.items))
	copy(out, s.items)
	return out
}

func (s *eventStore) pendingArchiveList() []EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventEnvelope, len(s.pendingArchive))
	copy(out, s.pendingArchive)
	return out
}

func (s *eventStore) markArchivePersisted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.archivedCount += len(s.pendingArchive)
	s.pendingArchive = []EventEnvelope{}
}

func (s *eventStore) list() []EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventEnvelope, 0, len(s.pendingArchive)+len(s.items))
	out = append(out, s.pendingArchive...)
	out = append(out, s.items...)
	return out
}

func (s *eventStore) append(item EventEnvelope) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	if s.limit > 0 && len(s.items) > s.limit {
		overflow := len(s.items) - s.limit
		s.pendingArchive = append(s.pendingArchive, s.items[:overflow]...)
		s.items = append([]EventEnvelope{}, s.items[overflow:]...)
	}
	return nil
}

func (s *eventStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.archivedCount + len(s.pendingArchive) + len(s.items)
}

func splitRetainedEvents(evidence []EventEnvelope, limit int) ([]EventEnvelope, []EventEnvelope) {
	if evidence == nil {
		evidence = []EventEnvelope{}
	}
	if limit <= 0 || len(evidence) <= limit {
		return append([]EventEnvelope{}, evidence...), []EventEnvelope{}
	}
	split := len(evidence) - limit
	archive := append([]EventEnvelope{}, evidence[:split]...)
	recent := append([]EventEnvelope{}, evidence[split:]...)
	return recent, archive
}

func dedupeEvents(evidence []EventEnvelope) []EventEnvelope {
	if evidence == nil {
		return []EventEnvelope{}
	}
	seen := map[string]struct{}{}
	out := make([]EventEnvelope, 0, len(evidence))
	for _, item := range evidence {
		key := strings.TrimSpace(item.EventID)
		if key == "" {
			out = append(out, item)
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func eventRetentionLimitFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("LAYER_OS_EVENT_RETENTION"))
	if raw == "" {
		return defaultEventRetentionLimit
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultEventRetentionLimit
	}
	if value < 0 {
		return defaultEventRetentionLimit
	}
	return value
}

func newEvent(kind string, actor string, surface Surface, workItemID string, stage Stage, data map[string]any) EventEnvelope {
	return EventEnvelope{
		EventID:    workItemID + ":" + kind + ":" + zeroSafeNow().Format("20060102T150405.000000000"),
		Kind:       kind,
		Actor:      actor,
		Surface:    surface,
		WorkItemID: workItemID,
		Stage:      stage,
		Timestamp:  zeroSafeNow(),
		Data:       data,
	}
}

func (e EventEnvelope) Validate() error {
	if strings.TrimSpace(e.EventID) == "" {
		return errors.New("event_id is required")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("event kind is required")
	}
	if strings.TrimSpace(e.Actor) == "" {
		return errors.New("event actor is required")
	}
	if !validSurface(e.Surface) {
		return errors.New("event surface is invalid")
	}
	if strings.TrimSpace(e.WorkItemID) == "" {
		return errors.New("event work_item_id is required")
	}
	if !validStage(e.Stage) {
		return errors.New("event stage is invalid")
	}
	if e.Timestamp.IsZero() {
		return errors.New("event timestamp is required")
	}
	if e.Data == nil {
		return errors.New("event data is required")
	}
	if err := validateJSONObject(e.Data, "event data"); err != nil {
		return err
	}
	return nil
}

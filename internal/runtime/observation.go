package runtime

import "sync"

type observationStore struct {
	mu    sync.RWMutex
	items []ObservationRecord
}

func newObservationStore(items []ObservationRecord) *observationStore {
	if items == nil {
		items = []ObservationRecord{}
	}
	return &observationStore{items: items}
}

func (s *observationStore) list() []ObservationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ObservationRecord, len(s.items))
	copy(out, s.items)
	return out
}

func (s *observationStore) create(item ObservationRecord) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

package runtime

import "sync"

type preflightStore struct {
	mu    sync.RWMutex
	items []PreflightRecord
}

func newPreflightStore(items []PreflightRecord) *preflightStore {
	if items == nil {
		items = []PreflightRecord{}
	}
	return &preflightStore{items: items}
}

func (s *preflightStore) list() []PreflightRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PreflightRecord, len(s.items))
	copy(out, s.items)
	return out
}

func (s *preflightStore) create(item PreflightRecord) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *preflightStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

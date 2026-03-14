package runtime

import "sync"

type workStore struct {
	mu    sync.RWMutex
	items []WorkItem
}

func newWorkStore(items []WorkItem) *workStore {
	if items == nil {
		items = []WorkItem{}
	}
	return &workStore{items: items}
}

func (s *workStore) list() []WorkItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]WorkItem, len(s.items))
	copy(out, s.items)
	return out
}

func (s *workStore) create(item WorkItem) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *workStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *workStore) get(workID string) (WorkItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ID == workID {
			return item, true
		}
	}
	return WorkItem{}, false
}

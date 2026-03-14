package runtime

import "sync"

type flowStore struct {
	mu    sync.RWMutex
	items []FlowRun
}

func newFlowStore(items []FlowRun) *flowStore {
	if items == nil {
		items = []FlowRun{}
	}
	return &flowStore{items: items}
}

func (s *flowStore) list() []FlowRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FlowRun, len(s.items))
	copy(out, s.items)
	return out
}

func (s *flowStore) create(item FlowRun) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *flowStore) update(item FlowRun) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.items {
		if existing.FlowID == item.FlowID {
			s.items[index] = item
			return nil
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *flowStore) get(flowID string) (FlowRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.FlowID == flowID {
			return item, true
		}
	}
	return FlowRun{}, false
}

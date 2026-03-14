package runtime

import "sync"

type gatewayStore struct {
	mu    sync.RWMutex
	items []GatewayCall
}

func newGatewayStore(items []GatewayCall) *gatewayStore {
	if items == nil {
		items = []GatewayCall{}
	}
	return &gatewayStore{items: items}
}

func (s *gatewayStore) list() []GatewayCall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]GatewayCall, len(s.items))
	copy(out, s.items)
	return out
}

func (s *gatewayStore) create(item GatewayCall) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *gatewayStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *gatewayStore) get(callID string) (GatewayCall, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.CallID == callID {
			return item, true
		}
	}
	return GatewayCall{}, false
}

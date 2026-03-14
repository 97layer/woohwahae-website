package runtime

import "sync"

type releaseStore struct {
	mu    sync.RWMutex
	items []ReleasePacket
}

func newReleaseStore(items []ReleasePacket) *releaseStore {
	if items == nil {
		items = []ReleasePacket{}
	}
	return &releaseStore{items: items}
}

func (s *releaseStore) list() []ReleasePacket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ReleasePacket, len(s.items))
	copy(out, s.items)
	return out
}

func (s *releaseStore) create(item ReleasePacket) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *releaseStore) get(releaseID string) (ReleasePacket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ReleaseID == releaseID {
			return item, true
		}
	}
	return ReleasePacket{}, false
}

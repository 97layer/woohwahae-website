package runtime

import (
	"errors"
	"sync"
	"time"
)

type approvalStore struct {
	mu    sync.RWMutex
	items []ApprovalItem
}

func newApprovalStore(items []ApprovalItem) *approvalStore {
	if items == nil {
		items = []ApprovalItem{}
	}
	return &approvalStore{items: items}
}

func (s *approvalStore) list() []ApprovalItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ApprovalItem, len(s.items))
	copy(out, s.items)
	return out
}

func (s *approvalStore) create(item ApprovalItem) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *approvalStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, item := range s.items {
		if item.Status == "pending" {
			count++
		}
	}
	return count
}

func (s *approvalStore) resolve(approvalID string, status string, resolvedAt time.Time) (ApprovalItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.items {
		if item.ApprovalID != approvalID {
			continue
		}
		item.Status = status
		item.ResolvedAt = &resolvedAt
		if err := item.Validate(); err != nil {
			return ApprovalItem{}, err
		}
		s.items[index] = item
		return item, nil
	}
	return ApprovalItem{}, errors.New("approval_id not found")
}

func (s *approvalStore) get(approvalID string) (ApprovalItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ApprovalID == approvalID {
			return item, true
		}
	}
	return ApprovalItem{}, false
}

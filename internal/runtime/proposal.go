package runtime

import (
	"errors"
	"sync"
	"time"
)

type proposalStore struct {
	mu    sync.RWMutex
	items []ProposalItem
}

func newProposalStore(items []ProposalItem) *proposalStore {
	if items == nil {
		items = []ProposalItem{}
	}
	return &proposalStore{items: items}
}

func (s *proposalStore) list() []ProposalItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProposalItem, len(s.items))
	copy(out, s.items)
	return out
}

func (s *proposalStore) create(item ProposalItem) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *proposalStore) promote(proposalID string, workItemID string, updatedAt time.Time) (ProposalItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.ProposalID != proposalID {
			continue
		}
		if item.Status == "promoted" {
			return ProposalItem{}, errors.New("proposal already promoted")
		}
		item.Status = "promoted"
		item.PromotedWorkItemID = &workItemID
		item.UpdatedAt = updatedAt
		if err := item.Validate(); err != nil {
			return ProposalItem{}, err
		}
		s.items[index] = item
		return item, nil
	}
	return ProposalItem{}, errors.New("proposal_id not found")
}

func (s *proposalStore) get(proposalID string) (ProposalItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ProposalID == proposalID {
			return item, true
		}
	}
	return ProposalItem{}, false
}

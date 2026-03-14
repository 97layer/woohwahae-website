package runtime

import (
	"errors"
	"sync"
	"time"
)

type agentJobStore struct {
	mu    sync.RWMutex
	items []AgentJob
}

func newAgentJobStore(items []AgentJob) *agentJobStore {
	if items == nil {
		items = []AgentJob{}
	}
	return &agentJobStore{items: items}
}

func (s *agentJobStore) list() []AgentJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AgentJob, len(s.items))
	copy(out, s.items)
	return out
}

func (s *agentJobStore) create(item AgentJob) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.items {
		if existing.JobID == item.JobID {
			return errors.New("job_id already exists")
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *agentJobStore) get(jobID string) (AgentJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.JobID == jobID {
			return item, true
		}
	}
	return AgentJob{}, false
}

func (s *agentJobStore) update(jobID string, status string, notes []string, result map[string]any, updatedAt time.Time) (AgentJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.JobID != jobID {
			continue
		}
		item.Status = status
		if notes == nil {
			notes = []string{}
		}
		item.Notes = append([]string{}, notes...)
		if result != nil {
			item.Result = cloneJSONObject(result)
		}
		item.UpdatedAt = updatedAt
		if err := item.Validate(); err != nil {
			return AgentJob{}, err
		}
		s.items[index] = item
		return item, nil
	}
	return AgentJob{}, errors.New("job_id not found")
}

package runtime

import (
	"sync"
	"time"
)

type stateStore struct {
	mu    sync.RWMutex
	state CompanyState
}

func newStateStore(state CompanyState) *stateStore {
	return &stateStore{
		state: state,
	}
}

func defaultCompanyState() CompanyState {
	return CompanyState{
		ShellMode:        "founder",
		ApprovalsPending: 0,
		WorkItemsActive:  0,
		MemoryHealth:     "ready",
		DeployHealth:     "ready",
		PrimarySurface:   SurfaceCockpit,
		ActiveSurfaces:   []Surface{SurfaceCockpit, SurfaceTelegram, SurfaceAPI},
	}
}

func (s *stateStore) current() CompanyState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *stateStore) updateWorkItemsActive(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.WorkItemsActive = count
}

func (s *stateStore) updateApprovalsPending(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.ApprovalsPending = count
}

func (s *stateStore) updateLastReleaseAt(releasedAt *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.LastReleaseAt = releasedAt
}

func (s *stateStore) updateDeployHealth(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch status {
	case "failed":
		s.state.DeployHealth = "degraded"
	default:
		s.state.DeployHealth = "ready"
	}
}

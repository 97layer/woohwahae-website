package runtime

import (
	"errors"
	"os/exec"
	"strings"
	"sync"
)

type deployStore struct {
	mu    sync.RWMutex
	items []DeployRun
}

type rollbackStore struct {
	mu    sync.RWMutex
	items []RollbackRun
}

type targetStore struct {
	mu    sync.RWMutex
	items []DeployTarget
}

func newDeployStore(items []DeployRun) *deployStore {
	if items == nil {
		items = []DeployRun{}
	}
	return &deployStore{items: items}
}

func (s *deployStore) list() []DeployRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]DeployRun, len(s.items))
	copy(out, s.items)
	return out
}

func (s *deployStore) create(item DeployRun) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *deployStore) get(deployID string) (DeployRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.DeployID == deployID {
			return item, true
		}
	}
	return DeployRun{}, false
}

func newRollbackStore(items []RollbackRun) *rollbackStore {
	if items == nil {
		items = []RollbackRun{}
	}
	return &rollbackStore{items: items}
}

func (s *rollbackStore) list() []RollbackRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RollbackRun, len(s.items))
	copy(out, s.items)
	return out
}

func (s *rollbackStore) create(item RollbackRun) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *rollbackStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func newTargetStore(items []DeployTarget) *targetStore {
	if items == nil {
		items = []DeployTarget{}
	}
	return &targetStore{items: items}
}

func (s *targetStore) list() []DeployTarget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]DeployTarget, len(s.items))
	copy(out, s.items)
	return out
}

func (s *targetStore) put(item DeployTarget) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.items {
		if existing.TargetID == item.TargetID {
			s.items[index] = item
			return nil
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *targetStore) get(targetID string) (DeployTarget, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.TargetID == targetID {
			return item, true
		}
	}
	return DeployTarget{}, false
}

func executeTargetCommand(target DeployTarget) (string, error) {
	cmd := exec.Command(target.Command[0], target.Command[1:]...)
	if target.Workdir != nil && strings.TrimSpace(*target.Workdir) != "" {
		cmd.Dir = *target.Workdir
	}
	raw, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(raw)), err
}

func trimNote(note string) string {
	const max = 240
	note = strings.TrimSpace(note)
	if len(note) <= max {
		return note
	}
	return note[:max]
}

func statusFromError(err error) string {
	if err != nil {
		return "failed"
	}
	return "succeeded"
}

func buildDeployNotes(base []string, output string, err error) []string {
	notes := append([]string{}, base...)
	if output != "" {
		notes = append(notes, trimNote("output: "+output))
	}
	if err != nil {
		notes = append(notes, trimNote("error: "+err.Error()))
	}
	if len(notes) == 0 {
		notes = []string{"adapter executed"}
	}
	return notes
}

var errDeployTargetNotFound = errors.New("deploy target not found")

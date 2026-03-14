package runtime

import "sync"

type executeStore struct {
	mu    sync.RWMutex
	items []ExecuteRun
}

func newExecuteStore(items []ExecuteRun) *executeStore {
	return &executeStore{items: normalizeExecuteRuns(items)}
}

func (s *executeStore) list() []ExecuteRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ExecuteRun, len(s.items))
	copy(out, s.items)
	return out
}

func (s *executeStore) create(item ExecuteRun) error {
	item = normalizeExecuteRun(item)
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *executeStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func executeStatusFromDecision(decision PolicyDecision) string {
	switch decision.Mode {
	case "local":
		return "succeeded"
	case "single":
		return "recorded"
	default:
		return "failed"
	}
}

func buildExecuteNotes(decision PolicyDecision, base []string) []string {
	notes := append([]string{}, base...)
	switch decision.Mode {
	case "local":
		notes = append(notes, "local execution completed inside the kernel")
	case "single":
		notes = append(notes, "execution recorded for later provider-backed handling")
	case "blocked":
		notes = append(notes, "execution blocked by policy")
	default:
		notes = append(notes, "execution failed due to unsupported policy mode")
	}
	if decision.Decision == "hold" {
		notes = append(notes, "policy decision is hold")
	}
	if len(notes) == 0 {
		notes = []string{"execution recorded"}
	}
	return notes
}

package runtime

import "sync"

type memoryStore struct {
	mu     sync.RWMutex
	memory SystemMemory
}

func newMemoryStore(memory SystemMemory) *memoryStore {
	return &memoryStore{memory: memory}
}

func defaultSystemMemory() SystemMemory {
	return SystemMemory{
		CurrentFocus: "bootstrap runtime",
		NextSteps:    []string{},
		OpenRisks:    []string{},
		UpdatedAt:    zeroSafeNow(),
	}
}

func (s *memoryStore) current() SystemMemory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.memory
}

func (s *memoryStore) replace(memory SystemMemory) error {
	if err := memory.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memory = memory
	return nil
}

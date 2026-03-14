package runtime

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type verificationStore struct {
	mu    sync.RWMutex
	items []VerificationRecord
}

func newVerificationStore(items []VerificationRecord) *verificationStore {
	if items == nil {
		items = []VerificationRecord{}
	}
	return &verificationStore{items: items}
}

func (s *verificationStore) list() []VerificationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VerificationRecord, len(s.items))
	copy(out, s.items)
	return out
}

func (s *verificationStore) create(item VerificationRecord) error {
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *verificationStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func executeVerificationCommand(command []string, workdir string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	if strings.TrimSpace(workdir) != "" {
		cmd.Dir = workdir
	}
	raw, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(raw)), err
}

func normalizeVerificationCommand(command []string) ([]string, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("verification command is required")
	}
	parts := append([]string{}, command...)
	executable := strings.TrimSpace(parts[0])
	if executable == "" {
		return nil, fmt.Errorf("verification executable is required")
	}
	if !filepath.IsAbs(executable) {
		return nil, fmt.Errorf("verification executable must be an absolute path")
	}
	absExecutable, err := filepath.Abs(executable)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(absExecutable) {
		return nil, fmt.Errorf("verification executable must be an absolute path")
	}
	parts[0] = filepath.Clean(absExecutable)
	return parts, nil
}

func buildVerificationNotes(base []string, output string, err error) []string {
	notes := append([]string{}, base...)
	if output != "" {
		notes = append(notes, trimNote("output: "+output))
	}
	if err != nil {
		notes = append(notes, trimNote("error: "+err.Error()))
	}
	if len(notes) == 0 {
		notes = []string{"verification executed"}
	}
	return notes
}

func verificationStatusFromError(err error) string {
	if err != nil {
		return "failed"
	}
	return "passed"
}

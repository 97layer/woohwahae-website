package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbsorbGeminiArtifactsCreatesObservationsAndCleanup(t *testing.T) {
	repoRoot := t.TempDir()
	mustGeminiServiceWrite(t, filepath.Join(repoRoot, "task.md.resolved"), "# Task\n\nproposal_001 remains open.")
	mustGeminiServiceWrite(t, filepath.Join(repoRoot, "task.md.metadata.json"), `{"summary":"Recovered Gemini task","updatedAt":"2026-03-08T10:29:56Z","version":"2"}`)
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := service.AbsorbGeminiArtifacts(0, true)
	if err != nil {
		t.Fatalf("absorb gemini artifacts: %v", err)
	}
	if result.Considered != 1 || result.Created != 1 || result.Cleaned < 2 || result.Failed != 0 {
		t.Fatalf("unexpected absorb result: %+v", result)
	}
	items := service.ListObservations(ObservationQuery{SourceChannel: "gemini"})
	if len(items) != 1 || !containsPrefix(items[0].Refs, "gemini_sync:") {
		t.Fatalf("unexpected gemini observations: %+v", items)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "task.md.resolved")); !os.IsNotExist(err) {
		t.Fatalf("expected task artifact removed, err=%v", err)
	}
}

func TestAbsorbGeminiArtifactsSkipsExistingSyncRef(t *testing.T) {
	repoRoot := t.TempDir()
	mustGeminiServiceWrite(t, filepath.Join(repoRoot, "walkthrough.md.resolved"), "# Walkthrough\n\nthread_001 stays open.")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	first, err := service.AbsorbGeminiArtifacts(0, false)
	if err != nil {
		t.Fatalf("first absorb: %v", err)
	}
	if first.Created != 1 {
		t.Fatalf("expected first absorb to create one observation, got %+v", first)
	}
	second, err := service.AbsorbGeminiArtifacts(0, false)
	if err != nil {
		t.Fatalf("second absorb: %v", err)
	}
	if second.Existing != 1 || second.Created != 0 {
		t.Fatalf("expected second absorb to skip existing sync ref, got %+v", second)
	}
}

func mustGeminiServiceWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsPrefix(items []string, prefix string) bool {
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}
	return false
}

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbsorbCorpusMarkdownCreatesCanonicalEntryAndCleanup(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source := filepath.Join(repoRoot, "knowledge", "corpus", "entries", "analysis.md")
	mustWriteRuntimeFile(t, source, "# Queue Drift Analysis\n\n- repeated failure in planner lane\n- review-room remains open\n")

	result, err := service.AbsorbCorpusMarkdown(0, true, false)
	if err != nil {
		t.Fatalf("absorb corpus markdown: %v", err)
	}
	if result.Considered != 1 || result.Created != 1 || result.Cleaned == 0 {
		t.Fatalf("unexpected corpus recover result: %+v", result)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("expected source markdown cleanup, err=%v", err)
	}
	entries := service.ListCapitalizationEntries()
	if len(entries) != 1 {
		t.Fatalf("expected one capitalization entry, got %+v", entries)
	}
	if entries[0].SourceKind != "corpus.markdown_recovered" || !strings.Contains(strings.Join(entries[0].Decision.Items, "\n"), "source_path:knowledge/corpus/entries/analysis.md") {
		t.Fatalf("unexpected recovered corpus entry: %+v", entries[0])
	}
}

func TestAbsorbCorpusMarkdownDryRunLeavesArtifactsUntouched(t *testing.T) {
	repoRoot := t.TempDir()
	dataDir := filepath.Join(repoRoot, ".layer-os")
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	source := filepath.Join(repoRoot, "knowledge", "corpus", "entries", "analysis.md")
	mustWriteRuntimeFile(t, source, "# Audit Design\n\nKeep this for canonical recovery.\n")

	result, err := service.AbsorbCorpusMarkdown(0, true, true)
	if err != nil {
		t.Fatalf("dry-run absorb corpus markdown: %v", err)
	}
	if !result.DryRun || result.Created != 0 || result.Cleaned != 0 || result.Considered != 1 {
		t.Fatalf("unexpected dry-run corpus recover result: %+v", result)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("expected source markdown to remain during dry-run: %v", err)
	}
	if entries := service.ListCapitalizationEntries(); len(entries) != 0 {
		t.Fatalf("expected dry-run to avoid corpus writes, got %+v", entries)
	}
}

func mustWriteRuntimeFile(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

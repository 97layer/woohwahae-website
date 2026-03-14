package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanGeminiArtifactsGroupsResolvedAndMetadata(t *testing.T) {
	root := t.TempDir()
	mustGeminiWrite(t, filepath.Join(root, "task.md"), "# Task draft\n\nproposal_001 stays open.")
	mustGeminiWrite(t, filepath.Join(root, "task.md.resolved"), "# Final task\n\nproposal_001 is ready.")
	mustGeminiWrite(t, filepath.Join(root, "task.md.metadata.json"), `{"summary":"Task summary from metadata","updatedAt":"2026-03-08T10:29:56Z","version":"2"}`)

	scan, err := ScanGeminiArtifacts(root, 0)
	if err != nil {
		t.Fatalf("scan gemini artifacts: %v", err)
	}
	if len(scan.Artifacts) != 1 {
		t.Fatalf("expected one grouped artifact, got %+v", scan.Artifacts)
	}
	item := scan.Artifacts[0]
	if item.RelativePath != "task.md.resolved" {
		t.Fatalf("expected resolved primary path, got %+v", item)
	}
	if item.Topic != "task" || item.ArtifactType != "task" {
		t.Fatalf("unexpected gemini artifact classification: %+v", item)
	}
	if item.Summary != "Task summary from metadata" {
		t.Fatalf("expected metadata summary, got %+v", item)
	}
	if item.SyncKey == "" || !containsString(item.RelatedRelativePaths, "task.md.metadata.json") {
		t.Fatalf("expected sync key and metadata relation, got %+v", item)
	}
	if !containsString(item.Refs, "proposal_001") || !containsString(item.Refs, "gemini_path:task.md.resolved") {
		t.Fatalf("unexpected gemini refs: %+v", item.Refs)
	}
}

func TestGeminiObservationUsesDeterministicSyncRef(t *testing.T) {
	obs := GeminiObservation(GeminiArtifact{
		Stem:         "walkthrough.md",
		Label:        "agent artifact markdown",
		Topic:        "walkthrough",
		ArtifactType: "walkthrough",
		Summary:      "Verifier walkthrough",
		RawExcerpt:   "Open thread thread_001 needs closure.",
		Confidence:   "high",
		SyncKey:      "sync-001",
		Refs:         []string{"thread_001", "gemini_path:walkthrough.md.resolved"},
		Severity:     "medium",
	})
	if obs.SourceChannel != "gemini" || obs.ObservationID == "" {
		t.Fatalf("unexpected gemini observation: %+v", obs)
	}
	if !containsString(obs.Refs, "gemini_sync:sync-001") || !containsString(obs.Refs, "source_family:agent_workspace") {
		t.Fatalf("expected deterministic sync/family refs, got %+v", obs.Refs)
	}
}

func mustGeminiWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

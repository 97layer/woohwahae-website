package runtime

import (
	"path/filepath"
	"strings"
	"testing"

	"os"
)

func TestScanAntigravityRunsPrefersResolvedAndMetadata(t *testing.T) {
	root := t.TempDir()
	runRoot := filepath.Join(root, "run-001")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatalf("mkdir run root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runRoot, "task.md"), []byte("# Original task\n\nlegacy"), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runRoot, "task.md.resolved.1"), []byte("# Older resolved\n\nintel_001"), 0o644); err != nil {
		t.Fatalf("write old resolved: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runRoot, "task.md.resolved"), []byte("# Primary task\n\nProposalItem proposal_001 remains open."), 0o644); err != nil {
		t.Fatalf("write primary resolved: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runRoot, "task.md.metadata.json"), []byte(`{"artifactType":"ARTIFACT_TYPE_TASK","summary":"Task summary from metadata","updatedAt":"2026-03-08T10:29:56.121373Z","version":"2"}`), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runRoot, "media__1771917569587.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("write media: %v", err)
	}

	scan, err := ScanAntigravityRuns(root, "", 0)
	if err != nil {
		t.Fatalf("scan antigravity runs: %v", err)
	}
	if len(scan.Runs) != 1 {
		t.Fatalf("expected 1 run, got %+v", scan.Runs)
	}
	run := scan.Runs[0]
	if run.RunID != "run-001" || len(run.MediaPaths) != 1 {
		t.Fatalf("unexpected run: %+v", run)
	}
	if len(run.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %+v", run.Artifacts)
	}
	artifact := run.Artifacts[0]
	if artifact.PrimaryPath != filepath.Join(runRoot, "task.md.resolved") {
		t.Fatalf("expected primary resolved path, got %q", artifact.PrimaryPath)
	}
	if artifact.Topic != "task" || artifact.Summary != "Task summary from metadata" {
		t.Fatalf("unexpected artifact topic/summary: %+v", artifact)
	}
	if artifact.Version == nil || *artifact.Version != "2" {
		t.Fatalf("expected metadata version, got %+v", artifact.Version)
	}
	if !containsString(artifact.Refs, "proposal_001") || !containsString(artifact.Refs, "antigravity_run:run-001") {
		t.Fatalf("expected artifact refs, got %+v", artifact.Refs)
	}
	if !containsString(artifact.Refs, "source:antigravity") || !containsString(artifact.Refs, "source_family:agent_workspace") || !containsString(artifact.Refs, "topic:task") {
		t.Fatalf("expected artifact taxonomy refs, got %+v", artifact.Refs)
	}
	if artifact.SyncKey == "" || run.SyncKey == "" {
		t.Fatalf("expected sync keys, got artifact=%q run=%q", artifact.SyncKey, run.SyncKey)
	}
}

func TestAntigravityObservationUsesDeterministicSyncRef(t *testing.T) {
	obs := AntigravityObservation(AntigravityArtifact{
		RunID:      "run-001",
		Topic:      "diagnosis",
		Summary:    "Integrated diagnosis summary",
		RawExcerpt: "Integrated diagnosis excerpt",
		Confidence: "high",
		CapturedAt: zeroSafeNow(),
		SyncKey:    "sync-001",
		Refs:       []string{"proposal_001"},
	})
	if obs.SourceChannel != "antigravity" || obs.Topic != "diagnosis" {
		t.Fatalf("unexpected observation: %+v", obs)
	}
	if !containsString(obs.Refs, "antigravity_sync:sync-001") {
		t.Fatalf("expected sync ref in observation refs, got %+v", obs.Refs)
	}
	if !containsString(obs.Refs, "source:antigravity") || !containsString(obs.Refs, "source_family:agent_workspace") || !containsString(obs.Refs, "topic:diagnosis") {
		t.Fatalf("expected taxonomy refs in observation, got %+v", obs.Refs)
	}
	if !strings.HasPrefix(obs.ObservationID, "observation_antigravity_") {
		t.Fatalf("unexpected observation id: %q", obs.ObservationID)
	}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

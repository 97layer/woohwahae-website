package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditGeminiPassesProjectPolicy(t *testing.T) {
	root := t.TempDir()
	mustAuditMkdir(t, filepath.Join(root, ".gemini"))
	mustAuditWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), `# Layer OS — Gemini Project Override
report-only
파일을 만들지 말고 채팅으로만 답한다
layer-osctl job report
layer-osctl ingest gemini|antigravity|telegram|content
layer-osctl audit gemini
layer-osctl audit residue
.layer-os/*.json
`)

	audit := AuditGemini(root)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit, got %+v", audit)
	}
	if !audit.ProjectPolicyPresent {
		t.Fatalf("expected project policy present, got %+v", audit)
	}
	if len(audit.PolicyIssues) != 0 || len(audit.ArtifactMatches) != 0 {
		t.Fatalf("expected clean gemini audit, got %+v", audit)
	}
}

func TestAuditGeminiFindsMissingPolicyAndArtifacts(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "council_room.md"), "legacy scratchpad")
	mustAuditWrite(t, filepath.Join(root, "task.md.resolved"), "artifact output")

	audit := AuditGemini(root)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if audit.ProjectPolicyPresent {
		t.Fatalf("expected missing project policy, got %+v", audit)
	}
	if len(audit.PolicyIssues) == 0 {
		t.Fatalf("expected policy issues, got %+v", audit)
	}
	joined := strings.Join(audit.ArtifactMatches, "\n")
	if !strings.Contains(joined, "council_room.md") || !strings.Contains(joined, "task.md.resolved") {
		t.Fatalf("expected stray artifact matches, got %+v", audit.ArtifactMatches)
	}
}

func mustAuditMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustAuditWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

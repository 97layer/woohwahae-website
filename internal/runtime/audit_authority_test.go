package runtime

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditAuthorityPassesCanonicalBoundary(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "AGENTS.md"), "Legacy material under `/Users/97layer/97layerOS` is reference-only in this workspace.\nauthority must come from this workspace's `AGENTS.md`, `constitution/`, `docs/`, and `contracts/`.\nreport back via `/api/layer-os/jobs/report`.\nDo not rewrite `.layer-os/review_room.json` directly.\n")
	mustAuditWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "레거시 `97layerOS`는 이 워크스페이스에서 참고 전용\n`review_room.json`은 런타임 상태다. 직접 수정하지 않는다.\n`layer-osctl job report ...` 또는 `/api/layer-os/jobs/report`\nlayer-osctl audit authority --strict\n")
	audit := AuditAuthority(root)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit, got %+v", audit)
	}
	if err := audit.Boundary.Validate(); err != nil {
		t.Fatalf("expected valid authority boundary, got %v", err)
	}
}

func TestAuditAuthorityFindsPolicyAndLegacyLeaks(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "AGENTS.md"), "legacy")
	mustAuditWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "verify only")
	mustAuditWrite(t, filepath.Join(root, "cmd", "layer-osctl", "main.go"), "package main\n// /Users/97layer/97layerOS legacy path\n")
	audit := AuditAuthority(root)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if len(audit.PolicyIssues) == 0 {
		t.Fatalf("expected policy issues, got %+v", audit)
	}
	if !strings.Contains(strings.Join(audit.LegacyCodeMatches, "\n"), "cmd/layer-osctl/main.go") {
		t.Fatalf("expected legacy code match, got %+v", audit.LegacyCodeMatches)
	}
}

func TestAuditAuthorityFindsDirectRuntimeJSONScriptAccess(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "AGENTS.md"), "Legacy material under `/Users/97layer/97layerOS` is reference-only in this workspace.\nauthority must come from this workspace's `AGENTS.md`, `constitution/`, `docs/`, and `contracts/`.\nreport back via `/api/layer-os/jobs/report`.\nDo not rewrite `.layer-os/review_room.json` directly.\n")
	mustAuditWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "레거시 `97layerOS`는 이 워크스페이스에서 참고 전용\n`review_room.json`은 런타임 상태다. 직접 수정하지 않는다.\n`layer-osctl job report ...` 또는 `/api/layer-os/jobs/report`\nlayer-osctl audit authority --strict\n")
	mustAuditWrite(t, filepath.Join(root, "scripts", "cross-check.sh"), "cat .layer-os/review_room.json\n")

	audit := AuditAuthority(root)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	joined := strings.Join(audit.PolicyIssues, "\n")
	if !strings.Contains(joined, "scripts/cross-check.sh:1 should use layer-osctl/api surfaces instead of direct runtime json access") {
		t.Fatalf("expected script runtime json issue, got %+v", audit.PolicyIssues)
	}
}

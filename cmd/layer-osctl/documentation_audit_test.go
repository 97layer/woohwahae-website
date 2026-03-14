package main

import (
	"path/filepath"
	"testing"

	"layer-os/internal/runtime"
)

func TestSelectAuditDocumentationReportsIssues(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n## Related Docs\n\n- `docs/missing.md` explains the missing dependency.\n")

	result, hasIssues, err := selectAudit("docs", root, filepath.Join(root, ".layer-os"))
	if err != nil {
		t.Fatalf("select audit docs: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected docs audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.DocumentationAudit)
	if !ok {
		t.Fatalf("expected DocumentationAudit result, got %T", result)
	}
	if audit.Status != "degraded" || len(audit.MissingRefs) == 0 {
		t.Fatalf("expected degraded documentation audit, got %+v", audit)
	}
}

func TestSelectAuditDocumentationIncludesInventoryBuckets(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "# Entry\n")
	mustWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n")
	mustWrite(t, filepath.Join(root, "docs", "archive", "history.md"), "# History\n")
	mustWrite(t, filepath.Join(root, "docs", "agent_integration.md"), "# Alias\n")
	mustWrite(t, filepath.Join(root, "docs", "brand-home", "node_modules", "pkg", "README.md"), "# Vendor\n")
	mustWrite(t, filepath.Join(root, "skills", "active", "verify", "SKILL.md"), "# Skill\n")
	mustWrite(t, filepath.Join(root, ".layer-os", "artifact.md"), "# Artifact\n")
	mustWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "# Local\n")

	result, hasIssues, err := selectAudit("docs", root, filepath.Join(root, ".layer-os"))
	if err != nil {
		t.Fatalf("select audit docs: %v", err)
	}
	if hasIssues {
		t.Fatalf("expected clean docs inventory, got issues result=%+v", result)
	}
	audit, ok := result.(runtime.DocumentationAudit)
	if !ok {
		t.Fatalf("expected DocumentationAudit result, got %T", result)
	}
	if audit.Inventory.TotalMarkdownPaths != 8 {
		t.Fatalf("expected full markdown inventory, got %+v", audit.Inventory)
	}
	if audit.Inventory.ActiveCorePaths != 3 || audit.Inventory.AuthorityPaths != 2 || audit.Inventory.CompatibilityPaths != 1 {
		t.Fatalf("expected active/authority/compatibility counts, got %+v", audit.Inventory)
	}
	if audit.Inventory.ArchivedPaths != 1 || audit.Inventory.VendoredPaths != 1 || audit.Inventory.SkillManualPaths != 1 {
		t.Fatalf("expected archive/vendor/skill counts, got %+v", audit.Inventory)
	}
	if audit.Inventory.RuntimeArtifactPaths != 1 || audit.Inventory.AgentLocalPaths != 1 {
		t.Fatalf("expected runtime/agent-local counts, got %+v", audit.Inventory)
	}
}

package runtime

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditDocumentationPassesHealthyDocs(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "AGENTS.md"), "workspace")
	mustAuditWrite(t, filepath.Join(root, "README.md"), "See `docs/guide.md`.")
	mustAuditWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n## Related Docs\n\n- `docs/reference.md` explains the adjacent policy.\n\nSee `constitution/charter.md`.\n")
	mustAuditWrite(t, filepath.Join(root, "docs", "reference.md"), "# Reference\n")
	mustAuditWrite(t, filepath.Join(root, "constitution", "charter.md"), "# Charter\n")

	audit := AuditDocumentation(root)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit, got %+v", audit)
	}
	if audit.CheckedLocalRefCount != 3 {
		t.Fatalf("expected 3 checked refs, got %+v", audit)
	}
}

func TestAuditDocumentationFindsMissingRefsLegacyLabelsAndHeadingDrift(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n## Related docs\n\n- `docs/missing.md` explains the lane.\n\nSA owns this lane.\n")

	audit := AuditDocumentation(root)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	joinedIssues := strings.Join(audit.Issues, "\n")
	for _, want := range []string{
		"docs/guide.md -> docs/missing.md",
		"docs/guide.md: use `## Related Docs` heading casing",
		"docs/guide.md: legacy label reference found:",
	} {
		if !strings.Contains(joinedIssues, want) {
			t.Fatalf("expected issue containing %q, got %+v", want, audit)
		}
	}
}

func TestAuditDocumentationSkipsExternalRefsButChecksRelativeMarkdownLinks(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\nSee [reference](./reference.md) and https://example.com/spec for the hosted mirror.\n")
	mustAuditWrite(t, filepath.Join(root, "docs", "reference.md"), "# Reference\n")

	audit := AuditDocumentation(root)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit, got %+v", audit)
	}
	if len(audit.ExternalRefsSkipped) != 1 || audit.ExternalRefsSkipped[0] != "https://example.com/spec" {
		t.Fatalf("expected external ref to be tracked as deferred, got %+v", audit.ExternalRefsSkipped)
	}
}

func TestAuditDocumentationRequiresControlTowerPromptRules(t *testing.T) {
	root := t.TempDir()
	mustAuditWrite(t, filepath.Join(root, "docs", "prompting.md"), "# Prompting\n")

	audit := AuditDocumentation(root)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	joinedIssues := strings.Join(audit.Issues, "\n")
	for _, want := range []string{
		"docs/prompting.md: required operating rule missing: lead with current system facts, changed files, runtime state, risks, and next watchpoints",
		"docs/prompting.md: required operating rule missing: do not answer a status/tracking request with an implementation plan or redesign pitch",
		"docs/prompting.md: required operating rule missing: avoid opening with phrases like `intent layer`, `router`, `canonical action`, or similar implementation shorthand",
	} {
		if !strings.Contains(joinedIssues, want) {
			t.Fatalf("expected issue containing %q, got %+v", want, audit)
		}
	}
}

func TestAuditDocumentationWorkspacePasses(t *testing.T) {
	root := filepath.Join("..", "..")
	audit := AuditDocumentation(root)
	if audit.Status != "ok" {
		t.Fatalf("expected workspace documentation audit ok, got %+v", audit)
	}
}

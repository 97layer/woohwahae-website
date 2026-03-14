package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalContractCatalogIsUniqueAndTiered(t *testing.T) {
	catalog := CanonicalContractCatalog()
	if len(catalog) != 42 {
		t.Fatalf("expected 42 canonical contracts, got %d", len(catalog))
	}
	if got := len(ContractTitlesByTier(ContractTierSpine)); got != 30 {
		t.Fatalf("expected 30 spine contracts, got %d", got)
	}
	if got := len(ContractTitlesByTier(ContractTierHarness)); got != 12 {
		t.Fatalf("expected 12 harness contracts, got %d", got)
	}

	seenSchema := map[string]bool{}
	seenTitle := map[string]bool{}
	for _, spec := range catalog {
		if seenSchema[spec.Schema] {
			t.Fatalf("duplicate schema in catalog: %s", spec.Schema)
		}
		if seenTitle[spec.Title] {
			t.Fatalf("duplicate title in catalog: %s", spec.Title)
		}
		seenSchema[spec.Schema] = true
		seenTitle[spec.Title] = true
	}
}

func TestAuditContractsFindsSchemaTitleAndArchitectureDrift(t *testing.T) {
	root := t.TempDir()
	seedCanonicalContractAuditTree(t, root)

	mustAuditWrite(t, filepath.Join(root, "contracts", "branch.schema.json"), fakeContractSchemaJSON("BranchLane"))
	mustAuditWrite(t, filepath.Join(root, "docs", "architecture.md"), strings.Replace(fakeArchitectureContractSource(), "- `Branch`", "- `SessionBootstrapPacket`", 1))

	audit := AuditContracts(root)
	if len(audit.TitleIssues) == 0 {
		t.Fatalf("expected title issues, got %+v", audit)
	}
	if len(audit.ArchitectureIssues) == 0 {
		t.Fatalf("expected architecture issues, got %+v", audit)
	}

	joinedTitles := strings.Join(audit.TitleIssues, "\n")
	if !strings.Contains(joinedTitles, "branch.schema.json") {
		t.Fatalf("expected branch title drift, got %+v", audit.TitleIssues)
	}

	joinedArchitecture := strings.Join(audit.ArchitectureIssues, "\n")
	for _, want := range []string{
		"contract architecture harness missing: Branch",
		"contract architecture harness unexpected: SessionBootstrapPacket",
		"contract architecture duplicate tier listing: SessionBootstrapPacket",
	} {
		if !strings.Contains(joinedArchitecture, want) {
			t.Fatalf("expected architecture issue containing %q, got %+v", want, audit.ArchitectureIssues)
		}
	}
}

func TestAuditContractsWorkspacePasses(t *testing.T) {
	root := filepath.Join("..", "..")
	audit := AuditContracts(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected workspace contract audit ok, got %+v", audit)
	}
}

func seedCanonicalContractAuditTree(t *testing.T, root string) {
	t.Helper()
	for _, spec := range CanonicalContractCatalog() {
		mustAuditWrite(t, filepath.Join(root, "contracts", spec.Schema), fakeContractSchemaJSON(spec.Title))
	}
	mustAuditWrite(t, filepath.Join(root, "docs", "architecture.md"), fakeArchitectureContractSource())
}

func fakeContractSchemaJSON(title string) string {
	return fmt.Sprintf("{\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\n  \"title\": %q,\n  \"type\": \"object\"\n}\n", title)
}

func fakeArchitectureContractSource() string {
	lines := []string{"# Architecture", "", "## System Spine", ""}
	for idx, title := range ContractTitlesByTier(ContractTierSpine) {
		lines = append(lines, fmt.Sprintf("%d. `%s`", idx+1, title))
	}
	lines = append(lines, "", "## Contract Tiers", "")
	for _, title := range ContractTitlesByTier(ContractTierHarness) {
		lines = append(lines, fmt.Sprintf("- `%s`", title))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestAuditStructurePassesCleanTree(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustMkdir(t, filepath.Join(root, ".gemini"))
	mustWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "project override")
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")

	audit := auditStructure(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %v", audit.Issues)
	}
}

func TestAuditStructureFindsUnexpectedEntries(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustMkdir(t, filepath.Join(root, "internal", "shared"))
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")
	mustWrite(t, filepath.Join(root, "test.tmp"), "")
	mustMkdir(t, filepath.Join(root, "tmp"))

	audit := auditStructure(root)
	if len(audit.Issues) == 0 {
		t.Fatal("expected issues, got none")
	}
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "unexpected root entry: tmp") {
		t.Fatalf("expected unexpected root entry, got %v", audit.Issues)
	}
	if !strings.Contains(joined, "unexpected internal entry: shared") {
		t.Fatalf("expected unexpected internal entry, got %v", audit.Issues)
	}
	if !strings.Contains(joined, "stray file: test.tmp") {
		t.Fatalf("expected stray file issue, got %v", audit.Issues)
	}
}

func TestAuditStructureIgnoresOperationalRootsAndGeneratedSubtrees(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs", "brand-home", "node_modules", "next", "dist", "shared", "lib", "utils"))
	mustMkdir(t, filepath.Join(root, "docs", "brand-home", ".next", "cache", "helpers"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustMkdir(t, filepath.Join(root, ".cache"))
	mustMkdir(t, filepath.Join(root, ".omx"))
	mustMkdir(t, filepath.Join(root, ".tmp"))
	mustMkdir(t, filepath.Join(root, "bin"))
	mustMkdir(t, filepath.Join(root, "scripts"))
	mustWrite(t, filepath.Join(root, "docs", "brand-home", "package.json"), "{}")
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")

	audit := auditStructure(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %v", audit.Issues)
	}
}

func TestAuditStructureAllowsLocalDeprecatedResidue(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")
	mustMkdir(t, filepath.Join(root, ".playwright-cli"))
	mustMkdir(t, filepath.Join(root, ".tmp-search-verify"))
	mustWrite(t, filepath.Join(root, "layer-osctl"), "binary placeholder")

	audit := auditStructure(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected local deprecated residue to be ignored, got %v", audit.Issues)
	}
}

func TestAuditStructureFlagsDatedReportsOutsideArchive(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs", "archive", "reports"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")
	mustWrite(t, filepath.Join(root, "docs", "system-scan-2026-03-13.md"), "scan")
	mustWrite(t, filepath.Join(root, "docs", "archive", "reports", "system-scan-2026-03-13.md"), "archived")

	audit := auditStructure(root)
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "dated report should live under docs/archive/reports: docs/system-scan-2026-03-13.md") {
		t.Fatalf("expected dated report issue, got %v", audit.Issues)
	}
}

func TestAuditStructureFlagsTrackedGeneratedTrees(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs", "brand-home", "node_modules"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")
	mustWrite(t, filepath.Join(root, ".gitignore"), "**/node_modules/\n")
	mustWrite(t, filepath.Join(root, ".playwright-cli", "log.txt"), "playwright")
	mustWrite(t, filepath.Join(root, "layer-osctl"), "binary")
	mustWrite(t, filepath.Join(root, "docs", "brand-home", "node_modules", "index.js"), "module.exports = {}\n")
	runGit(t, root, "init")
	runGit(t, root, "add", ".playwright-cli/log.txt")
	runGit(t, root, "add", "layer-osctl")
	runGit(t, root, "add", "-f", "docs/brand-home/node_modules/index.js")

	audit := auditStructure(root)
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "tracked deprecated path: .playwright-cli") {
		t.Fatalf("expected tracked deprecated path issue, got %v", audit.Issues)
	}
	if !strings.Contains(joined, "tracked deprecated path: layer-osctl") {
		t.Fatalf("expected tracked root binary issue, got %v", audit.Issues)
	}
	if !strings.Contains(joined, "tracked generated path: docs/brand-home/node_modules") {
		t.Fatalf("expected tracked generated path issue, got %v", audit.Issues)
	}
}

func TestAuditStructureFlagsHotspotBudget(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "internal", "runtime"))
	mustWrite(t, filepath.Join(root, "README.md"), "readme")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, "go.mod"), "module test")
	mustWrite(t, filepath.Join(root, "internal", "runtime", "types.go"), strings.Repeat("line\n", 401))

	audit := auditStructure(root)
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "hotspot file exceeds line budget: internal/runtime/types.go") {
		t.Fatalf("expected hotspot issue, got %v", audit.Issues)
	}
}

func TestAuditResidueFindsLegacyTerms(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustWrite(t, filepath.Join(root, "README.md"), "clean")
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "main.go"), `package main
// plan_council residue
`)

	audit := auditResidue(root)
	if len(audit.Matches) == 0 {
		t.Fatal("expected residue match, got none")
	}
}

func TestAuditResiduePassesCleanTree(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustWrite(t, filepath.Join(root, "README.md"), "clean")
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "main.go"), `package main
`)

	audit := auditResidue(root)
	if len(audit.Matches) != 0 {
		t.Fatalf("expected no residue, got %v", audit.Matches)
	}
}

func TestAuditResidueFindsStrayGeminiScratchpads(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustWrite(t, filepath.Join(root, "README.md"), "clean")
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "main.go"), `package main
`)
	mustWrite(t, filepath.Join(root, "council_room.md"), "legacy scratchpad")
	mustWrite(t, filepath.Join(root, "task.md.resolved"), "artifact output")

	audit := auditResidue(root)
	joined := strings.Join(audit.Matches, "\n")
	if !strings.Contains(joined, "council_room.md") || !strings.Contains(joined, "task.md.resolved") {
		t.Fatalf("expected stray artifact matches, got %v", audit.Matches)
	}
}

func TestSelectAuditGeminiReportsIssues(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "task.md.resolved"), "artifact output")

	result, hasIssues, err := selectAudit("gemini", root, filepath.Join(root, ".layer-os"))
	if err != nil {
		t.Fatalf("select audit gemini: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected gemini audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.GeminiAudit)
	if !ok {
		t.Fatalf("expected GeminiAudit result, got %T", result)
	}
	if audit.Status != "degraded" || len(audit.ArtifactMatches) == 0 {
		t.Fatalf("expected degraded gemini audit, got %+v", audit)
	}
}

func TestAuditContractsPassesExpectedTree(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "docs"))
	for _, spec := range runtime.CanonicalContractCatalog() {
		mustWrite(t, filepath.Join(root, "contracts", spec.Schema), fakeContractSchemaJSON(spec.Title))
	}
	mustWrite(t, filepath.Join(root, "docs", "architecture.md"), fakeArchitectureContractSource())
	audit := auditContracts(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %+v", audit)
	}
}

func TestSelectAuditSecurityReportsIssues(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, ".layer-os")
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "security_007"
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{
		Text:     "[security_007] 운영 보안 posture 상향",
		Kind:     "agenda",
		Severity: "high",
		Source:   "test.security",
		Ref:      &ref,
	}); err != nil {
		t.Fatalf("seed security review item: %v", err)
	}

	result, hasIssues, err := selectAudit("security", root, dataDir)
	if err != nil {
		t.Fatalf("select audit security: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected security audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.SecurityAudit)
	if !ok {
		t.Fatalf("expected SecurityAudit result, got %T", result)
	}
	if audit.Status != "degraded" || len(audit.Issues) == 0 {
		t.Fatalf("expected degraded security audit, got %+v", audit)
	}
}

func TestSelectAuditMCPReportsIssues(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.toml")
	mustWrite(t, configPath, `
[mcp_servers.omx_code_intel]
command = "node"
args = ["/tmp/intel.js"]
`)
	t.Setenv("CODEX_CONFIG", configPath)

	result, hasIssues, err := selectAudit("mcp", root, filepath.Join(root, ".layer-os"))
	if err != nil {
		t.Fatalf("select audit mcp: %v", err)
	}
	if hasIssues {
		t.Fatalf("expected optional mcp audit not to fail strict mode, got result=%+v", result)
	}
	audit, ok := result.(runtime.MCPAudit)
	if !ok {
		t.Fatalf("expected MCPAudit result, got %T", result)
	}
	if audit.Status != "ok" || len(audit.Issues) != 0 {
		t.Fatalf("expected ok mcp audit with optional-only findings, got %+v", audit)
	}
}

func TestSelectAuditRuntimeDataReportsIssues(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, ".layer-os")
	mustMkdir(t, dataDir)
	mustWrite(t, filepath.Join(dataDir, "proposals.json"), "{}\n{}")

	result, hasIssues, err := selectAudit("runtime-data", root, dataDir)
	if err != nil {
		t.Fatalf("select audit runtime-data: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected runtime-data audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.RuntimeDataAudit)
	if !ok {
		t.Fatalf("expected RuntimeDataAudit result, got %T", result)
	}
	if audit.Status != "degraded" || len(audit.Issues) == 0 {
		t.Fatalf("expected degraded runtime-data audit, got %+v", audit)
	}
}

func TestAuditSecurityPassesWithFreshReview(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, ".layer-os")
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	if err := service.CreatePreflight(runtime.PreflightRecord{
		RecordID:    "security_review_001",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create preflight: %v", err)
	}
	audit := auditSecurity(root, dataDir)
	if audit.Status != "ok" {
		t.Fatalf("expected ok security audit, got %+v", audit)
	}
}

func TestAuditReviewRoomReturnsSealStatus(t *testing.T) {
	dataDir := t.TempDir()
	audit := auditReviewRoom(dataDir)
	if !audit.Sealed && len(audit.Issues) == 0 {
		t.Fatalf("expected review-room audit to report missing seal or seal state, got %+v", audit)
	}
}

func TestAuditSurfacePassesExpectedSurface(t *testing.T) {
	root := t.TempDir()
	spec := runtime.CanonicalSurfaceSpec()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "root_dispatch.go"), fakeRootDispatchSource(spec.CLIDispatch))
	mustWrite(t, filepath.Join(root, "internal", "api", "router.go"), fakeRouterSource(spec.APIRoutes[:11]))
	mustWrite(t, filepath.Join(root, "internal", "api", "router_audit.go"), fakeAuditRouterSource(spec.APIRoutes[11:]))

	audit := auditSurface(root)
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %v\nUsage Missing: %v\nUsage Unexpected: %v\nDispatch Missing: %v\nDispatch Unexpected: %v",
			audit.Issues,
			audit.MissingCLIUsage,
			audit.UnexpectedCLIUsage,
			audit.MissingCLIDispatch,
			audit.UnexpectedCLIDispatch)
	}
}

func TestAuditSurfaceFindsDrift(t *testing.T) {
	root := t.TempDir()
	spec := runtime.CanonicalSurfaceSpec()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "root_dispatch.go"), fakeRootDispatchSource(spec.CLIDispatch[:1]))
	mustWrite(t, filepath.Join(root, "internal", "api", "router.go"), fakeRouterSource(spec.APIRoutes[:1]))
	mustWrite(t, filepath.Join(root, "internal", "api", "router_audit.go"), fakeAuditRouterSource(spec.APIRoutes[1:2]))

	audit := auditSurface(root)
	if len(audit.Issues) == 0 {
		t.Fatal("expected surface issues, got none")
	}
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "missing cli dispatch surface") || !strings.Contains(joined, "missing api route surface") {
		t.Fatalf("expected both cli/api surface issues, got %v", audit.Issues)
	}
}

func TestAuditSurfaceFindsUnexpectedLiveSurface(t *testing.T) {
	root := t.TempDir()
	spec := runtime.CanonicalSurfaceSpec()
	mustMkdir(t, filepath.Join(root, "cmd", "layer-osctl"))
	mustMkdir(t, filepath.Join(root, "internal", "api"))
	mustWrite(t, filepath.Join(root, "cmd", "layer-osctl", "root_dispatch.go"), fakeRootDispatchSource(append(spec.CLIDispatch, `"mystery":`)))
	mustWrite(t, filepath.Join(root, "internal", "api", "router.go"), fakeRouterSource(append(spec.APIRoutes[:11], "/api/layer-os/mystery")))
	mustWrite(t, filepath.Join(root, "internal", "api", "router_audit.go"), fakeAuditRouterSource(spec.APIRoutes[11:]))

	audit := auditSurface(root)
	joined := strings.Join(audit.Issues, "\n")
	for _, want := range []string{
		`unexpected cli dispatch surface: "mystery":`,
		`missing cli usage surface: mystery`,
		`unexpected api route surface: /api/layer-os/mystery`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected issue containing %q, got %v", want, audit.Issues)
		}
	}
}

func fakeRootDispatchSource(markers []string) string {
	lines := []string{"package main", "", "var rootCommandRegistry = map[string]rootCommandHandler{"}
	for _, marker := range markers {
		command := strings.TrimSpace(marker)
		command = strings.TrimPrefix(command, `"`)
		command = strings.TrimSuffix(command, `":`)
		lines = append(lines, `	"`+command+`": func(service cliService, args []string) {},`)
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func fakeRouterSource(routes []string) string {
	lines := []string{"package api", "", "func registerRoutes() {"}
	for _, route := range routes {
		lines = append(lines, `	mux.HandleFunc("`+route+`", nil)`)
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func fakeAuditRouterSource(routes []string) string {
	lines := []string{"package api", "", "var auditRouteSpecs = []auditRouteSpec{"}
	for _, route := range routes {
		lines = append(lines, `	{path: "`+route+`"},`)
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func fakeContractSchemaJSON(title string) string {
	return fmt.Sprintf("{\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\n  \"title\": %q,\n  \"type\": \"object\"\n}\n", title)
}

func fakeArchitectureContractSource() string {
	lines := []string{"# Architecture", "", "## System Spine", ""}
	for idx, title := range runtime.ContractTitlesByTier(runtime.ContractTierSpine) {
		lines = append(lines, fmt.Sprintf("%d. `%s`", idx+1, title))
	}
	lines = append(lines, "", "## Contract Tiers", "")
	for _, title := range runtime.ContractTitlesByTier(runtime.ContractTierHarness) {
		lines = append(lines, fmt.Sprintf("- `%s`", title))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func TestSelectAuditAuthorityReportsIssues(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "legacy")
	mustWrite(t, filepath.Join(root, ".gemini", "GEMINI.md"), "verify")
	result, hasIssues, err := selectAudit("authority", root, filepath.Join(root, ".layer-os"))
	if err != nil {
		t.Fatalf("select audit authority: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected authority audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.AuthorityAudit)
	if !ok {
		t.Fatalf("expected AuthorityAudit result, got %T", result)
	}
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded authority audit, got %+v", audit)
	}
}

func TestSelectAuditCorpusReportsIssues(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, ".layer-os")
	mustWrite(t, filepath.Join(root, "knowledge", "corpus", "entries", "analysis.md"), "# Drift\n")

	result, hasIssues, err := selectAudit("corpus", root, dataDir)
	if err != nil {
		t.Fatalf("select audit corpus: %v", err)
	}
	if !hasIssues {
		t.Fatalf("expected corpus audit issues, got result=%+v", result)
	}
	audit, ok := result.(runtime.CorpusAudit)
	if !ok {
		t.Fatalf("expected CorpusAudit result, got %T", result)
	}
	if audit.Status != "degraded" || len(audit.ArtifactMatches) == 0 {
		t.Fatalf("expected degraded corpus audit, got %+v", audit)
	}
}

package runtime

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type StructureAudit struct {
	RootFolders []string `json:"root_folders"`
	Issues      []string `json:"issues"`
}

type ResidueAudit struct {
	Matches []string `json:"matches"`
}

type ContractAudit struct {
	Expected           []string `json:"expected"`
	Missing            []string `json:"missing"`
	Unexpected         []string `json:"unexpected"`
	TitleIssues        []string `json:"title_issues"`
	ArchitectureIssues []string `json:"architecture_issues"`
	Issues             []string `json:"issues"`
}

type SurfaceAudit struct {
	ExpectedCLIUsage      []string `json:"expected_cli_usage"`
	MissingCLIUsage       []string `json:"missing_cli_usage"`
	UnexpectedCLIUsage    []string `json:"unexpected_cli_usage"`
	ExpectedCLIDispatch   []string `json:"expected_cli_dispatch"`
	MissingCLIDispatch    []string `json:"missing_cli_dispatch"`
	UnexpectedCLIDispatch []string `json:"unexpected_cli_dispatch"`
	ExpectedAPIRoutes     []string `json:"expected_api_routes"`
	MissingAPIRoutes      []string `json:"missing_api_routes"`
	UnexpectedAPIRoutes   []string `json:"unexpected_api_routes"`
	Issues                []string `json:"issues"`
}

type GeminiAudit struct {
	Status               string   `json:"status"`
	ProjectPolicyPath    string   `json:"project_policy_path"`
	ProjectPolicyPresent bool     `json:"project_policy_present"`
	PolicyIssues         []string `json:"policy_issues"`
	ArtifactMatches      []string `json:"artifact_matches"`
	SuggestedCommands    []string `json:"suggested_commands"`
}

type CorpusAudit struct {
	Status            string   `json:"status"`
	Root              string   `json:"root"`
	DataDir           string   `json:"data_dir"`
	ArtifactMatches   []string `json:"artifact_matches"`
	SuggestedCommands []string `json:"suggested_commands"`
}

type DaemonFreshnessAudit struct {
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	LatestSourceAt    *time.Time `json:"latest_source_at,omitempty"`
	LatestSourcePath  string     `json:"latest_source_path,omitempty"`
	Issues            []string   `json:"issues"`
	SuggestedCommands []string   `json:"suggested_commands"`
}

type AuthorityAudit struct {
	Status            string            `json:"status"`
	Boundary          AuthorityBoundary `json:"boundary"`
	PolicyIssues      []string          `json:"policy_issues"`
	LegacyCodeMatches []string          `json:"legacy_code_matches"`
	SuggestedCommands []string          `json:"suggested_commands"`
}

var (
	rootDispatchCommandRE = regexp.MustCompile(`(?m)^\s*"([^"]+)":\s*func`)
	apiHandleFuncRouteRE  = regexp.MustCompile(`HandleFunc\("([^"]+)"`)
	auditRoutePathRE      = regexp.MustCompile(`path:\s*"([^"]+)"`)
	contractNumberedItem  = regexp.MustCompile("^\\d+\\.\\s+`([^`]+)`")
	contractBulletItem    = regexp.MustCompile("^- `([^`]+)`")
	datedReportNameRE     = regexp.MustCompile(`20\d{2}-\d{2}-\d{2}`)
)

func AuditStructure(root string) StructureAudit {
	canonicalRoot := map[string]bool{
		".gemini":      true,
		".gitignore":   true,
		"AGENTS.md":    true,
		"README.md":    true,
		"go.mod":       true,
		"go.sum":       true,
		"constitution": true,
		"contracts":    true,
		"docs":         true,
		"skills":       true,
		"cmd":          true,
		"internal":     true,
		".layer-os":    true,
		"scripts":      true,
	}
	ignoredOperationalRoot := map[string]bool{
		".cache":             true,
		".omx":               true,
		".tmp":               true,
		".playwright-cli":    true,
		".tmp-search-verify": true,
		"bin":                true,
		"layer-osctl":        true,
	}
	ignoredGeneratedDirs := map[string]bool{
		".git":         true,
		".gemini":      true,
		".layer-os":    true,
		".cache":       true,
		".omx":         true,
		".tmp":         true,
		"node_modules": true,
		".next":        true,
		"bin":          true,
	}
	issues := []string{}
	rootFolders := []string{}
	entries, err := os.ReadDir(root)
	if err != nil {
		return StructureAudit{Issues: []string{"root readdir failed"}}
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			continue
		}
		rootFolders = append(rootFolders, name)
		if canonicalRoot[name] || ignoredOperationalRoot[name] || name == ".DS_Store" {
			continue
		}
		issues = append(issues, "unexpected root entry: "+name)
	}
	sort.Strings(rootFolders)

	internalRoot := filepath.Join(root, "internal")
	if internalEntries, err := os.ReadDir(internalRoot); err == nil {
		for _, entry := range internalEntries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || !entry.IsDir() {
				continue
			}
			if name != "api" && name != "runtime" {
				issues = append(issues, "unexpected internal entry: "+name)
			}
		}
	}

	forbiddenNames := []string{"shared", "common", "utils", "helpers", "misc", "compat", "bridge"}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			issues = append(issues, "walk error: "+walkErr.Error())
			return nil
		}
		name := d.Name()
		if d.IsDir() && ignoredGeneratedDirs[name] {
			return fs.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		for _, forbidden := range forbiddenNames {
			if name == forbidden {
				issues = append(issues, "forbidden path name: "+rel)
			}
		}
		if d.IsDir() {
			return nil
		}
		if name == ".DS_Store" {
			return nil
		}
		if strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".bak") || strings.HasSuffix(name, "~") {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				rel = path
			}
			issues = append(issues, "stray file: "+filepath.ToSlash(rel))
		}
		return nil
	})

	issues = append(issues, auditStructureTrackedLeaks(root)...)
	issues = append(issues, auditStructureHotspots(root)...)
	issues = append(issues, auditStructureDatedReports(root)...)
	issues = dedupeSortedIssues(issues)
	return StructureAudit{
		RootFolders: rootFolders,
		Issues:      issues,
	}
}

func dedupeSortedIssues(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func auditStructureTrackedLeaks(root string) []string {
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		return nil
	}
	cmd := exec.Command("git", "ls-files", "--",
		".playwright-cli",
		".tmp",
		".tmp-search-verify",
		".cache",
		"layer-osctl",
		"docs/brand-home/node_modules",
		"docs/brand-home/.next",
	)
	cmd.Dir = root
	raw, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.FieldsFunc(string(raw), func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	if len(lines) == 0 {
		return nil
	}
	issues := []string{}
	seen := map[string]bool{}
	for _, line := range lines {
		rel := filepath.ToSlash(strings.TrimSpace(line))
		if rel == "" {
			continue
		}
		bucket := trackedStructureLeakBucket(rel)
		if bucket == "" || seen[bucket] {
			continue
		}
		seen[bucket] = true
		switch bucket {
		case ".playwright-cli", ".tmp-search-verify", "layer-osctl":
			issues = append(issues, "tracked deprecated path: "+bucket)
		case ".tmp", ".cache":
			issues = append(issues, "tracked operational path: "+bucket)
		case "docs/brand-home/node_modules", "docs/brand-home/.next":
			issues = append(issues, "tracked generated path: "+bucket)
		default:
			issues = append(issues, "tracked structure leak: "+bucket)
		}
	}
	return issues
}

func trackedStructureLeakBucket(rel string) string {
	switch {
	case rel == ".playwright-cli" || strings.HasPrefix(rel, ".playwright-cli/"):
		return ".playwright-cli"
	case rel == ".tmp" || strings.HasPrefix(rel, ".tmp/"):
		return ".tmp"
	case rel == ".tmp-search-verify" || strings.HasPrefix(rel, ".tmp-search-verify/"):
		return ".tmp-search-verify"
	case rel == ".cache" || strings.HasPrefix(rel, ".cache/"):
		return ".cache"
	case rel == "layer-osctl":
		return "layer-osctl"
	case rel == "docs/brand-home/node_modules" || strings.HasPrefix(rel, "docs/brand-home/node_modules/"):
		return "docs/brand-home/node_modules"
	case rel == "docs/brand-home/.next" || strings.HasPrefix(rel, "docs/brand-home/.next/"):
		return "docs/brand-home/.next"
	default:
		return ""
	}
}

func auditStructureHotspots(root string) []string {
	thresholds := map[string]int{
		"internal/runtime/types.go":                400,
		"internal/runtime/disk.go":                 300,
		"internal/api/router_workflow_handlers.go": 350,
		"internal/api/source_intake_projection.go": 350,
	}
	issues := []string{}
	for rel, maxLines := range thresholds {
		path := filepath.Join(root, filepath.FromSlash(rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lineCount := 0
		if len(raw) > 0 {
			lineCount = strings.Count(string(raw), "\n") + 1
		}
		if lineCount > maxLines {
			issues = append(issues, fmt.Sprintf("hotspot file exceeds line budget: %s (%d > %d)", rel, lineCount, maxLines))
		}
	}
	return issues
}

func auditStructureDatedReports(root string) []string {
	docsRoot := filepath.Join(root, "docs")
	issues := []string{}
	_ = filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".next":
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" || !datedReportNameRE.MatchString(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "docs/archive/reports/") {
			return nil
		}
		issues = append(issues, "dated report should live under docs/archive/reports: "+rel)
		return nil
	})
	return issues
}

func AuditDaemonFreshness(root string, startedAt time.Time) DaemonFreshnessAudit {
	audit := DaemonFreshnessAudit{
		Status: "ok",
		Issues: []string{},
		SuggestedCommands: []string{
			"go run ./cmd/layer-osd",
			"go test ./...",
		},
	}
	root = strings.TrimSpace(root)
	if root == "" || startedAt.IsZero() {
		return audit
	}
	startedAt = startedAt.UTC()
	audit.StartedAt = &startedAt
	latestAt := time.Time{}
	latestPath := ""
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		name := d.Name()
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if name == ".git" || name == ".layer-os" || name == ".cache" || name == ".omx" || name == ".playwright-cli" || name == ".tmp" || name == ".tmp-search-verify" || name == "node_modules" || name == ".next" || name == "bin" {
				return fs.SkipDir
			}
			if rel == "docs/brand-home" {
				return fs.SkipDir
			}
			return nil
		}
		if rel == "layer-osctl" || strings.HasPrefix(rel, "docs/brand-home/") {
			return nil
		}
		if strings.HasPrefix(rel, ".cache/") || strings.HasPrefix(rel, ".omx/") || strings.HasPrefix(rel, ".playwright-cli/") || strings.HasPrefix(rel, ".tmp/") || strings.HasPrefix(rel, ".tmp-search-verify/") || strings.HasPrefix(rel, "bin/") || strings.HasPrefix(rel, ".layer-os/") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(latestAt) {
			latestAt = info.ModTime().UTC()
			latestPath = rel
		}
		return nil
	})
	if latestPath == "" {
		return audit
	}
	audit.LatestSourceAt = &latestAt
	audit.LatestSourcePath = latestPath
	if latestAt.After(startedAt.Add(2 * time.Second)) {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "source tree changed after current layer-osd start; restart the daemon to serve the latest code")
	}
	return audit
}

func AuditResidue(root string) ResidueAudit {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`97layerOS/`),
		regexp.MustCompile(`\bplan_council\b`),
		regexp.MustCompile(`\bqueue_manager\b`),
		regexp.MustCompile(`\bpipeline_orchestrator\b`),
		regexp.MustCompile(`\bweb-11ty-redesign\b`),
		regexp.MustCompile(`\bweb-lock-guardian\b`),
		regexp.MustCompile(`\bcode-audit\b`),
	}
	scopes := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "cmd"),
		filepath.Join(root, "contracts"),
		filepath.Join(root, "internal"),
	}
	matches := []string{}
	for _, scope := range scopes {
		info, err := os.Stat(scope)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			raw, err := os.ReadFile(scope)
			if err != nil {
				continue
			}
			for _, pattern := range patterns {
				if pattern.Match(raw) {
					matches = append(matches, filepath.Base(scope)+": "+pattern.String())
				}
			}
			continue
		}
		_ = filepath.WalkDir(scope, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			name := d.Name()
			if strings.HasSuffix(name, "_test.go") || name == "audit.go" {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			for _, pattern := range patterns {
				if pattern.Match(raw) {
					matches = append(matches, filepath.ToSlash(strings.TrimPrefix(path, root+"/"))+": "+pattern.String())
				}
			}
			return nil
		})
	}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == ".layer-os" || name == ".gemini" {
				return fs.SkipDir
			}
			return nil
		}
		rel := filepath.ToSlash(strings.TrimPrefix(path, root+"/"))
		if _, rule, ok := classifyGeminiArtifactPath(rel); ok {
			matches = append(matches, "stray agent artifact: "+rel+" ("+geminiArtifactResidueLabel(rel, rule)+")")
		}
		return nil
	})
	sort.Strings(matches)
	return ResidueAudit{Matches: matches}
}

func AuditGemini(root string) GeminiAudit {
	policyPath := filepath.Join(root, ".gemini", "GEMINI.md")
	policyRel, err := filepath.Rel(root, policyPath)
	if err != nil {
		policyRel = filepath.ToSlash(policyPath)
	} else {
		policyRel = filepath.ToSlash(policyRel)
	}
	audit := GeminiAudit{
		Status:            "ok",
		ProjectPolicyPath: policyRel,
		SuggestedCommands: []string{
			"layer-osctl audit gemini --strict",
			"layer-osctl ingest gemini --cleanup",
			"layer-osctl audit residue --strict",
			"layer-osctl job report --id <job-id> --status failed --notes gemini_drift",
		},
	}

	raw, err := os.ReadFile(policyPath)
	if err != nil {
		audit.PolicyIssues = append(audit.PolicyIssues, "missing project Gemini policy: "+policyRel)
	} else {
		audit.ProjectPolicyPresent = true
		content := string(raw)
		requiredMarkers := []struct {
			needle string
			issue  string
		}{
			{needle: "report-only", issue: "project Gemini policy should keep the default lane report-only"},
			{needle: "파일을 만들지 말고 채팅으로만 답한다", issue: "project Gemini policy should forbid ad-hoc file creation without an explicit target path"},
			{needle: "layer-osctl job report", issue: "project Gemini policy should route agent completion through layer-osctl job report"},
			{needle: "layer-osctl ingest gemini|antigravity|telegram|content", issue: "project Gemini policy should route external evidence through canonical ingest commands"},
			{needle: "layer-osctl audit gemini", issue: "project Gemini policy should require gemini containment audit in the verification loop"},
			{needle: "layer-osctl audit residue", issue: "project Gemini policy should require residue audit in the verification loop"},
			{needle: ".layer-os/*.json", issue: "project Gemini policy should forbid direct runtime json rewrites"},
		}
		for _, marker := range requiredMarkers {
			if !strings.Contains(content, marker.needle) {
				audit.PolicyIssues = append(audit.PolicyIssues, marker.issue)
			}
		}
	}

	scan, err := ScanGeminiArtifacts(root, 0)
	if err != nil {
		audit.PolicyIssues = append(audit.PolicyIssues, "gemini artifact scan failed: "+err.Error())
	} else {
		for _, artifact := range scan.Artifacts {
			audit.ArtifactMatches = append(audit.ArtifactMatches, GeminiArtifactResidueMatch(artifact))
		}
	}
	sort.Strings(audit.PolicyIssues)
	sort.Strings(audit.ArtifactMatches)
	if len(audit.PolicyIssues) > 0 || len(audit.ArtifactMatches) > 0 {
		audit.Status = "degraded"
	}
	return audit
}

func isGeminiArtifactResidueMatch(match string) bool {
	return strings.HasPrefix(strings.TrimSpace(match), "stray agent artifact: ")
}

func AuditCorpus(root string, dataDir string) CorpusAudit {
	audit := CorpusAudit{
		Status:  "ok",
		Root:    strings.TrimSpace(root),
		DataDir: strings.TrimSpace(dataDir),
		SuggestedCommands: []string{
			"layer-osctl audit corpus --strict",
			"layer-osctl ingest corpus --cleanup",
		},
		ArtifactMatches: []string{},
	}
	scan, err := ScanCorpusMarkdownArtifacts(root, dataDir, 0)
	if err != nil {
		audit.Status = "degraded"
		audit.ArtifactMatches = append(audit.ArtifactMatches, "scan failed: "+err.Error())
		return audit
	}
	for _, artifact := range scan.Artifacts {
		audit.ArtifactMatches = append(audit.ArtifactMatches, artifact.RelativePath)
	}
	sort.Strings(audit.ArtifactMatches)
	if len(audit.ArtifactMatches) > 0 {
		audit.Status = "degraded"
	}
	return audit
}

func AuditAuthority(root string) AuthorityAudit {
	audit := AuthorityAudit{
		Status:   "ok",
		Boundary: CanonicalAuthorityBoundary(),
		SuggestedCommands: []string{
			"layer-osctl audit authority --strict",
			"layer-osctl session bootstrap --full",
			"layer-osctl audit residue --strict",
		},
	}
	requiredFiles := []struct {
		path    string
		markers []struct {
			needle string
			issue  string
		}
	}{
		{
			path: filepath.Join(root, "AGENTS.md"),
			markers: []struct {
				needle string
				issue  string
			}{
				{needle: "Legacy material under `/Users/97layer/97layerOS` is reference-only", issue: "AGENTS.md should mark 97layerOS as reference-only"},
				{needle: "authority must come from this workspace's `AGENTS.md`, `constitution/`, `docs/`, and `contracts/`", issue: "AGENTS.md should point authority to current workspace roots"},
				{needle: "report back via `/api/layer-os/jobs/report`", issue: "AGENTS.md should keep external agent completion on /api/layer-os/jobs/report"},
				{needle: "Do not rewrite `.layer-os/review_room.json` directly", issue: "AGENTS.md should forbid direct review_room rewrites"},
			},
		},
		{
			path: filepath.Join(root, ".gemini", "GEMINI.md"),
			markers: []struct {
				needle string
				issue  string
			}{
				{needle: "레거시 `97layerOS`는 이 워크스페이스에서 참고 전용", issue: ".gemini/GEMINI.md should mark legacy 97layerOS as reference-only"},
				{needle: "`review_room.json`은 런타임 상태다. 직접 수정하지 않는다.", issue: ".gemini/GEMINI.md should forbid direct review_room edits"},
				{needle: "`layer-osctl job report ...` 또는 `/api/layer-os/jobs/report`", issue: ".gemini/GEMINI.md should keep agent completion on canonical job report surface"},
				{needle: "layer-osctl audit authority --strict", issue: ".gemini/GEMINI.md should include authority audit in the verification loop"},
			},
		},
	}
	for _, file := range requiredFiles {
		raw, err := os.ReadFile(file.path)
		if err != nil {
			audit.PolicyIssues = append(audit.PolicyIssues, "missing authority policy file: "+filepath.ToSlash(strings.TrimPrefix(file.path, root+"/")))
			continue
		}
		content := string(raw)
		for _, marker := range file.markers {
			if !strings.Contains(content, marker.needle) {
				audit.PolicyIssues = append(audit.PolicyIssues, marker.issue)
			}
		}
	}
	legacyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`/Users/97layer/97layerOS`),
		regexp.MustCompile(`97layerOS/constitution`),
	}
	legacyScopes := []string{
		filepath.Join(root, "cmd"),
		filepath.Join(root, "contracts"),
		filepath.Join(root, "internal"),
	}
	for _, scope := range legacyScopes {
		_ = filepath.WalkDir(scope, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			rel := filepath.ToSlash(strings.TrimPrefix(path, root+"/"))
			if strings.HasSuffix(d.Name(), "_test.go") || rel == "internal/runtime/audit.go" || rel == "internal/runtime/authority.go" {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			for _, pattern := range legacyPatterns {
				if pattern.Match(raw) {
					audit.LegacyCodeMatches = append(audit.LegacyCodeMatches, filepath.ToSlash(strings.TrimPrefix(path, root+"/"))+": "+pattern.String())
				}
			}
			return nil
		})
	}
	audit.PolicyIssues = append(audit.PolicyIssues, authorityDirectRuntimeJSONScriptIssues(root)...)
	sort.Strings(audit.PolicyIssues)
	sort.Strings(audit.LegacyCodeMatches)
	if len(audit.PolicyIssues) > 0 || len(audit.LegacyCodeMatches) > 0 {
		audit.Status = "degraded"
	}
	return audit
}

func authorityDirectRuntimeJSONScriptIssues(root string) []string {
	scriptsDir := filepath.Join(root, "scripts")
	info, err := os.Stat(scriptsDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	issues := map[string]bool{}
	_ = filepath.WalkDir(scriptsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil || d.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		for idx, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if !strings.Contains(trimmed, ".layer-os/") || !strings.Contains(trimmed, ".json") {
				continue
			}
			issues[fmt.Sprintf("%s:%d should use layer-osctl/api surfaces instead of direct runtime json access", rel, idx+1)] = true
		}
		return nil
	})
	return sortedKeys(issues)
}

func AuditContracts(root string) ContractAudit {
	expected := CanonicalContractFilenames()

	contractsDir := filepath.Join(root, "contracts")
	entries, _ := filepath.Glob(filepath.Join(contractsDir, "*.json"))
	actual := []string{}
	actualSet := map[string]bool{}
	for _, entry := range entries {
		name := filepath.Base(entry)
		actual = append(actual, name)
		actualSet[name] = true
	}
	sort.Strings(actual)

	expectedSet := map[string]bool{}
	for _, name := range expected {
		expectedSet[name] = true
	}

	missing := []string{}
	for _, name := range expected {
		if !actualSet[name] {
			missing = append(missing, name)
		}
	}

	unexpected := []string{}
	for _, name := range actual {
		if !expectedSet[name] {
			unexpected = append(unexpected, name)
		}
	}

	issues := []string{}
	for _, name := range missing {
		issues = append(issues, "missing contract: "+name)
	}
	for _, name := range unexpected {
		issues = append(issues, "unexpected contract: "+name)
	}
	titleIssues := contractSchemaTitleIssues(root)
	architectureIssues := contractArchitectureIssues(root)
	issues = append(issues, titleIssues...)
	issues = append(issues, architectureIssues...)
	sort.Strings(titleIssues)
	sort.Strings(architectureIssues)
	sort.Strings(issues)

	return ContractAudit{
		Expected:           expected,
		Missing:            missing,
		Unexpected:         unexpected,
		TitleIssues:        titleIssues,
		ArchitectureIssues: architectureIssues,
		Issues:             issues,
	}
}

func contractSchemaTitleIssues(root string) []string {
	issues := []string{}
	contractsDir := filepath.Join(root, "contracts")
	for _, spec := range CanonicalContractCatalog() {
		title, err := readContractSchemaTitle(filepath.Join(contractsDir, spec.Schema))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			issues = append(issues, fmt.Sprintf("contract title check failed: %s: %v", spec.Schema, err))
			continue
		}
		if title != spec.Title {
			issues = append(issues, fmt.Sprintf("contract title drift: %s expected %q got %q", spec.Schema, spec.Title, title))
		}
	}
	return sortedUniqueStrings(issues)
}

func readContractSchemaTitle(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var schema struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return "", err
	}
	return schema.Title, nil
}

func contractArchitectureIssues(root string) []string {
	path := filepath.Join(root, "docs", "architecture.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"contract architecture source missing: docs/architecture.md"}
		}
		return []string{fmt.Sprintf("contract architecture read failed: %v", err)}
	}

	text := string(raw)
	issues := []string{}
	spineTitles, spineFound := markdownContractListUnderHeading(text, "## System Spine", contractNumberedItem)
	if !spineFound {
		issues = append(issues, "contract architecture missing section: ## System Spine")
	} else {
		issues = append(issues, compareContractSection("spine", spineTitles, ContractTitlesByTier(ContractTierSpine))...)
	}

	harnessTitles, harnessFound := markdownContractListUnderHeading(text, "## Contract Tiers", contractBulletItem)
	if !harnessFound {
		issues = append(issues, "contract architecture missing section: ## Contract Tiers")
	} else {
		issues = append(issues, compareContractSection("harness", harnessTitles, ContractTitlesByTier(ContractTierHarness))...)
	}

	if spineFound && harnessFound {
		harnessSet := map[string]bool{}
		for _, title := range harnessTitles {
			harnessSet[title] = true
		}
		for _, title := range spineTitles {
			if harnessSet[title] {
				issues = append(issues, "contract architecture duplicate tier listing: "+title)
			}
		}
	}
	return sortedUniqueStrings(issues)
}

func markdownContractListUnderHeading(text string, heading string, itemRE *regexp.Regexp) ([]string, bool) {
	lines := strings.Split(text, "\n")
	inSection := false
	items := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == heading {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			break
		}
		if trimmed == "" {
			continue
		}
		match := itemRE.FindStringSubmatch(trimmed)
		if len(match) == 2 {
			items = append(items, match[1])
		}
	}
	return items, inSection
}

func compareContractSection(section string, actual []string, expected []string) []string {
	issues := []string{}
	for _, title := range setDifference(expected, actual) {
		issues = append(issues, fmt.Sprintf("contract architecture %s missing: %s", section, title))
	}
	for _, title := range setDifference(actual, expected) {
		issues = append(issues, fmt.Sprintf("contract architecture %s unexpected: %s", section, title))
	}
	if len(actual) == len(expected) {
		same := true
		for i := range actual {
			if actual[i] != expected[i] {
				same = false
				break
			}
		}
		if !same {
			issues = append(issues, fmt.Sprintf("contract architecture %s order drift", section))
		}
	}
	return issues
}

func AuditSurface(root string) SurfaceAudit {
	spec := CanonicalSurfaceSpec()
	usageRoots := surfaceCLIUsageRoots(spec.CLIUsage)

	expectedCLIDispatch := sortedUniqueStrings(spec.CLIDispatch)
	actualCLIDispatch, dispatchErr := extractActualCLIDispatch(filepath.Join(root, "cmd", "layer-osctl", "root_dispatch.go"))
	missingCLIDispatch := append([]string{}, expectedCLIDispatch...)
	unexpectedCLIDispatch := []string{}
	missingCLIUsage := []string{}
	unexpectedCLIUsage := []string{}
	if dispatchErr == nil {
		missingCLIDispatch = setDifference(expectedCLIDispatch, actualCLIDispatch)
		unexpectedCLIDispatch = setDifference(actualCLIDispatch, expectedCLIDispatch)
		actualCLICommands := dispatchMarkersToCommands(actualCLIDispatch)
		missingCLIUsage = setDifference(actualCLICommands, usageRoots)
		unexpectedCLIUsage = setDifference(usageRoots, actualCLICommands)
	}

	expectedAPIRoutes := sortedUniqueStrings(spec.APIRoutes)
	actualAPIRoutes, apiErr := extractActualAPIRoutes([]string{
		filepath.Join(root, "internal", "api", "router.go"),
		filepath.Join(root, "internal", "api", "router_audit.go"),
	})
	missingAPIRoutes := append([]string{}, expectedAPIRoutes...)
	unexpectedAPIRoutes := []string{}
	if apiErr == nil {
		missingAPIRoutes = setDifference(expectedAPIRoutes, actualAPIRoutes)
		unexpectedAPIRoutes = setDifference(actualAPIRoutes, expectedAPIRoutes)
	}

	issues := []string{}
	if dispatchErr != nil {
		issues = append(issues, "surface source missing: cmd/layer-osctl/root_dispatch.go")
	}
	if apiErr != nil {
		issues = append(issues, "surface source missing: internal/api/router.go or internal/api/router_audit.go")
	}
	for _, command := range missingCLIUsage {
		issues = append(issues, "missing cli usage surface: "+command)
	}
	for _, command := range unexpectedCLIUsage {
		issues = append(issues, "unexpected cli usage surface: "+command)
	}
	for _, snippet := range missingCLIDispatch {
		issues = append(issues, "missing cli dispatch surface: "+snippet)
	}
	for _, snippet := range unexpectedCLIDispatch {
		issues = append(issues, "unexpected cli dispatch surface: "+snippet)
	}
	for _, snippet := range missingAPIRoutes {
		issues = append(issues, "missing api route surface: "+snippet)
	}
	for _, snippet := range unexpectedAPIRoutes {
		issues = append(issues, "unexpected api route surface: "+snippet)
	}
	sort.Strings(issues)

	return SurfaceAudit{
		ExpectedCLIUsage:      spec.CLIUsage,
		MissingCLIUsage:       missingCLIUsage,
		UnexpectedCLIUsage:    unexpectedCLIUsage,
		ExpectedCLIDispatch:   expectedCLIDispatch,
		MissingCLIDispatch:    missingCLIDispatch,
		UnexpectedCLIDispatch: unexpectedCLIDispatch,
		ExpectedAPIRoutes:     expectedAPIRoutes,
		MissingAPIRoutes:      missingAPIRoutes,
		UnexpectedAPIRoutes:   unexpectedAPIRoutes,
		Issues:                issues,
	}
}

func extractActualCLIDispatch(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	matches := rootDispatchCommandRE.FindAllStringSubmatch(string(raw), -1)
	items := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		items = append(items, `"`+strings.TrimSpace(match[1])+`":`)
	}
	return sortedUniqueStrings(items), nil
}

func extractActualAPIRoutes(paths []string) ([]string, error) {
	items := []string{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		content := string(raw)
		for _, match := range apiHandleFuncRouteRE.FindAllStringSubmatch(content, -1) {
			if len(match) < 2 {
				continue
			}
			route := strings.TrimSpace(match[1])
			if strings.HasPrefix(route, "/api/layer-os/") {
				items = append(items, route)
			}
		}
		for _, match := range auditRoutePathRE.FindAllStringSubmatch(content, -1) {
			if len(match) < 2 {
				continue
			}
			route := strings.TrimSpace(match[1])
			if strings.HasPrefix(route, "/api/layer-os/") {
				items = append(items, route)
			}
		}
	}
	return sortedUniqueStrings(items), nil
}

func surfaceCLIUsageRoots(items []string) []string {
	roots := []string{}
	for _, item := range items {
		line := strings.TrimSpace(item)
		if !strings.HasPrefix(line, "layer-osctl ") {
			continue
		}
		command := strings.TrimSpace(strings.TrimPrefix(line, "layer-osctl "))
		if command == "" {
			continue
		}
		root := strings.Fields(command)[0]
		roots = append(roots, root)
	}
	return sortedUniqueStrings(roots)
}

func dispatchMarkersToCommands(items []string) []string {
	commands := []string{}
	for _, item := range items {
		command := strings.TrimSpace(item)
		command = strings.TrimPrefix(command, `"`)
		command = strings.TrimSuffix(command, `":`)
		if command == "" {
			continue
		}
		commands = append(commands, command)
	}
	return sortedUniqueStrings(commands)
}

func setDifference(left []string, right []string) []string {
	seen := map[string]bool{}
	for _, item := range right {
		seen[item] = true
	}
	diff := []string{}
	for _, item := range sortedUniqueStrings(left) {
		if !seen[item] {
			diff = append(diff, item)
		}
	}
	return diff
}

func sortedUniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

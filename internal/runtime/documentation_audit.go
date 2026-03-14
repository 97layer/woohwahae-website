package runtime

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type DocumentationAudit struct {
	Status               string                 `json:"status"`
	Inventory            DocumentationInventory `json:"inventory"`
	CheckedFiles         []string               `json:"checked_files"`
	CheckedLocalRefCount int                    `json:"checked_local_ref_count"`
	MissingRefs          []string               `json:"missing_refs"`
	RelatedDocsIssues    []string               `json:"related_docs_issues"`
	RequiredRuleIssues   []string               `json:"required_rule_issues"`
	LegacyLabelMatches   []string               `json:"legacy_label_matches"`
	ExternalRefsSkipped  []string               `json:"external_refs_skipped"`
	Issues               []string               `json:"issues"`
	SuggestedCommands    []string               `json:"suggested_commands"`
}

type DocumentationInventory struct {
	TotalMarkdownPaths   int `json:"total_markdown_paths"`
	ActiveCorePaths      int `json:"active_core_paths"`
	AuthorityPaths       int `json:"authority_paths"`
	CompatibilityPaths   int `json:"compatibility_paths"`
	ArchivedPaths        int `json:"archived_paths"`
	SkillManualPaths     int `json:"skill_manual_paths"`
	VendoredPaths        int `json:"vendored_paths"`
	RuntimeArtifactPaths int `json:"runtime_artifact_paths"`
	AgentLocalPaths      int `json:"agent_local_paths"`
}

var (
	relatedDocsHeaderRE = regexp.MustCompile(`(?m)^## Related Docs$`)
	relatedDocsLegacyRE = regexp.MustCompile(`(?m)^## Related docs$`)
	anyHeadingRE        = regexp.MustCompile(`(?m)^#{1,6} `)
	markdownLinkRE      = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	inlineCodeRE        = regexp.MustCompile("`([^`\\n]+)`")
	rawURLRE            = regexp.MustCompile(`https?://[^\s)>\"]+`)
	legacyLabelRE       = regexp.MustCompile(`\b(SA|CE|AD|RALPH|CD)\b`)
)

func AuditDocumentation(root string) DocumentationAudit {
	audit := DocumentationAudit{
		Status:              "ok",
		Inventory:           documentationInventory(root),
		CheckedFiles:        documentationScopeFiles(root),
		MissingRefs:         []string{},
		RelatedDocsIssues:   []string{},
		RequiredRuleIssues:  []string{},
		LegacyLabelMatches:  []string{},
		ExternalRefsSkipped: []string{},
		Issues:              []string{},
		SuggestedCommands: []string{
			"layer-osctl audit docs --strict",
			"go test ./...",
			"markdown-link-check docs/**/*.md constitution/**/*.md # run in a network-enabled environment for external URLs",
		},
	}
	if len(audit.CheckedFiles) == 0 {
		return audit
	}

	missingRefs := map[string]bool{}
	relatedIssues := map[string]bool{}
	requiredRuleIssues := map[string]bool{}
	legacyMatches := map[string]bool{}
	externalRefs := map[string]bool{}
	checkedRefs := map[string]bool{}

	for _, rel := range audit.CheckedFiles {
		path := filepath.Join(root, filepath.FromSlash(rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			issue := fmt.Sprintf("%s: read failed: %v", rel, err)
			audit.Issues = append(audit.Issues, issue)
			continue
		}
		text := string(raw)

		for _, issue := range documentationRelatedDocsIssues(rel, text) {
			relatedIssues[issue] = true
		}
		for _, issue := range documentationRequiredRuleIssues(rel, text) {
			requiredRuleIssues[issue] = true
		}
		for _, match := range documentationLegacyLabelMatches(rel, text) {
			legacyMatches[match] = true
		}
		for _, target := range documentationExternalRefs(text) {
			externalRefs[target] = true
		}
		for _, ref := range documentationLocalRefs(rel, text) {
			checkedRefs[ref] = true
			full := filepath.Join(root, filepath.FromSlash(ref))
			if _, err := os.Stat(full); err != nil {
				missingRefs[fmt.Sprintf("%s -> %s", rel, ref)] = true
			}
		}
	}

	audit.CheckedLocalRefCount = len(checkedRefs)
	audit.MissingRefs = sortedKeys(missingRefs)
	audit.RelatedDocsIssues = sortedKeys(relatedIssues)
	audit.RequiredRuleIssues = sortedKeys(requiredRuleIssues)
	audit.LegacyLabelMatches = sortedKeys(legacyMatches)
	audit.ExternalRefsSkipped = sortedKeys(externalRefs)
	audit.Issues = append(audit.Issues, audit.MissingRefs...)
	audit.Issues = append(audit.Issues, audit.RelatedDocsIssues...)
	audit.Issues = append(audit.Issues, audit.RequiredRuleIssues...)
	audit.Issues = append(audit.Issues, audit.LegacyLabelMatches...)
	sort.Strings(audit.Issues)
	if len(audit.Issues) > 0 {
		audit.Status = "degraded"
	}
	return audit
}

func documentationInventory(root string) DocumentationInventory {
	inventory := DocumentationInventory{}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == ".next" {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		inventory.TotalMarkdownPaths++
		switch documentationInventoryClass(rel) {
		case "vendored":
			inventory.VendoredPaths++
		case "archive":
			inventory.ArchivedPaths++
		case "skill":
			inventory.SkillManualPaths++
		case "runtime_artifact":
			inventory.RuntimeArtifactPaths++
		case "agent_local":
			inventory.AgentLocalPaths++
		case "compatibility":
			inventory.CompatibilityPaths++
			inventory.ActiveCorePaths++
		case "authority":
			inventory.AuthorityPaths++
			inventory.ActiveCorePaths++
		default:
			inventory.ActiveCorePaths++
		}
		return nil
	})
	return inventory
}

func documentationInventoryClass(rel string) string {
	switch {
	case strings.HasPrefix(rel, "docs/brand-home/node_modules/"):
		return "vendored"
	case strings.HasPrefix(rel, "docs/archive/"):
		return "archive"
	case strings.HasPrefix(rel, "skills/"):
		return "skill"
	case strings.HasPrefix(rel, ".omx/") || strings.HasPrefix(rel, ".layer-os/"):
		return "runtime_artifact"
	case strings.HasPrefix(rel, ".gemini/"):
		return "agent_local"
	case rel == "docs/agent_integration.md":
		return "compatibility"
	case rel == "AGENTS.md" || rel == "README.md" || strings.HasPrefix(rel, "constitution/") || strings.HasPrefix(rel, "docs/"):
		return "authority"
	default:
		return "active"
	}
}

func documentationScopeFiles(root string) []string {
	files := []string{}
	for _, name := range []string{"AGENTS.md", "README.md"} {
		path := filepath.Join(root, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files = append(files, filepath.ToSlash(name))
		}
	}
	for _, dir := range []string{"constitution", "docs", "skills"} {
		base := filepath.Join(root, dir)
		if info, err := os.Stat(base); err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d == nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".next" || name == "node_modules" {
					return fs.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			files = append(files, filepath.ToSlash(rel))
			return nil
		})
	}
	sort.Strings(files)
	return files
}

func documentationRelatedDocsIssues(rel string, text string) []string {
	issues := []string{}
	if relatedDocsLegacyRE.MatchString(text) {
		issues = append(issues, fmt.Sprintf("%s: use `## Related Docs` heading casing", rel))
	}
	loc := relatedDocsHeaderRE.FindStringIndex(text)
	if loc == nil {
		return issues
	}
	block := text[loc[1]:]
	end := len(block)
	if next := anyHeadingRE.FindStringIndex(block); next != nil {
		end = next[0]
	}
	block = strings.TrimSpace(block[:end])
	if block == "" {
		issues = append(issues, fmt.Sprintf("%s: `Related Docs` block is empty", rel))
		return issues
	}
	bullets := []string{}
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			bullets = append(bullets, line)
		}
	}
	if len(bullets) == 0 {
		issues = append(issues, fmt.Sprintf("%s: `Related Docs` block must use bullet links", rel))
		return issues
	}
	for _, bullet := range bullets {
		if len(documentationDocRefsFromLine(rel, bullet)) == 0 {
			issues = append(issues, fmt.Sprintf("%s: `Related Docs` bullet missing local doc ref: %s", rel, bullet))
		}
	}
	return issues
}

func documentationRequiredRuleIssues(rel string, text string) []string {
	if rel != "docs/prompting.md" {
		return nil
	}
	requiredSnippets := []string{
		"lead with current system facts, changed files, runtime state, risks, and next watchpoints",
		"do not answer a status/tracking request with an implementation plan or redesign pitch",
		"avoid opening with phrases like `intent layer`, `router`, `canonical action`, or similar implementation shorthand",
	}
	issues := []string{}
	for _, snippet := range requiredSnippets {
		if strings.Contains(text, snippet) {
			continue
		}
		issues = append(issues, fmt.Sprintf("%s: required operating rule missing: %s", rel, snippet))
	}
	return issues
}

func documentationLegacyLabelMatches(rel string, text string) []string {
	if !documentationScansLegacyLabels(rel) {
		return nil
	}
	matches := []string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		found := legacyLabelRE.FindAllString(line, -1)
		if len(found) == 0 {
			continue
		}
		matches = append(matches, fmt.Sprintf("%s: legacy label reference found: %s", rel, line))
		break
	}
	return matches
}

func documentationScansLegacyLabels(rel string) bool {
	base := filepath.Base(rel)
	if strings.HasPrefix(base, "legacy_") {
		return false
	}
	if rel == "docs/legacy_inventory.md" || rel == "constitution/lineage.md" {
		return false
	}
	return strings.HasPrefix(rel, "docs/") || strings.HasPrefix(rel, "constitution/")
}

func documentationExternalRefs(text string) []string {
	matches := rawURLRE.FindAllString(text, -1)
	items := map[string]bool{}
	for _, match := range matches {
		normalized := strings.TrimRight(match, "`,.;:)}]")
		normalized = strings.ReplaceAll(normalized, "}/", "/")
		items[normalized] = true
	}
	return sortedKeys(items)
}

func documentationLocalRefs(rel string, text string) []string {
	refs := map[string]bool{}
	for _, match := range markdownLinkRE.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		if resolved, ok := resolveDocumentationRef(rel, match[1]); ok {
			refs[resolved] = true
		}
	}
	for _, token := range inlineCodeRE.FindAllStringSubmatch(text, -1) {
		if len(token) < 2 {
			continue
		}
		if resolved, ok := resolveDocumentationRef(rel, token[1]); ok {
			refs[resolved] = true
		}
	}
	return sortedKeys(refs)
}

func documentationDocRefsFromLine(rel string, line string) []string {
	refs := map[string]bool{}
	for _, token := range inlineCodeRE.FindAllStringSubmatch(line, -1) {
		if len(token) < 2 {
			continue
		}
		if resolved, ok := resolveDocumentationRef(rel, token[1]); ok {
			refs[resolved] = true
		}
	}
	return sortedKeys(refs)
}

func resolveDocumentationRef(rel string, raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	raw = strings.Trim(raw, "<>")
	raw = strings.TrimSpace(strings.Fields(raw)[0])
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}
	if idx := strings.Index(raw, "#"); idx >= 0 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, "?"); idx >= 0 {
		raw = raw[:idx]
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "mailto:") {
		return "", false
	}
	if strings.HasPrefix(raw, "/api/") {
		return "", false
	}
	if strings.HasPrefix(raw, "/Users/97layer/layer OS/") {
		trimmed := strings.TrimPrefix(raw, "/Users/97layer/layer OS/")
		return filepath.ToSlash(filepath.Clean(trimmed)), documentationRefLooksLocal(trimmed)
	}
	if strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../") {
		joined := filepath.Clean(filepath.Join(filepath.Dir(rel), raw))
		joined = filepath.ToSlash(joined)
		return joined, documentationRefLooksLocal(joined)
	}
	cleaned := filepath.ToSlash(filepath.Clean(raw))
	return cleaned, documentationRefLooksLocal(cleaned)
}

func documentationRefLooksLocal(value string) bool {
	if value == "AGENTS.md" || value == "README.md" {
		return true
	}
	if !strings.HasSuffix(value, ".md") {
		return false
	}
	for _, prefix := range []string{"docs/", "constitution/", "contracts/", "skills/", "cmd/", "internal/", "scripts/"} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func sortedKeys(items map[string]bool) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

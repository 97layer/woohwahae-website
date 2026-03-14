package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type GeminiScan struct {
	Root      string           `json:"root"`
	Artifacts []GeminiArtifact `json:"artifacts"`
}

type GeminiArtifact struct {
	Stem                 string    `json:"stem"`
	Label                string    `json:"label"`
	Topic                string    `json:"topic"`
	ArtifactType         string    `json:"artifact_type"`
	Summary              string    `json:"summary"`
	RawExcerpt           string    `json:"raw_excerpt"`
	Confidence           string    `json:"confidence"`
	CapturedAt           time.Time `json:"captured_at"`
	SyncKey              string    `json:"sync_key"`
	PrimaryPath          string    `json:"primary_path"`
	RelativePath         string    `json:"relative_path"`
	RelatedPaths         []string  `json:"related_paths"`
	RelatedRelativePaths []string  `json:"related_relative_paths"`
	MetadataPath         *string   `json:"metadata_path,omitempty"`
	Version              *string   `json:"version,omitempty"`
	Hash                 string    `json:"hash"`
	Refs                 []string  `json:"refs"`
	Severity             string    `json:"severity"`
}

type geminiArtifactGroup struct {
	stem        string
	metadata    string
	candidates  []string
	related     []string
	metadataRaw antigravityMetadata
	rule        geminiArtifactRule
}

type geminiArtifactRule struct {
	pattern      *regexp.Regexp
	label        string
	topic        string
	artifactType string
	severity     string
}

var geminiArtifactRules = []geminiArtifactRule{
	{pattern: regexp.MustCompile(`^SESSION_SUMMARY_.*\.md$`), label: "session-summary scratchpad", topic: "session_summary", artifactType: "session_summary", severity: "medium"},
	{pattern: regexp.MustCompile(`^WAKEUP_REPORT\.md$`), label: "wake-up report scratchpad", topic: "wakeup_report", artifactType: "wakeup_report", severity: "medium"},
	{pattern: regexp.MustCompile(`^DEEP_WORK_PROGRESS\.md$`), label: "deep-work progress scratchpad", topic: "deep_work_progress", artifactType: "deep_work_progress", severity: "medium"},
	{pattern: regexp.MustCompile(`^DEPLOY_.*\.md$`), label: "deploy scratchpad", topic: "deploy_scratchpad", artifactType: "deploy_scratchpad", severity: "medium"},
	{pattern: regexp.MustCompile(`^DEPLOYMENT_CHECKLIST\.md$`), label: "deploy checklist scratchpad", topic: "deployment_checklist", artifactType: "deployment_checklist", severity: "medium"},
	{pattern: regexp.MustCompile(`^NEXT_STEPS\.md$`), label: "next-steps scratchpad", topic: "next_steps", artifactType: "next_steps", severity: "medium"},
	{pattern: regexp.MustCompile(`^council_room\.md$`), label: "legacy council scratchpad", topic: "legacy_council", artifactType: "legacy_council", severity: "high"},
	{pattern: regexp.MustCompile(`^state\.md$`), label: "legacy state scratchpad", topic: "legacy_state", artifactType: "legacy_state", severity: "high"},
	{pattern: regexp.MustCompile(`^task\.md$`), label: "agent artifact markdown", topic: "task", artifactType: "task", severity: "medium"},
	{pattern: regexp.MustCompile(`^walkthrough\.md$`), label: "agent artifact markdown", topic: "walkthrough", artifactType: "walkthrough", severity: "medium"},
	{pattern: regexp.MustCompile(`^implementation_plan\.md$`), label: "agent artifact markdown", topic: "implementation_plan", artifactType: "implementation_plan", severity: "medium"},
	{pattern: regexp.MustCompile(`^integrity_diagnosis\.md$`), label: "agent artifact markdown", topic: "integrity_diagnosis", artifactType: "integrity_diagnosis", severity: "high"},
}

func DefaultGeminiWorkspaceRoot() string {
	if value := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT")); value != "" {
		return value
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func ScanGeminiArtifacts(root string, limit int) (GeminiScan, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = DefaultGeminiWorkspaceRoot()
	}
	if root == "" {
		return GeminiScan{}, errors.New("gemini workspace root is required")
	}
	info, err := os.Stat(root)
	if err != nil {
		return GeminiScan{}, err
	}
	if !info.IsDir() {
		return GeminiScan{}, errors.New("gemini workspace root must be a directory")
	}

	groups := map[string]*geminiArtifactGroup{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			switch name {
			case ".git", ".layer-os", ".gemini":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		stem, rule, ok := classifyGeminiArtifactPath(rel)
		if !ok {
			return nil
		}
		group := groups[stem]
		if group == nil {
			group = &geminiArtifactGroup{stem: stem, rule: rule}
			groups[stem] = group
		}
		group.related = append(group.related, path)
		group.candidates = append(group.candidates, path)
		if strings.HasSuffix(rel, ".metadata.json") {
			group.metadata = path
			metadata, err := loadAntigravityMetadata(path)
			if err == nil {
				group.metadataRaw = metadata
			}
		}
		return nil
	})
	if err != nil {
		return GeminiScan{}, err
	}

	artifacts := make([]GeminiArtifact, 0, len(groups))
	for _, group := range groups {
		artifact, ok := buildGeminiArtifact(root, group)
		if !ok {
			continue
		}
		artifacts = append(artifacts, artifact)
	}
	sort.SliceStable(artifacts, func(i, j int) bool {
		ri := geminiArtifactPriority(artifacts[i])
		rj := geminiArtifactPriority(artifacts[j])
		if ri != rj {
			return ri > rj
		}
		if artifacts[i].CapturedAt.Equal(artifacts[j].CapturedAt) {
			return artifacts[i].RelativePath < artifacts[j].RelativePath
		}
		return artifacts[i].CapturedAt.After(artifacts[j].CapturedAt)
	})
	if limit > 0 && len(artifacts) > limit {
		artifacts = artifacts[:limit]
	}
	return GeminiScan{Root: root, Artifacts: artifacts}, nil
}

func GeminiObservation(item GeminiArtifact) ObservationRecord {
	refs := append([]string{}, item.Refs...)
	for _, ref := range ObservationMetadataRefs("gemini", item.Topic, map[string]string{
		"artifact_type":  item.ArtifactType,
		"artifact_stem":  item.Stem,
		"ingest_adapter": "gemini",
		"severity":       item.Severity,
	}) {
		refs = appendUniqueString(refs, ref)
	}
	refs = appendUniqueString(refs, "gemini_sync:"+strings.TrimSpace(item.SyncKey))
	return ObservationRecord{
		ObservationID:     geminiObservationID(item.SyncKey),
		SourceChannel:     "gemini",
		CapturedAt:        item.CapturedAt,
		Actor:             "gemini",
		Topic:             item.Topic,
		Refs:              refs,
		Confidence:        item.Confidence,
		RawExcerpt:        item.RawExcerpt,
		NormalizedSummary: item.Summary,
	}
}

func GeminiArtifactResidueMatch(item GeminiArtifact) string {
	return "stray agent artifact: " + strings.TrimSpace(item.RelativePath) + " (" + strings.TrimSpace(item.Label) + ")"
}

func classifyGeminiArtifactPath(rel string) (string, geminiArtifactRule, bool) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return "", geminiArtifactRule{}, false
	}
	stem := normalizeAntigravityStem(rel)
	if stem == "" {
		return "", geminiArtifactRule{}, false
	}
	rule, ok := classifyGeminiArtifactStem(stem)
	if !ok {
		return "", geminiArtifactRule{}, false
	}
	return stem, rule, true
}

func classifyGeminiArtifactStem(stem string) (geminiArtifactRule, bool) {
	base := filepath.Base(filepath.ToSlash(strings.TrimSpace(stem)))
	for _, rule := range geminiArtifactRules {
		if rule.pattern.MatchString(base) {
			return rule, true
		}
	}
	return geminiArtifactRule{}, false
}

func geminiArtifactResidueLabel(rel string, rule geminiArtifactRule) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	switch {
	case strings.HasSuffix(rel, ".metadata.json"):
		return "artifact metadata"
	case strings.Contains(filepath.Base(rel), ".resolved"):
		return "resolved artifact"
	default:
		return rule.label
	}
}

func buildGeminiArtifact(root string, group *geminiArtifactGroup) (GeminiArtifact, bool) {
	primary := selectAntigravityPrimary(group.stem, group.candidates)
	if primary == "" {
		return GeminiArtifact{}, false
	}
	raw, err := os.ReadFile(primary)
	if err != nil {
		return GeminiArtifact{}, false
	}
	info, err := os.Stat(primary)
	if err != nil {
		return GeminiArtifact{}, false
	}
	capturedAt := info.ModTime().UTC()
	if parsed, ok := parseAntigravityTime(group.metadataRaw.UpdatedAt); ok {
		capturedAt = parsed
	}
	content := string(raw)
	summary := strings.TrimSpace(group.metadataRaw.Summary)
	if summary == "" {
		summary = antigravitySummaryFromContent(content, filepath.Base(group.stem))
	}
	primaryRel, err := filepath.Rel(root, primary)
	if err != nil {
		primaryRel = primary
	}
	primaryRel = filepath.ToSlash(primaryRel)
	relatedPaths := append([]string{}, group.related...)
	sort.Strings(relatedPaths)
	relatedRel := make([]string, 0, len(relatedPaths))
	for _, path := range relatedPaths {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		relatedRel = append(relatedRel, filepath.ToSlash(rel))
	}
	sort.Strings(relatedRel)
	hash := antigravityHash(group.stem, string(raw))
	version := optionalAntigravityString(strings.TrimSpace(group.metadataRaw.Version))
	syncParts := []string{group.stem, hash}
	if version != nil {
		syncParts = append(syncParts, *version)
	}
	if !capturedAt.IsZero() {
		syncParts = append(syncParts, capturedAt.Format(time.RFC3339Nano))
	}
	severity := geminiArtifactSeverity(group.rule.severity, summary+"\n"+content)
	refs := []string{
		"gemini_artifact:" + group.stem,
		"gemini_path:" + primaryRel,
	}
	for _, ref := range ObservationMetadataRefs("gemini", group.rule.topic, map[string]string{
		"artifact_type":  group.rule.artifactType,
		"artifact_stem":  group.stem,
		"ingest_adapter": "gemini",
		"severity":       severity,
	}) {
		refs = appendUniqueString(refs, ref)
	}
	for _, match := range antigravityRefPattern.FindAllString(content+"\n"+summary, -1) {
		refs = appendUniqueString(refs, strings.TrimSpace(match))
	}
	return GeminiArtifact{
		Stem:                 group.stem,
		Label:                group.rule.label,
		Topic:                group.rule.topic,
		ArtifactType:         group.rule.artifactType,
		Summary:              limitText(summary, 180),
		RawExcerpt:           antigravityExcerpt(content),
		Confidence:           antigravityConfidence(group.metadata != "", strings.Contains(filepath.Base(primary), ".resolved")),
		CapturedAt:           capturedAt,
		SyncKey:              antigravityHash(syncParts...),
		PrimaryPath:          primary,
		RelativePath:         primaryRel,
		RelatedPaths:         relatedPaths,
		RelatedRelativePaths: relatedRel,
		MetadataPath:         optionalAntigravityString(group.metadata),
		Version:              version,
		Hash:                 hash,
		Refs:                 refs,
		Severity:             severity,
	}, true
}

func geminiArtifactPriority(item GeminiArtifact) int {
	switch strings.TrimSpace(item.Topic) {
	case "integrity_diagnosis":
		return 6
	case "task":
		return 5
	case "implementation_plan":
		return 4
	case "walkthrough":
		return 3
	case "legacy_council", "legacy_state":
		return 2
	default:
		return 1
	}
}

func geminiArtifactSeverity(base string, text string) string {
	current := normalizeObservationMetaToken(base)
	derived := antigravitySeverity(text)
	if geminiSeverityRank(derived) > geminiSeverityRank(current) {
		return derived
	}
	if current == "" {
		return "medium"
	}
	return current
}

func geminiSeverityRank(value string) int {
	switch normalizeObservationMetaToken(value) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func geminiObservationID(syncKey string) string {
	return "observation_gemini_" + antigravityHash(syncKey)
}

package runtime

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CorpusMarkdownArtifact struct {
	EntryID      string    `json:"entry_id"`
	RelativePath string    `json:"relative_path"`
	Summary      string    `json:"summary"`
	RawExcerpt   string    `json:"raw_excerpt"`
	SyncKey      string    `json:"sync_key"`
	Hash         string    `json:"hash"`
	CapturedAt   time.Time `json:"captured_at"`
	SourcePath   string    `json:"-"`
	InsightItems []string  `json:"insight_items,omitempty"`
}

type CorpusMarkdownScan struct {
	Root      string                   `json:"root"`
	DataDir   string                   `json:"data_dir"`
	Artifacts []CorpusMarkdownArtifact `json:"artifacts"`
}

type CorpusMarkdownRecoverResult struct {
	GeneratedAt     time.Time                `json:"generated_at"`
	Root            string                   `json:"root"`
	DataDir         string                   `json:"data_dir"`
	DryRun          bool                     `json:"dry_run"`
	Cleanup         bool                     `json:"cleanup"`
	Considered      int                      `json:"considered"`
	Created         int                      `json:"created"`
	Existing        int                      `json:"existing"`
	Cleaned         int                      `json:"cleaned"`
	Failed          int                      `json:"failed"`
	Artifacts       []CorpusMarkdownArtifact `json:"artifacts"`
	CleanedPaths    []string                 `json:"cleaned_paths,omitempty"`
	CreatedEntryIDs []string                 `json:"created_entry_ids,omitempty"`
	Errors          []string                 `json:"errors,omitempty"`
}

func ScanCorpusMarkdownArtifacts(repoRoot string, dataDir string, limit int) (CorpusMarkdownScan, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		repoRoot = "."
	}
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		dataDir = filepath.Join(repoRoot, ".layer-os")
	}
	searchRoots := []string{
		filepath.Join(repoRoot, "knowledge", "corpus"),
		filepath.Join(dataDir, "knowledge", "corpus"),
	}
	artifacts := []CorpusMarkdownArtifact{}
	seen := map[string]bool{}
	for _, searchRoot := range searchRoots {
		if _, err := os.Stat(searchRoot); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return CorpusMarkdownScan{}, err
		}
		err := filepath.WalkDir(searchRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
				return nil
			}
			cleanPath := filepath.Clean(path)
			if seen[cleanPath] {
				return nil
			}
			seen[cleanPath] = true
			artifact, err := buildCorpusMarkdownArtifact(repoRoot, cleanPath)
			if err != nil {
				return err
			}
			artifacts = append(artifacts, artifact)
			if limit > 0 && len(artifacts) >= limit {
				return fs.SkipAll
			}
			return nil
		})
		if err != nil && err != fs.SkipAll {
			return CorpusMarkdownScan{}, err
		}
		if limit > 0 && len(artifacts) >= limit {
			break
		}
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].RelativePath < artifacts[j].RelativePath
	})
	return CorpusMarkdownScan{Root: repoRoot, DataDir: dataDir, Artifacts: artifacts}, nil
}

func (s *Service) AbsorbCorpusMarkdown(limit int, cleanup bool, dryRun bool) (CorpusMarkdownRecoverResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repoRoot := strings.TrimSpace(s.repoRoot)
	if repoRoot == "" {
		repoRoot = "."
	}
	scan, err := ScanCorpusMarkdownArtifacts(repoRoot, s.disk.baseDir, limit)
	if err != nil {
		return CorpusMarkdownRecoverResult{}, err
	}
	result := CorpusMarkdownRecoverResult{
		GeneratedAt:     zeroSafeNow(),
		Root:            scan.Root,
		DataDir:         scan.DataDir,
		DryRun:          dryRun,
		Cleanup:         cleanup,
		Artifacts:       append([]CorpusMarkdownArtifact{}, scan.Artifacts...),
		CleanedPaths:    []string{},
		CreatedEntryIDs: []string{},
		Errors:          []string{},
	}
	existingEntries, err := s.disk.loadCapitalizationEntries()
	if err != nil {
		return CorpusMarkdownRecoverResult{}, err
	}
	existingByID := map[string]bool{}
	for _, entry := range existingEntries {
		existingByID[strings.TrimSpace(entry.EntryID)] = true
	}
	cleanupCandidates := make([]CorpusMarkdownArtifact, 0, len(scan.Artifacts))
	for _, artifact := range scan.Artifacts {
		result.Considered++
		if existingByID[artifact.EntryID] {
			result.Existing++
			cleanupCandidates = append(cleanupCandidates, artifact)
			continue
		}
		if dryRun {
			continue
		}
		entry := corpusMarkdownCapitalizationEntry(artifact, s.currentActor())
		if err := s.disk.appendCapitalizationEntry(entry); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, artifact.RelativePath+": append corpus entry: "+err.Error())
			continue
		}
		existingByID[artifact.EntryID] = true
		result.Created++
		result.CreatedEntryIDs = append(result.CreatedEntryIDs, entry.EntryID)
		cleanupCandidates = append(cleanupCandidates, artifact)
	}
	if cleanup && !dryRun {
		cleaned, cleanupErrors := cleanupCorpusMarkdownPaths(repoRoot, s.disk.baseDir, cleanupCandidates)
		result.CleanedPaths = append(result.CleanedPaths, cleaned...)
		result.Cleaned += len(cleaned)
		result.Errors = append(result.Errors, cleanupErrors...)
		result.Failed += len(cleanupErrors)
	}
	sort.Strings(result.CreatedEntryIDs)
	sort.Strings(result.CleanedPaths)
	sort.Strings(result.Errors)
	return result, nil
}

func buildCorpusMarkdownArtifact(repoRoot string, path string) (CorpusMarkdownArtifact, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return CorpusMarkdownArtifact{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return CorpusMarkdownArtifact{}, err
	}
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	hash := externalEvidenceHash(rel, string(raw))
	syncKey := externalEvidenceHash("corpus_markdown", rel, hash)
	insights := corpusMarkdownInsightItems(string(raw), 4)
	return CorpusMarkdownArtifact{
		EntryID:      "cap_corpus_markdown_" + externalEvidenceHash(syncKey),
		RelativePath: rel,
		Summary:      corpusMarkdownSummary(string(raw), path),
		RawExcerpt:   limitText(strings.TrimSpace(string(raw)), 600),
		SyncKey:      syncKey,
		Hash:         hash,
		CapturedAt:   info.ModTime().UTC(),
		SourcePath:   path,
		InsightItems: insights,
	}, nil
}

func corpusMarkdownCapitalizationEntry(artifact CorpusMarkdownArtifact, actor string) CapitalizationEntry {
	situationItems := append([]string{}, limitStrings(artifact.InsightItems, 4)...)
	if len(situationItems) == 0 {
		situationItems = []string{"path:" + artifact.RelativePath}
	}
	decisionItems := []string{
		"normalize_markdown_to_json",
		"source_path:" + artifact.RelativePath,
		"sync_key:" + artifact.SyncKey,
	}
	costItems := []string{
		"loader_skips_markdown",
		"cleanup_required:" + artifact.RelativePath,
	}
	resultItems := []string{
		"source_kind:corpus_markdown_recovered",
		"artifact_hash:" + artifact.Hash,
		"artifact_path:" + artifact.RelativePath,
	}
	return CapitalizationEntry{
		EntryID:       artifact.EntryID,
		CreatedAt:     artifact.CapturedAt,
		Actor:         normalizeActor(actor),
		SourceEventID: "corpus_markdown:" + artifact.SyncKey,
		SourceKind:    "corpus.markdown_recovered",
		Situation: CapitalizationFacet{
			Summary: artifact.Summary,
			Items:   situationItems,
		},
		Decision: CapitalizationFacet{
			Summary: "Recovered markdown corpus note into canonical JSON entry",
			Items:   decisionItems,
		},
		Cost: CapitalizationFacet{
			Summary: "Markdown artifact would be ignored by the corpus loader until normalized",
			Items:   costItems,
		},
		Result: CapitalizationFacet{
			Summary: "Recovered note is now preserved in the canonical corpus ledger",
			Items:   resultItems,
		},
	}
}

func corpusMarkdownSummary(raw string, path string) string {
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "-*0123456789.[]()"))
		if trimmed == "" {
			continue
		}
		return limitText(trimmed, 180)
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return limitText(strings.ReplaceAll(base, "_", " "), 180)
}

func corpusMarkdownInsightItems(raw string, max int) []string {
	if max <= 0 {
		return []string{}
	}
	items := []string{}
	seen := map[string]bool{}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "-*0123456789.[]()"))
		trimmed = limitText(trimmed, 160)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		items = append(items, trimmed)
		if len(items) >= max {
			break
		}
	}
	return items
}

func cleanupCorpusMarkdownPaths(repoRoot string, dataDir string, artifacts []CorpusMarkdownArtifact) ([]string, []string) {
	seen := map[string]bool{}
	cleaned := []string{}
	errors := []string{}
	for _, artifact := range artifacts {
		path := strings.TrimSpace(artifact.SourcePath)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			errors = append(errors, corpusMarkdownCleanupRef(repoRoot, path)+": "+err.Error())
			continue
		}
		cleaned = append(cleaned, corpusMarkdownCleanupRef(repoRoot, path))
		pruneCorpusMarkdownParents(repoRoot, dataDir, filepath.Dir(path))
	}
	sort.Strings(cleaned)
	sort.Strings(errors)
	return cleaned, errors
}

func corpusMarkdownCleanupRef(repoRoot string, path string) string {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func pruneCorpusMarkdownParents(repoRoot string, dataDir string, dir string) {
	repoRoot = filepath.Clean(strings.TrimSpace(repoRoot))
	dataDir = filepath.Clean(strings.TrimSpace(dataDir))
	for {
		dir = filepath.Clean(dir)
		if dir == "." || dir == string(filepath.Separator) || dir == repoRoot || dir == dataDir {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

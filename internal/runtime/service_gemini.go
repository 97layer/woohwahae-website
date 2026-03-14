package runtime

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type GeminiAbsorbResult struct {
	GeneratedAt time.Time `json:"generated_at"`
	Root        string    `json:"root"`
	Considered  int       `json:"considered"`
	Created     int       `json:"created"`
	Existing    int       `json:"existing"`
	Cleaned     int       `json:"cleaned"`
	Failed      int       `json:"failed"`
	Artifacts   []string  `json:"artifacts"`
	CleanedRefs []string  `json:"cleaned_refs"`
	Errors      []string  `json:"errors"`
}

func (s *Service) AbsorbGeminiArtifacts(limit int, cleanup bool) (GeminiAbsorbResult, error) {
	root := strings.TrimSpace(s.repoRoot)
	if root == "" {
		root = "."
	}
	scan, err := ScanGeminiArtifacts(root, limit)
	if err != nil {
		return GeminiAbsorbResult{}, err
	}
	result := GeminiAbsorbResult{
		GeneratedAt: zeroSafeNow(),
		Root:        scan.Root,
		Artifacts:   []string{},
		CleanedRefs: []string{},
		Errors:      []string{},
	}
	cleanupCandidates := make([]GeminiArtifact, 0, len(scan.Artifacts))
	for _, artifact := range scan.Artifacts {
		result.Considered++
		result.Artifacts = append(result.Artifacts, strings.TrimSpace(artifact.RelativePath))
		obs := GeminiObservation(artifact)
		syncRef := geminiObservationSyncRef(artifact.SyncKey)
		existing := s.ListObservations(ObservationQuery{SourceChannel: obs.SourceChannel, Ref: syncRef, Limit: 1})
		if len(existing) > 0 {
			result.Existing++
			cleanupCandidates = append(cleanupCandidates, artifact)
			continue
		}
		if _, err := s.CreateObservation(obs); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, artifact.RelativePath+": create observation: "+err.Error())
			continue
		}
		result.Created++
		cleanupCandidates = append(cleanupCandidates, artifact)
	}
	if cleanup {
		cleaned, cleanupErrors := cleanupGeminiArtifactPaths(scan.Root, cleanupCandidates)
		result.CleanedRefs = append(result.CleanedRefs, cleaned...)
		result.Cleaned += len(cleaned)
		result.Errors = append(result.Errors, cleanupErrors...)
		result.Failed += len(cleanupErrors)
	}
	sort.Strings(result.Artifacts)
	sort.Strings(result.CleanedRefs)
	sort.Strings(result.Errors)
	return result, nil
}

func geminiObservationSyncRef(syncKey string) string {
	return "gemini_sync:" + strings.TrimSpace(syncKey)
}

func cleanupGeminiArtifactPaths(root string, artifacts []GeminiArtifact) ([]string, []string) {
	seen := map[string]bool{}
	cleaned := []string{}
	cleanupErrors := []string{}
	for _, artifact := range artifacts {
		for _, path := range artifact.RelatedPaths {
			value := strings.TrimSpace(path)
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			if err := os.Remove(value); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				rel, relErr := filepath.Rel(root, value)
				if relErr != nil {
					rel = value
				}
				cleanupErrors = append(cleanupErrors, filepath.ToSlash(rel)+": "+err.Error())
				continue
			}
			rel, relErr := filepath.Rel(root, value)
			if relErr != nil {
				rel = value
			}
			cleaned = append(cleaned, filepath.ToSlash(rel))
		}
	}
	sort.Strings(cleaned)
	sort.Strings(cleanupErrors)
	return cleaned, cleanupErrors
}

package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

// Corpus defines the interface for retrieving knowledge assets.
type Corpus interface {
	Retrieve(query string, options RetrievalOptions) ([]KnowledgePacketResult, error)
	// AddEntry allows adding new capitalizations, typically with dedup handling.
	AddEntry(entry CapitalizationEntry) bool
}

// RetrievalOptions contains settings for knowledge retrieval.
type RetrievalOptions struct {
	Limit int
}

// KnowledgePacketResult represents a search result from the corpus.
// This is a view-optimized version of data for injection.
type KnowledgePacketResult struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// SimpleCorpus is an in-memory implementation of the Corpus interface.
// It includes deduplication logic to prevent flooding.
type SimpleCorpus struct {
	mu           sync.RWMutex
	entries      []CapitalizationEntry
	recentHashes map[string]time.Time
	dedupWindow  time.Duration
}

// NewSimpleCorpus creates a new SimpleCorpus.
// It initializes the dedup cache.
func NewSimpleCorpus(entries []CapitalizationEntry) *SimpleCorpus {
	corpus := &SimpleCorpus{
		entries:      make([]CapitalizationEntry, 0, len(entries)),
		recentHashes: make(map[string]time.Time),
		dedupWindow:  1 * time.Hour, // 1 hour dedup window
	}

	for _, entry := range entries {
		corpus.AddEntry(entry)
	}
	return corpus
}

// hashEntry generates a unique signature for an entry to detect duplicates.
func hashEntry(entry CapitalizationEntry) string {
	parts := []string{
		entry.SourceKind,
		entry.Actor,
		entry.Situation.Summary,
		entry.Decision.Summary,
	}
	raw := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// AddEntry adds an entry if it's not a recent duplicate.
// Returns true if added, false if deduplicated.
func (s *SimpleCorpus) AddEntry(entry CapitalizationEntry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := hashEntry(entry)
	now := time.Now()

	// Check deduplication window
	if lastSeen, exists := s.recentHashes[hash]; exists {
		if now.Sub(lastSeen) < s.dedupWindow {
			return false // Deduplicated
		}
	}

	// Clean up old hashes periodically to prevent memory leak (naive implementation)
	if len(s.recentHashes) > 1000 {
		for h, t := range s.recentHashes {
			if now.Sub(t) > s.dedupWindow {
				delete(s.recentHashes, h)
			}
		}
	}

	s.recentHashes[hash] = now
	s.entries = append(s.entries, entry)
	return true
}

// Retrieve searches capitalization entries for the given query.
func (s *SimpleCorpus) Retrieve(query string, options RetrievalOptions) ([]KnowledgePacketResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := []KnowledgePacketResult{}
	query = strings.ToLower(query)

	for _, entry := range s.entries {
		// Simple keyword match in situation, decision, and result summaries
		match := strings.Contains(strings.ToLower(entry.Situation.Summary), query) ||
			strings.Contains(strings.ToLower(entry.Decision.Summary), query) ||
			strings.Contains(strings.ToLower(entry.Result.Summary), query)

		if match {
			results = append(results, CapitalizationEntryToResult(entry))
			if options.Limit > 0 && len(results) >= options.Limit {
				break
			}
		}
	}

	return results, nil
}

// CapitalizationEntryToResult converts a CapitalizationEntry to a KnowledgePacketResult.
func CapitalizationEntryToResult(entry CapitalizationEntry) KnowledgePacketResult {
	content := strings.Join([]string{
		"Situation: " + entry.Situation.Summary,
		"Decision: " + entry.Decision.Summary,
		"Result: " + entry.Result.Summary,
	}, "\n")

	return KnowledgePacketResult{
		ID:      entry.EntryID,
		Content: content,
		Source:  entry.SourceKind,
		Metadata: map[string]interface{}{
			"actor":           entry.Actor,
			"source_event_id": entry.SourceEventID,
		},
		Timestamp: entry.CreatedAt,
	}
}

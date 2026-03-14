package runtime

import (
	"sort"
	"strings"
)

func (s *Service) SearchKnowledge(query string) KnowledgeSearchResponse {
	query = strings.TrimSpace(query)
	if query == "" {
		return KnowledgeSearchResponse{Query: "", Results: []KnowledgeSearchResult{}, AutoThreads: []OpenThread{}}
	}
	entries := s.ListCapitalizationEntries()
	results := make([]KnowledgeSearchResult, 0, len(entries))
	for _, entry := range entries {
		matchFields, matchCount := matchCapitalizationEntry(entry, query)
		if matchCount == 0 {
			continue
		}
		results = append(results, KnowledgeSearchResult{
			Entry:       entry,
			MatchFields: matchFields,
			MatchCount:  matchCount,
		})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].MatchCount == results[j].MatchCount {
			return results[i].Entry.CreatedAt.After(results[j].Entry.CreatedAt)
		}
		return results[i].MatchCount > results[j].MatchCount
	})
	autoThreads := []OpenThread{}
	if len(results) >= 3 {
		evidence := []string{"query:" + query}
		for _, result := range limitKnowledgeSearchResults(results, 3) {
			evidence = append(evidence, "entry:"+result.Entry.EntryID)
		}
		thread, err := s.EnsureOpenThread(newOpenThread(
			openThreadKindPattern,
			"Why does keyword \""+query+"\" recur across the knowledge corpus?",
			openThreadSourceCorpus,
			[]string{"keyword:" + normalizeSearchKeyword(query)},
			evidence,
		))
		if err == nil {
			autoThreads = append(autoThreads, thread)
		}
	}
	return KnowledgeSearchResponse{
		Query:       query,
		Results:     limitKnowledgeSearchResults(results, 20),
		AutoThreads: sortAndDedupeOpenThreads(autoThreads),
	}
}

func limitKnowledgeSearchResults(items []KnowledgeSearchResult, max int) []KnowledgeSearchResult {
	if max <= 0 || len(items) <= max {
		return append([]KnowledgeSearchResult{}, items...)
	}
	return append([]KnowledgeSearchResult{}, items[:max]...)
}

func matchCapitalizationEntry(entry CapitalizationEntry, query string) ([]string, int) {
	fields := []string{}
	total := 0
	for _, candidate := range []struct {
		name  string
		facet CapitalizationFacet
	}{
		{name: "situation", facet: entry.Situation},
		{name: "decision", facet: entry.Decision},
		{name: "outcome", facet: entry.Result},
	} {
		count := matchCapitalizationFacet(candidate.facet, query)
		if count == 0 {
			continue
		}
		fields = append(fields, candidate.name)
		total += count
	}
	return fields, total
}

func matchCapitalizationFacet(facet CapitalizationFacet, query string) int {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return 0
	}
	tokens := extractCorpusTokens(query)
	if len(tokens) == 0 {
		tokens = []string{query}
	}
	texts := append([]string{facet.Summary}, facet.Items...)
	count := 0
	for _, text := range texts {
		normalized := strings.ToLower(strings.TrimSpace(text))
		if normalized == "" {
			continue
		}
		if strings.Contains(normalized, query) {
			count += 2
			continue
		}
		for _, token := range tokens {
			if strings.Contains(normalized, token) {
				count++
			}
		}
	}
	return count
}

func normalizeSearchKeyword(query string) string {
	return strings.Join(extractCorpusTokens(query), "_")
}

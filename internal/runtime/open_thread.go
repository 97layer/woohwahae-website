package runtime

import (
	"errors"
	"strings"
)

func limitOpenThreads(items []OpenThread, max int) []OpenThread {
	if items == nil {
		return []OpenThread{}
	}
	if max <= 0 || len(items) <= max {
		return append([]OpenThread{}, items...)
	}
	return append([]OpenThread{}, items[:max]...)
}

func (d *diskStore) loadOpenThreads() ([]OpenThread, error) {
	return readJSON(d.openThreadsPath(), []OpenThread{})
}

func (d *diskStore) saveOpenThreads(items []OpenThread) error {
	return writeJSONAtomic(d.openThreadsPath(), sortAndDedupeOpenThreads(items))
}

func loadStoredOpenThreads(disk *diskStore) []OpenThread {
	if disk == nil {
		return []OpenThread{}
	}
	items, err := disk.loadOpenThreads()
	if err != nil {
		return []OpenThread{}
	}
	return sortAndDedupeOpenThreads(items)
}

func mergeOpenThreads(stored []OpenThread, groups ...[]OpenThread) []OpenThread {
	merged := map[string]OpenThread{}
	for _, item := range stored {
		normalized := normalizeOpenThread(item)
		merged[normalized.ThreadID] = normalized
	}
	for _, group := range groups {
		for _, item := range group {
			normalized := normalizeOpenThread(item)
			current, exists := merged[normalized.ThreadID]
			if !exists {
				merged[normalized.ThreadID] = normalized
				continue
			}
			if current.Status != openThreadStatusOpen {
				continue
			}
			merged[normalized.ThreadID] = mergeOpenThreadRecord(current, normalized)
		}
	}
	items := make([]OpenThread, 0, len(merged))
	for _, item := range merged {
		items = append(items, item)
	}
	return sortAndDedupeOpenThreads(items)
}

func mergeOpenThreadRecord(base OpenThread, incoming OpenThread) OpenThread {
	merged := normalizeOpenThread(base)
	if strings.TrimSpace(incoming.Question) != "" {
		merged.Question = strings.TrimSpace(incoming.Question)
	}
	if strings.TrimSpace(incoming.Status) != "" {
		merged.Status = strings.TrimSpace(strings.ToLower(incoming.Status))
	}
	if len(incoming.PatternRefs) > 0 {
		merged.PatternRefs = normalizeReviewRoomEvidence(append(append([]string{}, merged.PatternRefs...), incoming.PatternRefs...))
	}
	if len(incoming.Evidence) > 0 {
		merged.Evidence = normalizeReviewRoomEvidence(append(append([]string{}, merged.Evidence...), incoming.Evidence...))
	}
	if strings.TrimSpace(incoming.Source) != "" {
		merged.Source = strings.TrimSpace(strings.ToLower(incoming.Source))
	}
	return normalizeOpenThread(merged)
}

func filterOpenThreadStatus(items []OpenThread, status string) []OpenThread {
	status = strings.TrimSpace(strings.ToLower(status))
	filtered := make([]OpenThread, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Status)) == status {
			filtered = append(filtered, normalizeOpenThread(item))
		}
	}
	return sortAndDedupeOpenThreads(filtered)
}

func (s *Service) OpenThreads(limit int) []OpenThread {
	s.mu.Lock()
	founderView := s.deriveFounderViewLocked()
	fullReviewRoom := s.currentReviewRoomLocked()
	reviewRoom := SummarizeReviewRoom(fullReviewRoom)
	founderSummary := s.deriveFounderSummaryLocked(founderView, reviewRoom)
	memory := s.memory.current()
	observations := s.observation.list()
	disk := s.disk
	s.mu.Unlock()

	_, _, _, _, _, threads := loadKnowledgeContext(disk, observations, memory, founderSummary, reviewRoom, fullReviewRoom)
	return limitOpenThreads(threads, limit)
}

func (s *Service) Threads(limit int) []OpenThread {
	s.mu.Lock()
	founderView := s.deriveFounderViewLocked()
	fullReviewRoom := s.currentReviewRoomLocked()
	reviewRoom := SummarizeReviewRoom(fullReviewRoom)
	founderSummary := s.deriveFounderSummaryLocked(founderView, reviewRoom)
	memory := s.memory.current()
	observations := s.observation.list()
	disk := s.disk
	s.mu.Unlock()

	_, _, _, _, _, openThreads := loadKnowledgeContext(disk, observations, memory, founderSummary, reviewRoom, fullReviewRoom)
	allThreads := mergeOpenThreads(loadStoredOpenThreads(disk), openThreads)
	return limitOpenThreads(allThreads, limit)
}

func (s *Service) SaveOpenThread(item OpenThread) (OpenThread, error) {
	s.mu.Lock()
	disk := s.disk
	s.mu.Unlock()

	stored := loadStoredOpenThreads(disk)
	allThreads := s.Threads(0)
	base, found := findOpenThreadByID(allThreads, item.ThreadID)
	if found {
		item = mergeOpenThreadRecord(base, item)
	}
	item = normalizeOpenThread(item)
	if !found && strings.TrimSpace(item.Source) == "" {
		item.Source = openThreadSourceManual
	}
	if !found && strings.TrimSpace(item.Question) == "" {
		return OpenThread{}, errOpenThreadQuestionRequired()
	}
	if strings.TrimSpace(item.ThreadID) == "" {
		item.ThreadID = deriveOpenThreadID(openThreadKind(item), item.Question, item.PatternRefs)
	}
	if err := item.Validate(); err != nil {
		return OpenThread{}, err
	}
	stored = upsertStoredOpenThread(stored, item)
	if err := disk.saveOpenThreads(stored); err != nil {
		return OpenThread{}, err
	}
	return item, nil
}

func (s *Service) EnsureOpenThread(item OpenThread) (OpenThread, error) {
	item = normalizeOpenThread(item)
	if strings.TrimSpace(item.ThreadID) == "" {
		item.ThreadID = deriveOpenThreadID(openThreadKind(item), item.Question, item.PatternRefs)
	}
	existing, found := findOpenThreadByID(s.Threads(0), item.ThreadID)
	if found {
		return existing, nil
	}
	return s.SaveOpenThread(item)
}

func findOpenThreadByID(items []OpenThread, threadID string) (OpenThread, bool) {
	threadID = strings.TrimSpace(threadID)
	for _, item := range items {
		if item.ThreadID == threadID {
			return normalizeOpenThread(item), true
		}
	}
	return OpenThread{}, false
}

func upsertStoredOpenThread(items []OpenThread, item OpenThread) []OpenThread {
	updated := false
	result := make([]OpenThread, 0, len(items)+1)
	for _, current := range items {
		if current.ThreadID != item.ThreadID {
			result = append(result, normalizeOpenThread(current))
			continue
		}
		result = append(result, mergeOpenThreadRecord(current, item))
		updated = true
	}
	if !updated {
		result = append(result, normalizeOpenThread(item))
	}
	return sortAndDedupeOpenThreads(result)
}

func errOpenThreadQuestionRequired() error {
	return errors.New("open thread question is required")
}

func (s *Service) TransitionOpenThread(threadID string, status string) (OpenThread, error) {
	return s.SaveOpenThread(OpenThread{ThreadID: threadID, Status: status})
}

package runtime

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func limitStrings(items []string, max int) []string {
	if items == nil {
		return []string{}
	}
	if max <= 0 || len(items) <= max {
		return append([]string{}, items...)
	}
	return append([]string{}, items[:max]...)
}

func limitParallelCandidates(items []ParallelCandidate, max int) []ParallelCandidate {
	if items == nil {
		return []ParallelCandidate{}
	}
	if max <= 0 || len(items) <= max {
		return append([]ParallelCandidate{}, items...)
	}
	return append([]ParallelCandidate{}, items[:max]...)
}

func limitCorpusLessons(items []CorpusLesson, max int) []CorpusLesson {
	if items == nil {
		return []CorpusLesson{}
	}
	if max <= 0 || len(items) <= max {
		return append([]CorpusLesson{}, items...)
	}
	return append([]CorpusLesson{}, items[:max]...)
}

func limitObservationLinks(items []ObservationLink, max int) []ObservationLink {
	if items == nil {
		return []ObservationLink{}
	}
	if max <= 0 || len(items) <= max {
		return append([]ObservationLink{}, items...)
	}
	return append([]ObservationLink{}, items[:max]...)
}

func limitProposalCandidates(items []ProposalCandidate, max int) []ProposalCandidate {
	if items == nil {
		return []ProposalCandidate{}
	}
	if max <= 0 || len(items) <= max {
		return append([]ProposalCandidate{}, items...)
	}
	return append([]ProposalCandidate{}, items[:max]...)
}

func topReviewTexts(items []ReviewRoomItem, max int) []string {
	if items == nil {
		return []string{}
	}
	texts := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, item.Text)
	}
	return limitStrings(texts, max)
}

func bindingIDsByKind(bindings []CapabilityBinding, kind string) []string {
	items := []string{}
	for _, binding := range bindings {
		if binding.Kind == kind && binding.Active {
			items = append(items, binding.ID)
		}
	}
	return items
}

func knowledgeQueryTexts(memory SystemMemory, summary FounderSummary, review ReviewRoomSummary) []string {
	items := []string{memory.CurrentFocus, summary.PrimaryAction, summary.PrimaryRef}
	if memory.CurrentGoal != nil {
		items = append(items, *memory.CurrentGoal)
	}
	items = append(items, memory.NextSteps...)
	items = append(items, memory.OpenRisks...)
	items = append(items, topReviewTexts(review.TopOpen, 3)...)
	return items
}

func corpusTokenSet(items []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, item := range items {
		for _, token := range extractCorpusTokens(item) {
			set[token] = struct{}{}
		}
	}
	return set
}

func extractCorpusTokens(raw string) []string {
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(raw)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(fields) == 0 {
		return []string{}
	}
	items := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		if utf8.RuneCountInString(field) < 2 {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		items = append(items, field)
	}
	return items
}

func corpusEntryTexts(item CapitalizationEntry) []string {
	items := []string{}
	for _, value := range []string{
		item.Situation.Summary,
		item.Decision.Summary,
		item.Cost.Summary,
		item.Result.Summary,
	} {
		if text := strings.TrimSpace(value); text != "" {
			items = append(items, text)
		}
	}
	items = append(items, item.Situation.Items...)
	items = append(items, item.Decision.Items...)
	items = append(items, item.Cost.Items...)
	items = append(items, item.Result.Items...)
	return items
}

func scoreCorpusEntry(item CapitalizationEntry, query map[string]struct{}) int {
	if len(query) == 0 {
		return 0
	}
	score := 0
	seen := map[string]struct{}{}
	for _, text := range corpusEntryTexts(item) {
		for _, token := range extractCorpusTokens(text) {
			if _, done := seen[token]; done {
				continue
			}
			seen[token] = struct{}{}
			if _, ok := query[token]; ok {
				score++
			}
		}
	}
	return score
}

func corpusEntryRole(item CapitalizationEntry) string {
	for _, value := range item.Situation.Items {
		trimmed := strings.TrimSpace(value)
		if !strings.HasPrefix(strings.ToLower(trimmed), "role:") {
			continue
		}
		role := strings.TrimSpace(trimmed[len("role:"):])
		if role != "" {
			return strings.ToLower(role)
		}
	}
	return ""
}

func corpusLessonRole(summary string) string {
	trimmed := strings.TrimSpace(summary)
	if !strings.HasPrefix(strings.ToLower(trimmed), "role:") {
		return ""
	}
	rest := strings.TrimSpace(trimmed[len("role:"):])
	if index := strings.Index(rest, " -> "); index >= 0 {
		rest = rest[:index]
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return ""
	}
	return strings.ToLower(rest)
}

func summarizeCorpusLesson(item CapitalizationEntry) string {
	parts := []string{}
	if role := corpusEntryRole(item); role != "" && strings.HasPrefix(strings.TrimSpace(item.SourceKind), "agent_job.") {
		parts = append(parts, "role:"+role)
	}
	if text := strings.TrimSpace(item.Situation.Summary); text != "" {
		parts = append(parts, text)
	}
	decision := ""
	if len(item.Decision.Items) > 0 {
		decision = strings.Join(limitStrings(item.Decision.Items, 2), "; ")
	} else if text := strings.TrimSpace(item.Decision.Summary); text != "" {
		decision = text
	}
	if decision != "" {
		parts = append(parts, decision)
	}
	if len(parts) == 0 {
		if text := strings.TrimSpace(item.Result.Summary); text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		parts = append(parts, item.EntryID)
	}
	return limitText(strings.Join(limitStrings(parts, 3), " -> "), 180)
}

func limitText(text string, max int) string {
	text = strings.TrimSpace(text)
	if max <= 0 || utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return strings.TrimSpace(string(runes[:max-1])) + "…"
}

func retrieveCorpusLessons(entries []CapitalizationEntry, queries []string, max int) []CorpusLesson {
	if max <= 0 || len(entries) == 0 {
		return []CorpusLesson{}
	}
	query := corpusTokenSet(queries)
	type scoredLesson struct {
		lesson    CorpusLesson
		score     int
		createdAt time.Time
	}
	scored := []scoredLesson{}
	for _, entry := range entries {
		score := scoreCorpusEntry(entry, query)
		if score == 0 {
			continue
		}
		scored = append(scored, scoredLesson{
			lesson: CorpusLesson{
				EntryID:    entry.EntryID,
				SourceKind: entry.SourceKind,
				Summary:    summarizeCorpusLesson(entry),
				CreatedAt:  entry.CreatedAt,
			},
			score:     score,
			createdAt: entry.CreatedAt,
		})
	}
	if len(scored) == 0 {
		items := make([]CorpusLesson, 0, max)
		for _, entry := range entries {
			items = append(items, CorpusLesson{
				EntryID:    entry.EntryID,
				SourceKind: entry.SourceKind,
				Summary:    summarizeCorpusLesson(entry),
				CreatedAt:  entry.CreatedAt,
			})
			if len(items) >= max {
				break
			}
		}
		return limitCorpusLessons(items, max)
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].createdAt.After(scored[j].createdAt)
		}
		return scored[i].score > scored[j].score
	})
	items := make([]CorpusLesson, 0, len(scored))
	for _, item := range scored {
		items = append(items, item.lesson)
	}
	return limitCorpusLessons(items, max)
}

func buildSurprisingSignals(lessons []CorpusLesson, reviewOpenCount int, openRisks []string) []string {
	signals := []string{}
	if reviewOpenCount > 0 && len(openRisks) == 0 {
		signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("review_room open_count=%d but open_risks is empty", reviewOpenCount))
	}
	if reviewOpenCount >= 15 {
		signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("review_backlog_high: %d open items", reviewOpenCount))
	}
	for _, risk := range openRisks {
		normalized := strings.ToLower(strings.TrimSpace(risk))
		if strings.Contains(normalized, "dispatch") || strings.Contains(normalized, "provider") {
			signals = appendUniqueKnowledgeText(signals, "dispatch_risk_detected")
			break
		}
	}
	sessionSummaries := map[string]int{}
	failureSummaries := map[string]int{}
	succeededSummaries := map[string]int{}
	terminalEntries := 0
	succeededEntries := 0
	for _, lesson := range lessons {
		summary := strings.TrimSpace(lesson.Summary)
		sourceKind := strings.TrimSpace(lesson.SourceKind)
		switch sourceKind {
		case "session.finished":
			if summary != "" {
				sessionSummaries[summary]++
			}
		case "agent_job.failed":
			terminalEntries++
			if summary != "" {
				failureSummaries[summary]++
			}
		case "agent_job.canceled":
			terminalEntries++
		case "agent_job.succeeded":
			terminalEntries++
			succeededEntries++
			if summary != "" {
				succeededSummaries[summary]++
			}
		}
	}
	bestSessionSummary := ""
	bestSessionCount := 0
	for summary, count := range sessionSummaries {
		if count > bestSessionCount {
			bestSessionSummary = summary
			bestSessionCount = count
		}
	}
	if bestSessionCount >= 3 {
		signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("current_focus repeated across last %d session.finished entries: %s", bestSessionCount, limitText(bestSessionSummary, 80)))
	}
	bestFailureSummary := ""
	bestFailureCount := 0
	failureRoles := map[string]int{}
	succeededRoles := map[string]int{}
	for summary, count := range failureSummaries {
		if count > bestFailureCount {
			bestFailureSummary = summary
			bestFailureCount = count
		}
		if role := corpusLessonRole(summary); role != "" {
			failureRoles[role] += count
		}
	}
	for summary, count := range succeededSummaries {
		if role := corpusLessonRole(summary); role != "" {
			succeededRoles[role] += count
		}
	}
	if bestFailureCount >= 2 {
		signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("agent_job.failed repeated %d times for: %s", bestFailureCount, limitText(bestFailureSummary, 80)))
	}
	for role := range succeededRoles {
		if failureRoles[role] >= 2 {
			signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("agent_job.succeeded after repeated_failure pattern: role:%s", role))
			return limitStrings(signals, 5)
		}
	}
	for summary := range succeededSummaries {
		if failureSummaries[summary] >= 2 {
			signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("agent_job.succeeded after repeated_failure pattern: %s", limitText(summary, 80)))
			break
		}
	}
	if terminalEntries >= 3 && succeededEntries == 0 {
		signals = appendUniqueKnowledgeText(signals, fmt.Sprintf("agent_job terminal entries=%d but succeeded entries=0", terminalEntries))
	}
	return limitStrings(signals, 5)
}

func loadKnowledgeContext(disk *diskStore, observations []ObservationRecord, memory SystemMemory, summary FounderSummary, review ReviewRoomSummary, fullReviewRoom ReviewRoom) ([]CorpusLesson, []KnowledgePacketResult, []string, []ObservationLink, []ProposalCandidate, []OpenThread) {
	entries := []CapitalizationEntry{}
	if disk != nil {
		loaded, err := disk.loadCapitalizationEntries()
		if err == nil {
			entries = loaded
		}
	}
	corpusLessons := []CorpusLesson{}
	corpusResults := []KnowledgePacketResult{}
	if len(entries) > 0 {
		queries := knowledgeQueryTexts(memory, summary, review)
		corpusLessons = retrieveCorpusLessons(entries, queries, 3)

		// RAG: Retrieve richer results using the new Corpus interface with deduplication
		corpus := NewSimpleCorpus(entries)
		seenResultIDs := map[string]bool{}
		for _, q := range queries {
			results, err := corpus.Retrieve(q, RetrievalOptions{Limit: 1})
			if err == nil {
				for _, res := range results {
					if !seenResultIDs[res.ID] {
						corpusResults = append(corpusResults, res)
						seenResultIDs[res.ID] = true
					}
				}
			}
			if len(corpusResults) >= 3 {
				break
			}
		}
	}
	surprising := buildSurprisingSignals(corpusLessons, review.OpenCount, memory.OpenRisks)
	if review.OpenCount > 0 && len(memory.NextSteps) == 0 {
		surprising = appendUniqueKnowledgeText(surprising, fmt.Sprintf("review_room open_count=%d but next_steps is empty", review.OpenCount))
	}
	surprising = limitStrings(surprising, 5)
	observationLinks := deriveObservationLinks(observations)
	derivedThreads := append(deriveOpenThreads(fullReviewRoom, surprising), detectCorpusPatterns(corpusLessons)...)
	openThreads := filterOpenThreadStatus(mergeOpenThreads(loadStoredOpenThreads(disk), derivedThreads), openThreadStatusOpen)
	proposalCandidates := mergeProposalCandidates(deriveProposalCandidates(observationLinks, openThreads), deriveReviewProposalCandidates(fullReviewRoom, openThreads))
	return corpusLessons, corpusResults, surprising, observationLinks, proposalCandidates, openThreads
}

func copyStringRef(value *string) *string {
	if value == nil {
		return nil
	}
	copy := strings.TrimSpace(*value)
	if copy == "" {
		return nil
	}
	return &copy
}

func (s *Service) Knowledge() KnowledgePacket {
	s.mu.Lock()
	founderView := s.deriveFounderViewLocked()
	fullReviewRoom := s.currentReviewRoomLocked()
	reviewRoom := SummarizeReviewRoom(fullReviewRoom)
	founderSummary := s.deriveFounderSummaryLocked(founderView, reviewRoom)
	parallel := deriveParallelCandidates(founderView, founderSummary, fullReviewRoom)
	branches := s.branch.list()
	caps := capabilityRegistryFromAdapters(AdapterSummary{
		Gateway:                s.gatewayAdapter.Name(),
		GatewaySemantics:       s.gatewayAdapter.Semantics(),
		GatewayDispatchEnabled: s.gatewayAdapter.DispatchEnabled(),
		GatewayRequiredMode:    canonicalPolicyMode(s.gatewayAdapter.RequiredMode()),
		Verify:                 s.verifyAdapter.Name(),
		Deploy:                 s.deployAdapter.Name(),
		Rollback:               s.rollbackAdapter.Name(),
	})
	memory := s.memory.current()
	observations := s.observation.list()
	disk := s.disk
	s.mu.Unlock()

	corpusLessons, corpusResults, surprising, observationLinks, proposalCandidates, openThreads := loadKnowledgeContext(disk, observations, memory, founderSummary, reviewRoom, fullReviewRoom)
	actionRoutes := deriveKnowledgeActionRoutes(memory, reviewRoom, surprising, proposalCandidates, openThreads)
	actionHints := deriveKnowledgeActionHints(actionRoutes)

	return KnowledgePacket{
		GeneratedAt:         zeroSafeNow(),
		EnvironmentAdvisory: currentEnvironmentAdvisory(),
		Authority:           caps.Authority,
		CurrentFocus:        memory.CurrentFocus,
		CurrentGoal:         copyStringRef(memory.CurrentGoal),
		NextSteps:           limitStrings(memory.NextSteps, 3),
		OpenRisks:           limitStrings(memory.OpenRisks, 3),
		PrimaryAction:       founderSummary.PrimaryAction,
		PrimaryRef:          founderSummary.PrimaryRef,
		PriorityRationale:   founderSummary.PriorityRationale,
		ReviewOpenCount:     reviewRoom.OpenCount,
		ReviewTopOpen:       topReviewTexts(reviewRoom.TopOpen, 3),
		CorpusLessons:       corpusLessons,
		CorpusResults:       corpusResults,
		Surprising:          surprising,
		ObservationLinks:    limitObservationLinks(observationLinks, 3),
		ProposalCandidates:  limitProposalCandidates(proposalCandidates, 3),
		OpenThreads:         limitOpenThreads(openThreads, 3),
		ActionHints:         actionHints,
		ActionRoutes:        actionRoutes,
		ActiveBranches:      activeBranches(branches, 3),
		ParallelCandidates:  limitParallelCandidates(parallel, 3),
		DefaultActor:        caps.DefaultActor,
		Actors:              bindingIDsByKind(caps.Bindings, "actor"),
		Providers:           bindingIDsByKind(caps.Bindings, "provider"),
		GatewaySemantics:    s.gatewayAdapter.Semantics(),
	}
}

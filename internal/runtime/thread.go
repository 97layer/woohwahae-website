package runtime

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strings"
)

const (
	openThreadStatusOpen       = "open"
	openThreadStatusResolved   = "resolved"
	openThreadStatusDismissed  = "dismissed"
	openThreadSourceCorpus     = "corpus_signal"
	openThreadSourceReviewRoom = "review_room_drift"
	openThreadSourceManual     = "manual"
	openThreadKindPattern      = "pattern"
	openThreadKindContradict   = "contradiction"
	openThreadKindStale        = "stale"
	openThreadKindCuriosity    = "curiosity"
)

func deriveOpenThreads(reviewRoom ReviewRoom, surprising []string) []OpenThread {
	items := []OpenThread{}
	for _, signal := range surprising {
		thread, ok := openThreadFromSignal(signal)
		if !ok {
			continue
		}
		items = append(items, thread)
	}
	for _, item := range reviewRoom.Open {
		items = append(items, openThreadsFromReviewRoomItem("open", item)...)
	}
	for _, item := range reviewRoom.Deferred {
		items = append(items, openThreadsFromReviewRoomItem("deferred", item)...)
	}
	return sortAndDedupeOpenThreads(items)
}

func openThreadsFromReviewRoomItem(section string, item ReviewRoomItem) []OpenThread {
	item = normalizeReviewRoomItem(item)
	patternRefs := reviewItemRefs(item)
	evidence := normalizeReviewRoomEvidence(append([]string{"review_item:" + item.Text, "item_source:" + item.Source, "section:" + strings.TrimSpace(section)}, item.Evidence...))
	threads := []OpenThread{}
	curiosityKind := openThreadKindCuriosity
	if strings.EqualFold(strings.TrimSpace(section), "deferred") {
		curiosityKind = openThreadKindStale
	}
	if strings.TrimSpace(item.Why) != "" {
		threads = append(threads, newOpenThread(curiosityKind, item.Why, openThreadSourceReviewRoom, patternRefs, evidence))
	}
	if item.WhyUnresolved != nil {
		threads = append(threads, newOpenThread(curiosityKind, strings.TrimSpace(*item.WhyUnresolved), openThreadSourceReviewRoom, patternRefs, evidence))
	}
	for _, contradiction := range item.Contradictions {
		threads = append(threads, newOpenThread(openThreadKindContradict, strings.TrimSpace(contradiction), openThreadSourceReviewRoom, patternRefs, evidence))
	}
	if item.Contradiction != nil {
		threads = append(threads, newOpenThread(openThreadKindContradict, strings.TrimSpace(*item.Contradiction), openThreadSourceReviewRoom, patternRefs, evidence))
	}
	return sortAndDedupeOpenThreads(threads)
}

func openThreadFromSignal(signal string) (OpenThread, bool) {
	signal = strings.TrimSpace(signal)
	if signal == "" {
		return OpenThread{}, false
	}
	kind := openThreadKindPattern
	question := signal
	patternRefs := []string{}
	switch {
	case strings.Contains(signal, "open_risks is empty"):
		kind = openThreadKindStale
		question = "Why are review-room tensions open while open_risks is empty?"
		patternRefs = []string{"open_risks_gap"}
	case strings.Contains(signal, "next_steps is empty"):
		kind = openThreadKindStale
		question = "What concrete next step is missing for the current open review tensions?"
		patternRefs = []string{"next_steps_gap"}
	case strings.HasPrefix(signal, "current_focus repeated across last "):
		patternRefs = []string{"focus_repeat"}
	case strings.HasPrefix(signal, "agent_job.failed repeated "):
		patternRefs = []string{"failure_repeat"}
	case strings.HasPrefix(signal, "agent_job terminal entries="):
		patternRefs = []string{"success_gap"}
		question = "Why are agent jobs reaching terminal states without any succeeded outcomes?"
	case strings.HasPrefix(signal, "review_open carry-over persisted across last "):
		kind = openThreadKindStale
		patternRefs = []string{"review_carry_over"}
	}
	evidence := []string{"signal:" + signal}
	return newOpenThread(kind, question, openThreadSourceCorpus, patternRefs, evidence), true
}

func newOpenThread(kind string, question string, source string, patternRefs []string, evidence []string) OpenThread {
	item := OpenThread{
		Question:    strings.TrimSpace(question),
		Status:      openThreadStatusOpen,
		PatternRefs: append([]string{}, patternRefs...),
		Evidence:    append([]string{}, evidence...),
		Source:      strings.TrimSpace(strings.ToLower(source)),
	}
	item.ThreadID = deriveOpenThreadID(kind, item.Question, item.PatternRefs)
	return normalizeOpenThread(item)
}

func deriveOpenThreadID(kind string, question string, patternRefs []string) string {
	kind = strings.TrimSpace(strings.ToLower(kind))
	if kind == "" {
		kind = openThreadKindCuriosity
	}
	payload := strings.Join(append([]string{kind, strings.TrimSpace(question)}, patternRefs...), "|")
	return fmt.Sprintf("%s_%08x", kind, crc32.ChecksumIEEE([]byte(payload)))
}

func normalizeOpenThread(item OpenThread) OpenThread {
	item.ThreadID = strings.TrimSpace(item.ThreadID)
	item.Question = strings.TrimSpace(item.Question)
	item.Status = strings.TrimSpace(strings.ToLower(item.Status))
	item.PatternRefs = normalizeReviewRoomEvidence(item.PatternRefs)
	item.Evidence = normalizeReviewRoomEvidence(item.Evidence)
	item.Source = strings.TrimSpace(strings.ToLower(item.Source))
	if item.Status == "" {
		item.Status = openThreadStatusOpen
	}
	if item.PatternRefs == nil {
		item.PatternRefs = []string{}
	}
	if item.Evidence == nil {
		item.Evidence = []string{}
	}
	if item.Source == "" {
		item.Source = openThreadSourceManual
	}
	if item.ThreadID == "" && item.Question != "" {
		item.ThreadID = deriveOpenThreadID(openThreadKind(item), item.Question, item.PatternRefs)
	}
	return item
}

func openThreadKind(item OpenThread) string {
	threadID := strings.TrimSpace(item.ThreadID)
	if prefix, _, ok := strings.Cut(threadID, "_"); ok {
		switch prefix {
		case openThreadKindPattern, openThreadKindContradict, openThreadKindStale, openThreadKindCuriosity:
			return prefix
		}
	}
	if strings.TrimSpace(strings.ToLower(item.Source)) == openThreadSourceCorpus {
		return openThreadKindPattern
	}
	for _, ref := range item.PatternRefs {
		if strings.Contains(strings.ToLower(ref), "contradiction") {
			return openThreadKindContradict
		}
	}
	return openThreadKindCuriosity
}

func openThreadPriority(item OpenThread) string {
	switch openThreadKind(item) {
	case openThreadKindContradict:
		return "high"
	case openThreadKindPattern, openThreadKindStale:
		return "medium"
	default:
		return "low"
	}
}

func openThreadStatusRank(status string) int {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case openThreadStatusOpen:
		return 0
	case openThreadStatusResolved:
		return 1
	case openThreadStatusDismissed:
		return 2
	default:
		return 3
	}
}

func openThreadSourceRank(source string) int {
	switch strings.TrimSpace(strings.ToLower(source)) {
	case openThreadSourceCorpus:
		return 0
	case openThreadSourceReviewRoom:
		return 1
	case openThreadSourceManual:
		return 2
	default:
		return 3
	}
}

func sortAndDedupeOpenThreads(items []OpenThread) []OpenThread {
	if len(items) == 0 {
		return []OpenThread{}
	}
	seen := map[string]struct{}{}
	out := make([]OpenThread, 0, len(items))
	for _, item := range items {
		normalized := normalizeOpenThread(item)
		if normalized.ThreadID == "" {
			continue
		}
		if _, ok := seen[normalized.ThreadID]; ok {
			continue
		}
		seen[normalized.ThreadID] = struct{}{}
		out = append(out, normalized)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if openThreadStatusRank(out[i].Status) != openThreadStatusRank(out[j].Status) {
			return openThreadStatusRank(out[i].Status) < openThreadStatusRank(out[j].Status)
		}
		if openThreadSourceRank(out[i].Source) != openThreadSourceRank(out[j].Source) {
			return openThreadSourceRank(out[i].Source) < openThreadSourceRank(out[j].Source)
		}
		if openThreadPriority(out[i]) != openThreadPriority(out[j]) {
			return proposalPriorityRank(openThreadPriority(out[i])) > proposalPriorityRank(openThreadPriority(out[j]))
		}
		return out[i].ThreadID < out[j].ThreadID
	})
	return out
}

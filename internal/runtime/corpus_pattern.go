package runtime

import (
	"fmt"
	"strings"
)

func detectCorpusPatterns(lessons []CorpusLesson) []OpenThread {
	items := []OpenThread{}
	if thread, ok := detectRepeatedFocusPattern(lessons); ok {
		items = append(items, thread)
	}
	if thread, ok := detectRepeatedFailureRolePattern(lessons); ok {
		items = append(items, thread)
	}
	if thread, ok := detectNoSucceededJobsPattern(lessons); ok {
		items = append(items, thread)
	}
	if len(items) == 0 {
		return []OpenThread{}
	}
	return sortAndDedupeOpenThreads(items)
}

func detectRepeatedFocusPattern(lessons []CorpusLesson) (OpenThread, bool) {
	counts := map[string]int{}
	bestKind := ""
	bestCount := 0

	for _, lesson := range lessons {
		sourceKind := strings.TrimSpace(lesson.SourceKind)
		if sourceKind == "" {
			continue
		}
		counts[sourceKind]++
		if counts[sourceKind] > bestCount {
			bestKind = sourceKind
			bestCount = counts[sourceKind]
		}
	}

	if bestCount < 3 || bestKind == "" {
		return OpenThread{}, false
	}

	question := fmt.Sprintf("repeated_pattern: source_kind=%s repeated %d times in corpus lessons", bestKind, bestCount)
	thread := newOpenThread(
		openThreadKindPattern,
		question,
		openThreadSourceCorpus,
		[]string{"repeated_pattern", "source_kind:" + bestKind},
		[]string{"source_kind:" + bestKind, fmt.Sprintf("count:%d", bestCount)},
	)
	return thread, true
}

func detectRepeatedFailureRolePattern(lessons []CorpusLesson) (OpenThread, bool) {
	type roleStat struct {
		failed int
		total  int
	}

	stats := map[string]roleStat{}
	bestRole := ""
	bestRatio := 0.0
	bestTotal := 0
	bestFailed := 0

	for _, lesson := range lessons {
		role := corpusLessonRole(lesson.Summary)
		if role == "" {
			continue
		}

		var terminal bool
		var failed bool
		switch strings.TrimSpace(lesson.SourceKind) {
		case "agent_job.failed":
			terminal = true
			failed = true
		case "agent_job.succeeded", "agent_job.canceled":
			terminal = true
		}
		if !terminal {
			continue
		}

		stat := stats[role]
		stat.total++
		if failed {
			stat.failed++
		}
		stats[role] = stat

		ratio := float64(stat.failed) / float64(stat.total)
		if ratio > 0.6 {
			if bestRole == "" || ratio > bestRatio || (ratio == bestRatio && stat.total > bestTotal) {
				bestRole = role
				bestRatio = ratio
				bestTotal = stat.total
				bestFailed = stat.failed
			}
		}
	}

	if bestRole == "" {
		return OpenThread{}, false
	}
	question := fmt.Sprintf("failure_pattern: role:%s failed_ratio=%.2f (%d/%d terminal job lessons)", bestRole, bestRatio, bestFailed, bestTotal)
	thread := newOpenThread(
		openThreadKindPattern,
		question,
		openThreadSourceCorpus,
		[]string{"failure_pattern", "role:" + bestRole},
		[]string{"role:" + bestRole, fmt.Sprintf("failed:%d", bestFailed), fmt.Sprintf("total:%d", bestTotal), fmt.Sprintf("failed_ratio:%.2f", bestRatio)},
	)
	return thread, true
}

func detectNoSucceededJobsPattern(lessons []CorpusLesson) (OpenThread, bool) {
	terminalCount := 0
	succeededCount := 0

	for _, lesson := range lessons {
		switch strings.TrimSpace(lesson.SourceKind) {
		case "agent_job.failed", "agent_job.canceled", "agent_job.succeeded":
			terminalCount++
			if lesson.SourceKind == "agent_job.succeeded" {
				succeededCount++
			}
		}
	}

	if terminalCount == 0 || succeededCount > 0 {
		return OpenThread{}, false
	}
	question := fmt.Sprintf("no_completed_jobs: agent_job terminal entries=%d but succeeded entries=0", terminalCount)
	thread := newOpenThread(
		openThreadKindPattern,
		question,
		openThreadSourceCorpus,
		[]string{"no_completed_jobs"},
		[]string{fmt.Sprintf("terminal_entries:%d", terminalCount), "succeeded_entries:0"},
	)
	return thread, true
}

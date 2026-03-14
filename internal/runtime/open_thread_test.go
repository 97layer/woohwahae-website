package runtime

import (
	"strings"
	"testing"
)

func TestOpenThreadsDeriveFromSignalsAndReviewTension(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	why := "race root cause is still not isolated"
	contradiction := "accepted mitigation exists but the same failure persists"
	ref := "issue_001"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Fix data race.", Kind: "bug", Severity: "high", Source: "agent.codex", Ref: &ref, WhyUnresolved: &why, Contradiction: &contradiction, PatternRefs: []string{"issue_001", "session_003"}}); err != nil {
		t.Fatalf("add review room tension: %v", err)
	}
	threads := service.OpenThreads(10)
	if len(threads) < 3 {
		t.Fatalf("expected derived open threads, got %+v", threads)
	}
	questions := make([]string, 0, len(threads))
	for _, item := range threads {
		questions = append(questions, item.Question)
	}
	joined := strings.ToLower(strings.Join(questions, "\n"))
	for _, needle := range []string{"open_risks", "race root cause", "accepted mitigation exists"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected open thread signal %q, got %+v", needle, threads)
		}
	}
}

func TestOpenThreadsCarryLeadingReviewTextRef(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	why := "No explicit rationale was recorded for this agenda."
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "[intel_011] Review why-layer still lacks consistent adoption.", Kind: "agenda", Severity: "high", Source: "claude.arch.review.2026-03-08", WhyUnresolved: &why}); err != nil {
		t.Fatalf("add review room tension: %v", err)
	}
	threads := service.OpenThreads(10)
	if len(threads) == 0 {
		t.Fatalf("expected derived open threads, got %+v", threads)
	}
	found := false
	for _, thread := range threads {
		if !strings.HasPrefix(thread.ThreadID, openThreadKindCuriosity+"_") && !strings.HasPrefix(thread.ThreadID, openThreadKindStale+"_") {
			continue
		}
		for _, ref := range thread.PatternRefs {
			if ref == "intel_011" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected open thread pattern_refs to include intel_011, got %+v", threads)
	}
}

func TestOpenThreadsLimitAndValidation(t *testing.T) {
	item := newOpenThread(openThreadKindPattern, "Why is bootstrap runtime still repeating?", openThreadSourceCorpus, []string{"focus_repeat"}, []string{"signal:focus"})
	if err := item.Validate(); err != nil {
		t.Fatalf("validate open thread: %v", err)
	}
	limited := limitOpenThreads([]OpenThread{item, item, item}, 2)
	if len(limited) != 2 {
		t.Fatalf("expected limit 2, got %+v", limited)
	}
}

func TestSaveOpenThreadTransitionsDerivedThread(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	why := "race root cause is still not isolated"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{Text: "Fix data race.", Kind: "bug", Severity: "high", Source: "agent.codex", WhyUnresolved: &why}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	threads := service.OpenThreads(10)
	if len(threads) == 0 {
		t.Fatalf("expected derived thread, got %+v", threads)
	}
	updated, err := service.TransitionOpenThread(threads[0].ThreadID, openThreadStatusResolved)
	if err != nil {
		t.Fatalf("transition thread: %v", err)
	}
	if updated.Status != openThreadStatusResolved {
		t.Fatalf("expected resolved status, got %+v", updated)
	}
	all := service.Threads(10)
	stored, found := findOpenThreadByID(all, updated.ThreadID)
	if !found || stored.Status != openThreadStatusResolved {
		t.Fatalf("expected stored resolved thread, got %+v", all)
	}
}

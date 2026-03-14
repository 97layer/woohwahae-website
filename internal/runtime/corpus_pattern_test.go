package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestCorpusPatternsReturnsEmptySliceWhenNoPatterns(t *testing.T) {
	threads := detectCorpusPatterns([]CorpusLesson{})
	if threads == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(threads) != 0 {
		t.Fatalf("expected no threads, got %+v", threads)
	}
}

func TestCorpusPatternsDetectsRepeatedSourceKind(t *testing.T) {
	now := time.Now().UTC()
	lessons := []CorpusLesson{
		{EntryID: "entry_001", SourceKind: "session.finished", Summary: "focus one", CreatedAt: now.Add(-3 * time.Minute)},
		{EntryID: "entry_002", SourceKind: "session.finished", Summary: "focus two", CreatedAt: now.Add(-2 * time.Minute)},
		{EntryID: "entry_003", SourceKind: "session.finished", Summary: "focus three", CreatedAt: now.Add(-1 * time.Minute)},
	}

	thread, ok := detectRepeatedFocusPattern(lessons)
	if !ok {
		t.Fatal("expected repeated source kind pattern")
	}
	if !strings.HasPrefix(thread.ThreadID, openThreadKindPattern+"_") || thread.Source != openThreadSourceCorpus {
		t.Fatalf("unexpected thread envelope: %+v", thread)
	}
	if thread.PatternRefs[0] != "repeated_pattern" {
		t.Fatalf("expected repeated_pattern ref, got %+v", thread.PatternRefs)
	}
}

func TestCorpusPatternsDetectsRepeatedFailureRole(t *testing.T) {
	now := time.Now().UTC()
	lessons := []CorpusLesson{
		{EntryID: "entry_001", SourceKind: "agent_job.failed", Summary: "role:planner -> failed once", CreatedAt: now.Add(-3 * time.Minute)},
		{EntryID: "entry_002", SourceKind: "agent_job.failed", Summary: "role:planner -> failed twice", CreatedAt: now.Add(-2 * time.Minute)},
		{EntryID: "entry_003", SourceKind: "agent_job.canceled", Summary: "role:planner -> canceled", CreatedAt: now.Add(-1 * time.Minute)},
	}

	thread, ok := detectRepeatedFailureRolePattern(lessons)
	if !ok {
		t.Fatal("expected repeated failure role pattern")
	}
	if !strings.HasPrefix(thread.ThreadID, openThreadKindPattern+"_") || thread.Source != openThreadSourceCorpus {
		t.Fatalf("unexpected thread envelope: %+v", thread)
	}
	if len(thread.PatternRefs) < 2 || thread.PatternRefs[0] != "failure_pattern" || thread.PatternRefs[1] != "role:planner" {
		t.Fatalf("expected role-tagged failure pattern refs, got %+v", thread.PatternRefs)
	}
}

func TestCorpusPatternsDetectsNoCompletedJobs(t *testing.T) {
	now := time.Now().UTC()
	lessons := []CorpusLesson{
		{EntryID: "entry_001", SourceKind: "agent_job.failed", Summary: "role:planner -> fail", CreatedAt: now.Add(-3 * time.Minute)},
		{EntryID: "entry_002", SourceKind: "agent_job.canceled", Summary: "role:implementer -> cancel", CreatedAt: now.Add(-2 * time.Minute)},
		{EntryID: "entry_003", SourceKind: "agent_job.failed", Summary: "role:verifier -> fail", CreatedAt: now.Add(-1 * time.Minute)},
	}

	thread, ok := detectNoSucceededJobsPattern(lessons)
	if !ok {
		t.Fatal("expected no completed jobs pattern")
	}
	if !strings.HasPrefix(thread.ThreadID, openThreadKindPattern+"_") || thread.Source != openThreadSourceCorpus {
		t.Fatalf("unexpected thread envelope: %+v", thread)
	}
	if len(thread.PatternRefs) == 0 || thread.PatternRefs[0] != "no_completed_jobs" {
		t.Fatalf("expected no_completed_jobs ref, got %+v", thread.PatternRefs)
	}
}

func TestCorpusPatternsCombinesDetectedPatterns(t *testing.T) {
	now := time.Now().UTC()
	lessons := []CorpusLesson{
		{EntryID: "entry_001", SourceKind: "session.finished", Summary: "focus one", CreatedAt: now.Add(-6 * time.Minute)},
		{EntryID: "entry_002", SourceKind: "session.finished", Summary: "focus two", CreatedAt: now.Add(-5 * time.Minute)},
		{EntryID: "entry_003", SourceKind: "session.finished", Summary: "focus three", CreatedAt: now.Add(-4 * time.Minute)},
		{EntryID: "entry_004", SourceKind: "agent_job.failed", Summary: "role:planner -> fail one", CreatedAt: now.Add(-3 * time.Minute)},
		{EntryID: "entry_005", SourceKind: "agent_job.failed", Summary: "role:planner -> fail two", CreatedAt: now.Add(-2 * time.Minute)},
		{EntryID: "entry_006", SourceKind: "agent_job.canceled", Summary: "role:planner -> cancel", CreatedAt: now.Add(-1 * time.Minute)},
	}

	threads := detectCorpusPatterns(lessons)
	if len(threads) != 3 {
		t.Fatalf("expected three pattern threads, got %+v", threads)
	}
}

package main

import (
	"strings"
	"testing"

	"layer-os/internal/runtime"
)

type threadServiceStub struct {
	limit int
	items []runtime.OpenThread
}

func (s *threadServiceStub) ListOpenThreads(limit int) []runtime.OpenThread {
	s.limit = limit
	return s.items
}

func TestRunThreadListWritesItems(t *testing.T) {
	service := &threadServiceStub{items: []runtime.OpenThread{{ThreadID: "pattern_001", Question: "Why is bootstrap runtime still repeating?", Status: "open", Source: "corpus_signal", PatternRefs: []string{"focus_repeat"}, Evidence: []string{"signal:focus"}}}}
	raw := captureStdout(t, func() {
		runThread(service, []string{"list", "--limit", "2"})
	})
	if service.limit != 2 {
		t.Fatalf("expected limit 2, got %d", service.limit)
	}
	if !strings.Contains(raw, "thread_001") {
		if !strings.Contains(raw, "pattern_001") {
			t.Fatalf("expected thread output, got %s", raw)
		}
	}
}

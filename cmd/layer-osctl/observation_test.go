package main

import (
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type observationServiceStub struct {
	auth         runtime.AuthStatus
	query        runtime.ObservationQuery
	items        []runtime.ObservationRecord
	createdInput runtime.ObservationRecord
	created      runtime.ObservationRecord
}

func (s *observationServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *observationServiceStub) ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord {
	s.query = query
	return s.items
}
func (s *observationServiceStub) CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error) {
	s.createdInput = item
	if s.created.ObservationID == "" {
		s.created = item
		if s.created.ObservationID == "" {
			s.created.ObservationID = "observation_001"
		}
	}
	return s.created, nil
}

func TestRunObservationListWritesItems(t *testing.T) {
	service := &observationServiceStub{items: []runtime.ObservationRecord{{ObservationID: "observation_001", SourceChannel: "terminal", Topic: "queue drift", Actor: "system", Refs: []string{"job_001"}, Confidence: "medium", RawExcerpt: "Queue drift", NormalizedSummary: "Queue drift", CapturedAt: time.Now().UTC()}}}
	raw := captureStdout(t, func() {
		runObservation(service, []string{"list", "--channel", "terminal", "--topic", "queue", "--limit", "5"})
	})
	if service.query.SourceChannel != "terminal" || service.query.Topic != "queue" || service.query.Limit != 5 {
		t.Fatalf("unexpected observation query: %+v", service.query)
	}
	if !strings.Contains(raw, "observation_001") || !strings.Contains(raw, "queue drift") {
		t.Fatalf("unexpected observation list output: %s", raw)
	}
}

func TestRunObservationCreateWritesCreatedItem(t *testing.T) {
	service := &observationServiceStub{created: runtime.ObservationRecord{ObservationID: "observation_001", SourceChannel: "telegram", Topic: "migration", Actor: "founder", Refs: []string{"thread_001"}, Confidence: "high", RawExcerpt: "Legacy issue", NormalizedSummary: "Legacy issue", CapturedAt: time.Now().UTC()}}
	raw := captureStdout(t, func() {
		runObservation(service, []string{"create", "--channel", "telegram", "--topic", "migration", "--summary", "Legacy issue", "--refs", "thread_001", "--confidence", "high"})
	})
	if service.createdInput.SourceChannel != "telegram" || service.createdInput.Topic != "migration" || service.createdInput.Confidence != "high" {
		t.Fatalf("unexpected created observation input: %+v", service.createdInput)
	}
	if !strings.Contains(raw, "observation_001") || !strings.Contains(raw, "telegram") {
		t.Fatalf("unexpected observation create output: %s", raw)
	}
}

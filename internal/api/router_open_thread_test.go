package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestOpenThreadsRouteListsDerivedThreads(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	why := "race root cause is still not isolated"
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{Text: "Fix data race.", Kind: "bug", Severity: "high", Source: "agent.codex", WhyUnresolved: &why}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/open-threads?limit=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Items []runtime.OpenThread `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode open threads response: %v", err)
	}
	if len(response.Items) != 2 {
		t.Fatalf("expected limited open threads, got %+v", response.Items)
	}
}

func TestThreadsRouteCreatesManualThread(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	body := strings.NewReader(`{"question":"Why is queue drift repeating?","pattern_refs":["queue_drift"],"evidence":["manual:test"],"source":"manual","status":"open"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/threads", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var item runtime.OpenThread
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if item.Source != "manual" || item.Status != "open" || item.ThreadID == "" {
		t.Fatalf("unexpected created thread: %+v", item)
	}
}

func TestKnowledgeSearchRouteCreatesPatternThread(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	for i := 0; i < 3; i++ {
		entry := runtime.CapitalizationEntry{
			EntryID:       "entry_00" + string(rune('1'+i)),
			CreatedAt:     zeroTimeForTest(i),
			Actor:         "tester",
			SourceEventID: "event_00" + string(rune('1'+i)),
			SourceKind:    "session.finished",
			Situation:     runtime.CapitalizationFacet{Summary: "Queue drift blocks release", Items: []string{"queue drift repeated"}},
			Decision:      runtime.CapitalizationFacet{Summary: "Queue drift triage", Items: []string{"queue drift review"}},
			Cost:          runtime.CapitalizationFacet{Summary: "cost", Items: []string{"time"}},
			Result:        runtime.CapitalizationFacet{Summary: "Queue drift still open", Items: []string{"queue drift unresolved"}},
		}
		if err := service.DebugAppendCapitalizationEntry(entry); err != nil {
			t.Fatalf("append entry: %v", err)
		}
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/knowledge/search?q=queue%20drift", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response runtime.KnowledgeSearchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(response.Results) < 3 || len(response.AutoThreads) != 1 {
		t.Fatalf("expected search results and auto thread, got %+v", response)
	}
}

func zeroTimeForTest(offset int) time.Time {
	return time.Date(2026, time.March, 8, 12, offset, 0, 0, time.UTC)
}

func TestOpenThreadsRouteRejectsInvalidLimit(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/open-threads?limit=bad", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

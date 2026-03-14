package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestEventPostRouteCreatesEvent(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	raw := []byte(`{"kind":"session.finished","surface":"api","work_item_id":"system","stage":"discover","data":{"note":"resume"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/events", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var item runtime.EventEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if item.Actor != "gemini" || item.Kind != "session.finished" {
		t.Fatalf("unexpected event payload: %+v", item)
	}
}

func TestEventStreamRouteStreamsRecentEvents(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.CreateEvent(runtime.EventCreateInput{Kind: "session.finished", Surface: runtime.SurfaceAPI, WorkItemID: "system", Stage: runtime.StageDiscover, Data: map[string]any{"note": "resume"}}); err != nil {
		t.Fatalf("create event: %v", err)
	}
	router := NewRouter(service)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/events/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		router.ServeHTTP(rec, req)
	}()
	time.Sleep(900 * time.Millisecond)
	cancel()
	<-done

	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("expected event-stream content type, got %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: session.finished") {
		t.Fatalf("expected session event in stream, got %s", body)
	}
}

func TestSessionBootstrapRouteReturnsFullPacket(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Recover queue", NextSteps: []string{"resume"}, OpenRisks: []string{}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/session/bootstrap?full=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var response struct {
		Session runtime.SessionBootstrapPacket `json:"session"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if response.Session.Source != "daemon" || response.Session.Handoff == nil {
		t.Fatalf("unexpected session bootstrap packet: %+v", response.Session)
	}
	if response.Session.Continuity == nil {
		t.Fatalf("expected continuity bootstrap view, got %+v", response.Session)
	}
}

func TestSessionCheckpointRouteStoresResume(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	raw := []byte(`{"source":"terminal","current_focus":"Lock queue","next_steps":["sync"],"open_risks":["drift"],"handoff_note":"continue here","note":"mid-flight","refs":["thread:terminal-1"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/session/checkpoint", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var checkpoint runtime.SessionCheckpoint
	if err := json.NewDecoder(rec.Body).Decode(&checkpoint); err != nil {
		t.Fatalf("decode checkpoint: %v", err)
	}
	if checkpoint.Actor != "gemini" || checkpoint.CurrentFocus != "Lock queue" {
		t.Fatalf("unexpected checkpoint: %+v", checkpoint)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/api/layer-os/session/bootstrap", nil)
	bootstrapRec := httptest.NewRecorder()
	router.ServeHTTP(bootstrapRec, bootstrapReq)
	if bootstrapRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", bootstrapRec.Code)
	}
	var response struct {
		Session runtime.SessionBootstrapPacket `json:"session"`
	}
	if err := json.NewDecoder(bootstrapRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if response.Session.Resume == nil || response.Session.Resume.Actor != "gemini" {
		t.Fatalf("expected resume checkpoint, got %+v", response.Session)
	}
	if response.Session.Continuity == nil || response.Session.Continuity.Record == nil || response.Session.Continuity.Record.Actor != "gemini" {
		t.Fatalf("expected continuity record, got %+v", response.Session.Continuity)
	}
}

func TestSessionNoteRouteAppendsContinuityObservation(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	raw := []byte(`{"source":"terminal","kind":"todo","text":"check queue health","refs":["thread:terminal-1"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/session/note", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Layer-Actor", "gemini")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var result runtime.SessionNoteResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode note result: %v", err)
	}
	if result.Record.Actor != "gemini" || result.Note.Kind != "todo" {
		t.Fatalf("unexpected session note result: %+v", result)
	}
	if result.Observation.Topic != "continuity.note" {
		t.Fatalf("expected continuity observation, got %+v", result.Observation)
	}
	bootstrapReq := httptest.NewRequest(http.MethodGet, "/api/layer-os/session/bootstrap", nil)
	bootstrapRec := httptest.NewRecorder()
	router.ServeHTTP(bootstrapRec, bootstrapReq)
	var response struct {
		Session runtime.SessionBootstrapPacket `json:"session"`
	}
	if err := json.NewDecoder(bootstrapRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if response.Session.Continuity == nil || response.Session.Continuity.Record == nil || len(response.Session.Continuity.Record.Notes) != 1 {
		t.Fatalf("expected continuity note in bootstrap, got %+v", response.Session.Continuity)
	}
}

func TestSessionFinishRouteUpdatesMemoryAndEvent(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	raw := []byte(`{"current_focus":"Lock queue","current_goal":"close loop","next_steps":["sync"],"open_risks":["drift"],"handoff_note":"continue","note":"closed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/session/finish", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var result runtime.SessionFinishResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode finish result: %v", err)
	}
	if result.Memory.CurrentFocus != "Lock queue" || result.Event.Kind != "session.finished" {
		t.Fatalf("unexpected finish result: %+v", result)
	}
	items := service.ListObservations(runtime.ObservationQuery{SourceChannel: "session"})
	if len(items) != 1 || items[0].Topic != "session.finished" {
		t.Fatalf("expected auto session observation, got %+v", items)
	}
}

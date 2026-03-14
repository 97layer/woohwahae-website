package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"layer-os/internal/api"
	"layer-os/internal/runtime"
)

func TestFounderDailyLoopPersistsAcrossNextBootstrap(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{
		Text:     "Investigate founder backlog drift.",
		Kind:     "risk",
		Severity: "high",
		Source:   "founder.loop.test",
	}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{
		CurrentFocus: "Inspect founder queue",
		NextSteps:    []string{"review backlog"},
		OpenRisks:    []string{"drift"},
		UpdatedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed memory: %v", err)
	}

	client := daemonClientForService(service)

	bootstrap, err := client.SessionBootstrap(true)
	if err != nil {
		t.Fatalf("initial bootstrap: %v", err)
	}
	if bootstrap.Knowledge.CurrentFocus != "Inspect founder queue" {
		t.Fatalf("unexpected initial focus: %+v", bootstrap.Knowledge)
	}
	if bootstrap.ReviewRoom == nil || bootstrap.ReviewRoom.OpenCount != 1 {
		t.Fatalf("expected initial review room summary, got %+v", bootstrap.ReviewRoom)
	}
	if bootstrap.Handoff == nil || bootstrap.Handoff.SystemMemory.CurrentFocus != "Inspect founder queue" {
		t.Fatalf("expected initial handoff focus, got %+v", bootstrap.Handoff)
	}

	goal := "close founder daily loop"
	handoff := "resume with verification evidence"
	finish, err := client.SessionFinish(runtime.SessionFinishInput{
		CurrentFocus: "Close founder loop",
		CurrentGoal:  &goal,
		NextSteps:    []string{"verify backlog", "prepare next session"},
		OpenRisks:    []string{"approval drift"},
		HandoffNote:  &handoff,
	})
	if err != nil {
		t.Fatalf("session finish: %v", err)
	}
	if finish.Event.Kind != "session.finished" || finish.Memory.CurrentFocus != "Close founder loop" {
		t.Fatalf("unexpected finish result: %+v", finish)
	}

	reloaded, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	client = daemonClientForService(reloaded)

	nextBootstrap, err := client.SessionBootstrap(true)
	if err != nil {
		t.Fatalf("next bootstrap: %v", err)
	}
	if nextBootstrap.Knowledge.CurrentFocus != "Close founder loop" {
		t.Fatalf("expected persisted focus, got %+v", nextBootstrap.Knowledge)
	}
	if len(nextBootstrap.Knowledge.NextSteps) != 2 || nextBootstrap.Knowledge.NextSteps[0] != "verify backlog" {
		t.Fatalf("expected persisted next steps, got %+v", nextBootstrap.Knowledge.NextSteps)
	}
	if len(nextBootstrap.Knowledge.OpenRisks) != 1 || nextBootstrap.Knowledge.OpenRisks[0] != "approval drift" {
		t.Fatalf("expected persisted risks, got %+v", nextBootstrap.Knowledge.OpenRisks)
	}
	if nextBootstrap.Handoff == nil || nextBootstrap.Handoff.SystemMemory.HandoffNote == nil || *nextBootstrap.Handoff.SystemMemory.HandoffNote != "resume with verification evidence" {
		t.Fatalf("expected persisted handoff note, got %+v", nextBootstrap.Handoff)
	}
	if nextBootstrap.ReviewRoom == nil || nextBootstrap.ReviewRoom.OpenCount == 0 {
		t.Fatalf("expected open review room continuity, got %+v", nextBootstrap.ReviewRoom)
	}

	observations := reloaded.ListObservations(runtime.ObservationQuery{SourceChannel: "session", Topic: "session.finished"})
	if len(observations) == 0 {
		t.Fatalf("expected session.finished observation continuity, got %+v", observations)
	}
}

func TestFounderBootstrapSurfacesReviewRoomIntegrityDrift(t *testing.T) {
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{
		Text:     "Original founder integrity item.",
		Kind:     "risk",
		Severity: "high",
		Source:   "founder.integrity.test",
	}); err != nil {
		t.Fatalf("seed review room: %v", err)
	}

	tampered := service.ReviewRoom()
	tampered.Open = []runtime.ReviewRoomItem{}
	tampered.Accepted = []runtime.ReviewRoomItem{}
	raw, err := json.MarshalIndent(tampered, "", "  ")
	if err != nil {
		t.Fatalf("marshal tampered review room: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "review_room.json"), raw, 0o644); err != nil {
		t.Fatalf("tamper review room file: %v", err)
	}

	reloaded, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	client := daemonClientForService(reloaded)
	bootstrap, err := client.SessionBootstrap(true)
	if err != nil {
		t.Fatalf("bootstrap after integrity drift: %v", err)
	}
	if bootstrap.ReviewRoom == nil || bootstrap.ReviewRoom.OpenCount == 0 {
		t.Fatalf("expected review room integrity drift in bootstrap, got %+v", bootstrap.ReviewRoom)
	}
	if bootstrap.ReviewRoom.TopOpen[0].Source != "system.review_room_integrity" {
		t.Fatalf("expected integrity source, got %+v", bootstrap.ReviewRoom.TopOpen[0])
	}
	if bootstrap.Handoff == nil || len(bootstrap.Handoff.ReviewRoom.Issues) == 0 {
		t.Fatalf("expected handoff review-room issues after integrity drift, got %+v", bootstrap.Handoff)
	}
}

func daemonClientForService(service *runtime.Service) *daemonClient {
	router := api.NewRouterWithRuntime(service, api.DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)})
	return &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, r)
			return rec.Result(), nil
		})},
	}
}

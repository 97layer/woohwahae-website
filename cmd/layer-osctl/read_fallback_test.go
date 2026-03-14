package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

func TestEnvBoolRecognizesTruthyValues(t *testing.T) {
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "yes")
	if !envBool("LAYER_OS_ALLOW_LOCAL_FALLBACK") {
		t.Fatal("expected envBool to recognize truthy value")
	}
}

func TestStatusResultFallsBackWhenEnabled(t *testing.T) {
	seedLocalReadFallbackRuntime(t)
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "1")

	state, usedFallback, err := statusResult(unavailableReadClient(), false)
	if err != nil {
		t.Fatalf("status fallback: %v", err)
	}
	if !usedFallback {
		t.Fatal("expected status to use local fallback")
	}
	if state.MemoryHealth != "ready" || state.PrimarySurface != runtime.SurfaceCockpit {
		t.Fatalf("unexpected fallback status: %+v", state)
	}
}

func TestHandoffResultFallsBackWhenEnabled(t *testing.T) {
	seedLocalReadFallbackRuntime(t)
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "1")

	item, usedFallback, err := handoffResult(unavailableReadClient(), false)
	if err != nil {
		t.Fatalf("handoff fallback: %v", err)
	}
	if !usedFallback {
		t.Fatal("expected handoff to use local fallback")
	}
	if item.SystemMemory.CurrentFocus != "Recover queue" {
		t.Fatalf("unexpected fallback handoff memory: %+v", item.SystemMemory)
	}
	if item.FounderSummary.PrimaryAction != "review_room" || item.ReviewRoom.OpenCount < 1 {
		t.Fatalf("unexpected fallback handoff: %+v", item)
	}
}

func TestReviewRoomReadSnapshotFallsBackWhenEnabled(t *testing.T) {
	seedLocalReadFallbackRuntime(t)
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "1")

	result, usedFallback, err := reviewRoomReadSnapshot(unavailableReadClient(), false)
	if err != nil {
		t.Fatalf("review-room fallback: %v", err)
	}
	if !usedFallback {
		t.Fatal("expected review-room to use local fallback")
	}
	if result.summary.OpenCount < 1 || len(result.room.Open) < 1 {
		t.Fatalf("unexpected fallback review-room: %+v / %+v", result.summary, result.room)
	}
	if result.room.Open[0].Text != "Carry-over review is required before the next lane." {
		t.Fatalf("expected seeded review-room item first, got %+v", result.room.Open[0])
	}
}

func TestFounderSummaryResultFallsBackWhenEnabled(t *testing.T) {
	seedLocalReadFallbackRuntime(t)
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "1")

	summary, usedFallback, err := founderSummaryResult(unavailableReadClient(), false)
	if err != nil {
		t.Fatalf("founder summary fallback: %v", err)
	}
	if !usedFallback {
		t.Fatal("expected founder summary to use local fallback")
	}
	if summary.ReviewOpenCount < 1 || summary.PrimaryAction != "review_room" {
		t.Fatalf("unexpected fallback founder summary: %+v", summary)
	}
	if len(summary.ReviewTopOpen) == 0 || summary.ReviewTopOpen[0] != "Carry-over review is required before the next lane." {
		t.Fatalf("expected seeded founder top open item, got %+v", summary.ReviewTopOpen)
	}
}

func TestStatusResultHintsWhenFallbackDisabled(t *testing.T) {
	_, _, err := statusResult(unavailableReadClient(), false)
	if err == nil {
		t.Fatal("expected daemon unavailable error")
	}
	if !strings.Contains(err.Error(), "--allow-local-fallback") {
		t.Fatalf("expected fallback hint, got %v", err)
	}
}

func TestHandoffResultHintsWhenFallbackDisabled(t *testing.T) {
	_, _, err := handoffResult(unavailableReadClient(), false)
	if err == nil {
		t.Fatal("expected daemon unavailable error")
	}
	if !strings.Contains(err.Error(), "--allow-local-fallback") {
		t.Fatalf("expected fallback hint, got %v", err)
	}
}

func TestStatusFallbackDoesNotMutateRuntimeDataDir(t *testing.T) {
	seedLocalReadFallbackRuntime(t)
	t.Setenv("LAYER_OS_ALLOW_LOCAL_FALLBACK", "1")
	dataDir := localRuntimeDataDir()
	reviewRoomPath := filepath.Join(dataDir, "review_room.json")

	room := map[string]any{}
	raw, err := os.ReadFile(reviewRoomPath)
	if err != nil {
		t.Fatalf("read review room: %v", err)
	}
	if err := json.Unmarshal(raw, &room); err != nil {
		t.Fatalf("unmarshal review room: %v", err)
	}
	room["source"] = "/tmp/legacy/review_room.json"
	updated, err := json.MarshalIndent(room, "", "  ")
	if err != nil {
		t.Fatalf("marshal review room: %v", err)
	}
	if err := os.WriteFile(reviewRoomPath, updated, 0o644); err != nil {
		t.Fatalf("seed legacy source: %v", err)
	}

	_, usedFallback, err := statusResult(unavailableReadClient(), false)
	if err != nil {
		t.Fatalf("status fallback: %v", err)
	}
	if !usedFallback {
		t.Fatal("expected status fallback path")
	}

	rawAfter, err := os.ReadFile(reviewRoomPath)
	if err != nil {
		t.Fatalf("read review room after fallback: %v", err)
	}
	var roomAfter map[string]any
	if err := json.Unmarshal(rawAfter, &roomAfter); err != nil {
		t.Fatalf("unmarshal review room after fallback: %v", err)
	}
	if roomAfter["source"] != "/tmp/legacy/review_room.json" {
		t.Fatalf("expected real runtime file to stay untouched during read-only fallback, got source=%v", roomAfter["source"])
	}
}

func seedLocalReadFallbackRuntime(t *testing.T) {
	t.Helper()
	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	now := time.Now().UTC()
	if err := service.ReplaceMemory(runtime.SystemMemory{
		CurrentFocus: "Recover queue",
		NextSteps:    []string{"resume"},
		OpenRisks:    []string{"daemon_down"},
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if _, err := service.AddStructuredReviewRoomItem("open", runtime.ReviewRoomItem{
		Text:      "Carry-over review is required before the next lane.",
		Kind:      "runtime_signal",
		Severity:  "high",
		Source:    "session.finished",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("add review-room item: %v", err)
	}

	repoRoot, err := localReadFallbackRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	t.Setenv("LAYER_OS_DATA_DIR", dataDir)
}

func unavailableReadClient() *daemonClient {
	return &daemonClient{
		baseURL: "http://127.0.0.1:1",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, &url.Error{Op: r.Method, URL: r.URL.String(), Err: errors.New("operation not permitted")}
		})},
	}
}

func localReadFallbackRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(wd, "..", "..")), nil
}

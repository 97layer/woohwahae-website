package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEventRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	work := WorkItem{
		ID:               "work_001",
		Title:            "Initialize runtime",
		Intent:           "bootstrap",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_001",
		CreatedAt:        time.Now().UTC(),
	}
	if err := service.CreateWorkItem(work); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListEvents()); got != 1 {
		t.Fatalf("expected 1 reloaded event, got %d", got)
	}
	if reloaded.Handoff().Counts.Events != 1 {
		t.Fatalf("expected handoff event count 1, got %d", reloaded.Handoff().Counts.Events)
	}
}

func TestEventActorUsesConfiguredActor(t *testing.T) {
	oldActor := os.Getenv("LAYER_OS_ACTOR")
	defer os.Setenv("LAYER_OS_ACTOR", oldActor)
	os.Setenv("LAYER_OS_ACTOR", "claude")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.CreateWorkItem(WorkItem{
		ID:               "work_actor_001",
		Title:            "Actor test",
		Intent:           "verify actor",
		Stage:            StageDiscover,
		Surface:          SurfaceAPI,
		Pack:             "core",
		Priority:         "high",
		Risk:             "low",
		RequiresApproval: false,
		Payload:          map[string]any{"source": "test"},
		CorrelationID:    "corr_actor_001",
		CreatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	events := service.ListEvents()
	if len(events) == 0 {
		t.Fatal("expected events to be recorded")
	}
	if events[len(events)-1].Actor != "claude" {
		t.Fatalf("expected actor claude, got %q", events[len(events)-1].Actor)
	}
}

func TestEventRetentionKeepsEvidence(t *testing.T) {
	oldRetention := os.Getenv("LAYER_OS_EVENT_RETENTION")
	defer os.Setenv("LAYER_OS_EVENT_RETENTION", oldRetention)
	os.Setenv("LAYER_OS_EVENT_RETENTION", "2")

	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("work_retention_%03d", i)
		if err := service.CreateWorkItem(WorkItem{
			ID:               id,
			Title:            "Retention test",
			Intent:           "retain evidence",
			Stage:            StageDiscover,
			Surface:          SurfaceAPI,
			Pack:             "core",
			Priority:         "high",
			Risk:             "low",
			RequiresApproval: false,
			Payload:          map[string]any{"source": "test"},
			CorrelationID:    id,
			CreatedAt:        time.Now().UTC(),
		}); err != nil {
			t.Fatalf("create work item %d: %v", i, err)
		}
	}

	if got := len(service.ListEvents()); got != 2 {
		t.Fatalf("expected 2 retained recent events, got %d", got)
	}
	if got := service.Handoff().Counts.Events; got != 3 {
		t.Fatalf("expected total event count 3, got %d", got)
	}
	if got := len(service.Snapshot().Events); got != 3 {
		t.Fatalf("expected snapshot evidence count 3, got %d", got)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if got := len(reloaded.ListEvents()); got != 2 {
		t.Fatalf("expected 2 retained recent events after reload, got %d", got)
	}
	if got := reloaded.Handoff().Counts.Events; got != 3 {
		t.Fatalf("expected total event count 3 after reload, got %d", got)
	}
	if got := len(reloaded.Snapshot().Events); got != 3 {
		t.Fatalf("expected snapshot evidence count 3 after reload, got %d", got)
	}
	archiveRaw, err := os.ReadFile(dataDir + "/events_archive.json")
	if err != nil {
		t.Fatalf("read event archive: %v", err)
	}
	trimmed := strings.TrimSpace(string(archiveRaw))
	if strings.HasPrefix(trimmed, "[") {
		t.Fatalf("expected streamed archive format, got legacy json array: %s", trimmed)
	}
}

func TestEventRejectsUnsupportedDataValue(t *testing.T) {
	err := EventEnvelope{
		EventID:    "event_invalid_data",
		Kind:       "test.invalid",
		Actor:      "codex",
		Surface:    SurfaceAPI,
		WorkItemID: "work_001",
		Stage:      StageDiscover,
		Timestamp:  time.Now().UTC(),
		Data:       map[string]any{"invalid": invalidJSONPayload{Message: "no structs"}},
	}.Validate()
	if err == nil || !strings.Contains(err.Error(), "event data") {
		t.Fatalf("expected event data validation error, got %v", err)
	}
}

func TestResolveActorDefaultsToNeutralSystem(t *testing.T) {
	oldActor := os.Getenv("LAYER_OS_ACTOR")
	oldWriter := os.Getenv("LAYER_OS_WRITER_ID")
	oldDefault := os.Getenv("LAYER_OS_DEFAULT_ACTOR")
	defer os.Setenv("LAYER_OS_ACTOR", oldActor)
	defer os.Setenv("LAYER_OS_WRITER_ID", oldWriter)
	defer os.Setenv("LAYER_OS_DEFAULT_ACTOR", oldDefault)
	os.Unsetenv("LAYER_OS_ACTOR")
	os.Unsetenv("LAYER_OS_WRITER_ID")
	os.Unsetenv("LAYER_OS_DEFAULT_ACTOR")

	if got := ResolveActor(os.Getenv("LAYER_OS_ACTOR"), os.Getenv("LAYER_OS_WRITER_ID")); got != "system" {
		t.Fatalf("expected neutral default actor system, got %q", got)
	}
}

func TestEventArchiveSkipsMalformedNDJSONLines(t *testing.T) {
	dataDir := t.TempDir()
	disk, err := newDiskStore(dataDir)
	if err != nil {
		t.Fatalf("new disk store: %v", err)
	}

	first := EventEnvelope{
		EventID:    "event_first",
		Kind:       "test.event",
		Actor:      "system",
		Surface:    SurfaceAPI,
		Stage:      StageDiscover,
		Timestamp:  time.Now().UTC(),
		WorkItemID: "work_first",
	}
	second := EventEnvelope{
		EventID:    "event_second",
		Kind:       "test.event",
		Actor:      "system",
		Surface:    SurfaceAPI,
		Stage:      StageDiscover,
		Timestamp:  time.Now().UTC(),
		WorkItemID: "work_second",
	}

	firstRaw, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first event: %v", err)
	}
	secondRaw, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second event: %v", err)
	}

	body := strings.Join([]string{
		string(firstRaw),
		"{not valid json",
		string(secondRaw),
		"",
	}, "\n")

	if err := os.WriteFile(filepath.Join(dataDir, "events_archive.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	items, err := disk.readEventArchive()
	if err != nil {
		t.Fatalf("read event archive: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 valid events after skipping malformed line, got %d", len(items))
	}
	if items[0].EventID != first.EventID || items[1].EventID != second.EventID {
		t.Fatalf("expected valid events to survive malformed line, got %+v", items)
	}

	count, err := disk.countEventArchive()
	if err != nil {
		t.Fatalf("count event archive: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected archive count 2 after skipping malformed line, got %d", count)
	}
}

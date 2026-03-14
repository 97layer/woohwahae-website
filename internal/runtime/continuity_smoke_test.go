package runtime

import (
	"testing"
	"time"
)

func TestContinuityCheckpointRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	note := "capture queue risk"
	checkpoint := SessionCheckpoint{
		CheckpointID: "session_checkpoint:123",
		Source:       "terminal",
		Actor:        "system",
		CurrentFocus: "Stabilize queue",
		NextSteps:    []string{"cutover"},
		OpenRisks:    []string{"drift"},
		HandoffNote:  strPtr("handoff ref"),
		Note:         &note,
		Refs:         []string{"thread:queue"},
		UpdatedAt:    now,
	}

	record := continuityRecordFromSessionCheckpoint(checkpoint)
	if err := record.Validate(); err != nil {
		t.Fatalf("continuity record should validate: %v", err)
	}
	if record.Status != continuityStatusActive {
		t.Fatalf("expected active status, got %s", record.Status)
	}
	if len(record.Notes) != 1 || record.Notes[0].Text != note {
		t.Fatalf("expected note copy, got %+v", record.Notes)
	}

	roundTrip := sessionCheckpointFromContinuityRecord(record)
	if roundTrip.CurrentFocus != checkpoint.CurrentFocus {
		t.Fatalf("focus mismatch: %q vs %q", roundTrip.CurrentFocus, checkpoint.CurrentFocus)
	}
	if roundTrip.Note == nil || *roundTrip.Note != note {
		t.Fatalf("note mismatch: %+v", roundTrip.Note)
	}
}

func TestContinuityStatusStaleWhenOutdated(t *testing.T) {
	now := time.Now().UTC()
	record := ContinuityRecord{
		RecordID:     "continuity_record:stale",
		Source:       "terminal",
		Actor:        "system",
		CurrentFocus: "Aging focus",
		NextSteps:    []string{},
		OpenRisks:    []string{},
		Refs:         []string{},
		Notes:        []ContinuityNote{},
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now.Add(-48 * time.Hour),
	}
	record = normalizeContinuityRecordAt(record, now)
	if record.Status != continuityStatusStale {
		t.Fatalf("expected stale status, got %s", record.Status)
	}
}

func strPtr(value string) *string {
	return &value
}

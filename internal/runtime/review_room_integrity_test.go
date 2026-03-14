package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewRoomSealDetectsDirectRewrite(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Original agenda item."); err != nil {
		t.Fatalf("seed review room: %v", err)
	}
	room, err := service.TransitionStructuredReviewRoomItem("accept", "Original agenda item.", &ReviewRoomResolution{Action: "accept", Reason: "seed", Rule: "test.seed"})
	if err != nil {
		t.Fatalf("accept review room item: %v", err)
	}
	room.Open = []ReviewRoomItem{}
	room.Accepted = []ReviewRoomItem{}
	raw, err := json.MarshalIndent(room, "", "  ")
	if err != nil {
		t.Fatalf("marshal tampered room: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, reviewRoomRuntimeFile), raw, 0o644); err != nil {
		t.Fatalf("tamper review room: %v", err)
	}

	audit := AuditReviewRoomIntegrity(dataDir)
	if len(audit.Issues) == 0 {
		t.Fatalf("expected integrity issues after direct rewrite, got %+v", audit)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	summary := reloaded.ReviewRoomSummary()
	if summary.OpenCount == 0 {
		t.Fatalf("expected integrity drift item in open review room, got %+v", summary)
	}
	if summary.TopOpen[0].Source != "system.review_room_integrity" {
		t.Fatalf("expected integrity source, got %+v", summary.TopOpen[0])
	}
}

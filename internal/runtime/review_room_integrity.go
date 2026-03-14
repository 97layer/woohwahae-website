package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
)

type ReviewRoomIntegrityAudit struct {
	Sealed       bool     `json:"sealed"`
	ExpectedHash string   `json:"expected_hash"`
	ActualHash   string   `json:"actual_hash"`
	Issues       []string `json:"issues"`
}

type reviewRoomSeal struct {
	Hash          string    `json:"hash"`
	AcceptedCount int       `json:"accepted_count"`
	OpenCount     int       `json:"open_count"`
	DeferredCount int       `json:"deferred_count"`
	GeneratedAt   time.Time `json:"generated_at"`
}

func AuditReviewRoomIntegrity(dataDir string) ReviewRoomIntegrityAudit {
	disk, err := newDiskStore(dataDir)
	if err != nil {
		return ReviewRoomIntegrityAudit{Issues: []string{"review room integrity audit failed: " + err.Error()}}
	}
	room, err := disk.loadReviewRoom()
	if err != nil {
		return ReviewRoomIntegrityAudit{Issues: []string{"review room load failed: " + err.Error()}}
	}
	seal, err := disk.loadReviewRoomSeal()
	if err != nil {
		return ReviewRoomIntegrityAudit{Issues: []string{"review room seal load failed: " + err.Error()}}
	}
	return verifyReviewRoomSeal(room, seal)
}

func verifyReviewRoomSeal(room ReviewRoom, seal reviewRoomSeal) ReviewRoomIntegrityAudit {
	expected := computeReviewRoomSeal(room)
	audit := ReviewRoomIntegrityAudit{
		Sealed:       strings.TrimSpace(seal.Hash) != "",
		ExpectedHash: expected.Hash,
		ActualHash:   strings.TrimSpace(seal.Hash),
		Issues:       []string{},
	}
	if !audit.Sealed {
		audit.Issues = append(audit.Issues, "review room seal missing: direct file drift cannot be verified")
		return audit
	}
	if audit.ActualHash != audit.ExpectedHash {
		audit.Issues = append(audit.Issues, "review room integrity drift: runtime file no longer matches daemon seal")
	}
	if seal.AcceptedCount != expected.AcceptedCount || seal.OpenCount != expected.OpenCount || seal.DeferredCount != expected.DeferredCount {
		audit.Issues = append(audit.Issues, "review room integrity drift: section counts no longer match daemon seal")
	}
	return audit
}

func computeReviewRoomSeal(room ReviewRoom) reviewRoomSeal {
	room = normalizeReviewRoom(room)
	raw, _ := json.Marshal(struct {
		Source    string           `json:"source"`
		Accepted  []ReviewRoomItem `json:"accepted"`
		Open      []ReviewRoomItem `json:"open"`
		Deferred  []ReviewRoomItem `json:"deferred"`
		Issues    []string         `json:"issues"`
		UpdatedAt *time.Time       `json:"updated_at,omitempty"`
	}{
		Source:    room.Source,
		Accepted:  room.Accepted,
		Open:      room.Open,
		Deferred:  room.Deferred,
		Issues:    room.Issues,
		UpdatedAt: room.UpdatedAt,
	})
	sum := sha256.Sum256(raw)
	return reviewRoomSeal{
		Hash:          hex.EncodeToString(sum[:]),
		AcceptedCount: len(room.Accepted),
		OpenCount:     len(room.Open),
		DeferredCount: len(room.Deferred),
		GeneratedAt:   zeroSafeNow(),
	}
}

func applyReviewRoomIntegrityGuardrail(room ReviewRoom, audit ReviewRoomIntegrityAudit) ReviewRoom {
	if len(audit.Issues) == 0 {
		return normalizeReviewRoom(room)
	}
	room = copyReviewRoom(room)
	for _, issue := range audit.Issues {
		room.Issues = appendUniqueString(room.Issues, issue)
	}
	item := newStructuredReviewRoomItem(
		"리뷰룸 무결성 드리프트가 감지됐어. 직접 파일을 덮어썼거나 오래된 상태가 daemon 전환을 우회했을 가능성이 커.",
		"review_room_integrity",
		"high",
		"system.review_room_integrity",
		nil,
		&ReviewRoomRationale{Trigger: "review_room.integrity", Reason: "review room file diverged from daemon-issued seal and may have been rewritten outside the canonical transition path", Rule: "review_room.integrity.guardrail"},
		[]string{"seal_expected:" + audit.ExpectedHash, "seal_actual:" + audit.ActualHash, "path:" + filepath.ToSlash(reviewRoomRuntimeFile)},
	)
	updated, _, err := ensureReviewRoomItem(room, "open", item)
	if err == nil {
		room = updated
	}
	return normalizeReviewRoom(room)
}

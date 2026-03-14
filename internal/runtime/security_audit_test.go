package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditSecurityFlagsWriteAuthAndCadenceGaps(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ref := "security_007"
	if _, err := service.AddStructuredReviewRoomItem("open", ReviewRoomItem{
		Text:     "[security_007] 운영 보안 posture 상향",
		Kind:     "agenda",
		Severity: "high",
		Source:   "test.security",
		Ref:      &ref,
	}); err != nil {
		t.Fatalf("seed security review-room item: %v", err)
	}

	audit := AuditSecurity(".", dataDir)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if audit.WriteAuthEnabled {
		t.Fatalf("expected write auth disabled, got %+v", audit)
	}
	if len(audit.OpenSecurityItems) != 1 {
		t.Fatalf("expected one open security item, got %+v", audit.OpenSecurityItems)
	}
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "write auth disabled") {
		t.Fatalf("expected write auth issue, got %+v", audit.Issues)
	}
	if !strings.Contains(joined, "security review cadence missing") {
		t.Fatalf("expected cadence issue, got %+v", audit.Issues)
	}
}

func TestAuditSecurityPassesWithFreshReview(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreatePreflight(PreflightRecord{
		RecordID:    "security_review_001",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: now,
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}

	audit := AuditSecurity(".", dataDir)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit, got %+v", audit)
	}
	if !audit.WriteAuthEnabled {
		t.Fatalf("expected write auth enabled, got %+v", audit)
	}
	if audit.LastReviewAt == nil || audit.LastReviewAt.IsZero() {
		t.Fatalf("expected last review timestamp, got %+v", audit)
	}
	if len(audit.Issues) != 0 {
		t.Fatalf("expected no issues, got %+v", audit.Issues)
	}
	if joined := strings.Join(audit.Checks, "\n"); !strings.Contains(joined, "auth trust root separated") && !strings.Contains(joined, "write auth enabled") {
		t.Fatalf("expected auth checks to remain populated, got %+v", audit.Checks)
	}
}

func TestAuditSecuritySummarizesNonEscalatedSecuritySignals(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreatePreflight(PreflightRecord{
		RecordID:    "security_review_001a",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: now,
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}
	if err := service.RecordSecuritySignal(SecuritySignalInput{
		Signal:   "write_auth_rejected",
		Summary:  "Write auth rejected for actor `gemini` from remote `127.0.0.1`.",
		Severity: "medium",
		Source:   "api.auth",
		Data: map[string]any{
			"actor":       "gemini",
			"remote_host": "127.0.0.1",
			"count":       1,
		},
		Evidence: []string{"actor:gemini", "remote:127.0.0.1", "count:1"},
	}); err != nil {
		t.Fatalf("record security signal: %v", err)
	}

	audit := AuditSecurity(".", dataDir)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit for non-escalated signal, got %+v", audit)
	}
	if audit.RecentSecuritySignalCount != 1 {
		t.Fatalf("expected one recent signal, got %+v", audit)
	}
	if audit.RecentEscalatedSignalCount != 0 {
		t.Fatalf("expected no escalated signals, got %+v", audit)
	}
	if len(audit.RecentSecuritySignals) != 1 || audit.RecentSecuritySignals[0].Kind != "security.write_auth_rejected" {
		t.Fatalf("expected recent write_auth_rejected signal, got %+v", audit.RecentSecuritySignals)
	}
}

func TestAuditSecurityFlagsRecentEscalatedSignals(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreatePreflight(PreflightRecord{
		RecordID:    "security_review_001b",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: now,
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}
	if err := service.RecordSecuritySignal(SecuritySignalInput{
		Signal:       "write_auth_bruteforce_detected",
		Summary:      "Repeated write-auth failures reached the throttle threshold for actor `gemini` from remote `127.0.0.1`.",
		Severity:     "high",
		Source:       "api.auth",
		Promote:      true,
		ReviewReason: "repeated write-auth failures suggest brute-force or misconfigured agent traffic that needs founder review before trusting the caller",
		ReviewRule:   "review_room.auto.security_write_auth_bruteforce",
		Data: map[string]any{
			"actor":       "gemini",
			"remote_host": "127.0.0.1",
			"count":       5,
		},
		Evidence: []string{"actor:gemini", "remote:127.0.0.1", "count:5"},
	}); err != nil {
		t.Fatalf("record escalated security signal: %v", err)
	}

	audit := AuditSecurity(".", dataDir)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit for escalated signal, got %+v", audit)
	}
	if audit.RecentSecuritySignalCount != 1 || audit.RecentEscalatedSignalCount != 1 {
		t.Fatalf("expected one escalated signal, got %+v", audit)
	}
	if audit.LastSecuritySignalAt == nil || audit.LastSecuritySignalAt.IsZero() {
		t.Fatalf("expected last security signal time, got %+v", audit)
	}
	joined := strings.Join(audit.Issues, "\n")
	if !strings.Contains(joined, "recent escalated security signals detected") {
		t.Fatalf("expected escalated signal issue, got %+v", audit.Issues)
	}
}

func TestAuditSecurityFlagsReleaseGateStaleness(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	reviewAt := time.Now().UTC().Add(-2 * time.Hour)
	if err := service.CreatePreflight(PreflightRecord{
		RecordID:    "security_review_002",
		Task:        "security review: release",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"review release surface"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: reviewAt,
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}
	resolvedAt := reviewAt.Add(-10 * time.Minute)
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Approve release gate",
		Risks:           []string{},
		RollbackPlan:    "rollback if needed",
		DecisionSurface: SurfaceCockpit,
		Status:          "approved",
		RequestedAt:     reviewAt.Add(-20 * time.Minute),
		ResolvedAt:      &resolvedAt,
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	releasedAt := reviewAt.Add(time.Hour)
	if err := service.CreateRelease(ReleasePacket{
		ReleaseID:    "release_001",
		WorkItemID:   "work_001",
		Target:       "vm",
		Channel:      "cockpit",
		Artifacts:    []string{},
		Metrics:      map[string]any{},
		RollbackPlan: "rollback if needed",
		ApprovalRefs: []string{"approval_001"},
		ReleasedAt:   &releasedAt,
	}); err != nil {
		t.Fatalf("create release: %v", err)
	}

	audit := AuditSecurity(".", dataDir)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if joined := strings.Join(audit.Issues, "\n"); !strings.Contains(joined, "security release gate stale") {
		t.Fatalf("expected release gate issue, got %+v", audit.Issues)
	}
}

func TestAuditSecurityFlagsPlaintextRuntimeSecret(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("security-secret"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	now := time.Now().UTC()
	if err := service.CreatePreflight(PreflightRecord{
		RecordID:    "security_review_003",
		Task:        "security review: weekly",
		Mode:        "security_review",
		Status:      "ready",
		Decision:    "go",
		ModelsUsed:  []string{},
		Steps:       []string{"run audit security"},
		Risks:       []string{},
		Checks:      []string{"write_auth_enabled", "secret_plaintext_surface_minimized", "edge_tls_required", "edge_access_control_required"},
		GeneratedAt: now,
	}); err != nil {
		t.Fatalf("create security preflight: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "leak.json"), []byte(`{"api_key":"sk-live-secret"}`), 0o644); err != nil {
		t.Fatalf("write leak file: %v", err)
	}
	audit := AuditSecurity(".", dataDir)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if len(audit.RuntimeSecretFindings) == 0 {
		t.Fatalf("expected runtime secret findings, got %+v", audit)
	}
	if joined := strings.Join(audit.Issues, "\n"); !strings.Contains(joined, "runtime secret plaintext findings detected") {
		t.Fatalf("expected plaintext secret issue, got %+v", audit.Issues)
	}
}

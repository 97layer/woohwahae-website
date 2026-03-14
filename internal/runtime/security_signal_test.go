package runtime

import "testing"

func TestRecordSecuritySignalPromotesReviewRoom(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.RecordSecuritySignal(SecuritySignalInput{
		Signal:       "write_auth_bruteforce_detected",
		Summary:      "Repeated write-auth failures reached the throttle threshold for actor `gemini` from `127.0.0.1`; inspect for hostile or misconfigured agent activity.",
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
		t.Fatalf("record security signal: %v", err)
	}

	events := service.ListEvents()
	if len(events) == 0 {
		t.Fatal("expected security event to be recorded")
	}
	last := events[len(events)-1]
	if last.Kind != "security.write_auth_bruteforce_detected" {
		t.Fatalf("expected security.write_auth_bruteforce_detected, got %q", last.Kind)
	}
	if got := last.Data["source"]; got != "api.auth" {
		t.Fatalf("expected api.auth source, got %#v", got)
	}
	if got := last.Data["review_promoted"]; got != true {
		t.Fatalf("expected review_promoted=true, got %#v", got)
	}

	room := service.ReviewRoom()
	if len(room.Open) != 1 {
		t.Fatalf("expected 1 open review item, got %+v", room.Open)
	}
	if room.Open[0].Ref == nil || *room.Open[0].Ref != "security_write_auth_bruteforce_detected" {
		t.Fatalf("expected security review ref, got %+v", room.Open[0].Ref)
	}

	audit := service.SecurityAudit()
	if len(audit.OpenSecurityItems) != 1 {
		t.Fatalf("expected security audit to see promoted item, got %+v", audit.OpenSecurityItems)
	}
}

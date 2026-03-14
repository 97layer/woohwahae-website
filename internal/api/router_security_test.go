package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"layer-os/internal/runtime"
)

func TestCockpitRouteSetsSecurityHeaders(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "http://layer-os.test/cockpit", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got == "" {
		t.Fatal("expected cockpit CSP header")
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected frame deny header, got %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("expected no-referrer policy, got %q", got)
	}
	if got := rec.Header().Get("Permissions-Policy"); got == "" {
		t.Fatal("expected permissions policy header")
	}
	if got := rec.Header().Get("Cross-Origin-Opener-Policy"); got != "same-origin" {
		t.Fatalf("expected same-origin opener policy, got %q", got)
	}
	if got := rec.Header().Get("Cross-Origin-Resource-Policy"); got != "same-origin" {
		t.Fatalf("expected same-origin resource policy, got %q", got)
	}
}

func TestStatusRouteSetsHSTSOnHTTPS(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	req := httptest.NewRequest(http.MethodGet, "https://layer-os.test/api/layer-os/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=63072000; includeSubDomains" {
		t.Fatalf("expected HSTS header, got %q", got)
	}
}

func TestCreateWorkItemRejectsCrossOriginWrite(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_cross_origin_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_cross_origin_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "http://layer-os.test/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-origin write, got %d", rec.Code)
	}
}

func TestCreateWorkItemAcceptsSameOriginWrite(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_same_origin_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_same_origin_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "http://layer-os.test/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Origin", "http://layer-os.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for same-origin write, got %d", rec.Code)
	}
}

func TestActorScopedWriteTokenUsesRequestActor(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.WithActor("gemini", func(s *runtime.Service) error {
		_, err := s.SetWriteToken("gemini-token")
		return err
	}); err != nil {
		t.Fatalf("set gemini token: %v", err)
	}
	if err := service.WithActor("codex", func(s *runtime.Service) error {
		_, err := s.SetWriteToken("codex-token")
		return err
	}); err != nil {
		t.Fatalf("set codex token: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_actor_scope_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_actor_scope_001",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	req.Header.Set("Authorization", "Bearer codex-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for mismatched actor token, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	req.Header.Set("Authorization", "Bearer gemini-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for matching actor token, got %d", rec.Code)
	}
}

func TestRepeatedInvalidBearerTokenTriggersSecurityReviewAndThrottle(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.SetWriteToken("secret-token"); err != nil {
		t.Fatalf("set write token: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{
		"id":                "work_bruteforce_001",
		"title":             "Seed runtime",
		"intent":            "initialize runtime",
		"stage":             "discover",
		"surface":           "api",
		"pack":              "runtime",
		"priority":          "high",
		"risk":              "medium",
		"requires_approval": true,
		"payload":           map[string]any{"source": "test"},
		"correlation_id":    "corr_bruteforce_001",
	}
	raw, _ := json.Marshal(body)

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "http://layer-os.test/api/layer-os/work-items", bytes.NewReader(raw))
		req.Header.Set("X-Layer-Actor", "gemini")
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "http://layer-os.test/api/layer-os/work-items", bytes.NewReader(raw))
	req.Header.Set("X-Layer-Actor", "gemini")
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on threshold attempt, got %d", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got == "" {
		t.Fatal("expected Retry-After header")
	}

	room := service.ReviewRoom()
	if len(room.Open) == 0 {
		t.Fatal("expected review-room security item")
	}
	foundReview := false
	for _, item := range room.Open {
		if item.Ref != nil && *item.Ref == "security_write_auth_bruteforce_detected" {
			foundReview = true
			break
		}
	}
	if !foundReview {
		t.Fatalf("expected security_write_auth_bruteforce_detected review item, got %+v", room.Open)
	}

	foundEvent := false
	for _, item := range service.ListEvents() {
		if item.Kind == "security.write_auth_bruteforce_detected" {
			foundEvent = true
			break
		}
	}
	if !foundEvent {
		t.Fatalf("expected security.write_auth_bruteforce_detected event, got %+v", service.ListEvents())
	}
}

func TestAuthBootstrapBlockedAttemptRecordsSecurityEvent(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)

	body := map[string]any{"token": "secret-token"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/layer-os/auth", bytes.NewReader(raw))
	req.RemoteAddr = "203.0.113.8:41001"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	foundEvent := false
	for _, item := range service.ListEvents() {
		if item.Kind == "security.auth_bootstrap_blocked" {
			foundEvent = true
			break
		}
	}
	if !foundEvent {
		t.Fatalf("expected security.auth_bootstrap_blocked event, got %+v", service.ListEvents())
	}
}

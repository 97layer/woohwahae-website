package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"layer-os/internal/runtime"
)

func TestReviewRoomAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/review-room", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGeminiAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/gemini", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthorityAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/authority", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var audit runtime.AuthorityAudit
	if err := json.Unmarshal(rec.Body.Bytes(), &audit); err != nil {
		t.Fatalf("decode authority audit: %v", err)
	}
	if audit.Status == "" {
		t.Fatalf("expected authority audit payload, got %+v", audit)
	}
}

func TestSecurityAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/security", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var audit runtime.SecurityAudit
	if err := json.Unmarshal(rec.Body.Bytes(), &audit); err != nil {
		t.Fatalf("decode security audit: %v", err)
	}
	if audit.Status == "" {
		t.Fatalf("expected security audit payload, got %+v", audit)
	}
}

func TestMCPAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/mcp", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var audit runtime.MCPAudit
	if err := json.Unmarshal(rec.Body.Bytes(), &audit); err != nil {
		t.Fatalf("decode mcp audit: %v", err)
	}
	if audit.Status == "" {
		t.Fatalf("expected mcp audit payload, got %+v", audit)
	}
}

func TestRuntimeDataAuditRoute(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	router := NewRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/layer-os/audit/runtime-data", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var audit runtime.RuntimeDataAudit
	if err := json.Unmarshal(rec.Body.Bytes(), &audit); err != nil {
		t.Fatalf("decode runtime-data audit: %v", err)
	}
	if audit.Status == "" {
		t.Fatalf("expected runtime-data audit payload, got %+v", audit)
	}
}

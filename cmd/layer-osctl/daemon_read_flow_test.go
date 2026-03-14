package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"layer-os/internal/api"
	"layer-os/internal/runtime"
)

func TestDaemonReadFlowStatusHealthzBootstrapAndFallback(t *testing.T) {
	oldAttempts := sessionBootstrapRetryAttempts
	oldSleep := sessionBootstrapSleep
	defer func() {
		sessionBootstrapRetryAttempts = oldAttempts
		sessionBootstrapSleep = oldSleep
	}()
	sessionBootstrapRetryAttempts = 2
	sessionBootstrapSleep = func(time.Duration) {}

	dataDir := t.TempDir()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(runtime.SystemMemory{CurrentFocus: "Recover daemon read lane", NextSteps: []string{"verify daemon"}, OpenRisks: []string{}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}

	startedAt := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	router := api.NewRouterWithRuntime(service, api.DaemonRuntimeInfo{Address: "127.0.0.1:17808", StartedAt: startedAt})
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, r)
			return rec.Result(), nil
		})},
	}

	daemonStatus, err := client.TryDaemonStatus()
	if err != nil {
		t.Fatalf("daemon status route: %v", err)
	}
	if daemonStatus.Service != "layer-osd" || daemonStatus.Address != "127.0.0.1:17808" {
		t.Fatalf("unexpected daemon status: %+v", daemonStatus)
	}

	var healthz runtime.DaemonStatus
	if err := client.request(http.MethodGet, "/healthz", nil, &healthz); err != nil {
		t.Fatalf("healthz request: %v", err)
	}
	if err := healthz.Validate(); err != nil {
		t.Fatalf("invalid healthz payload: %v", err)
	}

	packet, err := sessionBootstrapPacket(client, false, false)
	if err != nil {
		t.Fatalf("daemon bootstrap: %v", err)
	}
	if packet.Source != "daemon" || !packet.ReadOnly || packet.Degraded {
		t.Fatalf("unexpected daemon bootstrap packet: %+v", packet)
	}
	if packet.Knowledge.CurrentFocus != "Recover daemon read lane" {
		t.Fatalf("unexpected daemon bootstrap knowledge: %+v", packet.Knowledge)
	}

	repoRoot := t.TempDir()
	if err := os.Setenv("LAYER_OS_REPO_ROOT", repoRoot); err != nil {
		t.Fatalf("set repo root: %v", err)
	}
	defer os.Unsetenv("LAYER_OS_REPO_ROOT")
	if err := os.MkdirAll(localRuntimeDataDir(), 0o755); err != nil {
		t.Fatalf("mkdir local runtime dir: %v", err)
	}
	copyDir(t, dataDir, localRuntimeDataDir())

	unavailableClient := &daemonClient{
		baseURL: "http://127.0.0.1:1",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, &url.Error{Op: r.Method, URL: r.URL.String(), Err: errors.New("connection refused")}
		})},
	}
	fallbackPacket, err := sessionBootstrapPacket(unavailableClient, false, true)
	if err != nil {
		t.Fatalf("fallback bootstrap: %v", err)
	}
	if fallbackPacket.Source != "local_fallback" || !fallbackPacket.ReadOnly || !fallbackPacket.Degraded {
		t.Fatalf("unexpected fallback packet: %+v", fallbackPacket)
	}
	if fallbackPacket.Knowledge.CurrentFocus != "Recover daemon read lane" {
		t.Fatalf("unexpected fallback knowledge: %+v", fallbackPacket.Knowledge)
	}
}

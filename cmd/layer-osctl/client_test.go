package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestDaemonClientAddsBearerToken(t *testing.T) {
	oldToken := os.Getenv("LAYER_OS_WRITE_TOKEN")
	oldActor := os.Getenv("LAYER_OS_ACTOR")
	oldModels := os.Getenv("LAYER_OS_MODELS")
	defer os.Setenv("LAYER_OS_WRITE_TOKEN", oldToken)
	defer os.Setenv("LAYER_OS_ACTOR", oldActor)
	defer os.Setenv("LAYER_OS_MODELS", oldModels)
	os.Setenv("LAYER_OS_WRITE_TOKEN", "secret-token")
	os.Setenv("LAYER_OS_ACTOR", "claude")
	os.Setenv("LAYER_OS_MODELS", "gpt-5.4,claude-sonnet-4.5")

	client := newDaemonClient()
	request, err := client.newRequest(http.MethodPost, "/api/layer-os/work-items", runtime.WorkItem{ID: "work_001"})
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if got := request.Header.Get("Authorization"); got != "Bearer secret-token" {
		t.Fatalf("expected bearer token header, got %q", got)
	}
	if got := request.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected json content type, got %q", got)
	}
	if got := request.Header.Get("X-Layer-Actor"); got != "claude" {
		t.Fatalf("expected actor header, got %q", got)
	}
	if got := request.Header.Get("X-Layer-Models"); got != "gpt-5.4,claude-sonnet-4.5" {
		t.Fatalf("expected models header, got %q", got)
	}
}

func TestDaemonClientRequestReturnsAPIError(t *testing.T) {
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Status:     "401 Unauthorized",
				Body:       io.NopCloser(strings.NewReader(`{"error":"invalid bearer token"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	}

	err := client.request(http.MethodGet, "/api/layer-os/status", nil, &map[string]any{})
	if err == nil {
		t.Fatal("expected api error")
	}
	if !strings.Contains(err.Error(), "invalid bearer token") {
		t.Fatalf("expected bearer token error, got %v", err)
	}
}

func TestDaemonClientRequestWithPartialDecodesItem(t *testing.T) {
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Status:     "502 Bad Gateway",
				Body:       io.NopCloser(bytes.NewBufferString(`{"execute":{"execute_id":"execute_001","work_item_id":"work_001","policy_decision_id":"policy_001","status":"failed","notes":[]},"error":"command failed"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	}

	var item runtime.ExecuteRun
	err := client.requestWithPartial(http.MethodPost, "/api/layer-os/execute-runs/run", map[string]any{"execute_id": "execute_001"}, "execute", &item)
	if err == nil {
		t.Fatal("expected partial error")
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Fatalf("expected command failure, got %v", err)
	}
	if item.ExecuteID != "execute_001" {
		t.Fatalf("expected partial execute item, got %+v", item)
	}
}

func TestDaemonClientSearchKnowledgeCallsSearchRoute(t *testing.T) {
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if got := r.URL.Path; got != "/api/layer-os/knowledge/search" {
				t.Fatalf("unexpected path: %s", got)
			}
			if got := r.URL.Query().Get("q"); got != "queue drift" {
				t.Fatalf("unexpected query: %s", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"query":"queue drift","results":[{"entry":{"entry_id":"cap_event_001","source_kind":"session.finished","created_at":"2026-03-08T00:00:00Z","situation":{"summary":"Queue drift"},"decision":{},"result":{}},"match_fields":["situation"],"match_count":2}],"auto_threads":[]}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	}

	response := client.SearchKnowledge("queue drift")
	if response.Query != "queue drift" || len(response.Results) != 1 || response.Results[0].Entry.EntryID != "cap_event_001" {
		t.Fatalf("unexpected search response: %+v", response)
	}
}

func TestRequestActorDefaultsToNeutralSystem(t *testing.T) {
	oldActor := os.Getenv("LAYER_OS_ACTOR")
	oldWriter := os.Getenv("LAYER_OS_WRITER_ID")
	oldDefault := os.Getenv("LAYER_OS_DEFAULT_ACTOR")
	defer os.Setenv("LAYER_OS_ACTOR", oldActor)
	defer os.Setenv("LAYER_OS_WRITER_ID", oldWriter)
	defer os.Setenv("LAYER_OS_DEFAULT_ACTOR", oldDefault)
	os.Unsetenv("LAYER_OS_ACTOR")
	os.Unsetenv("LAYER_OS_WRITER_ID")
	os.Unsetenv("LAYER_OS_DEFAULT_ACTOR")

	if got := requestActor(); got != "system" {
		t.Fatalf("expected neutral default actor system, got %q", got)
	}
}

func TestDaemonRequestTimeoutUsesLongBudgetForDeployRoutes(t *testing.T) {
	if got := daemonRequestTimeout(http.MethodGet, "/api/layer-os/status"); got != 30*time.Second {
		t.Fatalf("expected default timeout for status route, got %s", got)
	}
	if got := daemonRequestTimeout(http.MethodPost, "/api/layer-os/deploys/execute"); got != 5*time.Minute {
		t.Fatalf("expected long timeout for deploy execute, got %s", got)
	}
	if got := daemonRequestTimeout(http.MethodPost, "/api/layer-os/founder-actions/release"); got != 5*time.Minute {
		t.Fatalf("expected long timeout for founder release, got %s", got)
	}
}

func TestDaemonRequestTimeoutHonorsEnvOverrides(t *testing.T) {
	oldDefault := os.Getenv("LAYER_OS_HTTP_TIMEOUT")
	oldLong := os.Getenv("LAYER_OS_LONG_HTTP_TIMEOUT")
	defer os.Setenv("LAYER_OS_HTTP_TIMEOUT", oldDefault)
	defer os.Setenv("LAYER_OS_LONG_HTTP_TIMEOUT", oldLong)

	os.Setenv("LAYER_OS_HTTP_TIMEOUT", "45s")
	os.Setenv("LAYER_OS_LONG_HTTP_TIMEOUT", "9m")

	if got := daemonRequestTimeout(http.MethodGet, "/api/layer-os/status"); got != 45*time.Second {
		t.Fatalf("expected default env override 45s, got %s", got)
	}
	if got := daemonRequestTimeout(http.MethodPost, "/api/layer-os/deploys/execute"); got != 9*time.Minute {
		t.Fatalf("expected long env override 9m, got %s", got)
	}
}

type failingReadCloser struct {
	chunks []string
	err    error
	index  int
}

func (r *failingReadCloser) Read(p []byte) (int, error) {
	if r.index >= len(r.chunks) {
		return 0, r.err
	}
	chunk := r.chunks[r.index]
	r.index++
	copy(p, chunk)
	return len(chunk), nil
}

func (r *failingReadCloser) Close() error { return nil }

func TestDaemonClientRequestRetriesGetOnDaemonUnavailable(t *testing.T) {
	oldAttempts := daemonReadRetryAttempts
	oldDelay := daemonReadRetryDelay
	oldSleep := daemonRetrySleep
	t.Cleanup(func() {
		daemonReadRetryAttempts = oldAttempts
		daemonReadRetryDelay = oldDelay
		daemonRetrySleep = oldSleep
	})

	daemonReadRetryAttempts = 3
	daemonReadRetryDelay = time.Millisecond
	sleepCalls := 0
	daemonRetrySleep = func(time.Duration) { sleepCalls++ }

	transportCalls := 0
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			transportCalls++
			if transportCalls < 3 {
				return nil, &url.Error{Op: r.Method, URL: r.URL.String(), Err: errors.New("connection reset")}
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	}

	var response map[string]any
	if err := client.request(http.MethodGet, "/api/layer-os/status", nil, &response); err != nil {
		t.Fatalf("expected GET retry success, got %v", err)
	}
	if transportCalls != 3 {
		t.Fatalf("expected 3 GET attempts, got %d", transportCalls)
	}
	if sleepCalls != 2 {
		t.Fatalf("expected 2 retry sleeps, got %d", sleepCalls)
	}
}

func TestDaemonClientRequestDoesNotRetryPostOnDaemonUnavailable(t *testing.T) {
	oldAttempts := daemonReadRetryAttempts
	oldDelay := daemonReadRetryDelay
	oldSleep := daemonRetrySleep
	t.Cleanup(func() {
		daemonReadRetryAttempts = oldAttempts
		daemonReadRetryDelay = oldDelay
		daemonRetrySleep = oldSleep
	})

	daemonReadRetryAttempts = 3
	daemonReadRetryDelay = time.Millisecond
	sleepCalls := 0
	daemonRetrySleep = func(time.Duration) { sleepCalls++ }

	transportCalls := 0
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			transportCalls++
			return nil, &url.Error{Op: r.Method, URL: r.URL.String(), Err: errors.New("connection reset")}
		})},
	}

	err := client.request(http.MethodPost, "/api/layer-os/events", map[string]any{"kind": "session.finished"}, nil)
	if err == nil {
		t.Fatal("expected POST failure")
	}
	if transportCalls != 1 {
		t.Fatalf("expected single POST attempt, got %d", transportCalls)
	}
	if sleepCalls != 0 {
		t.Fatalf("expected no retry sleep for POST, got %d", sleepCalls)
	}
}

func TestDaemonClientStreamEventsReconnectsAfterReadDrop(t *testing.T) {
	oldAttempts := daemonReadRetryAttempts
	oldDelay := daemonReadRetryDelay
	oldSleep := daemonRetrySleep
	oldReconnects := daemonStreamReconnectAttempts
	t.Cleanup(func() {
		daemonReadRetryAttempts = oldAttempts
		daemonReadRetryDelay = oldDelay
		daemonRetrySleep = oldSleep
		daemonStreamReconnectAttempts = oldReconnects
	})

	daemonReadRetryAttempts = 1
	daemonReadRetryDelay = time.Millisecond
	daemonStreamReconnectAttempts = 2
	sleepCalls := 0
	daemonRetrySleep = func(time.Duration) { sleepCalls++ }

	transportCalls := 0
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			transportCalls++
			if transportCalls == 1 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       &failingReadCloser{chunks: []string{"event: connected\n\n"}, err: io.ErrUnexpectedEOF},
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader("event: resumed\n\n")),
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			}, nil
		})},
	}

	var out bytes.Buffer
	if err := client.StreamEvents(&out); err != nil {
		t.Fatalf("expected stream reconnect success, got %v", err)
	}
	if transportCalls != 2 {
		t.Fatalf("expected 2 stream connections, got %d", transportCalls)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected 1 reconnect sleep, got %d", sleepCalls)
	}
	got := out.String()
	if !strings.Contains(got, "event: connected") || !strings.Contains(got, "event: resumed") {
		t.Fatalf("unexpected stream output: %q", got)
	}
}

func TestDaemonClientListAgentJobsByStatusBuildsQuery(t *testing.T) {
	var gotPath string
	client := &daemonClient{
		baseURL: "http://layer-osd.test",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotPath = r.URL.RequestURI()
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"items":[{"job_id":"job_001","kind":"plan","role":"planner","summary":"Plan founder lane","status":"queued","source":"manual","surface":"api","stage":"discover","notes":[],"created_at":"2026-03-08T00:00:00Z","updated_at":"2026-03-08T00:00:00Z"}]}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	}

	items := client.ListAgentJobsByStatus("open", 1)
	if gotPath != "/api/layer-os/jobs?limit=1&status=open" {
		t.Fatalf("expected jobs query path, got %q", gotPath)
	}
	if len(items) != 1 || items[0].JobID != "job_001" {
		t.Fatalf("unexpected jobs response: %+v", items)
	}
}

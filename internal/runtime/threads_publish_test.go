package runtime

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type threadsRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn threadsRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestThreadsStatusDefaultsDisabledWithoutToken(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	status := service.ThreadsStatus()
	if status.PublishConfigured {
		t.Fatalf("expected threads publish to stay disabled, got %+v", status)
	}
	if status.Adapter == "" || len(status.Notes) == 0 {
		t.Fatalf("expected threads status metadata, got %+v", status)
	}
}

func TestPublishThreadsBrandDraftPublishesApprovedDraftAndRecordsReceipt(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "threads-token")

	requests := []struct {
		Path string
		Form url.Values
	}{}
	originalClient := threadsHTTPClient
	threadsHTTPClient = func() *http.Client {
		return &http.Client{
			Transport: threadsRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				raw, _ := io.ReadAll(r.Body)
				form, _ := url.ParseQuery(string(raw))
				requests = append(requests, struct {
					Path string
					Form url.Values
				}{
					Path: r.URL.Path,
					Form: form,
				})
				body := `{"id":"creation_001"}`
				switch r.URL.Path {
				case "/me/threads":
					body = `{"id":"creation_001"}`
				case "/me/threads_publish":
					body = `{"id":"thread_001"}`
				default:
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       io.NopCloser(bytes.NewBufferString(`{"error":{"message":"not found"}}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}),
		}
	}
	defer func() {
		threadsHTTPClient = originalClient
	}()
	t.Setenv("THREADS_API_BASE_URL", "https://threads.test")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	resolvedAt := zeroSafeNow()
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Brand publish · Threads · Launch note",
		Risks:           []string{"review wording"},
		RollbackPlan:    "do not publish",
		DecisionSurface: SurfaceCockpit,
		Status:          "approved",
		RequestedAt:     resolvedAt.Add(-time.Minute),
		ResolvedAt:      &resolvedAt,
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "cockpit",
		Actor:             "admin",
		Topic:             threadsBrandPrepTopic,
		Refs:              []string{"proposal_001", "work_001", "approval_001", "flow_001", "brand_story"},
		Confidence:        "high",
		RawExcerpt:        "channel=threads\ntarget_account=97layer\ntitle=Launch note\nsources=brand_story\ntopic_tag=vibe coding\ndraft:\nWe rebuilt the operating surface and the first social lane is ready.",
		NormalizedSummary: "Threads draft ready for founder review.",
	}); err != nil {
		t.Fatalf("create prep observation: %v", err)
	}

	receipt, err := service.PublishThreadsBrandDraft("approval_001")
	if err != nil {
		t.Fatalf("publish threads draft: %v", err)
	}
	if receipt.ThreadID != "thread_001" || receipt.CreationID != "creation_001" {
		t.Fatalf("unexpected threads receipt: %+v", receipt)
	}
	if receipt.TopicTag != "vibe coding" {
		t.Fatalf("unexpected topic tag in receipt: %+v", receipt)
	}
	if receipt.TargetAccount != "97layer" {
		t.Fatalf("unexpected target account in receipt: %+v", receipt)
	}
	if len(requests) != 2 {
		t.Fatalf("expected two threads api calls, got %+v", requests)
	}
	if requests[0].Path != "/me/threads" || requests[0].Form.Get("media_type") != "TEXT" || !strings.Contains(requests[0].Form.Get("text"), "first social lane") || requests[0].Form.Get("topic_tag") != "vibe coding" {
		t.Fatalf("unexpected create request: %+v", requests[0])
	}
	if requests[1].Path != "/me/threads_publish" || requests[1].Form.Get("creation_id") != "creation_001" {
		t.Fatalf("unexpected publish request: %+v", requests[1])
	}

	items := service.ListObservations(ObservationQuery{Topic: threadsBrandPublishTopic, Ref: "approval_001", Limit: 1})
	if len(items) != 1 {
		t.Fatalf("expected one threads publish observation, got %+v", items)
	}
	if !strings.Contains(items[0].RawExcerpt, "thread_id=thread_001") {
		t.Fatalf("expected receipt observation to include thread id, got %+v", items[0])
	}
	if !strings.Contains(items[0].RawExcerpt, "topic_tag=vibe coding") {
		t.Fatalf("expected receipt observation to include topic tag, got %+v", items[0])
	}
	if !strings.Contains(items[0].RawExcerpt, "target_account=97layer") {
		t.Fatalf("expected receipt observation to include target account, got %+v", items[0])
	}
}

func TestPublishThreadsBrandDraftRejectsDuplicateApproval(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "threads-token")
	originalClient := threadsHTTPClient
	threadsHTTPClient = func() *http.Client {
		return &http.Client{
			Transport: threadsRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				body := `{"id":"creation_001"}`
				if strings.Contains(r.URL.Path, "threads_publish") {
					body = `{"id":"thread_001"}`
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}),
		}
	}
	defer func() {
		threadsHTTPClient = originalClient
	}()
	t.Setenv("THREADS_API_BASE_URL", "https://threads.test")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	resolvedAt := zeroSafeNow()
	if err := service.CreateApproval(ApprovalItem{
		ApprovalID:      "approval_001",
		WorkItemID:      "work_001",
		Stage:           StageVerify,
		Summary:         "Brand publish · Threads · Launch note",
		Risks:           []string{"review wording"},
		RollbackPlan:    "do not publish",
		DecisionSurface: SurfaceCockpit,
		Status:          "approved",
		RequestedAt:     resolvedAt.Add(-time.Minute),
		ResolvedAt:      &resolvedAt,
	}); err != nil {
		t.Fatalf("create approval: %v", err)
	}
	if _, err := service.CreateObservation(ObservationRecord{
		SourceChannel:     "cockpit",
		Actor:             "admin",
		Topic:             threadsBrandPrepTopic,
		Refs:              []string{"proposal_001", "work_001", "approval_001", "flow_001"},
		Confidence:        "high",
		RawExcerpt:        "channel=threads\ntarget_account=97layer\ntitle=Launch note\ndraft:\nHello Threads",
		NormalizedSummary: "Threads draft ready for founder review.",
	}); err != nil {
		t.Fatalf("create prep observation: %v", err)
	}

	if _, err := service.PublishThreadsBrandDraft("approval_001"); err != nil {
		t.Fatalf("first publish: %v", err)
	}
	if _, err := service.PublishThreadsBrandDraft("approval_001"); !errors.Is(err, ErrThreadsAlreadyPublished) {
		t.Fatalf("expected duplicate publish to be rejected, got %v", err)
	}
}

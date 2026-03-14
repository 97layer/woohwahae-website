package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	threadsBrandPrepTopic      = "brand_publish_prep"
	threadsBrandPublishTopic   = "brand_publish_threads"
	threadsDefaultReplyControl = "everyone"
)

var (
	ErrThreadsPublishNotConfigured = errors.New("threads publish not configured")
	ErrThreadsApprovalNotReady     = errors.New("threads approval is not approved")
	ErrThreadsDraftNotFound        = errors.New("approved threads draft not found")
	ErrThreadsAlreadyPublished     = errors.New("approved threads draft is already published")
)

type threadsPublisher interface {
	Name() string
	Enabled() bool
	PublishText(text string, replyControl string, topicTag string) (threadsPublishResult, error)
}

type threadsPublishResult struct {
	CreationID string
	ThreadID   string
}

type brandThreadsDraft struct {
	Channel       string
	TargetAccount string
	Title         string
	Body          string
	TopicTag      string
	SourceIDs     []string
	ProposalID    string
	WorkItemID    string
	ApprovalID    string
	FlowID        string
}

type threadsAPIAdapter struct {
	baseURL string
	token   string
	client  *http.Client
}

type noopThreadsAdapter struct{}

type threadsAPIResponse struct {
	ID    string         `json:"id"`
	Error map[string]any `json:"error"`
	Meta  map[string]any `json:"meta"`
}

var threadsHTTPClient = func() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func threadsAccessToken() string {
	return strings.TrimSpace(os.Getenv("THREADS_ACCESS_TOKEN"))
}

func threadsReplyControl() string {
	switch value := strings.TrimSpace(strings.ToLower(os.Getenv("THREADS_REPLY_CONTROL"))); value {
	case "", threadsDefaultReplyControl:
		return threadsDefaultReplyControl
	case "accounts_you_follow", "mentioned_only":
		return value
	default:
		return threadsDefaultReplyControl
	}
}

func threadsAPIBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("THREADS_API_BASE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "https://graph.threads.net/v1.0"
}

func threadsPublisherFromEnv() threadsPublisher {
	token := threadsAccessToken()
	if token == "" {
		return noopThreadsAdapter{}
	}
	return threadsAPIAdapter{
		baseURL: threadsAPIBaseURL(),
		token:   token,
		client:  threadsHTTPClient(),
	}
}

func (a threadsAPIAdapter) Name() string  { return "threads_api" }
func (a threadsAPIAdapter) Enabled() bool { return strings.TrimSpace(a.token) != "" }

func (a threadsAPIAdapter) PublishText(text string, replyControl string, topicTag string) (threadsPublishResult, error) {
	if !a.Enabled() {
		return threadsPublishResult{}, ErrThreadsPublishNotConfigured
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return threadsPublishResult{}, errors.New("threads text is required")
	}
	form := url.Values{
		"media_type":    []string{"TEXT"},
		"text":          []string{text},
		"reply_control": []string{threadsReplyControlValue(replyControl)},
		"access_token":  []string{a.token},
	}
	if trimmed := strings.TrimSpace(topicTag); trimmed != "" {
		form.Set("topic_tag", trimmed)
	}
	creationID, err := a.postForm("/me/threads", form)
	if err != nil {
		return threadsPublishResult{}, fmt.Errorf("threads create failed: %w", err)
	}
	threadID, err := a.postForm("/me/threads_publish", url.Values{
		"creation_id":  []string{creationID},
		"access_token": []string{a.token},
	})
	if err != nil {
		return threadsPublishResult{}, fmt.Errorf("threads publish failed: %w", err)
	}
	return threadsPublishResult{
		CreationID: creationID,
		ThreadID:   threadID,
	}, nil
}

func (a threadsAPIAdapter) postForm(path string, form url.Values) (string, error) {
	request, err := http.NewRequest(http.MethodPost, a.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := a.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	var payload threadsAPIResponse
	_ = json.Unmarshal(body, &payload)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", threadsAPIError(response.StatusCode, payload, body)
	}
	if payload.ID == "" {
		return "", fmt.Errorf("threads api returned no id")
	}
	return payload.ID, nil
}

func (a noopThreadsAdapter) Name() string  { return "noop" }
func (a noopThreadsAdapter) Enabled() bool { return false }

func (a noopThreadsAdapter) PublishText(_ string, _ string, _ string) (threadsPublishResult, error) {
	return threadsPublishResult{}, ErrThreadsPublishNotConfigured
}

func threadsReplyControlValue(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "accounts_you_follow", "mentioned_only":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return threadsDefaultReplyControl
	}
}

func threadsAPIError(statusCode int, payload threadsAPIResponse, body []byte) error {
	if message, ok := payload.Error["message"].(string); ok && strings.TrimSpace(message) != "" {
		return fmt.Errorf("threads api returned %d: %s", statusCode, strings.TrimSpace(message))
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return fmt.Errorf("threads api returned %d", statusCode)
	}
	return fmt.Errorf("threads api returned %d: %s", statusCode, text)
}

func (s *Service) ThreadsStatus() ThreadsStatus {
	adapter := threadsPublisherFromEnv()
	status := ThreadsStatus{
		Adapter:           adapter.Name(),
		PublishConfigured: adapter.Enabled(),
		Notes:             []string{},
	}
	if !status.PublishConfigured {
		status.Notes = append(status.Notes, "THREADS_ACCESS_TOKEN is missing, so live Threads publish stays off.")
		return status
	}
	status.Notes = append(status.Notes, "Approved Threads drafts can publish from the canonical daemon route.")
	return status
}

func (s *Service) PublishThreadsBrandDraft(approvalID string) (ThreadsPublishReceipt, error) {
	approvalID = strings.TrimSpace(approvalID)
	if approvalID == "" {
		return ThreadsPublishReceipt{}, errors.New("approval_id is required")
	}

	adapter := threadsPublisherFromEnv()
	if !adapter.Enabled() {
		return ThreadsPublishReceipt{}, ErrThreadsPublishNotConfigured
	}

	s.mu.Lock()
	approval, ok := s.approval.get(approvalID)
	s.mu.Unlock()
	if !ok {
		return ThreadsPublishReceipt{}, errors.New("approval_id not found")
	}
	if approval.Status != "approved" {
		return ThreadsPublishReceipt{}, ErrThreadsApprovalNotReady
	}

	if items := s.ListObservations(ObservationQuery{Topic: threadsBrandPublishTopic, Ref: approvalID, Limit: 1}); len(items) > 0 {
		return ThreadsPublishReceipt{}, ErrThreadsAlreadyPublished
	}

	prepItems := s.ListObservations(ObservationQuery{Topic: threadsBrandPrepTopic, Ref: approvalID, Limit: 1})
	if len(prepItems) == 0 {
		return ThreadsPublishReceipt{}, ErrThreadsDraftNotFound
	}
	draft, err := parseBrandThreadsDraft(prepItems[0])
	if err != nil {
		return ThreadsPublishReceipt{}, err
	}
	if draft.Channel != "threads" {
		return ThreadsPublishReceipt{}, errors.New("approved draft is not a threads lane item")
	}

	result, err := adapter.PublishText(draft.Body, threadsReplyControl(), draft.TopicTag)
	if err != nil {
		return ThreadsPublishReceipt{}, err
	}
	publishedAt := zeroSafeNow()
	observation, err := s.CreateObservation(ObservationRecord{
		SourceChannel: "threads",
		Actor:         "brand",
		Topic:         threadsBrandPublishTopic,
		Refs: mergeObservationRefs(
			approvalID,
			draft.ProposalID,
			draft.WorkItemID,
			draft.FlowID,
			result.CreationID,
			result.ThreadID,
			draft.SourceIDs,
		),
		Confidence:        "high",
		RawExcerpt:        buildThreadsPublishExcerpt(draft, result, publishedAt),
		NormalizedSummary: fmt.Sprintf("Published approved Threads draft: %s.", draft.Title),
		CapturedAt:        publishedAt,
	})
	if err != nil {
		return ThreadsPublishReceipt{}, err
	}

	return ThreadsPublishReceipt{
		ApprovalID:    approvalID,
		ProposalID:    draft.ProposalID,
		WorkItemID:    draft.WorkItemID,
		FlowID:        draft.FlowID,
		ObservationID: observation.ObservationID,
		CreationID:    result.CreationID,
		ThreadID:      result.ThreadID,
		TargetAccount: draft.TargetAccount,
		Title:         draft.Title,
		TopicTag:      draft.TopicTag,
		SourceIDs:     append([]string{}, draft.SourceIDs...),
		PublishedAt:   publishedAt,
	}, nil
}

func parseBrandThreadsDraft(item ObservationRecord) (brandThreadsDraft, error) {
	lines := strings.Split(strings.ReplaceAll(item.RawExcerpt, "\r\n", "\n"), "\n")
	draft := brandThreadsDraft{
		SourceIDs: []string{},
	}
	bodyStart := -1
	for index, raw := range lines {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "channel="):
			draft.Channel = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "channel=")))
		case strings.HasPrefix(line, "target_account="):
			draft.TargetAccount = strings.TrimSpace(strings.TrimPrefix(line, "target_account="))
		case strings.HasPrefix(line, "title="):
			draft.Title = strings.TrimSpace(strings.TrimPrefix(line, "title="))
		case strings.HasPrefix(line, "topic_tag="):
			draft.TopicTag = strings.TrimSpace(strings.TrimPrefix(line, "topic_tag="))
		case strings.HasPrefix(line, "sources="):
			draft.SourceIDs = splitCSV(strings.TrimSpace(strings.TrimPrefix(line, "sources=")))
		case line == "draft:":
			bodyStart = index + 1
		}
	}
	if bodyStart >= 0 && bodyStart <= len(lines) {
		draft.Body = strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	}
	for _, ref := range item.Refs {
		switch {
		case strings.HasPrefix(ref, "proposal_"):
			draft.ProposalID = ref
		case strings.HasPrefix(ref, "work_"):
			draft.WorkItemID = ref
		case strings.HasPrefix(ref, "approval_"):
			draft.ApprovalID = ref
		case strings.HasPrefix(ref, "flow_"):
			draft.FlowID = ref
		}
	}
	if draft.Channel == "" || draft.Title == "" || draft.Body == "" {
		return brandThreadsDraft{}, ErrThreadsDraftNotFound
	}
	return draft, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" || strings.EqualFold(strings.TrimSpace(value), "none") {
		return []string{}
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		items = append(items, item)
	}
	return items
}

func mergeObservationRefs(approvalID string, proposalID string, workItemID string, flowID string, creationID string, threadID string, sourceIDs []string) []string {
	refs := []string{}
	for _, value := range []string{
		strings.TrimSpace(approvalID),
		strings.TrimSpace(proposalID),
		strings.TrimSpace(workItemID),
		strings.TrimSpace(flowID),
	} {
		if value != "" {
			refs = appendUniqueString(refs, value)
		}
	}
	if trimmed := strings.TrimSpace(creationID); trimmed != "" {
		refs = appendUniqueString(refs, "threads_creation:"+trimmed)
	}
	if trimmed := strings.TrimSpace(threadID); trimmed != "" {
		refs = appendUniqueString(refs, "threads_thread:"+trimmed)
	}
	for _, sourceID := range sourceIDs {
		if trimmed := strings.TrimSpace(sourceID); trimmed != "" {
			refs = appendUniqueString(refs, trimmed)
		}
	}
	return refs
}

func buildThreadsPublishExcerpt(draft brandThreadsDraft, result threadsPublishResult, publishedAt time.Time) string {
	lines := []string{
		"channel=threads",
		"target_account=" + draft.TargetAccount,
		"title=" + draft.Title,
		"published_at=" + publishedAt.Format(time.RFC3339),
		"creation_id=" + result.CreationID,
		"thread_id=" + result.ThreadID,
		"sources=" + strings.Join(draft.SourceIDs, ","),
	}
	if trimmed := strings.TrimSpace(draft.TopicTag); trimmed != "" {
		lines = append(lines, "topic_tag="+trimmed)
	}
	lines = append(lines, "draft:", draft.Body)
	return strings.Join(lines, "\n")
}

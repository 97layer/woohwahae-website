package runtime

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const claudeMessagesEndpoint = "https://api.anthropic.com/v1/messages"
const claudeAPIVersion = "2023-06-01"

type claudeDirectAdapter struct {
	httpClient   *http.Client
	endpointBase string // override for testing
}

func (a claudeDirectAdapter) Name() string          { return "claude" }
func (a claudeDirectAdapter) Semantics() string     { return "direct_llm" }
func (a claudeDirectAdapter) DispatchEnabled() bool { return true }
func (a claudeDirectAdapter) RequiredMode() string  { return "single" }

func (a claudeDirectAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call = prepareRecordedGatewayCall(call, decision)
	call.Notes = appendUniqueString(call.Notes, "adapter:claude")
	call.Notes = appendUniqueString(call.Notes, "dispatch:direct_llm")
	return call, nil
}

func (a claudeDirectAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.DispatchWithPayload(call, decision, nil)
}

func (a claudeDirectAdapter) DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error) {
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if apiKey == "" {
		call.Status = "failed"
		errText := "no_anthropic_api_key"
		call.LastError = &errText
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=no_anthropic_api_key")
		return call, nil
	}

	model := claudeModelForCall(call)
	prompt := claudePromptFromPayload(payload, call)

	requestBody := map[string]any{
		"model":      model,
		"max_tokens": call.TokenBudget,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	}

	raw, err := json.Marshal(requestBody)
	if err != nil {
		return call, err
	}

	endpoint := claudeMessagesEndpoint
	if strings.TrimSpace(a.endpointBase) != "" {
		endpoint = a.endpointBase + "/v1/messages"
	}

	client := a.httpClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return call, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)
	req.Header.Set("X-Layer-Call-ID", call.CallID)

	now := zeroSafeNow()
	call.DispatchedAt = &now
	call.AttemptCount = 1

	resp, err := client.Do(req)
	if err != nil {
		errText := strings.TrimSpace(err.Error())
		if errText == "" {
			errText = "http_request_failed"
		}
		call.LastError = &errText
		call.Status = "failed"
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=http_request_failed")
		return call, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	statusCode := resp.StatusCode
	call.LastHTTPStatus = &statusCode
	call.Notes = appendUniqueString(call.Notes, "dispatch:http_status="+strconvItoa(statusCode))

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		errText := "http_status_" + strconvItoa(statusCode)
		call.LastError = &errText
		call.Status = "failed"
		return call, nil
	}

	text := claudeExtractText(body)
	if text != "" {
		preview := text
		if len(preview) > 8000 {
			preview = preview[:8000]
		}
		call.ResponsePreview = &preview
	}
	call.Status = "sent"
	call.LastError = nil
	return call, nil
}

func claudeModelForCall(call GatewayCall) string {
	model := strings.TrimSpace(call.Model)
	if strings.HasPrefix(model, "claude") {
		return model
	}
	switch call.RequestKind {
	case "plan", "implement", "design":
		return "claude-opus-4-6"
	case "verify":
		return "claude-haiku-4-5-20251001"
	default:
		return "claude-opus-4-6"
	}
}

func claudePromptFromPayload(payload map[string]any, call GatewayCall) string {
	return buildAgentPrompt(payload, call)
}

func claudeExtractText(body []byte) string {
	var resp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}
	for _, block := range resp.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return strings.TrimSpace(block.Text)
		}
	}
	return ""
}

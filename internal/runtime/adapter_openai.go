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

const openaiChatEndpoint = "https://api.openai.com/v1/chat/completions"

type openaiDirectAdapter struct {
	httpClient   *http.Client
	endpointBase string // override for testing
}

func (a openaiDirectAdapter) Name() string          { return "openai" }
func (a openaiDirectAdapter) Semantics() string     { return "direct_llm" }
func (a openaiDirectAdapter) DispatchEnabled() bool { return true }
func (a openaiDirectAdapter) RequiredMode() string  { return "single" }

func (a openaiDirectAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call = prepareRecordedGatewayCall(call, decision)
	call.Notes = appendUniqueString(call.Notes, "adapter:openai")
	call.Notes = appendUniqueString(call.Notes, "dispatch:direct_llm")
	return call, nil
}

func (a openaiDirectAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.DispatchWithPayload(call, decision, nil)
}

func (a openaiDirectAdapter) DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		call.Status = "failed"
		errText := "no_openai_api_key"
		call.LastError = &errText
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=no_openai_api_key")
		return call, nil
	}

	model := openaiModelForCall(call)
	prompt := openaiPromptFromPayload(payload, call)

	requestBody := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
		"max_tokens": call.TokenBudget,
	}

	raw, err := json.Marshal(requestBody)
	if err != nil {
		return call, err
	}

	endpoint := openaiChatEndpoint
	if strings.TrimSpace(a.endpointBase) != "" {
		endpoint = a.endpointBase + "/v1/chat/completions"
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
	req.Header.Set("Authorization", "Bearer "+apiKey)
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

	text := openaiExtractText(body)
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

func openaiModelForCall(call GatewayCall) string {
	model := strings.TrimSpace(call.Model)
	// remap non-openai model names to sensible defaults
	switch {
	case model == "" || strings.HasPrefix(model, "gemini") || strings.HasPrefix(model, "claude"):
		switch call.RequestKind {
		case "plan", "implement", "design":
			return "o3-mini"
		case "verify":
			return "gpt-4o"
		default:
			return "o3-mini"
		}
	}
	return model
}

func openaiPromptFromPayload(payload map[string]any, call GatewayCall) string {
	return buildAgentPrompt(payload, call)
}

func openaiExtractText(body []byte) string {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}
	if len(resp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

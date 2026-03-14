package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const geminiGenerateEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type geminiDirectAdapter struct {
	httpClient   *http.Client
	endpointBase string // override for testing; empty = use default
}

func (a geminiDirectAdapter) Name() string          { return "gemini" }
func (a geminiDirectAdapter) Semantics() string     { return "direct_llm" }
func (a geminiDirectAdapter) DispatchEnabled() bool { return true }
func (a geminiDirectAdapter) RequiredMode() string  { return "single" }

func (a geminiDirectAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	call = prepareRecordedGatewayCall(call, decision)
	call.Notes = appendUniqueString(call.Notes, "adapter:gemini")
	call.Notes = appendUniqueString(call.Notes, "dispatch:direct_llm")
	return call, nil
}

func (a geminiDirectAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.DispatchWithPayload(call, decision, nil)
}

func (a geminiDirectAdapter) DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error) {
	apiKey, ok := ProviderCredentialValue("gemini")
	if !ok || apiKey == "" {
		call.Status = "failed"
		errText := "no_google_api_key"
		call.LastError = &errText
		call.Notes = appendUniqueString(call.Notes, "dispatch:error=no_google_api_key")
		return call, nil
	}

	model := strings.TrimSpace(call.Model)
	if model == "" || !strings.HasPrefix(model, "gemini") {
		model = "gemini-2.0-flash"
	}

	prompt := geminiPromptFromPayload(payload, call)

	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": call.TokenBudget,
		},
	}

	raw, err := json.Marshal(requestBody)
	if err != nil {
		return call, err
	}

	base := geminiGenerateEndpoint
	if strings.TrimSpace(a.endpointBase) != "" {
		base = a.endpointBase + "/v1beta/models/%s:generateContent?key=%s"
	}
	endpoint := fmt.Sprintf(base, model, apiKey)
	client := a.httpClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return call, err
	}
	req.Header.Set("Content-Type", "application/json")
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

	text := geminiExtractText(body)
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

func geminiPromptFromPayload(payload map[string]any, call GatewayCall) string {
	return buildAgentPrompt(payload, call)
}

func geminiExtractText(body []byte) string {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}
	if len(resp.Candidates) == 0 {
		return ""
	}
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Candidates[0].Content.Parts[0].Text)
}

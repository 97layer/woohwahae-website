package runtime

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestOpenAIDirectAdapterMeta(t *testing.T) {
	a := openaiDirectAdapter{}
	if a.Name() != "openai" {
		t.Fatalf("expected name=openai, got %q", a.Name())
	}
	if !a.DispatchEnabled() {
		t.Fatal("expected DispatchEnabled=true")
	}
	if a.Semantics() != "direct_llm" {
		t.Fatalf("expected semantics=direct_llm, got %q", a.Semantics())
	}
}

func TestOpenAIDirectAdapterFailsWithoutAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	a := openaiDirectAdapter{}
	call := GatewayCall{CallID: "test_001", Provider: "openai", Model: "gpt-4o"}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "failed" {
		t.Fatalf("expected status=failed, got %q", result.Status)
	}
	if result.LastError == nil || *result.LastError != "no_openai_api_key" {
		t.Fatalf("expected LastError=no_openai_api_key, got %v", result.LastError)
	}
}

func TestOpenAIDirectAdapterDispatchesAndExtractsText(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "Layer OS OpenAI agent acknowledgement.",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("OPENAI_API_KEY", "test_key")

	a := openaiDirectAdapter{
		httpClient: client,
	}

	call := GatewayCall{
		CallID:      "test_002",
		Provider:    "openai",
		Model:       "gpt-4o",
		TokenBudget: 256,
	}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "sent" {
		t.Fatalf("expected status=sent, got %q (err=%v)", result.Status, result.LastError)
	}
	if result.ResponsePreview == nil || *result.ResponsePreview != "Layer OS OpenAI agent acknowledgement." {
		t.Fatalf("unexpected ResponsePreview: %v", result.ResponsePreview)
	}
}

func TestOpenAIModelRemapsNonOpenAIModels(t *testing.T) {
	cases := []struct {
		model string
		kind  string
		want  string
	}{
		{"gemini-2.0-flash", "plan", "o3-mini"},
		{"claude-opus-4-6", "verify", "gpt-4o"},
		{"claude-sonnet-4-6", "implement", "o3-mini"},
		{"gpt-4o", "plan", "gpt-4o"},
		{"", "implement", "o3-mini"},
	}
	for _, c := range cases {
		call := GatewayCall{Model: c.model, RequestKind: c.kind}
		got := openaiModelForCall(call)
		if got != c.want {
			t.Errorf("model=%q kind=%q: expected %q got %q", c.model, c.kind, c.want, got)
		}
	}
}

package runtime

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestClaudeDirectAdapterMeta(t *testing.T) {
	a := claudeDirectAdapter{}
	if a.Name() != "claude" {
		t.Fatalf("expected name=claude, got %q", a.Name())
	}
	if !a.DispatchEnabled() {
		t.Fatal("expected DispatchEnabled=true")
	}
	if a.Semantics() != "direct_llm" {
		t.Fatalf("expected semantics=direct_llm, got %q", a.Semantics())
	}
}

func TestClaudeDirectAdapterFailsWithoutAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	a := claudeDirectAdapter{}
	call := GatewayCall{CallID: "test_001", Provider: "claude", Model: "claude-sonnet-4-6"}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "failed" {
		t.Fatalf("expected status=failed, got %q", result.Status)
	}
	if result.LastError == nil || *result.LastError != "no_anthropic_api_key" {
		t.Fatalf("expected LastError=no_anthropic_api_key, got %v", result.LastError)
	}
}

func TestClaudeDirectAdapterDispatchesAndExtractsText(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		resp := map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Layer OS Claude agent acknowledgement.",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("ANTHROPIC_API_KEY", "test_key")

	a := claudeDirectAdapter{
		httpClient: client,
	}

	call := GatewayCall{
		CallID:      "test_002",
		Provider:    "claude",
		Model:       "claude-sonnet-4-6",
		TokenBudget: 256,
	}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "sent" {
		t.Fatalf("expected status=sent, got %q (err=%v)", result.Status, result.LastError)
	}
	if result.ResponsePreview == nil || *result.ResponsePreview != "Layer OS Claude agent acknowledgement." {
		t.Fatalf("unexpected ResponsePreview: %v", result.ResponsePreview)
	}
}

func TestClaudeModelRemapsNonClaudeModels(t *testing.T) {
	cases := []struct {
		model string
		kind  string
		want  string
	}{
		{"gemini-2.0-flash", "plan", "claude-opus-4-6"},
		{"gpt-4o", "verify", "claude-haiku-4-5-20251001"},
		{"gpt-4o-mini", "implement", "claude-opus-4-6"},
		{"claude-opus-4-6", "plan", "claude-opus-4-6"},
		{"", "implement", "claude-opus-4-6"},
	}
	for _, c := range cases {
		call := GatewayCall{Model: c.model, RequestKind: c.kind}
		got := claudeModelForCall(call)
		if got != c.want {
			t.Errorf("model=%q kind=%q: expected %q got %q", c.model, c.kind, c.want, got)
		}
	}
}

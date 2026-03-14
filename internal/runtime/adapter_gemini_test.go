package runtime

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGeminiDirectAdapterMeta(t *testing.T) {
	a := geminiDirectAdapter{}
	if a.Name() != "gemini" {
		t.Fatalf("expected name=gemini, got %q", a.Name())
	}
	if !a.DispatchEnabled() {
		t.Fatal("expected DispatchEnabled=true")
	}
	if a.Semantics() != "direct_llm" {
		t.Fatalf("expected semantics=direct_llm, got %q", a.Semantics())
	}
}

func TestGeminiDirectAdapterFailsWithoutAPIKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	a := geminiDirectAdapter{}
	call := GatewayCall{CallID: "test_001", Provider: "gemini", Model: "gemini-2.0-flash"}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "failed" {
		t.Fatalf("expected status=failed, got %q", result.Status)
	}
	if result.LastError == nil || *result.LastError != "no_google_api_key" {
		t.Fatalf("expected LastError=no_google_api_key, got %v", result.LastError)
	}
}

func TestGeminiDirectAdapterDispatchesAndExtractsText(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Layer OS agent acknowledgement."},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("GOOGLE_API_KEY", "test_key")

	a := geminiDirectAdapter{
		httpClient: client,
	}

	call := GatewayCall{
		CallID:      "test_002",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		TokenBudget: 256,
	}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "sent" {
		t.Fatalf("expected status=sent, got %q (err=%v)", result.Status, result.LastError)
	}
	if result.ResponsePreview == nil || *result.ResponsePreview == "" {
		t.Fatal("expected non-empty ResponsePreview")
	}
	if *result.ResponsePreview != "Layer OS agent acknowledgement." {
		t.Fatalf("unexpected ResponsePreview: %q", *result.ResponsePreview)
	}
}

func TestGeminiDirectAdapterAcceptsGeminiAPIKeyAlias(t *testing.T) {
	client := newInMemoryTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Gemini alias path acknowledged."},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "alias_key")

	a := geminiDirectAdapter{httpClient: client}
	call := GatewayCall{
		CallID:      "test_003",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		TokenBudget: 256,
	}
	result, err := a.Dispatch(call, PolicyDecision{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "sent" {
		t.Fatalf("expected alias-backed status=sent, got %q (err=%v)", result.Status, result.LastError)
	}
	if result.ResponsePreview == nil || *result.ResponsePreview != "Gemini alias path acknowledged." {
		t.Fatalf("unexpected alias-backed ResponsePreview: %+v", result.ResponsePreview)
	}
}

func TestGeminiPromptExtractsInstructions(t *testing.T) {
	call := GatewayCall{RequestKind: "design"}
	payload := map[string]any{"instructions": "Build a landing page layout."}
	prompt := geminiPromptFromPayload(payload, call)
	if prompt != "Build a landing page layout." {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
}

func TestGeminiPromptFallsBackToCallFields(t *testing.T) {
	call := GatewayCall{RequestKind: "verify", Notes: []string{"check:integrity"}}
	prompt := geminiPromptFromPayload(nil, call)
	if prompt == "" {
		t.Fatal("expected non-empty fallback prompt")
	}
}

func TestGeminiPromptIncludesPromptingContract(t *testing.T) {
	call := GatewayCall{RequestKind: "implement"}
	payload := map[string]any{
		"job": map[string]any{
			"role":    "implementer",
			"kind":    "implement",
			"summary": "Ship one bounded patch",
		},
		"prompting": map[string]any{
			"directive":           "Act like a practical staff aide.",
			"cognition_mode":      "staff_advisor",
			"decision_scope":      "bounded",
			"autonomy_budget":     "multi_step",
			"mutation_policy":     "scoped_write",
			"escalation_triggers": []any{"contract change needed"},
			"open_questions":      []any{"Which lane is safest next?"},
		},
	}
	prompt := geminiPromptFromPayload(payload, call)
	for _, want := range []string{
		"Operating posture:",
		"Directive: Act like a practical staff aide.",
		"Autonomy budget: multi_step",
		"Mutation policy: scoped_write",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got %q", want, prompt)
		}
	}
}

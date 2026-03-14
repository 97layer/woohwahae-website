package runtime

import "strings"

// multiProviderAdapter routes each gateway call to the adapter matching
// call.Provider (gemini / openai / claude). Falls back to a default adapter
// when the provider is not recognised.
type multiProviderAdapter struct {
	providers map[string]GatewayAdapter
	fallback  GatewayAdapter
}

func defaultProviderAdapters() map[string]GatewayAdapter {
	return map[string]GatewayAdapter{
		"gemini": geminiDirectAdapter{},
		"openai": openaiDirectAdapter{},
		"claude": claudeDirectAdapter{},
	}
}

func newMultiProviderAdapter(fallback GatewayAdapter) multiProviderAdapter {
	adapters := defaultProviderAdapters()
	switch typed := fallback.(type) {
	case multiProviderAdapter:
		for provider, adapter := range typed.providers {
			adapters[provider] = adapter
		}
		fallback = typed.fallback
	case geminiDirectAdapter:
		adapters["gemini"] = typed
	case openaiDirectAdapter:
		adapters["openai"] = typed
	case claudeDirectAdapter:
		adapters["claude"] = typed
	}
	return multiProviderAdapter{
		providers: adapters,
		fallback:  fallback,
	}
}

func (a multiProviderAdapter) Name() string         { return "multi" }
func (a multiProviderAdapter) Semantics() string    { return "direct_llm" }
func (a multiProviderAdapter) RequiredMode() string { return "single" }

func (a multiProviderAdapter) DispatchEnabled() bool {
	for _, adapter := range a.providers {
		if adapter.DispatchEnabled() {
			return true
		}
	}
	return a.fallback != nil && a.fallback.DispatchEnabled()
}

func (a multiProviderAdapter) Prepare(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.adapterFor(call).Prepare(call, decision)
}

func (a multiProviderAdapter) Dispatch(call GatewayCall, decision PolicyDecision) (GatewayCall, error) {
	return a.DispatchWithPayload(call, decision, nil)
}

func (a multiProviderAdapter) DispatchWithPayload(call GatewayCall, decision PolicyDecision, payload map[string]any) (GatewayCall, error) {
	adapter := a.adapterFor(call)
	if dwa, ok := adapter.(gatewayPayloadAdapter); ok {
		return dwa.DispatchWithPayload(call, decision, payload)
	}
	return adapter.Dispatch(call, decision)
}

func (a multiProviderAdapter) adapterFor(call GatewayCall) GatewayAdapter {
	provider := strings.ToLower(strings.TrimSpace(call.Provider))
	if adapter, ok := a.providers[provider]; ok {
		return adapter
	}
	if a.fallback != nil {
		return a.fallback
	}
	return recordGatewayAdapter{}
}

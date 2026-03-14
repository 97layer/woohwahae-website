package runtime

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type agentJobCouncil struct {
	Providers       []string
	PrimaryProvider string
}

func agentJobCouncilFromPayload(payload map[string]any, fallbackPrimary string) (agentJobCouncil, bool) {
	if payload == nil {
		return agentJobCouncil{}, false
	}
	raw, ok := payload["council"].(map[string]any)
	if !ok || raw == nil {
		return agentJobCouncil{}, false
	}
	providers := normalizeCouncilProviders(raw["providers"])
	primary := normalizeActor(strings.TrimSpace(promptingString(raw["primary_provider"])))
	if primary != "" && !containsStringFold(providers, primary) {
		providers = append([]string{primary}, providers...)
	}
	if len(providers) == 0 {
		return agentJobCouncil{}, false
	}
	if primary == "" {
		if fallback := normalizeActor(strings.TrimSpace(fallbackPrimary)); fallback != "" && containsStringFold(providers, fallback) {
			primary = fallback
		} else {
			primary = providers[0]
		}
	}
	return agentJobCouncil{
		Providers:       providers,
		PrimaryProvider: primary,
	}, true
}

func deriveAgentJobAutoCouncil(job AgentJob, adapters AdapterSummary, providers []ProviderSummary, fallbackPrimary string) (agentJobCouncil, bool) {
	if !agentJobAutoCouncilEligible(job) {
		return agentJobCouncil{}, false
	}
	available := autoCouncilProvidersForDispatch(adapters, providers, job.Role)
	if len(available) < 2 {
		return agentJobCouncil{}, false
	}
	primary := available[0]
	if preferred := normalizeActor(strings.TrimSpace(fallbackPrimary)); preferred != "" && containsStringFold(available, preferred) {
		primary = preferred
	}
	return agentJobCouncil{
		Providers:       available,
		PrimaryProvider: primary,
	}, true
}

func normalizeCouncilProviders(value any) []string {
	items := []string{}
	seen := map[string]struct{}{}
	appendProvider := func(raw string) {
		provider := normalizeActor(strings.TrimSpace(raw))
		if !councilProviderAllowed(provider) {
			return
		}
		if _, ok := seen[provider]; ok {
			return
		}
		seen[provider] = struct{}{}
		items = append(items, provider)
	}
	switch typed := value.(type) {
	case string:
		for _, item := range splitRoleCSV(typed) {
			appendProvider(item)
		}
	case []string:
		for _, item := range typed {
			appendProvider(item)
		}
	case []any:
		for _, item := range typed {
			appendProvider(promptingString(item))
		}
	}
	return items
}

func agentJobAutoCouncilEligible(job AgentJob) bool {
	if strings.TrimSpace(job.Source) == "founder.manual" {
		return false
	}
	switch strings.TrimSpace(job.Role) {
	case "planner", "designer":
		return true
	default:
		return false
	}
}

func autoCouncilProvidersForDispatch(adapters AdapterSummary, providers []ProviderSummary, role string) []string {
	if strings.TrimSpace(adapters.GatewaySemantics) != "direct_llm" {
		return nil
	}
	available := []string{}
	gateway := normalizeActor(adapters.Gateway)
	switch gateway {
	case "multi":
		for _, summary := range providers {
			if providerEligibleForAutoCouncil(summary) {
				available = appendUniqueString(available, normalizeActor(summary.Provider))
			}
		}
	default:
		if councilProviderAllowed(gateway) && directProviderCredentialReady(gateway) {
			available = appendUniqueString(available, gateway)
		}
	}
	ordered := orderAgentJobCouncilProviders(available, councilPreferenceForAgentRole(role))
	if len(ordered) > 2 {
		return append([]string{}, ordered[:2]...)
	}
	return ordered
}

func providerEligibleForAutoCouncil(summary ProviderSummary) bool {
	if !summary.DispatchEnabled || strings.TrimSpace(summary.Semantics) != "direct_llm" {
		return false
	}
	provider := normalizeActor(summary.Provider)
	gateway := normalizeActor(summary.GatewayAdapter)
	if gateway == "multi" {
		return councilProviderAllowed(provider) && directProviderCredentialReady(provider)
	}
	return gateway == provider && directProviderCredentialReady(provider)
}

func directProviderCredentialReady(provider string) bool {
	return ProviderCredentialReady(provider)
}

func councilPreferenceForAgentRole(role string) []string {
	items := []string{}
	if preferred := normalizeActor(splitRoleBindingEnv(os.Getenv("LAYER_OS_AGENT_ROLE_PROVIDERS"))[strings.TrimSpace(role)]); councilProviderAllowed(preferred) {
		items = append(items, preferred)
	}
	for _, provider := range []string{"claude", "openai", "gemini"} {
		items = appendUniqueString(items, provider)
	}
	return items
}

func orderAgentJobCouncilProviders(available []string, preferred []string) []string {
	ordered := []string{}
	for _, provider := range preferred {
		if containsStringFold(available, provider) {
			ordered = appendUniqueString(ordered, normalizeActor(provider))
		}
	}
	for _, provider := range available {
		ordered = appendUniqueString(ordered, normalizeActor(provider))
	}
	return ordered
}

func councilProviderAllowed(provider string) bool {
	switch strings.TrimSpace(provider) {
	case "gemini", "openai", "claude":
		return true
	default:
		return false
	}
}

func containsStringFold(items []string, want string) bool {
	want = normalizeActor(strings.TrimSpace(want))
	for _, item := range items {
		if normalizeActor(strings.TrimSpace(item)) == want {
			return true
		}
	}
	return false
}

func (s *Service) dispatchGatewayCall(call GatewayCall, adapter GatewayAdapter, payload map[string]any) (GatewayCall, error) {
	if err := s.createGatewayCallWithPayload(call, adapter, payload); err != nil {
		return GatewayCall{}, err
	}
	gateway, ok := s.gateway.get(call.CallID)
	if !ok {
		return GatewayCall{}, errors.New("gateway call not found after dispatch")
	}
	return gateway, nil
}

func (s *Service) dispatchAgentJobCouncil(job AgentJob, decision PolicyDecision, council agentJobCouncil, payload map[string]any) ([]GatewayCall, GatewayCall, error) {
	adapter := newMultiProviderAdapter(s.gatewayAdapter)
	calls := make([]GatewayCall, 0, len(council.Providers))
	for idx, provider := range council.Providers {
		model := modelForAgentDispatch(job.Role, job.Kind, provider)
		if model == "" {
			model = provider
		}
		call := GatewayCall{
			CallID:      fmt.Sprintf("gateway_job_%s_%s_%02d", job.JobID, provider, idx+1),
			DecisionID:  decision.DecisionID,
			Provider:    provider,
			Model:       model,
			RequestKind: job.Kind,
			Status:      "recorded",
			TokenBudget: tokenBudgetForAgentRole(job.Role),
			Notes: []string{
				"agent_job:" + job.JobID,
				"role:" + job.Role,
				"source:" + job.Source,
				"dispatch:council",
				"council:member",
			},
			CreatedAt: zeroSafeNow(),
		}
		if provider == council.PrimaryProvider {
			call.Notes = appendUniqueString(call.Notes, "council:primary")
		}
		gateway, err := s.dispatchGatewayCall(call, adapter, payload)
		if err != nil {
			return nil, GatewayCall{}, err
		}
		calls = append(calls, gateway)
	}
	primary, err := selectCouncilPrimaryGateway(calls, council.PrimaryProvider)
	if err != nil {
		return nil, GatewayCall{}, err
	}
	return calls, primary, nil
}

func selectCouncilPrimaryGateway(calls []GatewayCall, primaryProvider string) (GatewayCall, error) {
	if len(calls) == 0 {
		return GatewayCall{}, errors.New("council produced no gateway calls")
	}
	for _, call := range calls {
		if normalizeActor(call.Provider) == normalizeActor(primaryProvider) && strings.TrimSpace(call.Status) == "sent" {
			return call, nil
		}
	}
	for _, call := range calls {
		if strings.TrimSpace(call.Status) == "sent" {
			return call, nil
		}
	}
	for _, call := range calls {
		if normalizeActor(call.Provider) == normalizeActor(primaryProvider) {
			return call, nil
		}
	}
	return calls[0], nil
}

func councilResultPayload(calls []GatewayCall, primary GatewayCall) map[string]any {
	items := make([]map[string]any, 0, len(calls))
	callIDs := make([]string, 0, len(calls))
	providers := make([]string, 0, len(calls))
	succeeded := 0
	failed := 0
	for _, call := range calls {
		callIDs = append(callIDs, call.CallID)
		providers = append(providers, call.Provider)
		item := map[string]any{
			"call_id":      call.CallID,
			"provider":     call.Provider,
			"model":        call.Model,
			"status":       call.Status,
			"request_kind": call.RequestKind,
		}
		if call.ResponsePreview != nil {
			item["response_preview"] = *call.ResponsePreview
		}
		if call.LastError != nil {
			item["last_error"] = *call.LastError
		}
		if call.LastHTTPStatus != nil {
			item["last_http_status"] = *call.LastHTTPStatus
		}
		switch strings.TrimSpace(call.Status) {
		case "sent":
			succeeded++
		case "failed":
			failed++
		}
		items = append(items, item)
	}
	return map[string]any{
		"mode":             "sequential",
		"providers":        providers,
		"gateway_call_ids": callIDs,
		"primary_provider": primary.Provider,
		"primary_call_id":  primary.CallID,
		"primary_model":    primary.Model,
		"primary_status":   primary.Status,
		"requested_count":  len(calls),
		"succeeded_count":  succeeded,
		"failed_count":     failed,
		"calls":            items,
	}
}

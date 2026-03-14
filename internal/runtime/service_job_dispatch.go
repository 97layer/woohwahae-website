package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

func (s *Service) DispatchAgentJob(jobID string) (AgentDispatchResult, error) {
	job, ok := s.agentJob.get(jobID)
	if !ok {
		return AgentDispatchResult{}, errors.New("job_id not found")
	}
	if job.Status != "queued" && job.Status != "failed" {
		return AgentDispatchResult{}, errors.New("job must be queued or failed before dispatch")
	}
	adapters := s.Adapters()
	providers := s.Providers()
	defaultProvider := providerForAgentRole(job.Role, providers)
	council, hasCouncil := agentJobCouncilFromPayload(job.Payload, defaultProvider)
	autoCouncil := false
	if !hasCouncil {
		if derived, ok := deriveAgentJobAutoCouncil(job, adapters, providers, defaultProvider); ok {
			council = derived
			hasCouncil = true
			autoCouncil = true
		}
	}
	provider := defaultProvider
	if hasCouncil && strings.TrimSpace(council.PrimaryProvider) != "" {
		provider = council.PrimaryProvider
	}
	if provider == "" {
		return AgentDispatchResult{}, errors.New("no provider available for agent role")
	}
	model := modelForAgentDispatch(job.Role, job.Kind, provider)
	if model == "" {
		model = provider
	}
	fallbackReason := providerDispatchFallbackReason(provider, providers, s.gatewayAdapter)
	startNotes := append([]string{}, job.Notes...)
	startNotes = appendUniqueString(startNotes, "dispatching")
	dispatchActor := s.currentActor()
	executionOrigin := dispatchExecutionOrigin(fallbackReason, s.gatewayAdapter)
	updated, err := s.UpdateAgentJob(jobID, "running", startNotes, map[string]any{
		"dispatch_state":          "starting",
		"provider":                provider,
		"model":                   model,
		"provider_dispatch_ready": fallbackReason == "",
		"dispatch_actor":          dispatchActor,
		"execution_origin":        executionOrigin,
	})
	if err != nil {
		return AgentDispatchResult{}, err
	}
	decisionID := "policy_job_" + jobID
	policy, err := s.EvaluatePolicy(decisionID, updated.Summary, "agent_job", riskForAgentRole(updated.Role), noveltyForAgentRole(updated.Role), tokenClassForAgentRole(updated.Role), false)
	if err != nil {
		return AgentDispatchResult{}, err
	}
	callID := "gateway_job_" + jobID
	call := GatewayCall{
		CallID:      callID,
		DecisionID:  decisionID,
		Provider:    provider,
		Model:       model,
		RequestKind: updated.Kind,
		Status:      "recorded",
		TokenBudget: tokenBudgetForAgentRole(updated.Role),
		Notes: []string{
			"agent_job:" + updated.JobID,
			"role:" + updated.Role,
			"source:" + updated.Source,
		},
		CreatedAt: zeroSafeNow(),
	}
	var transportPayload map[string]any
	if hasCouncil || fallbackReason == "" {
		packet, packetErr := s.agentRunPacketForTransport(updated.JobID, "http_push")
		if packetErr != nil {
			return AgentDispatchResult{}, packetErr
		}
		transportPayload, err = agentRunPacketPayload(packet)
		if err != nil {
			return AgentDispatchResult{}, err
		}
		call.Notes = appendUniqueString(call.Notes, "dispatch:transport=http_push")
	}
	var gateway GatewayCall
	result := map[string]any{
		"provider":         provider,
		"model":            model,
		"dispatch_actor":   dispatchActor,
		"execution_origin": executionOrigin,
	}
	if hasCouncil {
		executionOrigin = "direct_llm_council"
		result["execution_origin"] = executionOrigin
		calls, primary, councilErr := s.dispatchAgentJobCouncil(updated, policy, council, transportPayload)
		if councilErr != nil {
			return AgentDispatchResult{}, councilErr
		}
		gateway = primary
		result["provider"] = primary.Provider
		result["model"] = primary.Model
		result["gateway_call_id"] = primary.CallID
		result["gateway_state"] = primary.Status
		result["attempt_count"] = primary.AttemptCount
		result["provider_dispatch_ready"] = primary.Status == "sent"
		result["council"] = councilResultPayload(calls, primary)
		if autoCouncil {
			result["council_auto"] = true
		}
	} else {
		selectedAdapter := s.gatewayAdapter
		if fallbackReason == "no_provider_endpoint" && s.gatewayAdapter != nil && s.gatewayAdapter.DispatchEnabled() {
			selectedAdapter = packetGatewayAdapter{base: s.gatewayAdapter, reason: fallbackReason}
		}
		gateway, err = s.dispatchGatewayCall(call, selectedAdapter, transportPayload)
		if err != nil {
			return AgentDispatchResult{}, err
		}
		result["gateway_call_id"] = gateway.CallID
		result["gateway_state"] = gateway.Status
		result["attempt_count"] = gateway.AttemptCount
		result["provider_dispatch_ready"] = fallbackReason == "" && gateway.Status == "sent"
	}
	if gateway.LastHTTPStatus != nil {
		result["last_http_status"] = *gateway.LastHTTPStatus
	}
	if gateway.LastError != nil {
		result["last_error"] = *gateway.LastError
	}
	if gateway.ResponsePreview != nil {
		result["response_preview"] = *gateway.ResponsePreview
	}
	status := "running"
	notes := append([]string{}, updated.Notes...)
	switch gateway.Status {
	case "failed":
		result["dispatch_state"] = "failed"
		status = "failed"
		notes = appendUniqueString(notes, "dispatch_failed")
	case "recorded":
		result["dispatch_state"] = "packet_ready"
		result["dispatch_transport"] = "job_packet"
		result["packet_ready"] = true
		result["job_packet_path"] = "/api/layer-os/jobs/packet?job_id=" + updated.JobID
		result["job_packet_command"] = "layer-osctl job packet --id " + updated.JobID
		if fallbackReason != "" && !hasCouncil {
			result["dispatch_reason"] = fallbackReason
		}
		notes = appendUniqueString(notes, "dispatch_packet_ready")
	default:
		result["dispatch_state"] = gateway.Status
		result["dispatch_transport"] = "http_push"
		if hasCouncil {
			if autoCouncil {
				notes = appendUniqueString(notes, "dispatch_council_auto")
			} else {
				notes = appendUniqueString(notes, "dispatch_council")
			}
		} else {
			notes = appendUniqueString(notes, "dispatch_sent")
		}
	}
	finalJob, err := s.UpdateAgentJob(jobID, status, notes, result)
	if err != nil {
		return AgentDispatchResult{}, err
	}
	if jobRequestsDispatchTelegram(finalJob) && s.TelegramEnabled() {
		if telegramErr := s.SendTelegramPacket(jobDispatchTelegramPacket(finalJob, gateway)); telegramErr != nil {
			s.recordAutomationWarning(dispatchTelegramFailureReviewItem(finalJob, gateway, telegramErr))
		}
	}

	// Auto-close only for text-native direct LLM lanes. Implementer/verifier
	// jobs still need an explicit report so an acknowledgement does not look
	// like a completed code change or audit.
	if jobAllowsDirectLLMAutoClose(finalJob, gateway) {
		summary := strings.TrimSpace(finalJob.Summary)
		if summary == "" {
			summary = "Direct LLM response captured for " + strings.TrimSpace(finalJob.Role) + " lane."
		}
		autoResult := map[string]any{
			"summary":         summary,
			"artifacts":       []string{"gateway_call:" + gateway.CallID},
			"response":        *gateway.ResponsePreview,
			"gateway_call_id": gateway.CallID,
			"provider":        gateway.Provider,
			"model":           gateway.Model,
			"completion_mode": "direct_llm_auto_close",
		}
		if councilValue, ok := result["council"]; ok {
			autoResult["council"] = councilValue
		}
		if auto, ok := result["council_auto"]; ok {
			autoResult["council_auto"] = auto
		}
		autoNotes := appendUniqueString(append([]string{}, finalJob.Notes...), "auto_closed")
		if report, err := s.ReportAgentJob(jobID, "succeeded", autoNotes, autoResult); err == nil {
			finalJob = report.Job
		}
	}

	s.maybeNotifyFounderAttention()
	return AgentDispatchResult{Job: finalJob, Policy: policy, Gateway: gateway}, nil
}

func jobAllowsDirectLLMAutoClose(job AgentJob, gateway GatewayCall) bool {
	if gateway.Status != "sent" || gateway.ResponsePreview == nil {
		return false
	}
	if !gatewayUsesDirectLLM(gateway) {
		return false
	}
	switch strings.TrimSpace(job.Role) {
	case "planner", "designer":
		return true
	default:
		return false
	}
}

func dispatchExecutionOrigin(fallbackReason string, adapter GatewayAdapter) string {
	if strings.TrimSpace(fallbackReason) == "no_provider_endpoint" {
		return "job_packet"
	}
	if adapter == nil {
		return "dispatch_unknown"
	}
	switch strings.TrimSpace(adapter.Semantics()) {
	case "direct_llm":
		return "direct_llm"
	case "dispatch":
		return "provider_api"
	default:
		return "dispatch_unknown"
	}
}

func gatewayUsesDirectLLM(gateway GatewayCall) bool {
	for _, note := range gateway.Notes {
		if strings.TrimSpace(note) == "dispatch:direct_llm" {
			return true
		}
	}
	return false
}

func agentRunPacketPayload(packet AgentRunPacket) (map[string]any, error) {
	raw, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func providerForAgentRole(role string, providers []ProviderSummary) string {
	m := splitRoleBindingEnv(os.Getenv("LAYER_OS_AGENT_ROLE_PROVIDERS"))
	if value := strings.TrimSpace(m[strings.TrimSpace(role)]); value != "" {
		return value
	}
	for _, provider := range providers {
		if provider.DispatchEnabled {
			return provider.Provider
		}
	}
	for _, provider := range providers {
		if provider.Declared {
			return provider.Provider
		}
	}
	return ""
}

func providerSummaryForName(provider string, providers []ProviderSummary) (ProviderSummary, bool) {
	trimmed := strings.TrimSpace(provider)
	for _, summary := range providers {
		if strings.TrimSpace(summary.Provider) == trimmed {
			return summary, true
		}
	}
	return ProviderSummary{}, false
}

func providerDispatchFallbackReason(provider string, providers []ProviderSummary, adapter GatewayAdapter) string {
	trimmed := strings.TrimSpace(provider)
	if trimmed == "" {
		return "no_provider_bound"
	}
	if summary, ok := providerSummaryForName(trimmed, providers); ok {
		if !summary.DispatchEnabled {
			return "gateway_dispatch_disabled"
		}
		if strings.TrimSpace(summary.Semantics) == "dispatch" && summary.Endpoint == nil {
			return "no_provider_endpoint"
		}
		return ""
	}
	if adapter != nil && adapter.DispatchEnabled() {
		// direct_llm adapters carry their own API key — no endpoint URL needed
		if adapter.Semantics() == "direct_llm" {
			return ""
		}
		if endpoint, ok := providerEndpointMapFromEnv()[trimmed]; ok && strings.TrimSpace(endpoint) != "" {
			return ""
		}
		return "no_provider_endpoint"
	}
	if adapter != nil {
		return "gateway_dispatch_disabled"
	}
	return "provider_unresolved"
}

func providerDispatchReady(provider string, providers []ProviderSummary, adapter GatewayAdapter) bool {
	return providerDispatchFallbackReason(provider, providers, adapter) == ""
}

func providerDispatchFallbackNote(reason string) string {
	switch strings.TrimSpace(reason) {
	case "gateway_dispatch_disabled":
		return "gateway adapter is record-only; dispatch falls back to job packet"
	case "no_provider_endpoint":
		return "provider endpoint missing; dispatch falls back to job packet"
	case "provider_unresolved":
		return "provider dispatch is unresolved; inspect provider roster before automation"
	default:
		return "dispatch routing requires review"
	}
}

func modelForAgentDispatch(role string, requestKind string, provider string) string {
	m := splitRoleBindingEnv(os.Getenv("LAYER_OS_AGENT_ROLE_MODELS"))
	if value := strings.TrimSpace(m[strings.TrimSpace(role)]); value != "" {
		return value
	}
	models := runtimeModelsFromEnv()
	if len(models) > 0 {
		return models[0]
	}
	return defaultModelForProvider(provider, requestKind)
}

func defaultModelForProvider(provider string, requestKind string) string {
	switch strings.TrimSpace(provider) {
	case "gemini":
		switch strings.TrimSpace(requestKind) {
		case "verify":
			return "gemini-2.5-pro-preview-03-25"
		default:
			return "gemini-2.0-flash"
		}
	case "claude":
		switch strings.TrimSpace(requestKind) {
		case "verify":
			return "claude-haiku-4-5-20251001"
		default:
			return "claude-opus-4-6"
		}
	case "openai":
		switch strings.TrimSpace(requestKind) {
		case "verify":
			return "gpt-4o"
		default:
			return "gpt-4o-mini"
		}
	default:
		return ""
	}
}

func splitRoleBindingEnv(raw string) map[string]string {
	items := map[string]string{}
	for _, part := range splitRoleCSV(raw) {
		chunks := strings.SplitN(part, "=", 2)
		if len(chunks) != 2 {
			continue
		}
		key := strings.TrimSpace(chunks[0])
		value := strings.TrimSpace(chunks[1])
		if key == "" || value == "" {
			continue
		}
		items[key] = value
	}
	return items
}

func riskForAgentRole(role string) string {
	switch strings.TrimSpace(role) {
	case "implementer":
		return "medium"
	case "verifier":
		return "medium"
	case "designer":
		return "medium"
	default:
		return "low"
	}
}

func noveltyForAgentRole(role string) string {
	switch strings.TrimSpace(role) {
	case "planner", "implementer", "designer":
		return "medium"
	default:
		return "low"
	}
}

func tokenClassForAgentRole(role string) string {
	switch strings.TrimSpace(role) {
	case "planner":
		return "large"
	case "implementer":
		return "medium"
	case "verifier":
		return "medium"
	case "designer":
		return "medium"
	default:
		return "medium"
	}
}

func tokenBudgetForAgentRole(role string) int {
	switch strings.TrimSpace(role) {
	case "planner":
		return 32000
	case "implementer":
		return 16000
	case "verifier":
		return 12000
	case "designer":
		return 12000
	default:
		return 16000
	}
}

func splitRoleCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

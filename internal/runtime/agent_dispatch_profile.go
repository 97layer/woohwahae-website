package runtime

func (s *Service) AgentDispatchProfiles() []AgentDispatchProfile {
	providers := s.Providers()
	roles := []string{"planner", "implementer", "verifier", "designer"}
	items := make([]AgentDispatchProfile, 0, len(roles))
	for _, role := range roles {
		provider := providerForAgentRole(role, providers)
		model := modelForAgentDispatch(role, defaultAgentRequestKind(role), provider)
		notes := []string{}
		ready := providerDispatchReady(provider, providers, s.gatewayAdapter)
		if provider == "" {
			notes = append(notes, "no provider bound")
		} else {
			notes = append(notes, "provider:"+provider)
			if reason := providerDispatchFallbackReason(provider, providers, s.gatewayAdapter); reason != "" {
				notes = append(notes, providerDispatchFallbackNote(reason))
			}
		}
		if model == "" {
			notes = append(notes, "model unresolved; adapter or provider default will be used at dispatch time")
		} else {
			notes = append(notes, "model:"+model)
		}
		items = append(items, AgentDispatchProfile{
			Role:          role,
			Provider:      provider,
			Model:         model,
			Risk:          riskForAgentRole(role),
			Novelty:       noveltyForAgentRole(role),
			TokenClass:    tokenClassForAgentRole(role),
			TokenBudget:   tokenBudgetForAgentRole(role),
			DispatchReady: ready,
			Notes:         notes,
		})
	}
	return items
}

func defaultAgentRequestKind(role string) string {
	switch role {
	case "planner":
		return "plan"
	case "verifier":
		return "verify"
	case "designer":
		return "design"
	default:
		return "implement"
	}
}

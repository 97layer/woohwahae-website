package runtime

import "strings"

const defaultPromptingDirective = "Operate like a practical staff aide: trust the official packet and runtime state, infer the founder's intended next move, take one safe bounded step, verify it, and report in plain language. Do not widen scope, invent side channels, or hide blockers."

const defaultJobDirective = "Operate like a practical staff aide and control tower: trust the official packet and runtime state, infer the founder's intended next move, and keep advancing the current lane until the objective is complete or an escalation trigger fires. Verify meaningful changes, report in plain language, and do not invent side channels or hide blockers."

var defaultPromptingEscalationTriggers = []string{
	"hot seam or blocked write boundary detected",
	"contract, schema, or authority change is required",
	"runtime state or evidence conflicts with the expected lane",
}

func defaultSessionPrompting(knowledge KnowledgePacket) PromptingContract {
	return PromptingContract{
		Directive:          defaultPromptingDirective,
		CognitionMode:      "staff_advisor",
		DecisionScope:      "full",
		AutonomyBudget:     "single_step",
		MutationPolicy:     "read_only",
		EscalationTriggers: append([]string{}, defaultPromptingEscalationTriggers...),
		OpenQuestions:      promptingOpenQuestions(knowledge.OpenThreads, 3),
	}
}

func defaultJobPrompting(job AgentJob, knowledge KnowledgePacket) PromptingContract {
	base := PromptingContract{
		Directive:          defaultJobDirective,
		CognitionMode:      "staff_advisor",
		DecisionScope:      "bounded",
		AutonomyBudget:     "multi_step",
		MutationPolicy:     defaultJobMutationPolicy(job),
		EscalationTriggers: append([]string{}, defaultPromptingEscalationTriggers...),
		OpenQuestions:      promptingOpenQuestions(knowledge.OpenThreads, 3),
	}
	return mergePayloadPrompting(base, job.Payload)
}

func defaultJobMutationPolicy(job AgentJob) string {
	if len(payloadStringList(job.Payload, "allowed_paths")) > 0 {
		return "scoped_write"
	}
	switch job.Stage {
	case StageCompose, StageExperience:
		return "scoped_write"
	default:
		return "read_only"
	}
}

func mergePayloadPrompting(base PromptingContract, payload map[string]any) PromptingContract {
	if payload == nil {
		return ensurePromptingDefaults(base)
	}
	raw, ok := payload["prompting"].(map[string]any)
	if !ok || raw == nil {
		return ensurePromptingDefaults(base)
	}
	if value := strings.TrimSpace(promptingString(raw["directive"])); value != "" {
		base.Directive = value
	}
	if value := strings.TrimSpace(promptingString(raw["cognition_mode"])); value != "" {
		base.CognitionMode = value
	}
	if value := strings.TrimSpace(promptingString(raw["decision_scope"])); value != "" {
		base.DecisionScope = value
	}
	if value := strings.TrimSpace(promptingString(raw["autonomy_budget"])); value != "" {
		base.AutonomyBudget = value
	}
	if value := strings.TrimSpace(promptingString(raw["mutation_policy"])); value != "" {
		base.MutationPolicy = value
	}
	if items := promptingStringList(raw["escalation_triggers"]); len(items) > 0 {
		base.EscalationTriggers = items
	}
	if items := promptingStringList(raw["open_questions"]); len(items) > 0 {
		base.OpenQuestions = items
	}
	return ensurePromptingDefaults(base)
}

func ensurePromptingDefaults(item PromptingContract) PromptingContract {
	if strings.TrimSpace(item.Directive) == "" {
		item.Directive = defaultPromptingDirective
	}
	if strings.TrimSpace(item.CognitionMode) == "" {
		item.CognitionMode = "staff_advisor"
	}
	if strings.TrimSpace(item.DecisionScope) == "" {
		item.DecisionScope = "bounded"
	}
	if strings.TrimSpace(item.AutonomyBudget) == "" {
		item.AutonomyBudget = "single_step"
	}
	if strings.TrimSpace(item.MutationPolicy) == "" {
		item.MutationPolicy = "read_only"
	}
	if len(item.EscalationTriggers) == 0 {
		item.EscalationTriggers = append([]string{}, defaultPromptingEscalationTriggers...)
	}
	if item.OpenQuestions == nil {
		item.OpenQuestions = []string{}
	}
	return item
}

func promptingOpenQuestions(items []OpenThread, max int) []string {
	if len(items) == 0 || max == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	questions := []string{}
	for _, item := range items {
		question := strings.TrimSpace(item.Question)
		if question == "" || seen[question] {
			continue
		}
		seen[question] = true
		questions = append(questions, question)
		if len(questions) >= max {
			break
		}
	}
	return questions
}

func (s *Service) HandoffSummary() HandoffSummary {
	knowledge := s.Knowledge()
	handoff := s.Handoff()
	return HandoffSummary{
		GeneratedAt:        zeroSafeNow(),
		CurrentFocus:       knowledge.CurrentFocus,
		NextSteps:          append([]string{}, knowledge.NextSteps...),
		OpenRisks:          append([]string{}, knowledge.OpenRisks...),
		ReviewTopOpen:      append([]ReviewRoomItem{}, handoff.ReviewRoom.TopOpen...),
		ActiveBranches:     append([]Branch{}, handoff.ActiveBranches...),
		ParallelCandidates: limitParallelCandidates(handoff.ParallelCandidates, 3),
		Auth:               handoff.Auth,
		WriteLease:         handoff.WriteLease,
		Counts:             handoff.Counts,
	}
}

func agentPacketHandoff(role string, s *Service) (*HandoffPacket, *HandoffSummary) {
	switch strings.TrimSpace(role) {
	case "planner":
		handoff := s.Handoff()
		return &handoff, nil
	default:
		summary := s.HandoffSummary()
		return nil, &summary
	}
}

func promptingString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func promptingStringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string{}, typed...)
	case []any:
		items := []string{}
		for _, item := range typed {
			text := strings.TrimSpace(promptingString(item))
			if text != "" {
				items = append(items, text)
			}
		}
		return items
	case string:
		if strings.TrimSpace(typed) == "" {
			return []string{}
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return []string{}
	}
}

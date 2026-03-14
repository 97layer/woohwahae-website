package runtime

import (
	"strings"
)

// buildAgentPrompt constructs a focused, role-aware prompt from an AgentRunPacket
// payload (the full packet serialised as map[string]any). It extracts role,
// task summary, explicit instructions, chain context, and top corpus lessons +
// open threads from the embedded KnowledgePacket.
func buildAgentPrompt(payload map[string]any, call GatewayCall) string {
	if payload == nil {
		return fallbackPrompt(call)
	}

	// Extract from AgentRunPacket structure: payload["job"] is the AgentJob.
	jobMap, _ := payload["job"].(map[string]any)
	if jobMap == nil {
		if prompt := strings.TrimSpace(agentPromptString(payload["instructions"])); prompt != "" {
			return prompt
		}
		if prompt := strings.TrimSpace(agentPromptString(payload["prompt"])); prompt != "" {
			return prompt
		}
		return fallbackPrompt(call)
	}

	role := agentPromptString(jobMap["role"])
	kind := agentPromptString(jobMap["kind"])
	summary := agentPromptString(jobMap["summary"])

	// job.payload may contain explicit instructions set by the founder.
	explicitInstructions := []string{}
	if prompt := strings.TrimSpace(agentPromptString(payload["instructions"])); prompt != "" {
		explicitInstructions = append(explicitInstructions, prompt)
	}
	if prompt := strings.TrimSpace(agentPromptString(payload["prompt"])); prompt != "" {
		explicitInstructions = append(explicitInstructions, prompt)
	}
	if jobPayload, ok := jobMap["payload"].(map[string]any); ok {
		if prompt := strings.TrimSpace(agentPromptString(jobPayload["instructions"])); prompt != "" {
			explicitInstructions = append(explicitInstructions, prompt)
		}
		if prompt := strings.TrimSpace(agentPromptString(jobPayload["prompt"])); prompt != "" {
			explicitInstructions = append(explicitInstructions, prompt)
		}
	}

	// Chain context: parent job result, if present.
	var chainContext string
	if jobPayload, ok := jobMap["payload"].(map[string]any); ok {
		if parentResult, ok := jobPayload["chain_parent_result"].(map[string]any); ok {
			if resp, ok := parentResult["response"].(string); ok && strings.TrimSpace(resp) != "" {
				chainContext = strings.TrimSpace(resp)
				if len(chainContext) > 2000 {
					chainContext = chainContext[:2000]
				}
			}
		}
		if chainContext == "" {
			if parentSummary, ok := jobPayload["chain_parent_summary"].(string); ok {
				chainContext = strings.TrimSpace(parentSummary)
			}
		}
	}

	// Corpus lessons: top 3 recent decisions from KnowledgePacket.
	var corpusLines []string
	var openQuestions []string
	if knowledge, ok := payload["knowledge"].(map[string]any); ok {
		if lessons, ok := knowledge["corpus_lessons"].([]any); ok {
			for i, raw := range lessons {
				if i >= 3 {
					break
				}
				if lesson, ok := raw.(map[string]any); ok {
					if s, ok := lesson["summary"].(string); ok && strings.TrimSpace(s) != "" {
						line := strings.TrimSpace(s)
						if len(line) > 300 {
							line = line[:300]
						}
						corpusLines = append(corpusLines, "- "+line)
					}
				}
			}
		}
		// Open threads: surface unanswered questions as hints.
		if threads, ok := knowledge["open_threads"].([]any); ok {
			for i, raw := range threads {
				if i >= 2 {
					break
				}
				if thread, ok := raw.(map[string]any); ok {
					if q, ok := thread["question"].(string); ok && strings.TrimSpace(q) != "" {
						openQuestions = append(openQuestions, strings.TrimSpace(q))
					}
				}
			}
		}
	}
	if prompting, ok := payload["prompting"].(map[string]any); ok {
		if lines := agentPromptPromptingLines(prompting, openQuestions); len(lines) > 0 {
			corpusLines = append([]string{"Operating posture:"}, lines...)
		}
		if items := agentPromptStringList(prompting["open_questions"]); len(items) > 0 {
			openQuestions = items
		}
	}
	if len(openQuestions) > 0 && len(corpusLines) > 0 && corpusLines[0] == "Operating posture:" {
		corpusLines = append(corpusLines, "- Open questions:")
		for _, item := range openQuestions {
			corpusLines = append(corpusLines, "  - "+item)
		}
	} else if len(openQuestions) > 0 {
		for _, item := range openQuestions {
			corpusLines = append(corpusLines, "- Open question: "+item)
		}
	}

	var parts []string

	// Role line
	roleLabel := strings.TrimSpace(role)
	if roleLabel == "" {
		roleLabel = strings.TrimSpace(kind)
	}
	if roleLabel == "" {
		roleLabel = "agent"
	}
	parts = append(parts, "You are a Layer OS "+roleLabel+".")

	// Task
	if summary := strings.TrimSpace(summary); summary != "" {
		parts = append(parts, "\nTask: "+summary)
	}

	if len(explicitInstructions) > 0 {
		parts = append(parts, "\nInstructions:\n"+strings.Join(explicitInstructions, "\n\n"))
	}

	// Chain context from parent job.
	if chainContext != "" {
		parts = append(parts, "\nContext from previous agent:\n"+chainContext)
	}

	// Corpus context: recent lessons and open questions from Layer OS memory.
	if len(corpusLines) > 0 {
		parts = append(parts, "\nLayer OS recent context:\n"+strings.Join(corpusLines, "\n"))
	}

	// Output constraint
	parts = append(parts, "\nRespond concisely and directly. Do not summarise unrelated system state.")

	return strings.Join(parts, "\n")
}

func fallbackPrompt(call GatewayCall) string {
	if call.RequestKind != "" {
		return "You are a Layer OS agent. Task kind: " + call.RequestKind + ". Respond concisely."
	}
	return "You are a Layer OS agent. Respond concisely."
}

func agentPromptPromptingLines(prompting map[string]any, fallbackQuestions []string) []string {
	if prompting == nil {
		return nil
	}
	lines := []string{}
	if value := strings.TrimSpace(agentPromptString(prompting["directive"])); value != "" {
		lines = append(lines, "- Directive: "+value)
	}
	if value := strings.TrimSpace(agentPromptString(prompting["cognition_mode"])); value != "" {
		lines = append(lines, "- Cognition mode: "+value)
	}
	if value := strings.TrimSpace(agentPromptString(prompting["decision_scope"])); value != "" {
		lines = append(lines, "- Decision scope: "+value)
	}
	if value := strings.TrimSpace(agentPromptString(prompting["autonomy_budget"])); value != "" {
		lines = append(lines, "- Autonomy budget: "+value)
	}
	if value := strings.TrimSpace(agentPromptString(prompting["mutation_policy"])); value != "" {
		lines = append(lines, "- Mutation policy: "+value)
	}
	if items := agentPromptStringList(prompting["escalation_triggers"]); len(items) > 0 {
		lines = append(lines, "- Escalate when: "+strings.Join(items, "; "))
	}
	if items := agentPromptStringList(prompting["open_questions"]); len(items) > 0 {
		fallbackQuestions = items
	}
	if len(fallbackQuestions) > 0 {
		lines = append(lines, "- Open questions: "+strings.Join(fallbackQuestions, " | "))
	}
	return lines
}

func agentPromptString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func agentPromptStringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string{}, typed...)
	case []any:
		items := []string{}
		for _, item := range typed {
			text := strings.TrimSpace(agentPromptString(item))
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

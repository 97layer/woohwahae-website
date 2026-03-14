package runtime

import "sync"

type policyStore struct {
	mu    sync.RWMutex
	items []PolicyDecision
}

func newPolicyStore(items []PolicyDecision) *policyStore {
	return &policyStore{items: normalizePolicyDecisions(items)}
}

func (s *policyStore) list() []PolicyDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PolicyDecision, len(s.items))
	copy(out, s.items)
	return out
}

func (s *policyStore) create(item PolicyDecision) error {
	item = normalizePolicyDecision(item)
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	return nil
}

func (s *policyStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *policyStore) get(decisionID string) (PolicyDecision, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.DecisionID == decisionID {
			return item, true
		}
	}
	return PolicyDecision{}, false
}

func decidePolicy(intent string, risk string, novelty string, tokenClass string, requiresApproval bool) (string, string, []string) {
	reasons := []string{}

	if tokenClass == "tiny" && (risk == "high" || novelty == "high") {
		reasons = append(reasons, "token_class too small for high-risk or high-novelty work")
		return "blocked", "hold", reasons
	}

	if risk == "high" || novelty == "high" {
		reasons = append(reasons, "high-risk or high-novelty work stays in the strongest single review lane")
		if requiresApproval {
			reasons = append(reasons, "approval-sensitive work remains daemon-mediated instead of direct local execution")
		}
		return "single", "go", reasons
	}

	if requiresApproval || risk == "medium" || novelty == "medium" || tokenClass == "medium" || tokenClass == "large" {
		reasons = append(reasons, "moderate ambiguity or approval sensitivity routes to single mode")
		return "single", "go", reasons
	}

	reasons = append(reasons, "low-risk work stays local")
	if intent != "" {
		reasons = append(reasons, "intent is specific enough for deterministic local handling")
	}
	return "local", "go", reasons
}

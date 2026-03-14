package runtime

import "testing"

func TestFounderActionEventsCarryRationaleAndEvidence(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.PutTarget(DeployTarget{TargetID: "vm", Command: []string{"/usr/bin/true"}}); err != nil {
		t.Fatalf("put target: %v", err)
	}

	if _, err := service.StartFounderFlow("flow_001", "work_001", "approval_001", "Founder loop", "close founder loop", []string{"start"}); err != nil {
		t.Fatalf("start founder flow: %v", err)
	}
	if _, err := service.ApproveFounderFlow("flow_001", []string{"approve"}); err != nil {
		t.Fatalf("approve founder flow: %v", err)
	}
	if _, err := service.ReleaseFounderFlow("flow_001", "release_001", "deploy_001", "vm", "cockpit", []string{"release"}); err != nil {
		t.Fatalf("release founder flow: %v", err)
	}
	if _, err := service.RollbackFounderFlow("flow_001", "rollback_001", []string{"rollback"}); err != nil {
		t.Fatalf("rollback founder flow: %v", err)
	}

	events := service.ListEvents()
	checks := []struct {
		kind           string
		rule           string
		reason         string
		evidencePrefix string
	}{
		{kind: "founder.action.started", rule: "founder_action.start", reason: "founder opened a new approval-gated flow", evidencePrefix: "approval:approval_001"},
		{kind: "founder.action.approved", rule: "founder_action.approve", reason: "founder approved the pending flow gate", evidencePrefix: "approval:approval_001"},
		{kind: "founder.action.released", rule: "founder_action.release", reason: "founder advanced the approved flow into release and deploy", evidencePrefix: "release:release_001"},
		{kind: "founder.action.rolled_back", rule: "founder_action.rollback", reason: "founder initiated rollback to contain release risk", evidencePrefix: "rollback:rollback_001"},
	}
	for _, check := range checks {
		var found *EventEnvelope
		for i := range events {
			if events[i].Kind == check.kind {
				found = &events[i]
				break
			}
		}
		if found == nil {
			t.Fatalf("expected event %q in %+v", check.kind, events)
		}
		if found.Actor == "" {
			t.Fatalf("expected non-empty actor for %q", check.kind)
		}
		if found.Surface != SurfaceCockpit {
			t.Fatalf("expected cockpit surface for %q, got %q", check.kind, found.Surface)
		}
		if found.WorkItemID != "work_001" {
			t.Fatalf("expected work_001 for %q, got %q", check.kind, found.WorkItemID)
		}
		rule, _ := found.Data["rule"].(string)
		if rule != check.rule {
			t.Fatalf("expected rule %q for %q, got %q", check.rule, check.kind, rule)
		}
		reason, _ := found.Data["reason"].(string)
		if reason != check.reason {
			t.Fatalf("expected reason %q for %q, got %q", check.reason, check.kind, reason)
		}
		matched := false
		switch evidence := found.Data["evidence"].(type) {
		case []string:
			for _, value := range evidence {
				if value == check.evidencePrefix {
					matched = true
					break
				}
			}
		case []any:
			for _, item := range evidence {
				if value, ok := item.(string); ok && value == check.evidencePrefix {
					matched = true
					break
				}
			}
		default:
			t.Fatalf("expected evidence array for %q, got %#v", check.kind, found.Data["evidence"])
		}
		if !matched {
			t.Fatalf("expected evidence %q for %q, got %#v", check.evidencePrefix, check.kind, found.Data["evidence"])
		}
	}
}

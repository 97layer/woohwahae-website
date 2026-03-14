package runtime

type founderPriorityRule struct {
	lane   string
	action string
	source string
	reason string
	rule   string
}

type founderPrioritySelection struct {
	action     string
	primaryRef string
	lastChange *FounderItem
	rationale  *PriorityRationale
}

type founderParallelLaneSpec struct {
	lane     string
	source   string
	reason   string
	priority string
}

var founderPriorityRules = []founderPriorityRule{
	{lane: "review_open", action: "review_room", source: "review_room.open", reason: "open agenda overrides execution lanes until reviewed", rule: "founder_priority.review_open"},
	{lane: "risk", action: "rollback_or_fix", source: "founder_view.risk", reason: "risk lane outranks planning, waiting, and active execution", rule: "founder_priority.risk"},
	{lane: "proposal", action: "shape_or_promote", source: "founder_view.waiting", reason: "open proposals outrank approvals because planning must become an explicit work item before execution", rule: "founder_priority.proposal"},
	{lane: "job", action: "dispatch_job", source: "founder_view.waiting", reason: "queued agent jobs should be dispatched or triaged before generic waiting approvals", rule: "founder_priority.job"},
	{lane: "waiting", action: "approve", source: "founder_view.waiting", reason: "waiting approvals outrank active execution once planning is shaped", rule: "founder_priority.waiting"},
	{lane: "now", action: "continue", source: "founder_view.now", reason: "active execution is the next lane when no review or approval blocker exists", rule: "founder_priority.now"},
	{lane: "idle", action: "idle", source: "founder_view.idle", reason: "no open review, risk, waiting, or active lanes remain", rule: "founder_priority.idle"},
}

var founderParallelLaneSpecs = []founderParallelLaneSpec{
	{lane: "review_open", reason: "secondary review-room agenda"},
	{lane: "risk", source: "founder_view.risk", reason: "secondary risk lane", priority: "high"},
	{lane: "waiting", source: "founder_view.waiting", reason: "secondary waiting lane", priority: "medium"},
	{lane: "now", source: "founder_view.now", reason: "secondary active lane", priority: "medium"},
}

func selectFounderPriority(view FounderView, review ReviewRoomSummary) founderPrioritySelection {
	for _, rule := range founderPriorityRules {
		switch rule.lane {
		case "review_open":
			if review.OpenCount == 0 || len(review.TopOpen) == 0 {
				continue
			}
			item := FounderItem{
				Kind:      "review_room",
				Ref:       review.TopOpen[0].Text,
				Summary:   review.TopOpen[0].Text,
				Status:    "open",
				UpdatedAt: derefTime(review.UpdatedAt),
			}
			return founderPrioritySelection{action: rule.action, primaryRef: review.TopOpen[0].Text, lastChange: &item, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
		case "risk":
			if len(view.Risk) == 0 {
				continue
			}
			item := view.Risk[0]
			return founderPrioritySelection{action: rule.action, primaryRef: item.Ref, lastChange: &item, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
		case "proposal":
			for _, item := range view.Waiting {
				if item.Kind != "proposal" {
					continue
				}
				copy := item
				return founderPrioritySelection{action: rule.action, primaryRef: item.Ref, lastChange: &copy, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
			}
		case "job":
			for _, item := range view.Waiting {
				if item.Kind != "agent_job" {
					continue
				}
				copy := item
				return founderPrioritySelection{action: rule.action, primaryRef: item.Ref, lastChange: &copy, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
			}
		case "waiting":
			for _, item := range view.Waiting {
				if item.Kind == "proposal" {
					continue
				}
				copy := item
				return founderPrioritySelection{action: rule.action, primaryRef: item.Ref, lastChange: &copy, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
			}
		case "now":
			if len(view.Now) == 0 {
				continue
			}
			item := view.Now[0]
			return founderPrioritySelection{action: rule.action, primaryRef: item.Ref, lastChange: &item, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
		case "idle":
			return founderPrioritySelection{action: rule.action, rationale: &PriorityRationale{Lane: rule.lane, Source: rule.source, Reason: rule.reason, Rule: rule.rule}}
		}
	}
	return founderPrioritySelection{action: "idle", rationale: &PriorityRationale{Lane: "idle", Source: "founder_view.idle", Reason: "no open review, risk, waiting, or active lanes remain", Rule: "founder_priority.idle"}}
}

func deriveParallelCandidates(view FounderView, summary FounderSummary, review ReviewRoom) []ParallelCandidate {
	candidates := []ParallelCandidate{}
	for _, spec := range founderParallelLaneSpecs {
		switch spec.lane {
		case "review_open":
			for _, item := range review.Open {
				if item.Text == summary.PrimaryRef {
					continue
				}
				candidates = appendParallelCandidate(candidates, ParallelCandidate{
					Text:     item.Text,
					Kind:     item.Kind,
					Source:   item.Source,
					Reason:   spec.reason,
					Priority: item.Severity,
					Ref:      item.Ref,
				})
			}
		case "risk":
			candidates = appendFounderLaneCandidates(candidates, view.Risk, summary.PrimaryRef, spec)
		case "waiting":
			candidates = appendFounderLaneCandidates(candidates, view.Waiting, summary.PrimaryRef, spec)
		case "now":
			candidates = appendFounderLaneCandidates(candidates, view.Now, summary.PrimaryRef, spec)
		}
	}
	if len(candidates) > 5 {
		return append([]ParallelCandidate{}, candidates[:5]...)
	}
	return candidates
}

func appendFounderLaneCandidates(items []ParallelCandidate, lane []FounderItem, primaryRef string, spec founderParallelLaneSpec) []ParallelCandidate {
	for _, item := range lane {
		if item.Ref == primaryRef {
			continue
		}
		ref := item.Ref
		items = appendParallelCandidate(items, ParallelCandidate{
			Text:     item.Summary,
			Kind:     item.Kind,
			Source:   spec.source,
			Reason:   spec.reason,
			Priority: spec.priority,
			Ref:      &ref,
		})
	}
	return items
}

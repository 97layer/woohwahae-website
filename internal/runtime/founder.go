package runtime

import "time"

func (s *Service) FounderView() FounderView {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deriveFounderViewLocked()
}

func (s *Service) FounderSummary() FounderSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deriveFounderSummaryLocked(s.deriveFounderViewLocked(), SummarizeReviewRoom(s.currentReviewRoomLocked()))
}

func (s *Service) deriveFounderViewLocked() FounderView {
	view := FounderView{
		Now:     []FounderItem{},
		Waiting: []FounderItem{},
		Risk:    []FounderItem{},
	}

	flowApprovalRefs := map[string]struct{}{}

	for _, proposal := range reverseProposals(s.proposal.list()) {
		if proposal.Status != "proposed" {
			continue
		}
		view.Waiting = append(view.Waiting, FounderItem{
			Kind:       "proposal",
			Ref:        proposal.ProposalID,
			WorkItemID: "",
			Summary:    proposal.Summary,
			Status:     proposal.Status,
			UpdatedAt:  proposal.UpdatedAt,
		})
	}

	for _, job := range reverseAgentJobs(s.agentJob.list()) {
		card := FounderItem{
			Kind:       "agent_job",
			Ref:        job.JobID,
			WorkItemID: "",
			Summary:    job.Summary,
			Status:     job.Status,
			UpdatedAt:  job.UpdatedAt,
		}
		switch job.Status {
		case "queued":
			view.Waiting = append(view.Waiting, card)
		case "running":
			view.Now = append(view.Now, card)
		case "failed":
			view.Risk = append(view.Risk, card)
		}
	}

	for _, flow := range reverseFlows(s.flow.list()) {
		card := s.flowCardLocked(flow)
		switch flow.Status {
		case "active":
			view.Now = append(view.Now, card)
		case "waiting":
			view.Waiting = append(view.Waiting, card)
		case "blocked", "rolled_back":
			view.Risk = append(view.Risk, card)
		}
		if flow.ApprovalID != nil {
			flowApprovalRefs[*flow.ApprovalID] = struct{}{}
		}
	}

	for _, approval := range reverseApprovals(s.approval.list()) {
		if _, ok := flowApprovalRefs[approval.ApprovalID]; ok {
			continue
		}
		card := FounderItem{
			Kind:       "approval",
			Ref:        approval.ApprovalID,
			WorkItemID: approval.WorkItemID,
			Summary:    approval.Summary,
			Status:     approval.Status,
			UpdatedAt:  approvalTimestamp(approval),
		}
		switch approval.Status {
		case "pending":
			view.Waiting = append(view.Waiting, card)
		case "rejected":
			view.Risk = append(view.Risk, card)
		}
	}

	releases := s.release.list()
	if len(releases) > 0 {
		last := releases[len(releases)-1]
		view.LastRelease = &FounderItem{
			Kind:       "release",
			Ref:        last.ReleaseID,
			WorkItemID: last.WorkItemID,
			Summary:    last.Channel + " -> " + last.Target,
			Status:     "released",
			UpdatedAt:  derefTime(last.ReleasedAt),
		}
	}

	rollbacks := s.rollback.list()
	if len(rollbacks) > 0 {
		last := rollbacks[len(rollbacks)-1]
		view.LastRollback = &FounderItem{
			Kind:       "rollback",
			Ref:        last.RollbackID,
			WorkItemID: workIDForRelease(s.release.list(), last.ReleaseID),
			Summary:    "rollback -> " + last.Target,
			Status:     last.Status,
			UpdatedAt:  derefTime(last.FinishedAt),
		}
	}

	return view
}

func (s *Service) deriveFounderSummaryLocked(view FounderView, review ReviewRoomSummary) FounderSummary {
	reviewTopOpen := make([]string, 0, len(review.TopOpen))
	for _, item := range review.TopOpen {
		reviewTopOpen = append(reviewTopOpen, item.Text)
	}
	summary := FounderSummary{
		EnvironmentAdvisory: currentEnvironmentAdvisory(),
		NowCount:            len(view.Now),
		WaitingCount:        len(view.Waiting),
		RiskCount:           len(view.Risk),
		ReviewOpenCount:     review.OpenCount,
		ReviewTopOpen:       reviewTopOpen,
		LastRelease:         view.LastRelease,
		LastRollback:        view.LastRollback,
	}

	selection := selectFounderPriority(view, review)
	summary.PrimaryAction = selection.action
	summary.PrimaryRef = selection.primaryRef
	summary.PriorityRationale = selection.rationale
	summary.LastChange = selection.lastChange
	if summary.LastChange == nil {
		summary.LastChange = newestFounderItem(view.LastRelease, view.LastRollback)
	}

	return summary
}

func (s *Service) flowCardLocked(flow FlowRun) FounderItem {
	work, _ := s.work.get(flow.WorkItemID)
	summary := work.Title
	if summary == "" {
		summary = flow.WorkItemID
	}
	return FounderItem{
		Kind:       "flow",
		Ref:        flow.FlowID,
		WorkItemID: flow.WorkItemID,
		Summary:    summary,
		Status:     flow.Status,
		UpdatedAt:  flow.UpdatedAt,
	}
}

func reverseProposals(items []ProposalItem) []ProposalItem {
	out := make([]ProposalItem, len(items))
	copy(out, items)
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}
	return out
}

func reverseAgentJobs(items []AgentJob) []AgentJob {
	out := make([]AgentJob, len(items))
	copy(out, items)
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}
	return out
}

func reverseFlows(items []FlowRun) []FlowRun {
	out := make([]FlowRun, len(items))
	copy(out, items)
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}
	return out
}

func reverseApprovals(items []ApprovalItem) []ApprovalItem {
	out := make([]ApprovalItem, len(items))
	copy(out, items)
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}
	return out
}

func approvalTimestamp(item ApprovalItem) time.Time {
	if item.ResolvedAt != nil && !item.ResolvedAt.IsZero() {
		return *item.ResolvedAt
	}
	return item.RequestedAt
}

func derefTime(value *time.Time) time.Time {
	if value == nil || value.IsZero() {
		return zeroSafeNow()
	}
	return *value
}

func workIDForRelease(items []ReleasePacket, releaseID string) string {
	for _, item := range items {
		if item.ReleaseID == releaseID {
			return item.WorkItemID
		}
	}
	return ""
}

func newestFounderItem(items ...*FounderItem) *FounderItem {
	var newest *FounderItem
	for _, item := range items {
		if item == nil {
			continue
		}
		if newest == nil || item.UpdatedAt.After(newest.UpdatedAt) {
			copy := *item
			newest = &copy
		}
	}
	return newest
}

func appendParallelCandidate(items []ParallelCandidate, candidate ParallelCandidate) []ParallelCandidate {
	if candidate.Text == "" {
		return items
	}
	for _, item := range items {
		if item.Text == candidate.Text && item.Source == candidate.Source {
			return items
		}
	}
	return append(items, candidate)
}

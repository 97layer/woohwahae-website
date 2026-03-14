package runtime

import (
	"errors"
	"strings"
)

type founderActionRecord struct {
	kind     string
	stage    Stage
	reason   string
	rule     string
	evidence []string
	extras   map[string]any
}

func founderActionEvidence(values ...string) []string {
	evidence := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		evidence = append(evidence, trimmed)
	}
	return evidence
}

func founderActionPayload(record founderActionRecord, flow FlowRun, notes []string) map[string]any {
	payload := map[string]any{
		"status":   flow.Status,
		"reason":   record.reason,
		"rule":     record.rule,
		"evidence": append([]string{}, record.evidence...),
		"notes":    append([]string{}, notes...),
	}
	if flow.ApprovalID != nil {
		payload["approval_id"] = *flow.ApprovalID
	}
	if flow.ReleaseID != nil {
		payload["release_id"] = *flow.ReleaseID
	}
	if flow.DeployID != nil {
		payload["deploy_id"] = *flow.DeployID
	}
	if flow.RollbackID != nil {
		payload["rollback_id"] = *flow.RollbackID
	}
	for key, value := range record.extras {
		payload[key] = value
	}
	return payload
}

func (s *Service) appendFounderActionEvent(flow FlowRun, record founderActionRecord, notes []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldEventState := s.event.state()
	if err := s.event.append(newEvent(record.kind, s.currentActor(), SurfaceCockpit, flow.WorkItemID, record.stage, founderActionPayload(record, flow, notes))); err != nil {
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) StartFounderFlow(flowID string, workID string, approvalID string, title string, intent string, notes []string) (FlowRun, error) {
	work := WorkItem{
		ID:               workID,
		Title:            title,
		Intent:           intent,
		Stage:            StageDiscover,
		Surface:          SurfaceCockpit,
		Pack:             "founder",
		Priority:         "high",
		Risk:             "medium",
		RequiresApproval: true,
		Payload:          map[string]any{"source": "founder_action"},
		CorrelationID:    workID,
		CreatedAt:        zeroSafeNow(),
	}
	if err := s.CreateWorkItem(work); err != nil {
		return FlowRun{}, err
	}
	if err := s.CreateApproval(ApprovalItem{
		ApprovalID:      approvalID,
		WorkItemID:      workID,
		Stage:           StageVerify,
		Summary:         title,
		Risks:           []string{"founder approval required"},
		RollbackPlan:    "hold release",
		DecisionSurface: SurfaceCockpit,
		Status:          "pending",
		RequestedAt:     zeroSafeNow(),
	}); err != nil {
		return FlowRun{}, err
	}
	flow, err := s.SyncFlow(flowID, workID, &approvalID, nil, nil, nil, nil, nil, nil, notes)
	if err != nil {
		return FlowRun{}, err
	}
	record := founderActionRecord{
		kind:   "founder.action.started",
		stage:  StageDiscover,
		reason: "founder opened a new approval-gated flow",
		rule:   "founder_action.start",
		evidence: founderActionEvidence(
			"flow:"+flowID,
			"work:"+workID,
			"approval:"+approvalID,
		),
		extras: map[string]any{
			"action":       "start",
			"flow_id":      flowID,
			"work_item_id": workID,
			"title":        title,
			"intent":       intent,
		},
	}
	if err := s.appendFounderActionEvent(flow, record, notes); err != nil {
		return flow, err
	}
	return flow, nil
}

func (s *Service) ApproveFounderFlow(flowID string, notes []string) (FlowRun, error) {
	s.mu.Lock()
	flow, ok := s.flow.get(flowID)
	s.mu.Unlock()
	if !ok {
		return FlowRun{}, errors.New("flow_id not found")
	}
	if flow.ApprovalID == nil {
		return FlowRun{}, errors.New("flow approval_id not set")
	}
	if _, err := s.ResolveApproval(*flow.ApprovalID, "approved"); err != nil {
		return FlowRun{}, err
	}
	updated, err := s.SyncFlow(flow.FlowID, flow.WorkItemID, flow.ApprovalID, flow.PolicyDecisionID, flow.ExecuteID, flow.VerificationID, flow.ReleaseID, flow.DeployID, flow.RollbackID, notes)
	if err != nil {
		return FlowRun{}, err
	}
	record := founderActionRecord{
		kind:   "founder.action.approved",
		stage:  StageVerify,
		reason: "founder approved the pending flow gate",
		rule:   "founder_action.approve",
		evidence: founderActionEvidence(
			"flow:"+flow.FlowID,
			"work:"+flow.WorkItemID,
			"approval:"+*flow.ApprovalID,
		),
		extras: map[string]any{
			"action":  "approve",
			"flow_id": flow.FlowID,
		},
	}
	if err := s.appendFounderActionEvent(updated, record, notes); err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) ReleaseFounderFlow(flowID string, releaseID string, deployID string, target string, channel string, notes []string) (FlowRun, error) {
	s.mu.Lock()
	flow, ok := s.flow.get(flowID)
	s.mu.Unlock()
	if !ok {
		return FlowRun{}, errors.New("flow_id not found")
	}
	if flow.ApprovalID == nil {
		return FlowRun{}, errors.New("flow approval_id not set")
	}
	approval, ok := s.approval.get(*flow.ApprovalID)
	if !ok {
		return FlowRun{}, errors.New("flow approval_id not found")
	}
	if approval.Status != "approved" {
		return FlowRun{}, errors.New("flow approval is not approved")
	}
	release := ReleasePacket{
		ReleaseID:    releaseID,
		WorkItemID:   flow.WorkItemID,
		Target:       target,
		Channel:      channel,
		Artifacts:    []string{"founder-flow"},
		Metrics:      map[string]any{},
		RollbackPlan: "run founder rollback",
		ApprovalRefs: []string{approval.ApprovalID},
	}
	if err := s.CreateRelease(release); err != nil {
		return FlowRun{}, err
	}
	deploy, err := s.ExecuteDeploy(deployID, releaseID, notes)
	deployRef := &deploy.DeployID
	record := founderActionRecord{
		kind:   "founder.action.released",
		stage:  StageRelease,
		reason: "founder advanced the approved flow into release and deploy",
		rule:   "founder_action.release",
		evidence: founderActionEvidence(
			"flow:"+flow.FlowID,
			"work:"+flow.WorkItemID,
			"approval:"+approval.ApprovalID,
			"release:"+releaseID,
			"deploy:"+deployID,
			"target:"+target,
			"channel:"+channel,
		),
		extras: map[string]any{
			"action":  "release",
			"flow_id": flow.FlowID,
			"target":  target,
			"channel": channel,
		},
	}
	if err != nil {
		updated, syncErr := s.SyncFlow(flow.FlowID, flow.WorkItemID, flow.ApprovalID, flow.PolicyDecisionID, flow.ExecuteID, flow.VerificationID, &releaseID, deployRef, flow.RollbackID, notes)
		if syncErr != nil {
			return FlowRun{}, syncErr
		}
		record.extras["deploy_status"] = deploy.Status
		record.extras["deploy_error"] = err.Error()
		if eventErr := s.appendFounderActionEvent(updated, record, notes); eventErr != nil {
			return updated, eventErr
		}
		return updated, err
	}
	updated, syncErr := s.SyncFlow(flow.FlowID, flow.WorkItemID, flow.ApprovalID, flow.PolicyDecisionID, flow.ExecuteID, flow.VerificationID, &releaseID, &deploy.DeployID, flow.RollbackID, notes)
	if syncErr != nil {
		return FlowRun{}, syncErr
	}
	record.extras["deploy_status"] = deploy.Status
	if err := s.appendFounderActionEvent(updated, record, notes); err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *Service) RollbackFounderFlow(flowID string, rollbackID string, notes []string) (FlowRun, error) {
	s.mu.Lock()
	flow, ok := s.flow.get(flowID)
	s.mu.Unlock()
	if !ok {
		return FlowRun{}, errors.New("flow_id not found")
	}
	if flow.ReleaseID == nil {
		return FlowRun{}, errors.New("flow release_id not set")
	}
	deployID := ""
	if flow.DeployID != nil {
		deployID = *flow.DeployID
	}
	rollback, err := s.ExecuteRollback(rollbackID, *flow.ReleaseID, deployID, notes)
	rollbackRef := &rollback.RollbackID
	record := founderActionRecord{
		kind:   "founder.action.rolled_back",
		stage:  StageRelease,
		reason: "founder initiated rollback to contain release risk",
		rule:   "founder_action.rollback",
		evidence: founderActionEvidence(
			"flow:"+flow.FlowID,
			"work:"+flow.WorkItemID,
			"release:"+*flow.ReleaseID,
			"deploy:"+deployID,
			"rollback:"+rollbackID,
		),
		extras: map[string]any{
			"action":  "rollback",
			"flow_id": flow.FlowID,
		},
	}
	if err != nil {
		updated, syncErr := s.SyncFlow(flow.FlowID, flow.WorkItemID, flow.ApprovalID, flow.PolicyDecisionID, flow.ExecuteID, flow.VerificationID, flow.ReleaseID, flow.DeployID, rollbackRef, notes)
		if syncErr != nil {
			return FlowRun{}, syncErr
		}
		record.extras["rollback_status"] = rollback.Status
		record.extras["rollback_error"] = err.Error()
		if eventErr := s.appendFounderActionEvent(updated, record, notes); eventErr != nil {
			return updated, eventErr
		}
		return updated, err
	}
	updated, syncErr := s.SyncFlow(flow.FlowID, flow.WorkItemID, flow.ApprovalID, flow.PolicyDecisionID, flow.ExecuteID, flow.VerificationID, flow.ReleaseID, flow.DeployID, &rollback.RollbackID, notes)
	if syncErr != nil {
		return FlowRun{}, syncErr
	}
	record.extras["rollback_status"] = rollback.Status
	if err := s.appendFounderActionEvent(updated, record, notes); err != nil {
		return updated, err
	}
	return updated, nil
}

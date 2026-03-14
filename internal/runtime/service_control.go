package runtime

import (
	"errors"
	"strings"
)

func (s *Service) CreateApproval(item ApprovalItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.approval.list()
	oldState := s.state.current()
	oldEventState := s.event.state()

	if err := s.approval.create(item); err != nil {
		return err
	}
	s.state.updateApprovalsPending(s.approval.count())

	if err := s.event.append(newEvent("approval.created", s.currentActor(), item.DecisionSurface, item.WorkItemID, item.Stage, map[string]any{
		"summary": item.Summary,
	})); err != nil {
		s.approval = newApprovalStore(oldItems)
		s.state = newStateStore(oldState)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.approval = newApprovalStore(oldItems)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) CreatePreflight(item PreflightRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.ModelsUsed = mergeRuntimeModels(item.ModelsUsed, runtimeModelsFromEnv())

	oldItems := s.preflight.list()
	oldEventState := s.event.state()
	if err := s.preflight.create(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("preflight.created", map[string]any{
		"record_id": item.RecordID,
		"status":    item.Status,
		"decision":  item.Decision,
	})); err != nil {
		s.preflight = newPreflightStore(oldItems)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.preflight = newPreflightStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) CreatePolicy(item PolicyDecision) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.policy.list()
	oldEventState := s.event.state()
	if err := s.policy.create(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("policy.created", map[string]any{
		"decision_id": item.DecisionID,
		"decision":    item.Decision,
		"mode":        item.Mode,
	})); err != nil {
		s.policy = newPolicyStore(oldItems)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.policy = newPolicyStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (PolicyDecision, error) {
	mode, decision, reasons := decidePolicy(intent, risk, novelty, tokenClass, requiresApproval)
	item := PolicyDecision{
		DecisionID:       decisionID,
		Intent:           intent,
		Scope:            scope,
		Risk:             risk,
		Novelty:          novelty,
		TokenClass:       tokenClass,
		RequiresApproval: requiresApproval,
		Mode:             mode,
		Decision:         decision,
		Reasons:          reasons,
		CreatedAt:        zeroSafeNow(),
	}
	if err := s.CreatePolicy(item); err != nil {
		return PolicyDecision{}, err
	}
	return item, nil
}

func (s *Service) CreateGatewayCall(item GatewayCall) error {
	return s.createGatewayCallWithPayload(item, s.gatewayAdapter, nil)
}

func (s *Service) createGatewayCallWithAdapter(item GatewayCall, adapter GatewayAdapter) error {
	return s.createGatewayCallWithPayload(item, adapter, nil)
}

func (s *Service) createGatewayCallWithPayload(item GatewayCall, adapter GatewayAdapter, payload map[string]any) error {
	s.mu.Lock()
	decision, ok := s.policy.get(item.DecisionID)
	if !ok {
		s.mu.Unlock()
		return errors.New("gateway decision_id not found")
	}
	requiredMode := canonicalPolicyMode(adapter.RequiredMode())
	if decision.Decision != "go" || canonicalPolicyMode(decision.Mode) != requiredMode {
		s.mu.Unlock()
		return errors.New("gateway requires a " + requiredMode + "-mode go policy decision")
	}
	s.mu.Unlock()

	prepared, err := adapter.Prepare(item, decision)
	if err != nil {
		return err
	}
	if adapter.DispatchEnabled() {
		if payloadAdapter, ok := adapter.(gatewayPayloadAdapter); ok {
			prepared, err = payloadAdapter.DispatchWithPayload(prepared, decision, payload)
		} else {
			prepared, err = adapter.Dispatch(prepared, decision)
		}
		if err != nil {
			return err
		}
	}
	item = prepared

	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.gateway.list()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()
	if err := s.gateway.create(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("gateway.created", gatewayEventData(item))); err != nil {
		s.gateway = newGatewayStore(oldItems)
		return err
	}
	if item.Status == "sent" {
		if err := s.event.append(s.newSystemEvent("gateway.sent", gatewayEventData(item))); err != nil {
			s.gateway = newGatewayStore(oldItems)
			s.event = newEventStoreFromState(oldEventState)
			return err
		}
	}
	if item.Status == "failed" {
		if err := s.event.append(s.newSystemEvent("gateway.failed", gatewayEventData(item))); err != nil {
			s.gateway = newGatewayStore(oldItems)
			s.event = newEventStoreFromState(oldEventState)
			return err
		}
		if suggestion, ok := reviewRoomSuggestionForGatewayCall(item); ok {
			if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
				s.gateway = newGatewayStore(oldItems)
				s.event = newEventStoreFromState(oldEventState)
				s.reviewRoom = newReviewRoomStore(oldRoom)
				return err
			}
		}
	}
	if err := s.persistLocked(); err != nil {
		s.gateway = newGatewayStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func gatewayEventData(item GatewayCall) map[string]any {
	data := map[string]any{
		"call_id":       item.CallID,
		"decision_id":   item.DecisionID,
		"provider":      item.Provider,
		"request_kind":  item.RequestKind,
		"status":        item.Status,
		"attempt_count": item.AttemptCount,
	}
	if item.LastHTTPStatus != nil {
		data["last_http_status"] = *item.LastHTTPStatus
	}
	if item.LastError != nil {
		data["last_error"] = *item.LastError
	}
	if item.ResponsePreview != nil {
		data["response_preview"] = *item.ResponsePreview
	}
	return data
}

func (s *Service) CreateExecute(item ExecuteRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.execute.list()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()
	if err := s.execute.create(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("execute.created", map[string]any{
		"execute_id":         item.ExecuteID,
		"work_item_id":       item.WorkItemID,
		"policy_decision_id": item.PolicyDecisionID,
		"status":             item.Status,
		"mode":               item.Mode,
	})); err != nil {
		s.execute = newExecuteStore(oldItems)
		return err
	}
	if suggestion, ok := reviewRoomSuggestionForExecute(item); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.execute = newExecuteStore(oldItems)
			s.event = newEventStoreFromState(oldEventState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			return err
		}
	}
	if err := s.persistLocked(); err != nil {
		s.execute = newExecuteStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func (s *Service) RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (ExecuteRun, error) {
	s.mu.Lock()
	workItem, workFound := s.work.get(workItemID)
	policyDecision, decisionFound := s.policy.get(policyDecisionID)
	s.mu.Unlock()

	if !workFound {
		return ExecuteRun{}, errors.New("execute work_item_id not found")
	}
	if !decisionFound {
		return ExecuteRun{}, errors.New("execute policy_decision_id not found")
	}
	if policyDecision.Decision == "hold" || policyDecision.Mode == "blocked" {
		finishedAt := zeroSafeNow()
		item := ExecuteRun{
			ExecuteID:        executeID,
			WorkItemID:       workItem.ID,
			PolicyDecisionID: policyDecision.DecisionID,
			Mode:             policyDecision.Mode,
			Status:           "failed",
			Notes:            buildExecuteNotes(policyDecision, baseNotes),
			StartedAt:        finishedAt,
			FinishedAt:       &finishedAt,
		}
		if err := s.CreateExecute(item); err != nil {
			return ExecuteRun{}, err
		}
		return item, errors.New("execute blocked by policy")
	}

	startedAt := zeroSafeNow()
	finishedAt := startedAt
	status := executeStatusFromDecision(policyDecision)
	item := ExecuteRun{
		ExecuteID:        executeID,
		WorkItemID:       workItem.ID,
		PolicyDecisionID: policyDecision.DecisionID,
		Mode:             policyDecision.Mode,
		Status:           status,
		Notes:            buildExecuteNotes(policyDecision, baseNotes),
		StartedAt:        startedAt,
		FinishedAt:       &finishedAt,
	}
	if status == "recorded" {
		item.FinishedAt = nil
	}
	if err := s.CreateExecute(item); err != nil {
		return ExecuteRun{}, err
	}
	return item, nil
}

func (s *Service) CreateVerification(item VerificationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.verify.list()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()
	if err := s.verify.create(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("verification.created", map[string]any{
		"record_id": item.RecordID,
		"scope":     item.Scope,
		"status":    item.Status,
	})); err != nil {
		s.verify = newVerificationStore(oldItems)
		return err
	}
	if suggestion, ok := reviewRoomSuggestionForVerification(item); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.verify = newVerificationStore(oldItems)
			s.event = newEventStoreFromState(oldEventState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			return err
		}
	} else if item.Status == "passed" {
		scopeEvidence := "scope:" + strings.TrimSpace(item.Scope)
		if strings.TrimSpace(item.Scope) != "" {
			room, resolved, err := autoResolveReviewRoomItemsByEvidenceLocked(s.reviewRoom, "verification.failed", scopeEvidence, &ReviewRoomResolution{
				Action:   "resolve",
				Reason:   "later verification passed for the same scope",
				Rule:     "review_room.auto.verification_failed_clear",
				Evidence: []string{"verification:" + item.RecordID, scopeEvidence, "status:passed"},
			})
			if err != nil {
				s.verify = newVerificationStore(oldItems)
				s.event = newEventStoreFromState(oldEventState)
				s.reviewRoom = newReviewRoomStore(oldRoom)
				return err
			}
			if resolved > 0 {
				s.reviewRoom = newReviewRoomStore(room)
			}
		}
	}
	if err := s.persistLocked(); err != nil {
		s.verify = newVerificationStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func (s *Service) RunVerification(recordID string, scope string, workdir string, command []string, baseNotes []string) (VerificationRecord, error) {
	normalizedCommand, err := normalizeVerificationCommand(command)
	if err != nil {
		return VerificationRecord{}, err
	}
	startedAt := zeroSafeNow()
	output, runErr := s.verifyAdapter.Run(normalizedCommand, workdir)
	finishedAt := zeroSafeNow()
	item := VerificationRecord{
		RecordID:   recordID,
		Scope:      scope,
		Command:    append([]string{}, normalizedCommand...),
		Status:     verificationStatusFromError(runErr),
		Notes:      append(buildVerificationNotes(baseNotes, output, runErr), "adapter:"+s.verifyAdapter.Name()),
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}
	createErr := s.CreateVerification(item)
	if runErr != nil {
		if createErr != nil {
			return VerificationRecord{}, createErr
		}
		return item, runErr
	}
	if createErr != nil {
		return VerificationRecord{}, createErr
	}
	return item, nil
}

func (s *Service) ResolveApproval(approvalID string, status string) (ApprovalItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status != "approved" && status != "rejected" {
		return ApprovalItem{}, errors.New("approval resolve status must be approved or rejected")
	}

	oldItems := s.approval.list()
	oldState := s.state.current()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()

	item, err := s.approval.resolve(approvalID, status, zeroSafeNow())
	if err != nil {
		return ApprovalItem{}, err
	}
	s.state.updateApprovalsPending(s.approval.count())

	if err := s.event.append(newEvent("approval.resolved", s.currentActor(), item.DecisionSurface, item.WorkItemID, item.Stage, map[string]any{
		"status": item.Status,
	})); err != nil {
		s.approval = newApprovalStore(oldItems)
		s.state = newStateStore(oldState)
		return ApprovalItem{}, err
	}
	if suggestion, ok := reviewRoomSuggestionForApproval(item); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.approval = newApprovalStore(oldItems)
			s.state = newStateStore(oldState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return ApprovalItem{}, err
		}
	}
	if item.Status == "approved" {
		resolution := &ReviewRoomResolution{Action: "resolve", Reason: "approved resolution cleared the prior rejection review item", Rule: "review_room.auto.approval_rejection_cleared", Evidence: []string{"approval:" + item.ApprovalID}}
		updated, _, changed, err := autoResolveReviewRoomItemLocked(s.reviewRoom, "approval.rejected", &item.ApprovalID, resolution)
		if err != nil {
			s.approval = newApprovalStore(oldItems)
			s.state = newStateStore(oldState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return ApprovalItem{}, err
		}
		if changed {
			s.reviewRoom = newReviewRoomStore(updated)
			if err := s.event.append(s.newSystemEvent("review_room.item_auto_resolved", map[string]any{"source": "approval.rejected", "ref": item.ApprovalID, "action": "resolve"})); err != nil {
				s.approval = newApprovalStore(oldItems)
				s.state = newStateStore(oldState)
				s.reviewRoom = newReviewRoomStore(oldRoom)
				s.event = newEventStoreFromState(oldEventState)
				return ApprovalItem{}, err
			}
		}
	}
	if err := s.persistLocked(); err != nil {
		s.approval = newApprovalStore(oldItems)
		s.state = newStateStore(oldState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return ApprovalItem{}, err
	}
	return item, nil
}

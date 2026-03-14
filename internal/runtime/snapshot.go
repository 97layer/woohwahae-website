package runtime

import (
	"errors"
	"os"
)

func (s *Service) Snapshot() SnapshotPacket {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state.updateWorkItemsActive(s.work.count())
	s.state.updateApprovalsPending(s.approval.count())
	archive, err := s.disk.readEventArchive()
	if err != nil {
		archive = []EventEnvelope{}
	}
	continuityRecords := s.continuity.list()
	sessionCheckpoint := continuityRecordForCompatibility(continuityRecords)
	evidence := dedupeEvents(append(append(archive, s.event.pendingArchiveList()...), s.event.recent()...))
	return SnapshotPacket{
		GeneratedAt:       zeroSafeNow(),
		CompanyState:      s.state.current(),
		SystemMemory:      s.memory.current(),
		ContinuityRecords: continuityRecords,
		SessionCheckpoint: sessionCheckpoint,
		Auth:              s.auth.status(),
		Branches:          s.branch.list(),
		Proposals:         s.proposal.list(),
		AgentJobs:         s.agentJob.list(),
		WorkItems:         s.work.list(),
		Flows:             s.flow.list(),
		Approvals:         s.approval.list(),
		Releases:          s.release.list(),
		Deploys:           s.deploy.list(),
		Rollbacks:         s.rollback.list(),
		Targets:           s.target.list(),
		Events:            evidence,
		Observations:      s.observation.list(),
		ReviewRoom:        s.reviewRoom.current(),
		Preflights:        s.preflight.list(),
		Policies:          s.policy.list(),
		GatewayCalls:      s.gateway.list(),
		Executes:          s.execute.list(),
		Verifications:     s.verify.list(),
	}
}

func (s *Service) ImportSnapshot(packet SnapshotPacket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	packet = normalizeSnapshot(packet)
	if len(packet.ContinuityRecords) == 0 && packet.SessionCheckpoint != nil {
		packet.ContinuityRecords = []ContinuityRecord{continuityRecordFromSessionCheckpoint(*packet.SessionCheckpoint)}
	}
	if err := validateSnapshot(packet); err != nil {
		return err
	}

	state := packet.CompanyState
	state.WorkItemsActive = len(packet.WorkItems)
	pending := 0
	for _, item := range packet.Approvals {
		if item.Status == "pending" {
			pending++
		}
	}
	state.ApprovalsPending = pending
	if len(packet.Releases) > 0 {
		last := packet.Releases[len(packet.Releases)-1]
		state.LastReleaseAt = last.ReleasedAt
	}
	if len(packet.Deploys) > 0 {
		last := packet.Deploys[len(packet.Deploys)-1]
		if last.Status == "failed" {
			state.DeployHealth = "degraded"
		} else {
			state.DeployHealth = "ready"
		}
	}

	s.state = newStateStore(state)
	s.branch = newBranchStore(packet.Branches)
	s.proposal = newProposalStore(packet.Proposals)
	s.agentJob = newAgentJobStore(packet.AgentJobs)
	s.work = newWorkStore(packet.WorkItems)
	s.flow = newFlowStore(packet.Flows)
	s.approval = newApprovalStore(packet.Approvals)
	s.release = newReleaseStore(packet.Releases)
	s.deploy = newDeployStore(packet.Deploys)
	s.rollback = newRollbackStore(packet.Rollbacks)
	s.target = newTargetStore(packet.Targets)
	_ = os.Remove(s.disk.eventArchivePath())
	s.event = newEventStoreFromEvidence(packet.Events)
	s.observation = newObservationStore(packet.Observations)
	s.continuity = newContinuityStore(packet.ContinuityRecords)
	s.reviewRoom = newReviewRoomStore(packet.ReviewRoom)
	s.preflight = newPreflightStore(packet.Preflights)
	s.policy = newPolicyStore(packet.Policies)
	s.gateway = newGatewayStore(packet.GatewayCalls)
	s.execute = newExecuteStore(packet.Executes)
	s.verify = newVerificationStore(packet.Verifications)
	s.memory = newMemoryStore(packet.SystemMemory)
	s.auth = newAuthStore(defaultAuthConfig())

	if err := s.persistLocked(); err != nil {
		return err
	}
	return nil
}

func validateSnapshot(packet SnapshotPacket) error {
	branchIDs := map[string]bool{}
	for _, item := range packet.Branches {
		if err := item.Validate(); err != nil {
			return err
		}
		branchIDs[item.BranchID] = true
	}
	for _, item := range packet.Proposals {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.BranchID != nil && !branchIDs[*item.BranchID] {
			return errors.New("snapshot proposal branch_id not found")
		}
	}
	for _, item := range packet.AgentJobs {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.BranchID != nil && !branchIDs[*item.BranchID] {
			return errors.New("snapshot agent job branch_id not found")
		}
	}
	for _, item := range packet.WorkItems {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.BranchID != nil && !branchIDs[*item.BranchID] {
			return errors.New("snapshot work item branch_id not found")
		}
	}
	for _, item := range packet.Observations {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.ContinuityRecords {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if err := packet.ReviewRoom.Validate(); err != nil {
		return err
	}
	for _, item := range packet.Flows {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Approvals {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Releases {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Deploys {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Rollbacks {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Targets {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Events {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Preflights {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Policies {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.GatewayCalls {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Executes {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range packet.Verifications {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if err := packet.CompanyState.Validate(); err != nil {
		return err
	}
	if err := packet.SystemMemory.Validate(); err != nil {
		return err
	}
	if packet.SessionCheckpoint != nil {
		if err := packet.SessionCheckpoint.Validate(); err != nil {
			return err
		}
	}
	if packet.Auth.WriteAuthEnabled {
		return errors.New("snapshot import does not accept active auth tokens")
	}
	return nil
}

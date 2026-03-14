package runtime

import (
	"errors"
	"strings"
)

func (s *Service) CreateRelease(item ReleasePacket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item.ReleasedAt == nil || item.ReleasedAt.IsZero() {
		releasedAt := zeroSafeNow()
		item.ReleasedAt = &releasedAt
	}
	if item.Artifacts == nil {
		item.Artifacts = []string{}
	}
	if item.Metrics == nil {
		item.Metrics = map[string]any{}
	}
	if item.ApprovalRefs == nil {
		item.ApprovalRefs = []string{}
	}
	if len(item.ApprovalRefs) == 0 {
		return errors.New("release approval_refs must not be empty")
	}
	for _, approvalID := range item.ApprovalRefs {
		approval, ok := s.approval.get(approvalID)
		if !ok {
			return errors.New("release approval_ref not found")
		}
		if approval.Status != "approved" {
			return errors.New("release requires approved approval refs")
		}
		if approval.WorkItemID != item.WorkItemID {
			return errors.New("release approval_ref work_item mismatch")
		}
	}

	oldItems := s.release.list()
	oldState := s.state.current()
	oldEventState := s.event.state()

	if err := s.release.create(item); err != nil {
		return err
	}
	s.state.updateLastReleaseAt(item.ReleasedAt)

	if err := s.event.append(newEvent("release.created", s.currentActor(), SurfaceAPI, item.WorkItemID, StageRelease, map[string]any{
		"target":  item.Target,
		"channel": item.Channel,
	})); err != nil {
		s.release = newReleaseStore(oldItems)
		s.state = newStateStore(oldState)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.release = newReleaseStore(oldItems)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) CreateDeploy(item DeployRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	release, ok := s.release.get(item.ReleaseID)
	if !ok {
		return errors.New("deploy release_id not found")
	}
	if item.Target == "" {
		item.Target = release.Target
	}
	if item.Status == "" {
		item.Status = "succeeded"
	}
	if item.Notes == nil {
		item.Notes = []string{}
	}
	if item.StartedAt.IsZero() {
		item.StartedAt = zeroSafeNow()
	}
	if item.FinishedAt == nil || item.FinishedAt.IsZero() {
		finishedAt := zeroSafeNow()
		item.FinishedAt = &finishedAt
	}

	oldItems := s.deploy.list()
	oldState := s.state.current()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()

	if err := s.deploy.create(item); err != nil {
		return err
	}
	s.state.updateDeployHealth(item.Status)

	if err := s.event.append(newEvent("deploy.created", s.currentActor(), SurfaceAPI, release.WorkItemID, StageRelease, map[string]any{
		"target": item.Target,
		"status": item.Status,
	})); err != nil {
		s.deploy = newDeployStore(oldItems)
		s.state = newStateStore(oldState)
		return err
	}
	if suggestion, ok := reviewRoomSuggestionForDeploy(item); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.deploy = newDeployStore(oldItems)
			s.state = newStateStore(oldState)
			s.event = newEventStoreFromState(oldEventState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			return err
		}
	}
	if err := s.persistLocked(); err != nil {
		s.deploy = newDeployStore(oldItems)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func (s *Service) CreateRollback(item RollbackRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	release, ok := s.release.get(item.ReleaseID)
	if !ok {
		return errors.New("rollback release_id not found")
	}
	if item.DeployID != nil && strings.TrimSpace(*item.DeployID) != "" {
		if _, ok := s.deploy.get(strings.TrimSpace(*item.DeployID)); !ok {
			return errors.New("rollback deploy_id not found")
		}
	}
	if item.Target == "" {
		item.Target = release.Target
	}
	if item.Status == "" {
		item.Status = "succeeded"
	}
	if item.Notes == nil {
		item.Notes = []string{}
	}
	if item.StartedAt.IsZero() {
		item.StartedAt = zeroSafeNow()
	}
	if item.FinishedAt == nil || item.FinishedAt.IsZero() {
		finishedAt := zeroSafeNow()
		item.FinishedAt = &finishedAt
	}

	oldItems := s.rollback.list()
	oldState := s.state.current()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()

	if err := s.rollback.create(item); err != nil {
		return err
	}
	if item.Status == "succeeded" {
		s.state.updateDeployHealth("succeeded")
	} else if item.Status == "failed" {
		s.state.updateDeployHealth("failed")
	}
	if err := s.event.append(newEvent("rollback.created", s.currentActor(), SurfaceAPI, release.WorkItemID, StageRelease, map[string]any{
		"target": item.Target,
		"status": item.Status,
	})); err != nil {
		s.rollback = newRollbackStore(oldItems)
		s.state = newStateStore(oldState)
		return err
	}
	if suggestion, ok := reviewRoomSuggestionForRollback(item); ok {
		if err := s.autoOpenReviewRoomItemLocked(suggestion); err != nil {
			s.rollback = newRollbackStore(oldItems)
			s.state = newStateStore(oldState)
			s.event = newEventStoreFromState(oldEventState)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			return err
		}
	}
	if err := s.persistLocked(); err != nil {
		s.rollback = newRollbackStore(oldItems)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func (s *Service) PutTarget(item DeployTarget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldItems := s.target.list()
	oldEventState := s.event.state()
	if err := s.target.put(item); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("target.updated", map[string]any{
		"target_id": item.TargetID,
	})); err != nil {
		s.target = newTargetStore(oldItems)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.target = newTargetStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) ExecuteDeploy(deployID string, releaseID string, baseNotes []string) (DeployRun, error) {
	s.mu.Lock()
	release, ok := s.release.get(releaseID)
	if !ok {
		s.mu.Unlock()
		return DeployRun{}, errors.New("deploy release_id not found")
	}
	target, ok := s.target.get(release.Target)
	s.mu.Unlock()
	if !ok {
		return DeployRun{}, errDeployTargetNotFound
	}

	startedAt := zeroSafeNow()
	output, execErr := s.deployAdapter.Run(target)
	finishedAt := zeroSafeNow()
	item := DeployRun{
		DeployID:   deployID,
		ReleaseID:  releaseID,
		Target:     release.Target,
		Status:     statusFromError(execErr),
		Notes:      append(buildDeployNotes(baseNotes, output, execErr), "adapter:"+s.deployAdapter.Name()),
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}
	createErr := s.CreateDeploy(item)
	if execErr != nil {
		if createErr != nil {
			return DeployRun{}, createErr
		}
		return item, execErr
	}
	if createErr != nil {
		return DeployRun{}, createErr
	}
	return item, nil
}

func (s *Service) ExecuteRollback(rollbackID string, releaseID string, deployID string, baseNotes []string) (RollbackRun, error) {
	s.mu.Lock()
	release, ok := s.release.get(releaseID)
	if !ok {
		s.mu.Unlock()
		return RollbackRun{}, errors.New("rollback release_id not found")
	}
	target, ok := s.target.get(release.Target)
	s.mu.Unlock()
	if !ok {
		return RollbackRun{}, errDeployTargetNotFound
	}

	startedAt := zeroSafeNow()
	output, execErr := s.rollbackAdapter.Run(target)
	finishedAt := zeroSafeNow()
	var deployRef *string
	if strings.TrimSpace(deployID) != "" {
		value := strings.TrimSpace(deployID)
		deployRef = &value
	}
	item := RollbackRun{
		RollbackID: rollbackID,
		ReleaseID:  releaseID,
		DeployID:   deployRef,
		Target:     release.Target,
		Status:     statusFromError(execErr),
		Notes:      append(buildDeployNotes(baseNotes, output, execErr), "adapter:"+s.rollbackAdapter.Name()),
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}
	createErr := s.CreateRollback(item)
	if execErr != nil {
		if createErr != nil {
			return RollbackRun{}, createErr
		}
		return item, execErr
	}
	if createErr != nil {
		return RollbackRun{}, createErr
	}
	return item, nil
}

func (s *Service) ListReleases() []ReleasePacket {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.release.list()
}

func (s *Service) ListDeploys() []DeployRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deploy.list()
}

func (s *Service) ListRollbacks() []RollbackRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rollback.list()
}

func (s *Service) ListTargets() []DeployTarget {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.target.list()
}

package runtime

import (
	"errors"
	"strings"
)

func (s *Service) CreateAgentJob(item AgentJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureActiveBranchLocked(item.BranchID); err != nil {
		return err
	}
	oldJobs := s.agentJob.list()
	oldEventState := s.event.state()
	if _, err := s.createAgentJobLocked(item); err != nil {
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.agentJob = newAgentJobStore(oldJobs)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func normalizeAgentJobItem(item AgentJob) AgentJob {
	if item.Status == "" {
		item.Status = "queued"
	}
	if item.Notes == nil {
		item.Notes = []string{}
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = zeroSafeNow()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	return item
}

func (s *Service) createAgentJobLocked(item AgentJob) (AgentJob, error) {
	item = normalizeAgentJobItem(item)
	if err := s.agentJob.create(item); err != nil {
		return AgentJob{}, err
	}
	if err := s.event.append(newEvent("agent_job.created", s.currentActor(), item.Surface, item.JobID, item.Stage, map[string]any{
		"kind":    item.Kind,
		"role":    item.Role,
		"status":  item.Status,
		"summary": item.Summary,
	})); err != nil {
		return AgentJob{}, err
	}
	return item, nil
}

func (s *Service) UpdateAgentJob(jobID string, status string, notes []string, result map[string]any) (AgentJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	updated, _, err := s.updateAgentJobLocked(jobID, status, notes, result, "agent_job.updated")
	if err != nil {
		return AgentJob{}, err
	}
	return updated, nil
}

func (s *Service) updateAgentJobLocked(jobID string, status string, notes []string, result map[string]any, eventKind string) (AgentJob, EventEnvelope, error) {
	if _, ok := s.agentJob.get(jobID); !ok {
		return AgentJob{}, EventEnvelope{}, errors.New("job_id not found")
	}
	oldJobs := s.agentJob.list()
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()
	updated, err := s.agentJob.update(jobID, status, notes, result, zeroSafeNow())
	if err != nil {
		return AgentJob{}, EventEnvelope{}, err
	}
	data := map[string]any{
		"kind":   updated.Kind,
		"role":   updated.Role,
		"status": updated.Status,
	}
	if updated.Result != nil {
		data["result"] = updated.Result
		copyAgentJobResultProvenance(data, updated.Result)
	}
	event := newEvent(eventKind, s.currentActor(), updated.Surface, updated.JobID, updated.Stage, data)
	if err := s.event.append(event); err != nil {
		s.agentJob = newAgentJobStore(oldJobs)
		return AgentJob{}, EventEnvelope{}, err
	}
	if suggestion, ok := reviewRoomSuggestionForAgentJob(updated); ok {
		room, changed, err := ensureReviewRoomItem(s.reviewRoom.current(), "open", suggestion)
		if err != nil {
			s.agentJob = newAgentJobStore(oldJobs)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return AgentJob{}, EventEnvelope{}, err
		}
		if changed {
			s.reviewRoom = newReviewRoomStore(room)
		}
	} else if updated.Status != "failed" {
		ref := updated.JobID
		room, _, changed, err := autoResolveReviewRoomItemLocked(s.reviewRoom, "agent_job.failed", &ref, &ReviewRoomResolution{
			Action:   "resolve",
			Reason:   "job left failed state after status transition",
			Rule:     "review_room.auto.agent_job_failed_clear",
			Evidence: []string{"job:" + updated.JobID, "status:" + updated.Status},
		})
		if err != nil {
			s.agentJob = newAgentJobStore(oldJobs)
			s.reviewRoom = newReviewRoomStore(oldRoom)
			s.event = newEventStoreFromState(oldEventState)
			return AgentJob{}, EventEnvelope{}, err
		}
		if changed {
			s.reviewRoom = newReviewRoomStore(room)
		}
	}
	if err := s.persistLocked(); err != nil {
		s.agentJob = newAgentJobStore(oldJobs)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		return AgentJob{}, EventEnvelope{}, err
	}
	return updated, event, nil
}

func (s *Service) applyAgentJobResultPatchLocked(jobID string, patch map[string]any) (AgentJob, error) {
	current, ok := s.agentJob.get(jobID)
	if !ok {
		return AgentJob{}, errors.New("job_id not found")
	}
	oldJobs := s.agentJob.list()
	merged := cloneJSONObject(current.Result)
	if merged == nil {
		merged = map[string]any{}
	}
	for key, value := range cloneJSONObject(patch) {
		merged[key] = value
	}
	updated, err := s.agentJob.update(jobID, current.Status, current.Notes, merged, zeroSafeNow())
	if err != nil {
		return AgentJob{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.agentJob = newAgentJobStore(oldJobs)
		return AgentJob{}, err
	}
	return updated, nil
}

func copyAgentJobResultProvenance(data map[string]any, result map[string]any) {
	if data == nil || result == nil {
		return
	}
	for _, item := range []struct {
		From string
		To   string
	}{
		{From: "dispatch_actor", To: "dispatch_actor"},
		{From: "reported_by_actor", To: "report_actor"},
		{From: "completion_mode", To: "completion_mode"},
		{From: "execution_origin", To: "execution_origin"},
		{From: "dispatch_transport", To: "dispatch_transport"},
	} {
		value := strings.TrimSpace(agentJobResultString(result[item.From]))
		if value == "" {
			continue
		}
		data[item.To] = value
	}
}

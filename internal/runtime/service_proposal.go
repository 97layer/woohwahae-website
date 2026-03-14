package runtime

import "errors"

func (s *Service) CreateProposal(item ProposalItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureActiveBranchLocked(item.BranchID); err != nil {
		return err
	}
	oldItems := s.proposal.list()
	oldEventState := s.event.state()
	if _, err := s.createProposalLocked(item); err != nil {
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.proposal = newProposalStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func normalizeProposalItem(item ProposalItem) ProposalItem {
	if item.Summary == "" {
		item.Summary = item.Title
	}
	if item.Status == "" {
		item.Status = "proposed"
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

func (s *Service) createProposalLocked(item ProposalItem) (ProposalItem, error) {
	item = normalizeProposalItem(item)
	if err := s.proposal.create(item); err != nil {
		return ProposalItem{}, err
	}
	if err := s.event.append(newEvent("proposal.created", s.currentActor(), item.Surface, item.ProposalID, StageDiscover, map[string]any{
		"title":   item.Title,
		"status":  item.Status,
		"summary": item.Summary,
	})); err != nil {
		return ProposalItem{}, err
	}
	return item, nil
}

func (s *Service) PromoteProposal(proposalID string, workID string) (ProposalItem, WorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	proposal, ok := s.proposal.get(proposalID)
	if !ok {
		return ProposalItem{}, WorkItem{}, errors.New("proposal_id not found")
	}
	if proposal.Status != "proposed" {
		return ProposalItem{}, WorkItem{}, errors.New("proposal is not open")
	}
	oldProposals := s.proposal.list()
	oldWork := s.work.list()
	oldState := s.state.current()
	oldEventState := s.event.state()

	work := WorkItem{
		ID:               workID,
		BranchID:         normalizeRef(proposal.BranchID),
		Title:            proposal.Title,
		Intent:           proposal.Intent,
		Stage:            StageDiscover,
		Surface:          proposal.Surface,
		Pack:             "proposal",
		Priority:         proposal.Priority,
		Risk:             proposal.Risk,
		RequiresApproval: false,
		Payload:          map[string]any{"proposal_id": proposal.ProposalID},
		CorrelationID:    proposal.ProposalID,
		CreatedAt:        zeroSafeNow(),
	}
	if err := s.work.create(work); err != nil {
		return ProposalItem{}, WorkItem{}, err
	}
	updatedProposal, err := s.proposal.promote(proposalID, workID, zeroSafeNow())
	if err != nil {
		s.work = newWorkStore(oldWork)
		return ProposalItem{}, WorkItem{}, err
	}
	s.state.updateWorkItemsActive(s.work.count())
	if err := s.event.append(newEvent("proposal.promoted", s.currentActor(), proposal.Surface, proposal.ProposalID, StageDiscover, map[string]any{
		"proposal_id":  proposal.ProposalID,
		"work_item_id": work.ID,
	})); err != nil {
		s.proposal = newProposalStore(oldProposals)
		s.work = newWorkStore(oldWork)
		s.state = newStateStore(oldState)
		return ProposalItem{}, WorkItem{}, err
	}
	if err := s.event.append(newEvent("work_item.created", s.currentActor(), work.Surface, work.ID, work.Stage, map[string]any{
		"title":  work.Title,
		"source": "proposal.promote",
	})); err != nil {
		s.proposal = newProposalStore(oldProposals)
		s.work = newWorkStore(oldWork)
		s.state = newStateStore(oldState)
		return ProposalItem{}, WorkItem{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.proposal = newProposalStore(oldProposals)
		s.work = newWorkStore(oldWork)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		return ProposalItem{}, WorkItem{}, err
	}
	return updatedProposal, work, nil
}

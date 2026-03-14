package runtime

import (
	"errors"
	"strings"
)

func (s *Service) ListBranches() []Branch {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.branch.list()
}

func (s *Service) CreateBranch(item Branch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item = normalizeBranch(item)
	if item.ParentBranchID != nil {
		parent, ok := s.branch.get(strings.TrimSpace(*item.ParentBranchID))
		if !ok {
			return errors.New("branch parent_branch_id not found")
		}
		item.RootBranchID = parent.RootBranchID
	}
	if item.RootBranchID == "" {
		item.RootBranchID = item.BranchID
	}

	oldBranches := s.branch.list()
	oldEventState := s.event.state()
	if err := s.branch.create(item); err != nil {
		return err
	}
	if err := s.event.append(newEvent("branch.created", s.currentActor(), item.Surface, item.BranchID, item.Stage, map[string]any{
		"branch_id":      item.BranchID,
		"root_branch_id": item.RootBranchID,
		"status":         item.Status,
		"title":          item.Title,
	})); err != nil {
		s.branch = newBranchStore(oldBranches)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.branch = newBranchStore(oldBranches)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) MergeBranch(branchID string, targetBranchID string, notes []string) (Branch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldBranches := s.branch.list()
	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()
	updated, err := s.branch.merge(branchID, targetBranchID, notes, zeroSafeNow())
	if err != nil {
		return Branch{}, err
	}
	if err := s.event.append(newEvent("branch.merged", s.currentActor(), updated.Surface, updated.BranchID, updated.Stage, map[string]any{
		"branch_id":              updated.BranchID,
		"merge_target_branch_id": strings.TrimSpace(targetBranchID),
		"root_branch_id":         updated.RootBranchID,
		"status":                 updated.Status,
	})); err != nil {
		s.branch = newBranchStore(oldBranches)
		return Branch{}, err
	}
	ref := "branch:" + updated.BranchID
	item := newStructuredReviewRoomItem(
		"Branch `"+updated.BranchID+"` merged into `"+strings.TrimSpace(targetBranchID)+"`; reconcile branch-bound proposal/job/work references before archival cleanup.",
		"agenda",
		"low",
		"branch.merge",
		&ref,
		&ReviewRoomRationale{
			Trigger: "branch.merged",
			Reason:  "merged branches should close their attached execution lane cleanly before archival cleanup",
			Rule:    "review_room.branch_merge_reconcile",
		},
		[]string{"branch:" + updated.BranchID, "merge_target:" + strings.TrimSpace(targetBranchID)},
	)
	if err := s.autoOpenReviewRoomItemLocked(item); err != nil {
		s.branch = newBranchStore(oldBranches)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return Branch{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.branch = newBranchStore(oldBranches)
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return Branch{}, err
	}
	return updated, nil
}

func (s *Service) ensureActiveBranchLocked(branchID *string) error {
	if branchID == nil {
		return nil
	}
	value := strings.TrimSpace(*branchID)
	if value == "" {
		return errors.New("branch_id must not be empty")
	}
	branch, ok := s.branch.get(value)
	if !ok {
		return errors.New("branch_id not found")
	}
	if branch.Status != "active" {
		return errors.New("branch_id must reference an active branch")
	}
	return nil
}

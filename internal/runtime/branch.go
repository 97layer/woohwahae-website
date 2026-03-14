package runtime

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type branchStore struct {
	mu    sync.RWMutex
	items []Branch
}

func newBranchStore(items []Branch) *branchStore {
	return &branchStore{items: normalizeBranches(items)}
}

func normalizeBranch(item Branch) Branch {
	if item.RootBranchID == "" {
		item.RootBranchID = item.BranchID
	}
	if item.Summary == "" {
		item.Summary = item.Title
	}
	if item.Stage == "" {
		item.Stage = StageCompose
	}
	if item.Surface == "" {
		item.Surface = SurfaceAPI
	}
	if item.Status == "" {
		item.Status = "active"
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
	item.ParentBranchID = normalizeRef(item.ParentBranchID)
	item.BasisRef = normalizeRef(item.BasisRef)
	item.MergeTargetBranchID = normalizeRef(item.MergeTargetBranchID)
	return item
}

func normalizeBranches(items []Branch) []Branch {
	if items == nil {
		return []Branch{}
	}
	out := make([]Branch, len(items))
	for i, item := range items {
		out[i] = normalizeBranch(item)
	}
	return out
}

func cloneBranches(items []Branch) []Branch {
	out := make([]Branch, len(items))
	copy(out, items)
	return out
}

func activeBranches(items []Branch, max int) []Branch {
	filtered := make([]Branch, 0, len(items))
	for _, item := range items {
		if item.Status == "active" {
			filtered = append(filtered, item)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].UpdatedAt.Equal(filtered[j].UpdatedAt) {
			return filtered[i].BranchID < filtered[j].BranchID
		}
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})
	if max > 0 && len(filtered) > max {
		filtered = filtered[:max]
	}
	return cloneBranches(filtered)
}

func (s *branchStore) list() []Branch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneBranches(s.items)
}

func (s *branchStore) get(branchID string) (Branch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	needle := strings.TrimSpace(branchID)
	for _, item := range s.items {
		if item.BranchID == needle {
			return item, true
		}
	}
	return Branch{}, false
}

func (s *branchStore) create(item Branch) error {
	item = normalizeBranch(item)
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.items {
		if existing.BranchID == item.BranchID {
			return errors.New("branch_id already exists")
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *branchStore) update(item Branch) error {
	item = normalizeBranch(item)
	if err := item.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.items {
		if existing.BranchID == item.BranchID {
			s.items[index] = item
			return nil
		}
	}
	return errors.New("branch_id not found")
}

func (s *branchStore) merge(branchID string, targetBranchID string, notes []string, now time.Time) (Branch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceIndex := -1
	targetIndex := -1
	branchID = strings.TrimSpace(branchID)
	targetBranchID = strings.TrimSpace(targetBranchID)
	for index, item := range s.items {
		switch item.BranchID {
		case branchID:
			sourceIndex = index
		case targetBranchID:
			targetIndex = index
		}
	}
	if sourceIndex < 0 {
		return Branch{}, errors.New("branch_id not found")
	}
	if targetIndex < 0 {
		return Branch{}, errors.New("merge target branch_id not found")
	}
	if branchID == targetBranchID {
		return Branch{}, errors.New("branch merge target must differ from branch_id")
	}
	item := s.items[sourceIndex]
	if item.Status != "active" {
		return Branch{}, errors.New("branch must be active before merge")
	}
	item.Status = "merged"
	item.MergeTargetBranchID = &targetBranchID
	item.MergedAt = &now
	item.UpdatedAt = now
	item.Notes = append([]string{}, notes...)
	if err := item.Validate(); err != nil {
		return Branch{}, err
	}
	s.items[sourceIndex] = item
	return item, nil
}

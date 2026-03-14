package runtime

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const reviewRoomRuntimeFile = "review_room.json"

var (
	ErrInvalidReviewRoomSection = errors.New("review room section is invalid")
	ErrMissingReviewRoomItem    = errors.New("review room item is required")
	ErrDuplicateReviewRoomItem  = errors.New("review room item already exists in section")
	ErrInvalidReviewRoomAction  = errors.New("review room action is invalid")
	ErrReviewRoomItemNotFound   = errors.New("review room item not found")
)

type reviewRoomStore struct {
	mu   sync.RWMutex
	room ReviewRoom
}

func defaultReviewRoom(source string) ReviewRoom {
	return ReviewRoom{
		Source:   source,
		Accepted: []ReviewRoomItem{},
		Open:     []ReviewRoomItem{},
		Deferred: []ReviewRoomItem{},
		Issues:   []string{},
	}
}

func newReviewRoomStore(room ReviewRoom) *reviewRoomStore {
	room = normalizeReviewRoom(room)
	return &reviewRoomStore{room: room}
}

func (s *reviewRoomStore) current() ReviewRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyReviewRoom(s.room)
}

func (s *reviewRoomStore) add(section string, item ReviewRoomItem) (ReviewRoom, error) {
	section = normalizeReviewSection(section)
	if section == "" {
		return ReviewRoom{}, ErrInvalidReviewRoomSection
	}
	item = normalizeReviewRoomItem(item)
	if item.Text == "" {
		return ReviewRoom{}, ErrMissingReviewRoomItem
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	room := copyReviewRoom(s.room)
	items := reviewRoomSection(&room, section)
	for _, existing := range *items {
		if existing.Text == item.Text {
			return ReviewRoom{}, ErrDuplicateReviewRoomItem
		}
	}
	updated, added, err := ensureReviewRoomItem(room, section, item)
	if err != nil {
		return ReviewRoom{}, err
	}
	if !added {
		return ReviewRoom{}, ErrDuplicateReviewRoomItem
	}
	s.room = updated
	return copyReviewRoom(updated), nil
}

func normalizeReviewRoom(room ReviewRoom) ReviewRoom {
	if strings.TrimSpace(room.Source) == "" {
		room.Source = reviewRoomRuntimeFile
	}
	room.Accepted = normalizeReviewRoomItems(room.Accepted)
	room.Open = normalizeReviewRoomItems(room.Open)
	room.Deferred = normalizeReviewRoomItems(room.Deferred)
	if room.Issues == nil {
		room.Issues = []string{}
	}
	return room
}

func copyReviewRoom(room ReviewRoom) ReviewRoom {
	room = normalizeReviewRoom(room)
	out := room
	out.Accepted = normalizeReviewRoomItems(room.Accepted)
	out.Open = normalizeReviewRoomItems(room.Open)
	out.Deferred = normalizeReviewRoomItems(room.Deferred)
	out.Issues = append([]string{}, room.Issues...)
	return out
}

func reviewRoomSection(room *ReviewRoom, section string) *[]ReviewRoomItem {
	switch section {
	case "accepted":
		return &room.Accepted
	case "open":
		return &room.Open
	case "deferred":
		return &room.Deferred
	default:
		return nil
	}
}

func normalizeReviewSection(section string) string {
	section = strings.TrimSpace(strings.ToLower(section))
	switch section {
	case "accepted", "open", "deferred":
		return section
	default:
		return ""
	}
}

func normalizeReviewAction(action string) string {
	action = strings.TrimSpace(strings.ToLower(action))
	switch action {
	case "accept", "defer", "resolve":
		return action
	default:
		return ""
	}
}

func (s *reviewRoomStore) transitionStructured(action string, item string, resolution *ReviewRoomResolution) (ReviewRoom, *ReviewRoomItem, error) {
	action = normalizeReviewAction(action)
	if action == "" {
		return ReviewRoom{}, nil, ErrInvalidReviewRoomAction
	}
	item = strings.TrimSpace(item)
	if item == "" {
		return ReviewRoom{}, nil, ErrMissingReviewRoomItem
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	updated, transitioned, changed, err := transitionReviewRoomItem(s.room, action, item, resolution)
	if err != nil {
		return ReviewRoom{}, nil, err
	}
	if !changed {
		return ReviewRoom{}, nil, ErrReviewRoomItemNotFound
	}
	s.room = updated
	return copyReviewRoom(updated), transitioned, nil
}

func SummarizeReviewRoom(room ReviewRoom) ReviewRoomSummary {
	room = normalizeReviewRoom(room)
	return ReviewRoomSummary{
		Source:        room.Source,
		AcceptedCount: len(room.Accepted),
		OpenCount:     len(room.Open),
		DeferredCount: len(room.Deferred),
		TopAccepted:   topReviewItems(room.Accepted, 3),
		TopOpen:       topReviewItems(room.Open, 3),
		TopDeferred:   topReviewItems(room.Deferred, 3),
		Issues:        append([]string{}, room.Issues...),
		UpdatedAt:     room.UpdatedAt,
	}
}

func topReviewItems(items []ReviewRoomItem, limit int) []ReviewRoomItem {
	if len(items) <= limit {
		return append([]ReviewRoomItem{}, items...)
	}
	return append([]ReviewRoomItem{}, items[:limit]...)
}

func stampReviewRoom(room ReviewRoom) ReviewRoom {
	room = normalizeReviewRoom(room)
	if room.UpdatedAt == nil || room.UpdatedAt.IsZero() {
		now := time.Now().UTC()
		room.UpdatedAt = &now
	}
	return room
}

func canonicalReviewRoomSource(repoRoot string, runtimePath string) string {
	runtimePath = filepath.Clean(strings.TrimSpace(runtimePath))
	if runtimePath == "." || runtimePath == "" {
		return reviewRoomRuntimeFile
	}
	if repoRoot != "" {
		if rel, err := filepath.Rel(repoRoot, runtimePath); err == nil && rel != "" && rel != "." && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(runtimePath)
}

func normalizeRuntimeReviewRoomSource(room ReviewRoom, repoRoot string, runtimePath string) (ReviewRoom, bool) {
	room = normalizeReviewRoom(room)
	canonical := canonicalReviewRoomSource(repoRoot, runtimePath)
	current := strings.TrimSpace(filepath.ToSlash(room.Source))
	if current == "" {
		current = reviewRoomRuntimeFile
	}
	if current == canonical {
		return room, false
	}
	if current == reviewRoomRuntimeFile || filepath.Base(current) == reviewRoomRuntimeFile {
		room.Source = canonical
		return room, true
	}
	return room, false
}

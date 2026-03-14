package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func (d *diskStore) loadReviewRoom() (ReviewRoom, error) {
	path := d.reviewRoomPath()
	fallback := defaultReviewRoom(filepath.ToSlash(path))
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fallback, nil
		}
		return ReviewRoom{}, err
	}
	var room ReviewRoom
	decodeErr := json.Unmarshal(raw, &room)
	if decodeErr == nil {
		return normalizeReviewRoom(room), nil
	}
	var legacy reviewRoomDiskLegacy
	if err := json.Unmarshal(raw, &legacy); err == nil {
		room := ReviewRoom{
			Source:    legacy.Source,
			Accepted:  reviewRoomLegacyItems(legacy.Accepted, "legacy_import", legacy.UpdatedAt),
			Open:      reviewRoomLegacyItems(legacy.Open, "legacy_import", legacy.UpdatedAt),
			Deferred:  reviewRoomLegacyItems(legacy.Deferred, "legacy_import", legacy.UpdatedAt),
			Issues:    append([]string{}, legacy.Issues...),
			UpdatedAt: legacy.UpdatedAt,
		}
		return normalizeReviewRoom(room), nil
	}
	return ReviewRoom{}, decodeErr
}

func (d *diskStore) saveReviewRoom(room ReviewRoom) error {
	return writeJSONAtomic(d.reviewRoomPath(), normalizeReviewRoom(room))
}

func (d *diskStore) loadReviewRoomSeal() (reviewRoomSeal, error) {
	return readJSON(d.reviewRoomSealPath(), reviewRoomSeal{})
}

func (d *diskStore) saveReviewRoomSeal(room ReviewRoom) error {
	return writeJSONAtomic(d.reviewRoomSealPath(), computeReviewRoomSeal(room))
}

func (d *diskStore) reviewRoomPath() string {
	return filepath.Join(d.baseDir, reviewRoomRuntimeFile)
}

func (d *diskStore) reviewRoomSealPath() string {
	return filepath.Join(d.baseDir, "review_room.seal.json")
}

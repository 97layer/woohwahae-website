package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type diskStore struct {
	baseDir string
}

func newDiskStore(baseDir string) (*diskStore, error) {
	if baseDir == "" {
		return nil, errors.New("data dir is required")
	}
	cleanBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}
	cleanBaseDir = filepath.Clean(cleanBaseDir)
	if err := os.MkdirAll(cleanBaseDir, 0o755); err != nil {
		return nil, err
	}
	return &diskStore{baseDir: cleanBaseDir}, nil
}

func readJSON[T any](path string, fallback T) (T, error) {
	value, _, err := readJSONIfPresent(path, fallback)
	return value, err
}

func readJSONIfPresent[T any](path string, fallback T) (T, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fallback, false, nil
		}
		return fallback, false, err
	}

	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return fallback, false, err
	}
	return value, true, nil
}

func writeJSONAtomic(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

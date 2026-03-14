package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (d *diskStore) loadAuthConfig() (authConfig, error) {
	path := d.authPath()
	if config, found, err := readJSONIfPresent(path, defaultAuthConfig()); err != nil {
		return defaultAuthConfig(), err
	} else if found {
		return config, nil
	}

	legacyPath := d.legacyAuthPath()
	if legacyPath == path {
		return defaultAuthConfig(), nil
	}
	config, found, err := readJSONIfPresent(legacyPath, defaultAuthConfig())
	if err != nil {
		return defaultAuthConfig(), err
	}
	if !found {
		return defaultAuthConfig(), nil
	}
	if err := d.saveAuthConfig(config); err != nil {
		return defaultAuthConfig(), err
	}
	return config, nil
}

func (d *diskStore) saveAuthConfig(config authConfig) error {
	path := d.authPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return writeJSONAtomic(path, config)
}

func (d *diskStore) authPath() string {
	return resolveAuthStoragePath(d.baseDir)
}

func (d *diskStore) legacyAuthPath() string {
	return filepath.Join(d.baseDir, "auth.json")
}

func (d *diskStore) leasePath() string {
	return filepath.Join(d.baseDir, "write_lease.json")
}

func (d *diskStore) ensureWriterLease() error {
	owner := strings.TrimSpace(os.Getenv("LAYER_OS_WRITER_ID"))
	if owner == "" {
		owner = "layer-osctl"
	}
	now := time.Now().UTC()
	current := WriteLease{
		Owner:      owner,
		PID:        os.Getpid(),
		AcquiredAt: &now,
		UpdatedAt:  &now,
		Status:     "active",
	}

	file, err := os.OpenFile(d.leasePath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		raw, marshalErr := json.MarshalIndent(current, "", "  ")
		if marshalErr != nil {
			_ = file.Close()
			return marshalErr
		}
		if _, writeErr := file.Write(raw); writeErr != nil {
			_ = file.Close()
			return writeErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
		return nil
	}
	if !os.IsExist(err) {
		return err
	}

	existing, readErr := d.currentLeaseWithErr()
	if readErr != nil {
		return readErr
	}
	if existing.Owner == owner {
		current.AcquiredAt = existing.AcquiredAt
		return writeJSONAtomic(d.leasePath(), current)
	}
	if !processAlive(existing.PID) {
		return writeJSONAtomic(d.leasePath(), current)
	}
	return errors.New("write lease held by " + existing.Owner)
}

func (d *diskStore) currentLease() WriteLease {
	lease, err := d.currentLeaseWithErr()
	if err != nil {
		return WriteLease{Status: "none"}
	}
	if lease.Owner == "" {
		lease.Status = "none"
	}
	if lease.Status == "" {
		lease.Status = "active"
	}
	return lease
}

func (d *diskStore) currentLeaseWithErr() (WriteLease, error) {
	return readJSON(d.leasePath(), WriteLease{Status: "none"})
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func authStorageShouldExternalize(baseDir string) bool {
	baseDir = filepath.Clean(strings.TrimSpace(baseDir))
	if baseDir == "" {
		return false
	}
	tempDir := filepath.Clean(os.TempDir())
	if tempDir == "" || tempDir == "." {
		return true
	}
	return !pathWithinDir(baseDir, tempDir)
}

func resolveAuthStoragePath(baseDir string) string {
	baseDir = filepath.Clean(strings.TrimSpace(baseDir))
	if raw := strings.TrimSpace(os.Getenv("LAYER_OS_AUTH_FILE")); raw != "" {
		if filepath.IsAbs(raw) {
			return filepath.Clean(raw)
		}
		return filepath.Join(baseDir, filepath.Clean(raw))
	}
	if authStorageShouldExternalize(baseDir) {
		if configDir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(configDir) != "" {
			return filepath.Join(configDir, "layer-os", "auth", authStorageNamespace(baseDir)+".json")
		}
	}
	return filepath.Join(baseDir, "auth.json")
}

func authStorageNamespace(baseDir string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(strings.TrimSpace(baseDir))))
	return "runtime-" + hex.EncodeToString(sum[:6])
}

func pathWithinDir(path string, dir string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(dir) == "" {
		return false
	}
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}
	cleanDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(cleanDir, cleanPath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

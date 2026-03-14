package api

import (
	"errors"
	"path/filepath"
	"strings"

	"layer-os/internal/runtime"
)

var (
	errPathTraversal   = errors.New("file path must not contain ..")
	errPathOutsideBase = errors.New("file path must stay within repo root")
)

func sanitizeSnapshotPaths(packet *runtime.SnapshotPacket, baseDir string) error {
	for i := range packet.Targets {
		if err := sanitizeDeployTargetPaths(&packet.Targets[i], baseDir); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeDeployTargetPaths(item *runtime.DeployTarget, baseDir string) error {
	if item.Workdir == nil {
		return nil
	}
	cleanPath, err := sanitizeRepoPath(*item.Workdir, baseDir)
	if err != nil {
		return err
	}
	if cleanPath == "" {
		item.Workdir = nil
		return nil
	}
	item.Workdir = &cleanPath
	return nil
}

func sanitizeRepoPath(raw string, baseDir string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if containsTraversalSegment(value) {
		return "", errPathTraversal
	}
	cleanBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	cleanBase = filepath.Clean(cleanBase)

	cleanPath := filepath.Clean(value)
	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(cleanBase, cleanPath)
	}
	cleanPath = filepath.Clean(cleanPath)
	if !hasDirPrefix(cleanPath, cleanBase) {
		return "", errPathOutsideBase
	}
	return cleanPath, nil
}

func containsTraversalSegment(raw string) bool {
	for _, part := range strings.Split(filepath.ToSlash(strings.TrimSpace(raw)), "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func hasDirPrefix(path string, prefix string) bool {
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)
	if path == prefix {
		return true
	}
	if prefix == string(filepath.Separator) {
		return strings.HasPrefix(path, prefix)
	}
	return strings.HasPrefix(path, prefix+string(filepath.Separator))
}

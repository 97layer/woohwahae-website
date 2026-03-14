package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditStructureAllowsCacheDir(t *testing.T) {
	root := t.TempDir()

	// Create allowed root directories
	mustMkdir(t, filepath.Join(root, "cmd"))
	mustMkdir(t, filepath.Join(root, "internal"))
	mustMkdir(t, filepath.Join(root, "docs"))
	mustMkdir(t, filepath.Join(root, "scripts"))
	mustMkdir(t, filepath.Join(root, "skills"))
	mustMkdir(t, filepath.Join(root, "contracts"))
	mustMkdir(t, filepath.Join(root, "constitution"))
	mustMkdir(t, filepath.Join(root, ".layer-os"))

	// Create allowed files
	mustWriteLocal(t, filepath.Join(root, "README.md"), "")
	mustWriteLocal(t, filepath.Join(root, "AGENTS.md"), "")
	mustWriteLocal(t, filepath.Join(root, "go.mod"), "")
	mustWriteLocal(t, filepath.Join(root, ".gitignore"), "")

	// Create the .cache directory that we are testing
	mustMkdir(t, filepath.Join(root, ".cache"))
	// Put something inside .cache to ensure its contents are ignored correctly
	mustMkdir(t, filepath.Join(root, ".cache", "00"))
	mustWriteLocal(t, filepath.Join(root, ".cache", "00", "test.bin"), "")

	audit := AuditStructure(root)
	if len(audit.Issues) > 0 {
		t.Fatalf("expected no issues, got %v", audit.Issues)
	}
}

// mustMkdir is a helper function to create directories and fail the test on error.
func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// mustWriteLocal is a helper function to write files and fail the test on error.
func mustWriteLocal(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

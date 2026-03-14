package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

func DefaultGoTestVerificationCommand() ([]string, error) {
	goroot := strings.TrimSpace(goruntime.GOROOT())
	if goroot == "" {
		return nil, fmt.Errorf("verification default command requires a Go toolchain installation")
	}
	binary := filepath.Clean(filepath.Join(goroot, "bin", "go"))
	info, err := os.Stat(binary)
	if err != nil {
		return nil, fmt.Errorf("verification default command requires absolute go binary: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("verification default command requires absolute go binary")
	}
	return []string{binary, "test", "./..."}, nil
}

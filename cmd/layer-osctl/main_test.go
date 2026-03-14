package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	got := splitCSV("alpha, beta , ,gamma")
	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV mismatch: got=%v want=%v", got, want)
	}
}

func TestSplitPairs(t *testing.T) {
	got := splitPairs("latency_ms=120, target = vm , invalid")
	want := map[string]any{"latency_ms": "120", "target": "vm"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitPairs mismatch: got=%v want=%v", got, want)
	}
}

func TestSplitCSVEmpty(t *testing.T) {
	got := splitCSV("")
	want := []string{}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV empty mismatch: got=%v want=%v", got, want)
	}
}

func TestDaemonBaseURLDefault(t *testing.T) {
	oldBase := os.Getenv("LAYER_OS_BASE_URL")
	oldAddr := os.Getenv("LAYER_OS_ADDR")
	defer os.Setenv("LAYER_OS_BASE_URL", oldBase)
	defer os.Setenv("LAYER_OS_ADDR", oldAddr)

	os.Unsetenv("LAYER_OS_BASE_URL")
	os.Unsetenv("LAYER_OS_ADDR")
	if got := daemonBaseURL(); got != "http://127.0.0.1:17808" {
		t.Fatalf("expected default daemon url, got %q", got)
	}
}

func TestDaemonBaseURLFromEnv(t *testing.T) {
	oldBase := os.Getenv("LAYER_OS_BASE_URL")
	oldAddr := os.Getenv("LAYER_OS_ADDR")
	defer os.Setenv("LAYER_OS_BASE_URL", oldBase)
	defer os.Setenv("LAYER_OS_ADDR", oldAddr)

	os.Setenv("LAYER_OS_BASE_URL", "http://127.0.0.1:9090/")
	os.Unsetenv("LAYER_OS_ADDR")
	if got := daemonBaseURL(); got != "http://127.0.0.1:9090" {
		t.Fatalf("expected base url from env, got %q", got)
	}
}

func TestDaemonBaseURLFromAddr(t *testing.T) {
	oldBase := os.Getenv("LAYER_OS_BASE_URL")
	oldAddr := os.Getenv("LAYER_OS_ADDR")
	defer os.Setenv("LAYER_OS_BASE_URL", oldBase)
	defer os.Setenv("LAYER_OS_ADDR", oldAddr)

	os.Unsetenv("LAYER_OS_BASE_URL")
	os.Setenv("LAYER_OS_ADDR", ":9191")
	if got := daemonBaseURL(); got != "http://127.0.0.1:9191" {
		t.Fatalf("expected base url from addr, got %q", got)
	}
}

func TestRequestModelsFromEnv(t *testing.T) {
	oldModel := os.Getenv("LAYER_OS_MODEL")
	oldModels := os.Getenv("LAYER_OS_MODELS")
	defer os.Setenv("LAYER_OS_MODEL", oldModel)
	defer os.Setenv("LAYER_OS_MODELS", oldModels)

	os.Setenv("LAYER_OS_MODEL", "gemini-2.0-flash")
	os.Setenv("LAYER_OS_MODELS", "gpt-5.4,claude-sonnet-4.5")
	got := requestModels()
	want := []string{"gpt-5.4", "claude-sonnet-4.5", "gemini-2.0-flash"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requestModels mismatch: got=%v want=%v", got, want)
	}
}

func TestLocalRuntimeDataDirDefault(t *testing.T) {
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	oldDataDir := os.Getenv("LAYER_OS_DATA_DIR")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	defer os.Setenv("LAYER_OS_DATA_DIR", oldDataDir)

	repoRoot := t.TempDir()
	os.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	os.Unsetenv("LAYER_OS_DATA_DIR")

	got := localRuntimeDataDir()
	want := filepath.Join(repoRoot, ".layer-os")
	if got != want {
		t.Fatalf("expected default local runtime dir %q, got %q", want, got)
	}
}

func TestLocalRuntimeDataDirFromRelativeEnv(t *testing.T) {
	oldRoot := os.Getenv("LAYER_OS_REPO_ROOT")
	oldDataDir := os.Getenv("LAYER_OS_DATA_DIR")
	defer os.Setenv("LAYER_OS_REPO_ROOT", oldRoot)
	defer os.Setenv("LAYER_OS_DATA_DIR", oldDataDir)

	repoRoot := t.TempDir()
	os.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	os.Setenv("LAYER_OS_DATA_DIR", ".layer-os-dev")

	got := localRuntimeDataDir()
	want := filepath.Join(repoRoot, ".layer-os-dev")
	if got != want {
		t.Fatalf("expected env local runtime dir %q, got %q", want, got)
	}
}

func TestLocalRuntimeDataDirFromAbsoluteEnv(t *testing.T) {
	oldDataDir := os.Getenv("LAYER_OS_DATA_DIR")
	defer os.Setenv("LAYER_OS_DATA_DIR", oldDataDir)

	absDir := filepath.Join(t.TempDir(), "runtime")
	os.Setenv("LAYER_OS_DATA_DIR", absDir)

	got := localRuntimeDataDir()
	if got != absDir {
		t.Fatalf("expected absolute env local runtime dir %q, got %q", absDir, got)
	}
}

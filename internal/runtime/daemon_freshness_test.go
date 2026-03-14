package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditDaemonFreshnessFlagsRepoNewerThanDaemonStart(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	source := filepath.Join(repoRoot, "internal", "api", "router.go")
	mustRuntimeWrite(t, source, "package api")
	startedAt := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	baseline := startedAt.Add(-3 * time.Minute)
	for _, item := range []string{filepath.Join(repoRoot, "README.md"), filepath.Join(repoRoot, "AGENTS.md"), filepath.Join(repoRoot, "go.mod")} {
		if err := os.Chtimes(item, baseline, baseline); err != nil {
			t.Fatalf("chtimes baseline %s: %v", item, err)
		}
	}
	newer := startedAt.Add(3 * time.Minute)
	if err := os.Chtimes(source, newer, newer); err != nil {
		t.Fatalf("chtimes source: %v", err)
	}
	audit := AuditDaemonFreshness(repoRoot, startedAt)
	if audit.Status != "degraded" {
		t.Fatalf("expected degraded audit, got %+v", audit)
	}
	if strings.TrimSpace(audit.LatestSourcePath) != "internal/api/router.go" {
		t.Fatalf("expected latest source path, got %+v", audit)
	}
	if joined := strings.Join(audit.Issues, "\n"); !strings.Contains(joined, "source tree changed after current layer-osd start") {
		t.Fatalf("expected freshness issue, got %+v", audit.Issues)
	}
}

func TestAuditDaemonFreshnessIgnoresBrandHomeLane(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs", "brand-home", "app"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	webFile := filepath.Join(repoRoot, "docs", "brand-home", "app", "page.js")
	mustRuntimeWrite(t, webFile, "export default function Page() {}")
	startedAt := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	baseline := startedAt.Add(-3 * time.Minute)
	if err := os.Chtimes(filepath.Join(repoRoot, "go.mod"), baseline, baseline); err != nil {
		t.Fatalf("chtimes baseline go.mod: %v", err)
	}
	newer := startedAt.Add(3 * time.Minute)
	if err := os.Chtimes(webFile, newer, newer); err != nil {
		t.Fatalf("chtimes web file: %v", err)
	}
	audit := AuditDaemonFreshness(repoRoot, startedAt)
	if audit.Status != "ok" {
		t.Fatalf("expected ok audit for excluded brand-home lane, got %+v", audit)
	}
}

func TestReviewRoomSurfacesDaemonFreshnessAuditIssue(t *testing.T) {
	repoRoot := t.TempDir()
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "constitution"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "contracts"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "docs"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "skills"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "cmd"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "api"))
	mustRuntimeMkdir(t, filepath.Join(repoRoot, "internal", "runtime"))
	mustRuntimeWrite(t, filepath.Join(repoRoot, "README.md"), "readme")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "agents")
	mustRuntimeWrite(t, filepath.Join(repoRoot, "go.mod"), "module test")
	source := filepath.Join(repoRoot, "internal", "runtime", "service.go")
	mustRuntimeWrite(t, source, "package runtime")
	startedAt := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	baseline := startedAt.Add(-3 * time.Minute)
	for _, item := range []string{filepath.Join(repoRoot, "README.md"), filepath.Join(repoRoot, "AGENTS.md"), filepath.Join(repoRoot, "go.mod")} {
		if err := os.Chtimes(item, baseline, baseline); err != nil {
			t.Fatalf("chtimes baseline %s: %v", item, err)
		}
	}
	newer := startedAt.Add(3 * time.Minute)
	if err := os.Chtimes(source, newer, newer); err != nil {
		t.Fatalf("chtimes source: %v", err)
	}
	t.Setenv("LAYER_OS_REPO_ROOT", repoRoot)
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.BindDaemonRuntime(startedAt)
	room := service.ReviewRoom()
	for _, item := range room.Open {
		if item.Source == "audit.daemon" && strings.Contains(item.Text, "Daemon freshness drift") {
			if item.Rationale == nil || item.Rationale.Rule != "review_room.audit.daemon" {
				t.Fatalf("unexpected daemon rationale: %+v", item.Rationale)
			}
			return
		}
	}
	t.Fatalf("expected daemon freshness review-room item, got %+v", room.Open)
}

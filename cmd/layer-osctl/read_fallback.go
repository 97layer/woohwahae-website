package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"layer-os/internal/runtime"
)

type daemonRequester interface {
	request(method string, path string, payload any, out any) error
}

type reviewRoomReadResult struct {
	room    runtime.ReviewRoom
	summary runtime.ReviewRoomSummary
}

func envBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func localFallbackEnabled(explicit bool) bool {
	return explicit || envBool("LAYER_OS_ALLOW_LOCAL_FALLBACK")
}

func daemonReadUnavailableError(command string, err error) error {
	if !isDaemonUnavailable(err) {
		return err
	}
	return fmt.Errorf("%w; rerun with localhost permission, or use `%s --allow-local-fallback` / `LAYER_OS_ALLOW_LOCAL_FALLBACK=1` for read-only fallback", err, command)
}

func writeLocalFallbackWarning(command string) {
	fmt.Fprintf(os.Stderr, "warning: layer-osd unavailable; using read-only local fallback for %s from %s\n", command, localRuntimeDataDir())
}

func withLocalReadService[T any](read func(*runtime.Service) (T, error)) (T, error) {
	var zero T
	dataDir := localRuntimeDataDir()
	shadowDir, cleanup, err := cloneReadOnlyRuntimeDataDir(dataDir)
	if err != nil {
		return zero, err
	}
	defer cleanup()
	service, err := runtime.NewService(shadowDir)
	if err != nil {
		return zero, err
	}
	return read(service)
}

func cloneReadOnlyRuntimeDataDir(dataDir string) (string, func(), error) {
	src := filepath.Clean(strings.TrimSpace(dataDir))
	shadowRoot, err := os.MkdirTemp("", "layer-os-read-fallback-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(shadowRoot)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		srcPath := filepath.Join(src, name)
		dstPath := filepath.Join(shadowRoot, name)
		if entry.IsDir() {
			if err := copyDirTree(srcPath, dstPath); err != nil {
				cleanup()
				return "", nil, err
			}
			continue
		}
		if err := copyRegularFile(srcPath, dstPath); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return shadowRoot, cleanup, nil
}

func copyDirTree(srcDir string, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyRegularFile(path, target)
	})
}

func copyRegularFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func resolveReadWithLocalFallback[T any](allow bool, command string, daemonRead func() (T, error), localRead func() (T, error)) (T, bool, error) {
	var zero T
	item, err := daemonRead()
	if err == nil {
		return item, false, nil
	}
	if !localFallbackEnabled(allow) || !isDaemonUnavailable(err) {
		return zero, false, daemonReadUnavailableError(command, err)
	}
	item, err = localRead()
	if err != nil {
		return zero, false, err
	}
	return item, true, nil
}

func runStatus(service cliService, args []string) {
	cmd := flag.NewFlagSet("status", flag.ExitOnError)
	allowLocalFallback := cmd.Bool("allow-local-fallback", false, "use read-only local fallback when daemon is unavailable")
	parseArgs(cmd, args)

	item, usedFallback, err := statusResult(service, *allowLocalFallback)
	if err != nil {
		log.Fatal(err)
	}
	if usedFallback {
		writeLocalFallbackWarning("layer-osctl status")
	}
	writeJSON(item)
}

func statusResult(service cliService, allowLocalFallback bool) (runtime.CompanyState, bool, error) {
	return resolveReadWithLocalFallback(allowLocalFallback, "layer-osctl status", func() (runtime.CompanyState, error) {
		requester, ok := service.(daemonRequester)
		if !ok {
			return service.Status(), nil
		}
		var response struct {
			CompanyState runtime.CompanyState `json:"company_state"`
		}
		if err := requester.request(http.MethodGet, "/api/layer-os/status", nil, &response); err != nil {
			return runtime.CompanyState{}, err
		}
		return response.CompanyState, nil
	}, func() (runtime.CompanyState, error) {
		return withLocalReadService(func(service *runtime.Service) (runtime.CompanyState, error) {
			return service.Status(), nil
		})
	})
}

func runHandoff(service cliService, args []string) {
	cmd := flag.NewFlagSet("handoff", flag.ExitOnError)
	allowLocalFallback := cmd.Bool("allow-local-fallback", false, "use read-only local fallback when daemon is unavailable")
	parseArgs(cmd, args)

	item, usedFallback, err := handoffResult(service, *allowLocalFallback)
	if err != nil {
		log.Fatal(err)
	}
	if usedFallback {
		writeLocalFallbackWarning("layer-osctl handoff")
	}
	writeJSON(item)
}

func handoffResult(service cliService, allowLocalFallback bool) (runtime.HandoffPacket, bool, error) {
	return resolveReadWithLocalFallback(allowLocalFallback, "layer-osctl handoff", func() (runtime.HandoffPacket, error) {
		requester, ok := service.(daemonRequester)
		if !ok {
			return service.Handoff(), nil
		}
		var response struct {
			Handoff runtime.HandoffPacket `json:"handoff"`
		}
		if err := requester.request(http.MethodGet, "/api/layer-os/handoff", nil, &response); err != nil {
			return runtime.HandoffPacket{}, err
		}
		return response.Handoff, nil
	}, func() (runtime.HandoffPacket, error) {
		return withLocalReadService(func(service *runtime.Service) (runtime.HandoffPacket, error) {
			return service.Handoff(), nil
		})
	})
}

func runReviewRoomRead(service reviewRoomService, args []string) {
	cmd := flag.NewFlagSet("review-room", flag.ExitOnError)
	allowLocalFallback := cmd.Bool("allow-local-fallback", false, "use read-only local fallback when daemon is unavailable")
	parseArgs(cmd, args)

	item, usedFallback, err := reviewRoomReadSnapshot(service, *allowLocalFallback)
	if err != nil {
		log.Fatal(err)
	}
	if usedFallback {
		writeLocalFallbackWarning("layer-osctl review-room")
	}
	writeJSON(map[string]any{
		"review_room": item.room,
		"summary":     item.summary,
	})
}

func reviewRoomReadSnapshot(service reviewRoomService, allowLocalFallback bool) (reviewRoomReadResult, bool, error) {
	return resolveReadWithLocalFallback(allowLocalFallback, "layer-osctl review-room", func() (reviewRoomReadResult, error) {
		requester, ok := service.(daemonRequester)
		if !ok {
			return reviewRoomReadResult{room: service.ReviewRoom(), summary: service.ReviewRoomSummary()}, nil
		}
		var response struct {
			ReviewRoom runtime.ReviewRoom        `json:"review_room"`
			Summary    runtime.ReviewRoomSummary `json:"summary"`
		}
		if err := requester.request(http.MethodGet, "/api/layer-os/review-room", nil, &response); err != nil {
			return reviewRoomReadResult{}, err
		}
		return reviewRoomReadResult{room: response.ReviewRoom, summary: response.Summary}, nil
	}, func() (reviewRoomReadResult, error) {
		return withLocalReadService(func(service *runtime.Service) (reviewRoomReadResult, error) {
			return reviewRoomReadResult{room: service.ReviewRoom(), summary: service.ReviewRoomSummary()}, nil
		})
	})
}

func runFounderSummary(service founderService, args []string) {
	cmd := flag.NewFlagSet("founder summary", flag.ExitOnError)
	allowLocalFallback := cmd.Bool("allow-local-fallback", false, "use read-only local fallback when daemon is unavailable")
	parseArgs(cmd, args)

	item, usedFallback, err := founderSummaryResult(service, *allowLocalFallback)
	if err != nil {
		log.Fatal(err)
	}
	if usedFallback {
		writeLocalFallbackWarning("layer-osctl founder summary")
	}
	writeJSON(item)
}

func founderSummaryResult(service founderService, allowLocalFallback bool) (runtime.FounderSummary, bool, error) {
	return resolveReadWithLocalFallback(allowLocalFallback, "layer-osctl founder summary", func() (runtime.FounderSummary, error) {
		requester, ok := service.(daemonRequester)
		if !ok {
			return service.FounderSummary(), nil
		}
		var response struct {
			FounderSummary runtime.FounderSummary `json:"founder_summary"`
		}
		if err := requester.request(http.MethodGet, "/api/layer-os/founder-summary", nil, &response); err != nil {
			return runtime.FounderSummary{}, err
		}
		return response.FounderSummary, nil
	}, func() (runtime.FounderSummary, error) {
		return withLocalReadService(func(service *runtime.Service) (runtime.FounderSummary, error) {
			return service.FounderSummary(), nil
		})
	})
}

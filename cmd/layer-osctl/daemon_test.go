package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type daemonStatusStub struct {
	status runtime.DaemonStatus
	tryFn  func() (runtime.DaemonStatus, error)
}

func (s *daemonStatusStub) DaemonStatus() runtime.DaemonStatus { return s.status }
func (s *daemonStatusStub) TryDaemonStatus() (runtime.DaemonStatus, error) {
	if s.tryFn != nil {
		return s.tryFn()
	}
	return s.status, nil
}

func TestRunDaemonStatusWritesDaemonPacket(t *testing.T) {
	service := &daemonStatusStub{status: runtime.DaemonStatus{Service: "layer-osd", Status: "ok", Address: "127.0.0.1:17808", StartedAt: time.Now().UTC(), UptimeSeconds: 42, MemoryHealth: "ready", DeployHealth: "ready", DegradedReasons: []string{}}}
	raw := captureStdout(t, func() {
		runDaemonCommand(service, []string{"status"})
	})
	if !strings.Contains(raw, "layer-osd") || !strings.Contains(raw, "uptime_seconds") {
		t.Fatalf("unexpected daemon status output: %s", raw)
	}
}

func TestWaitForDaemonStatusRetriesUntilAvailable(t *testing.T) {
	oldNow := daemonWaitNow
	oldSleep := daemonWaitSleep
	defer func() {
		daemonWaitNow = oldNow
		daemonWaitSleep = oldSleep
	}()

	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	daemonWaitNow = func() time.Time { return now }
	sleepCalls := 0
	daemonWaitSleep = func(delay time.Duration) {
		sleepCalls++
		now = now.Add(delay)
	}

	calls := 0
	service := &daemonStatusStub{}
	service.tryFn = func() (runtime.DaemonStatus, error) {
		calls++
		if calls < 3 {
			return runtime.DaemonStatus{}, errDaemonUnavailable
		}
		return runtime.DaemonStatus{Service: "layer-osd", Status: "ok", Address: "127.0.0.1:17808", StartedAt: now.Add(-time.Minute), UptimeSeconds: 60, MemoryHealth: "ready", DeployHealth: "ready", DegradedReasons: []string{}}, nil
	}

	item, err := waitForDaemonStatus(service, 2*time.Second, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("waitForDaemonStatus: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
	if sleepCalls != 2 {
		t.Fatalf("expected 2 sleeps, got %d", sleepCalls)
	}
	if item.Service != "layer-osd" || item.Address != "127.0.0.1:17808" {
		t.Fatalf("unexpected daemon item: %+v", item)
	}
}

func TestWaitForDaemonStatusStopsOnGenericError(t *testing.T) {
	oldNow := daemonWaitNow
	oldSleep := daemonWaitSleep
	defer func() {
		daemonWaitNow = oldNow
		daemonWaitSleep = oldSleep
	}()

	daemonWaitNow = func() time.Time { return time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC) }
	daemonWaitSleep = func(time.Duration) { t.Fatal("unexpected sleep") }

	service := &daemonStatusStub{tryFn: func() (runtime.DaemonStatus, error) {
		return runtime.DaemonStatus{}, errors.New("daemon request failed: 500 Internal Server Error")
	}}

	_, err := waitForDaemonStatus(service, 2*time.Second, 100*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Fatalf("expected generic daemon error, got %v", err)
	}
}

func TestWaitForDaemonStatusTimesOutOnDaemonUnavailable(t *testing.T) {
	oldNow := daemonWaitNow
	oldSleep := daemonWaitSleep
	defer func() {
		daemonWaitNow = oldNow
		daemonWaitSleep = oldSleep
	}()

	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	daemonWaitNow = func() time.Time { return now }
	daemonWaitSleep = func(delay time.Duration) { now = now.Add(delay) }

	service := &daemonStatusStub{tryFn: func() (runtime.DaemonStatus, error) {
		return runtime.DaemonStatus{}, errDaemonUnavailable
	}}

	_, err := waitForDaemonStatus(service, 500*time.Millisecond, 250*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out waiting for daemon") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

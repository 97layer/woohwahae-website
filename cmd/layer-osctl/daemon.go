package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"layer-os/internal/runtime"
)

var daemonWaitTimeoutDefault = 10 * time.Second
var daemonWaitIntervalDefault = 250 * time.Millisecond
var daemonWaitSleep = time.Sleep
var daemonWaitNow = time.Now

type daemonControlService interface {
	DaemonStatus() runtime.DaemonStatus
	TryDaemonStatus() (runtime.DaemonStatus, error)
}

func runDaemonCommand(service daemonControlService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl daemon <status|wait>")
	}
	switch args[0] {
	case "status":
		writeJSON(service.DaemonStatus())
	case "wait":
		runDaemonWait(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl daemon <status|wait>")
	}
}

func runDaemonWait(service daemonControlService, args []string) {
	cmd := flag.NewFlagSet("daemon wait", flag.ExitOnError)
	timeout := cmd.Duration("timeout", daemonWaitTimeoutDefault, "maximum time to wait for daemon reachability")
	interval := cmd.Duration("interval", daemonWaitIntervalDefault, "poll interval while daemon is unavailable")
	parseArgs(cmd, args)

	item, err := waitForDaemonStatus(service, *timeout, *interval)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func waitForDaemonStatus(service daemonControlService, timeout time.Duration, interval time.Duration) (runtime.DaemonStatus, error) {
	if timeout <= 0 {
		return runtime.DaemonStatus{}, errors.New("daemon wait timeout must be positive")
	}
	if interval <= 0 {
		return runtime.DaemonStatus{}, errors.New("daemon wait interval must be positive")
	}
	deadline := daemonWaitNow().Add(timeout)
	var lastErr error
	for {
		item, err := service.TryDaemonStatus()
		if err == nil {
			return item, nil
		}
		if !isDaemonUnavailable(err) {
			return runtime.DaemonStatus{}, err
		}
		lastErr = err
		if !daemonWaitNow().Before(deadline) {
			return runtime.DaemonStatus{}, fmt.Errorf("timed out waiting for daemon: %w", lastErr)
		}
		daemonWaitSleep(interval)
	}
}

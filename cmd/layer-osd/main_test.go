package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

type daemonServerStub struct {
	started       chan struct{}
	serveDone     chan error
	shutdownErr   error
	shutdownCalls int
	shutdownCtx   context.Context
}

func (s *daemonServerStub) Serve(_ net.Listener) error {
	close(s.started)
	return <-s.serveDone
}

func (s *daemonServerStub) Shutdown(ctx context.Context) error {
	s.shutdownCalls++
	s.shutdownCtx = ctx
	select {
	case s.serveDone <- http.ErrServerClosed:
	default:
	}
	return s.shutdownErr
}

func TestNewDaemonServerSetsTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	server := newDaemonServer(":17808", handler)
	if server.Addr != ":17808" {
		t.Fatalf("unexpected server addr: %q", server.Addr)
	}
	if server.Handler != handler {
		t.Fatalf("unexpected server handler: %#v", server.Handler)
	}
	if server.ReadHeaderTimeout != daemonReadHeaderTimeout {
		t.Fatalf("unexpected read header timeout: %s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != daemonReadTimeout {
		t.Fatalf("unexpected read timeout: %s", server.ReadTimeout)
	}
	if server.WriteTimeout != daemonWriteTimeout {
		t.Fatalf("unexpected write timeout: %s", server.WriteTimeout)
	}
	if server.IdleTimeout != daemonIdleTimeout {
		t.Fatalf("unexpected idle timeout: %s", server.IdleTimeout)
	}
}

func TestRunDaemonShutsDownOnContextCancel(t *testing.T) {
	oldTimeout := daemonShutdownTimeout
	defer func() { daemonShutdownTimeout = oldTimeout }()
	daemonShutdownTimeout = 2 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	server := &daemonServerStub{started: make(chan struct{}), serveDone: make(chan error, 1)}
	done := make(chan error, 1)
	go func() {
		done <- runDaemon(ctx, server, nil)
	}()
	<-server.started
	cancel()

	err := <-done
	if err != nil {
		t.Fatalf("runDaemon returned error: %v", err)
	}
	if server.shutdownCalls != 1 {
		t.Fatalf("expected one shutdown call, got %d", server.shutdownCalls)
	}
	if server.shutdownCtx == nil {
		t.Fatal("expected shutdown context")
	}
	if _, ok := server.shutdownCtx.Deadline(); !ok {
		t.Fatal("expected shutdown deadline")
	}
}

func TestRunDaemonReturnsListenError(t *testing.T) {
	server := &daemonServerStub{started: make(chan struct{}), serveDone: make(chan error, 1)}
	server.serveDone <- errors.New("bind failed")

	err := runDaemon(context.Background(), server, nil)
	if err == nil || err.Error() != "bind failed" {
		t.Fatalf("expected bind failure, got %v", err)
	}
	if server.shutdownCalls != 0 {
		t.Fatalf("expected no shutdown call, got %d", server.shutdownCalls)
	}
}

func TestRunDaemonReturnsShutdownError(t *testing.T) {
	oldTimeout := daemonShutdownTimeout
	defer func() { daemonShutdownTimeout = oldTimeout }()
	daemonShutdownTimeout = 2 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	server := &daemonServerStub{started: make(chan struct{}), serveDone: make(chan error, 1), shutdownErr: errors.New("shutdown failed")}
	done := make(chan error, 1)
	go func() {
		done <- runDaemon(ctx, server, nil)
	}()
	<-server.started
	cancel()

	err := <-done
	if err == nil || err.Error() != "shutdown failed" {
		t.Fatalf("expected shutdown failure, got %v", err)
	}
	if server.shutdownCalls != 1 {
		t.Fatalf("expected one shutdown call, got %d", server.shutdownCalls)
	}
}

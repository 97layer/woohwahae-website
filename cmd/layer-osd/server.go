package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

var daemonReadHeaderTimeout = 5 * time.Second
var daemonReadTimeout = 30 * time.Second
var daemonWriteTimeout = 5 * time.Minute
var daemonIdleTimeout = 60 * time.Second
var daemonShutdownTimeout = 5 * time.Second

type daemonHTTPServer interface {
	Serve(net.Listener) error
	Shutdown(ctx context.Context) error
}

func newDaemonServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: daemonReadHeaderTimeout,
		ReadTimeout:       daemonReadTimeout,
		WriteTimeout:      daemonWriteTimeout,
		IdleTimeout:       daemonIdleTimeout,
	}
}

func listenDaemon(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}

func runDaemon(ctx context.Context, server daemonHTTPServer, listener net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), daemonShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

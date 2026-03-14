package main

import (
	"errors"
	"time"

	"layer-os/internal/runtime"
)

var sessionBootstrapRetryAttempts = 3
var sessionBootstrapRetryDelay = 150 * time.Millisecond
var sessionBootstrapSleep = time.Sleep
var sessionBootstrapFallback = localSessionBootstrap

func sessionBootstrapPacket(service sessionService, full bool, allowLocalFallback bool) (runtime.SessionBootstrapPacket, error) {
	packet, err := retrySessionBootstrap(service, full)
	if err == nil {
		return packet, nil
	}
	if !allowLocalFallback || !isDaemonUnavailable(err) {
		return runtime.SessionBootstrapPacket{}, err
	}
	fallback, fallbackErr := sessionBootstrapFallback(full)
	if fallbackErr != nil {
		return runtime.SessionBootstrapPacket{}, fallbackErr
	}
	return fallback, nil
}

func retrySessionBootstrap(service sessionService, full bool) (runtime.SessionBootstrapPacket, error) {
	attempts := sessionBootstrapRetryAttempts
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		packet, err := service.SessionBootstrap(full)
		if err == nil {
			return packet, nil
		}
		if !isDaemonUnavailable(err) || attempt == attempts {
			return runtime.SessionBootstrapPacket{}, err
		}
		sessionBootstrapSleep(sessionBootstrapRetryDelay)
	}
	return runtime.SessionBootstrapPacket{}, nil
}

func isDaemonUnavailable(err error) bool {
	return errors.Is(err, errDaemonUnavailable)
}

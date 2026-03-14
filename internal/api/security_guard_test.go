package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWriteSecurityGuardPrunesExpiredCountersEvenWhenMapIsSmall(t *testing.T) {
	guard := &writeSecurityGuard{
		counters: map[string]*writeSecurityCounter{},
	}
	req := httptest.NewRequest(http.MethodPost, "http://localhost/api/layer-os/auth", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	now := time.Now().UTC()
	key := guard.counterKey("write_auth", req)

	guard.mu.Lock()
	guard.counters[key] = &writeSecurityCounter{
		Count:     1,
		LastAt:    now.Add(-(2*writeProbeWindow + time.Second)),
		FirstSeen: true,
	}
	guard.pruneExpiredLocked(now)
	_, exists := guard.counters[key]
	guard.mu.Unlock()

	if exists {
		t.Fatalf("expected expired counter to be pruned even when the map is small")
	}
}

func TestWriteSecurityGuardKeepsActiveBlockedCounter(t *testing.T) {
	guard := &writeSecurityGuard{
		counters: map[string]*writeSecurityCounter{},
	}
	req := httptest.NewRequest(http.MethodPost, "http://localhost/api/layer-os/auth", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	now := time.Now().UTC()
	key := guard.counterKey("write_auth", req)

	guard.mu.Lock()
	guard.counters[key] = &writeSecurityCounter{
		Count:        writeAuthFailureThreshold,
		LastAt:       now.Add(-(2 * writeProbeWindow)),
		BlockedUntil: now.Add(time.Minute),
		FirstSeen:    true,
		Escalated:    true,
	}
	guard.pruneExpiredLocked(now)
	_, exists := guard.counters[key]
	guard.mu.Unlock()

	if !exists {
		t.Fatalf("expected active blocked counter to survive pruning")
	}
}

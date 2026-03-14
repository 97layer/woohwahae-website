package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"layer-os/internal/runtime"
)

const (
	writeAuthFailureThreshold = 5
	writeAuthFailureWindow    = 2 * time.Minute
	writeAuthThrottleWindow   = 90 * time.Second

	writeProbeThreshold = 3
	writeProbeWindow    = 10 * time.Minute
)

type writeSecurityGuard struct {
	service  *runtime.Service
	now      func() time.Time
	mu       sync.Mutex
	counters map[string]*writeSecurityCounter
}

type writeSecurityCounter struct {
	Count        int
	LastAt       time.Time
	BlockedUntil time.Time
	FirstSeen    bool
	Escalated    bool
}

type writeSecurityGuardContextKey struct{}

func newWriteSecurityGuard(service *runtime.Service) *writeSecurityGuard {
	return &writeSecurityGuard{
		service:  service,
		now:      func() time.Time { return time.Now().UTC() },
		counters: map[string]*writeSecurityCounter{},
	}
}

func withWriteSecurityGuard(ctx context.Context, guard *writeSecurityGuard) context.Context {
	if guard == nil {
		return ctx
	}
	return context.WithValue(ctx, writeSecurityGuardContextKey{}, guard)
}

func writeSecurityGuardFromContext(ctx context.Context) *writeSecurityGuard {
	guard, _ := ctx.Value(writeSecurityGuardContextKey{}).(*writeSecurityGuard)
	return guard
}

func (g *writeSecurityGuard) authTemporarilyBlocked(r *http.Request) (bool, time.Duration) {
	return g.blocked("write_auth", r)
}

func (g *writeSecurityGuard) clearAuthFailures(r *http.Request) {
	if g == nil {
		return
	}
	g.clear("write_auth", r)
}

func (g *writeSecurityGuard) recordAuthFailure(r *http.Request, reason string) (bool, time.Duration) {
	if g == nil {
		return false, 0
	}
	count, blockedUntil, first, escalated, now := g.observeFailure("write_auth", r, writeAuthFailureWindow, writeAuthFailureThreshold, writeAuthThrottleWindow)
	if first {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:   "write_auth_rejected",
			Summary:  "Write auth rejected for " + securitySubjectLabel(r) + ".",
			Severity: "medium",
			Source:   "api.auth",
			Data:     securitySignalData(r, reason, count, blockedUntil),
			Evidence: securitySignalEvidence(r, reason, count),
		})
	}
	if escalated {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:       "write_auth_bruteforce_detected",
			Summary:      "Repeated write-auth failures reached the throttle threshold for " + securitySubjectLabel(r) + "; inspect for hostile or misconfigured agent activity.",
			Severity:     "high",
			Source:       "api.auth",
			Promote:      true,
			ReviewReason: "repeated write-auth failures suggest brute-force or misconfigured agent traffic that needs founder review before trusting the caller",
			ReviewRule:   "review_room.auto.security_write_auth_bruteforce",
			Data:         securitySignalData(r, reason, count, blockedUntil),
			Evidence:     securitySignalEvidence(r, reason, count),
		})
	}
	if !blockedUntil.IsZero() && blockedUntil.After(now) {
		return true, blockedUntil.Sub(now)
	}
	return false, 0
}

func (g *writeSecurityGuard) recordCrossOriginWrite(r *http.Request) {
	if g == nil {
		return
	}
	count, _, first, escalated, _ := g.observeFailure("cross_origin_write", r, writeProbeWindow, writeProbeThreshold, 0)
	if first {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:   "cross_origin_write_blocked",
			Summary:  "Cross-origin write blocked for " + securitySubjectLabel(r) + ".",
			Severity: "medium",
			Source:   "api.same_origin",
			Data:     securitySignalData(r, "cross_origin_write_blocked", count, time.Time{}),
			Evidence: securitySignalEvidence(r, "cross_origin_write_blocked", count),
		})
	}
	if escalated {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:       "cross_origin_write_probe_detected",
			Summary:      "Repeated cross-origin write probes detected for " + securitySubjectLabel(r) + "; inspect for browser-origin abuse or hostile automation.",
			Severity:     "high",
			Source:       "api.same_origin",
			Promote:      true,
			ReviewReason: "repeated cross-origin write probes indicate hostile or misconfigured automation that needs founder review before trusting the caller",
			ReviewRule:   "review_room.auto.security_cross_origin_write_probe",
			Data:         securitySignalData(r, "cross_origin_write_blocked", count, time.Time{}),
			Evidence:     securitySignalEvidence(r, "cross_origin_write_blocked", count),
		})
	}
}

func (g *writeSecurityGuard) recordBootstrapProbe(r *http.Request) {
	if g == nil {
		return
	}
	count, _, first, escalated, _ := g.observeFailure("auth_bootstrap", r, writeProbeWindow, writeProbeThreshold, 0)
	if first {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:   "auth_bootstrap_blocked",
			Summary:  "Non-loopback auth bootstrap blocked for " + securitySubjectLabel(r) + ".",
			Severity: "medium",
			Source:   "api.auth_bootstrap",
			Data:     securitySignalData(r, "auth_bootstrap_requires_loopback", count, time.Time{}),
			Evidence: securitySignalEvidence(r, "auth_bootstrap_requires_loopback", count),
		})
	}
	if escalated {
		_ = g.service.RecordSecuritySignal(runtime.SecuritySignalInput{
			Signal:       "auth_bootstrap_probe_detected",
			Summary:      "Repeated non-loopback auth bootstrap probes detected for " + securitySubjectLabel(r) + "; inspect for hostile agent bootstrap attempts.",
			Severity:     "high",
			Source:       "api.auth_bootstrap",
			Promote:      true,
			ReviewReason: "repeated non-loopback bootstrap probes indicate hostile or misconfigured automation that needs founder review before trusting the caller",
			ReviewRule:   "review_room.auto.security_auth_bootstrap_probe",
			Data:         securitySignalData(r, "auth_bootstrap_requires_loopback", count, time.Time{}),
			Evidence:     securitySignalEvidence(r, "auth_bootstrap_requires_loopback", count),
		})
	}
}

func (g *writeSecurityGuard) blocked(scope string, r *http.Request) (bool, time.Duration) {
	if g == nil {
		return false, 0
	}
	now := g.now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()

	tracker := g.counters[g.counterKey(scope, r)]
	if tracker == nil || tracker.BlockedUntil.IsZero() || !now.Before(tracker.BlockedUntil) {
		return false, 0
	}
	return true, tracker.BlockedUntil.Sub(now)
}

func (g *writeSecurityGuard) clear(scope string, r *http.Request) {
	now := g.now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.counters, g.counterKey(scope, r))
	g.pruneExpiredLocked(now)
}

func (g *writeSecurityGuard) observeFailure(scope string, r *http.Request, window time.Duration, threshold int, blockDuration time.Duration) (int, time.Time, bool, bool, time.Time) {
	now := g.now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()

	key := g.counterKey(scope, r)
	tracker := g.counters[key]
	if tracker == nil {
		tracker = &writeSecurityCounter{}
		g.counters[key] = tracker
	}
	if tracker.BlockedUntil.IsZero() || !now.Before(tracker.BlockedUntil) {
		if tracker.LastAt.IsZero() || now.Sub(tracker.LastAt) > window {
			*tracker = writeSecurityCounter{}
		}
	}

	tracker.Count++
	tracker.LastAt = now

	first := false
	if !tracker.FirstSeen {
		tracker.FirstSeen = true
		first = true
	}

	escalated := false
	if threshold > 0 && tracker.Count >= threshold && !tracker.Escalated {
		tracker.Escalated = true
		escalated = true
		if blockDuration > 0 {
			tracker.BlockedUntil = now.Add(blockDuration)
		}
	}
	g.pruneExpiredLocked(now)
	return tracker.Count, tracker.BlockedUntil, first, escalated, now
}

func (g *writeSecurityGuard) counterKey(scope string, r *http.Request) string {
	return strings.TrimSpace(scope) + "|" + securitySubjectKey(r)
}

func (g *writeSecurityGuard) pruneExpiredLocked(now time.Time) {
	for key, tracker := range g.counters {
		if tracker == nil {
			delete(g.counters, key)
			continue
		}
		if !tracker.BlockedUntil.IsZero() && now.Before(tracker.BlockedUntil) {
			continue
		}
		if tracker.LastAt.IsZero() || now.Sub(tracker.LastAt) > 2*writeProbeWindow {
			delete(g.counters, key)
		}
	}
}

func securitySubjectKey(r *http.Request) string {
	actor := requestActor(r)
	if actor == "" {
		actor = "anonymous"
	}
	host := remoteHost(r.RemoteAddr)
	if host == "" {
		host = "unknown"
	}
	return strings.ToLower(strings.TrimSpace(actor)) + "|" + strings.ToLower(strings.TrimSpace(host))
}

func securitySubjectLabel(r *http.Request) string {
	actor := requestActor(r)
	if actor == "" {
		actor = "anonymous actor"
	} else {
		actor = "actor `" + actor + "`"
	}
	host := remoteHost(r.RemoteAddr)
	if host == "" {
		host = "unknown remote"
	} else {
		host = "remote `" + host + "`"
	}
	return actor + " from " + host
}

func securitySignalData(r *http.Request, reason string, count int, blockedUntil time.Time) map[string]any {
	data := map[string]any{
		"method":          strings.TrimSpace(r.Method),
		"path":            strings.TrimSpace(r.URL.Path),
		"actor":           requestActor(r),
		"remote_host":     remoteHost(r.RemoteAddr),
		"request_host":    requestHost(r),
		"request_scheme":  requestScheme(r),
		"count":           count,
		"expected_origin": requestScheme(r) + "://" + requestHost(r),
	}
	if reason != "" {
		data["reason"] = reason
	}
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		data["origin"] = origin
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		data["referer"] = referer
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		data["forwarded_for"] = forwarded
	}
	if !blockedUntil.IsZero() {
		data["blocked_until"] = blockedUntil.UTC().Format(time.RFC3339)
	}
	return data
}

func securitySignalEvidence(r *http.Request, reason string, count int) []string {
	evidence := []string{
		"method:" + strings.TrimSpace(r.Method),
		"path:" + strings.TrimSpace(r.URL.Path),
		"remote:" + blankToUnknown(remoteHost(r.RemoteAddr)),
		"actor:" + blankToUnknown(requestActor(r)),
		fmt.Sprintf("count:%d", count),
	}
	if reason != "" {
		evidence = append(evidence, "reason:"+reason)
	}
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		evidence = append(evidence, "origin:"+origin)
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		evidence = append(evidence, "referer:"+referer)
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		evidence = append(evidence, "forwarded_for:"+forwarded)
	}
	return evidence
}

func blankToUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return value
}

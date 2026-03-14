package api

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

func NewRouter(service *runtime.Service) http.Handler {
	return NewRouterWithRuntime(service, DaemonRuntimeInfo{})
}

func NewRouterWithRuntime(service *runtime.Service, daemonInfo DaemonRuntimeInfo) http.Handler {
	daemonInfo = normalizeDaemonRuntimeInfo(daemonInfo)
	guard := newWriteSecurityGuard(service)
	mux := http.NewServeMux()

	mux.HandleFunc("/cockpit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		serveCockpit(w)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		item := daemonStatus(service, daemonInfo)
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("/api/layer-os/daemon", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"daemon": daemonStatus(service, daemonInfo),
		})
	})

	mux.HandleFunc("/api/layer-os/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		item := daemonStatus(service, daemonInfo)
		security := service.SecurityAudit()
		writeJSON(w, http.StatusOK, map[string]any{
			"status":        item.Status,
			"company_state": service.Status(),
			"daemon":        item,
			"security":      security,
		})
	})

	mux.HandleFunc("/api/layer-os/handoff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"handoff": service.Handoff(),
		})
	})

	mux.HandleFunc("/api/layer-os/knowledge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"knowledge": service.Knowledge(),
		})
	})

	mux.HandleFunc("/api/layer-os/telegram", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{
				"telegram": service.Telegram(),
				"enabled":  service.TelegramEnabled(),
				"adapter":  service.TelegramAdapterName(),
				"status":   service.TelegramStatus(),
			})
		case http.MethodPost:
			if !requireWriteAuth(service, w, r) {
				return
			}
			if err := service.SendTelegram(); err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sent": true})
		default:
			methodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/layer-os/social/threads", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			status := service.ThreadsStatus()
			writeJSON(w, http.StatusOK, map[string]any{
				"enabled": status.PublishConfigured,
				"adapter": status.Adapter,
				"status":  status,
			})
		case http.MethodPost:
			if !requireWriteAuth(service, w, r) {
				return
			}
			var payload struct {
				ApprovalID string `json:"approval_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "invalid threads publish payload")
				return
			}
			receipt, err := service.PublishThreadsBrandDraft(payload.ApprovalID)
			if err != nil {
				switch {
				case errors.Is(err, runtime.ErrThreadsPublishNotConfigured):
					writeError(w, http.StatusServiceUnavailable, err.Error())
				case errors.Is(err, runtime.ErrThreadsApprovalNotReady), errors.Is(err, runtime.ErrThreadsDraftNotFound):
					writeError(w, http.StatusBadRequest, err.Error())
				case errors.Is(err, runtime.ErrThreadsAlreadyPublished):
					writeError(w, http.StatusConflict, err.Error())
				default:
					writeError(w, http.StatusBadGateway, err.Error())
				}
				return
			}
			writeJSON(w, http.StatusOK, receipt)
		default:
			methodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/layer-os/founder-view", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"founder_view": service.FounderView(),
		})
	})

	mux.HandleFunc("/api/layer-os/founder-summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"founder_summary": service.FounderSummary(),
		})
	})

	mux.HandleFunc("/api/layer-os/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"providers": publicProviderSummaries(service.Providers()),
		})
	})

	mux.HandleFunc("/api/layer-os/corpus/entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"entries": service.ListCapitalizationEntries(),
		})
	})

	mux.HandleFunc("/api/layer-os/corpus/recover", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		if !requireWriteAuth(service, w, r) {
			return
		}
		var input struct {
			Limit   int  `json:"limit"`
			Cleanup bool `json:"cleanup"`
			DryRun  bool `json:"dry_run"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid corpus recover payload")
			return
		}
		if input.Limit < 0 {
			writeError(w, http.StatusBadRequest, "invalid corpus recover limit")
			return
		}
		result, err := service.AbsorbCorpusMarkdown(input.Limit, input.Cleanup, input.DryRun)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("/api/layer-os/knowledge/search", func(w http.ResponseWriter, r *http.Request) {
		handleKnowledgeSearchRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/conversation", func(w http.ResponseWriter, r *http.Request) {
		handleConversationRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/observations", func(w http.ResponseWriter, r *http.Request) {
		handleObservationsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/source-intake", func(w http.ResponseWriter, r *http.Request) {
		handleSourceIntakeRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/source-intake/drafts/prep", func(w http.ResponseWriter, r *http.Request) {
		handleSourceDraftPrepRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/open-threads", func(w http.ResponseWriter, r *http.Request) {
		handleOpenThreadsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/threads", func(w http.ResponseWriter, r *http.Request) {
		handleThreadsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"capabilities": service.Capabilities(),
		})
	})

	mux.HandleFunc("/api/layer-os/session/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		handleSessionBootstrap(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/session/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		handleSessionCheckpoint(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/session/note", func(w http.ResponseWriter, r *http.Request) {
		handleSessionNote(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/session/finish", func(w http.ResponseWriter, r *http.Request) {
		handleSessionFinish(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/review-room", func(w http.ResponseWriter, r *http.Request) {
		handleReviewRoomRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/adapters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"adapters": service.Adapters(),
		})
	})

	mux.HandleFunc("/api/layer-os/writer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"write_lease": service.WriteLease(),
		})
	})

	mux.HandleFunc("/api/layer-os/founder-actions/start", func(w http.ResponseWriter, r *http.Request) {
		handleFounderStartRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/founder-actions/approve", func(w http.ResponseWriter, r *http.Request) {
		handleFounderApproveRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/founder-actions/release", func(w http.ResponseWriter, r *http.Request) {
		handleFounderReleaseRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/founder-actions/rollback", func(w http.ResponseWriter, r *http.Request) {
		handleFounderRollbackRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/cockpit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		full := r.URL.Query().Get("full") == "1" || strings.EqualFold(r.URL.Query().Get("full"), "true")
		writeJSON(w, http.StatusOK, buildCockpitPayload(service, daemonInfo, full))
	})

	mux.HandleFunc("/api/layer-os/snapshot", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"snapshot": service.Snapshot()})
		case http.MethodPost:
			if !requireWriteAuth(service, w, r) {
				return
			}
			var packet runtime.SnapshotPacket
			if err := json.NewDecoder(r.Body).Decode(&packet); err != nil {
				writeError(w, http.StatusBadRequest, "invalid snapshot payload")
				return
			}
			if err := sanitizeSnapshotPaths(&packet, repoRoot()); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if err := service.ImportSnapshot(packet); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"snapshot": service.Snapshot()})
		default:
			methodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/layer-os/audit/security", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, service.SecurityAudit())
	})

	registerAuditRoutes(mux)

	mux.HandleFunc("/api/layer-os/events", func(w http.ResponseWriter, r *http.Request) {
		handleEvents(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/events/stream", func(w http.ResponseWriter, r *http.Request) {
		handleEventStream(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/flows", func(w http.ResponseWriter, r *http.Request) {
		handleFlowsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/flows/sync", func(w http.ResponseWriter, r *http.Request) {
		handleFlowSyncRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/auth", func(w http.ResponseWriter, r *http.Request) {
		handleAuthRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/preflights", func(w http.ResponseWriter, r *http.Request) {
		handlePreflightsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/policies", func(w http.ResponseWriter, r *http.Request) {
		handlePoliciesRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/policies/evaluate", func(w http.ResponseWriter, r *http.Request) {
		handlePolicyEvaluateRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/gateway-calls", func(w http.ResponseWriter, r *http.Request) {
		handleGatewayCallsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/execute-runs", func(w http.ResponseWriter, r *http.Request) {
		handleExecuteRunsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/execute-runs/run", func(w http.ResponseWriter, r *http.Request) {
		handleExecuteRunRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/verifications", func(w http.ResponseWriter, r *http.Request) {
		handleVerificationsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/verifications/run", func(w http.ResponseWriter, r *http.Request) {
		handleVerificationRunRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/branches", func(w http.ResponseWriter, r *http.Request) {
		handleBranchesRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/branches/merge", func(w http.ResponseWriter, r *http.Request) {
		handleBranchMergeRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/profiles", func(w http.ResponseWriter, r *http.Request) {
		handleJobProfilesRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/packet", func(w http.ResponseWriter, r *http.Request) {
		handleJobPacketRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs", func(w http.ResponseWriter, r *http.Request) {
		handleJobsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/promote", func(w http.ResponseWriter, r *http.Request) {
		handleJobPromoteRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/dispatch", func(w http.ResponseWriter, r *http.Request) {
		handleJobDispatchRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/update", func(w http.ResponseWriter, r *http.Request) {
		handleJobUpdateRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/jobs/report", func(w http.ResponseWriter, r *http.Request) {
		handleJobReportRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/proposals", func(w http.ResponseWriter, r *http.Request) {
		handleProposalsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/proposals/promote", func(w http.ResponseWriter, r *http.Request) {
		handleProposalPromoteRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/work-items", func(w http.ResponseWriter, r *http.Request) {
		handleWorkItemsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/approval-inbox", func(w http.ResponseWriter, r *http.Request) {
		handleApprovalInboxRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/approval-inbox/resolve", func(w http.ResponseWriter, r *http.Request) {
		handleApprovalResolveRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/releases", func(w http.ResponseWriter, r *http.Request) {
		handleReleasesRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/deploys", func(w http.ResponseWriter, r *http.Request) {
		handleDeploysRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/deploys/execute", func(w http.ResponseWriter, r *http.Request) {
		handleDeployExecuteRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/rollbacks", func(w http.ResponseWriter, r *http.Request) {
		handleRollbacksRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/rollbacks/execute", func(w http.ResponseWriter, r *http.Request) {
		handleRollbackExecuteRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/deploy-targets", func(w http.ResponseWriter, r *http.Request) {
		handleDeployTargetsRoute(service, w, r)
	})

	mux.HandleFunc("/api/layer-os/memory", func(w http.ResponseWriter, r *http.Request) {
		handleMemoryRoute(service, w, r)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyResponseSecurityHeaders(w, r)
		r = r.WithContext(withWriteSecurityGuard(r.Context(), guard))
		actor := requestActor(r)
		if actor == "" {
			mux.ServeHTTP(w, r)
			return
		}
		if err := service.WithActor(actor, func(*runtime.Service) error {
			mux.ServeHTTP(w, r)
			return nil
		}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
	})
}

func requestActor(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-Layer-Actor"))
}

func requestModels(r *http.Request) []string {
	return parseCSVHeader(r.Header.Get("X-Layer-Models"))
}

func parseCSVHeader(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

func mergeModels(groups ...[]string) []string {
	seen := map[string]bool{}
	items := []string{}
	for _, group := range groups {
		for _, raw := range group {
			value := strings.TrimSpace(raw)
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			items = append(items, value)
		}
	}
	return items
}

func requireWriteAuth(service *runtime.Service, w http.ResponseWriter, r *http.Request) bool {
	if !requireSameOriginWrite(w, r) {
		return false
	}
	status := service.AuthStatus()
	if !status.WriteAuthEnabled {
		return true
	}
	guard := writeSecurityGuardFromContext(r.Context())
	if blocked, retryAfter := guard.authTemporarilyBlocked(r); blocked {
		writeThrottledAuthError(w, retryAfter)
		return false
	}
	const prefix = "Bearer "
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(header, prefix) {
		if blocked, retryAfter := guard.recordAuthFailure(r, "missing_bearer_token"); blocked {
			writeThrottledAuthError(w, retryAfter)
			return false
		}
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if !service.AuthorizeWriteToken(token) {
		if blocked, retryAfter := guard.recordAuthFailure(r, "invalid_bearer_token"); blocked {
			writeThrottledAuthError(w, retryAfter)
			return false
		}
		writeError(w, http.StatusUnauthorized, "invalid bearer token")
		return false
	}
	guard.clearAuthFailures(r)
	return true
}

func requireAuthBootstrap(service *runtime.Service, w http.ResponseWriter, r *http.Request) bool {
	if service.AuthStatus().WriteAuthEnabled {
		return requireWriteAuth(service, w, r)
	}
	if !requireSameOriginWrite(w, r) {
		return false
	}
	if !requestFromLoopback(r) {
		if guard := writeSecurityGuardFromContext(r.Context()); guard != nil {
			guard.recordBootstrapProbe(r)
		}
		writeError(w, http.StatusForbidden, "auth bootstrap requires loopback request")
		return false
	}
	return true
}

func requireSameOriginWrite(w http.ResponseWriter, r *http.Request) bool {
	if originAllowed(r) {
		return true
	}
	if guard := writeSecurityGuardFromContext(r.Context()); guard != nil {
		guard.recordCrossOriginWrite(r)
	}
	writeError(w, http.StatusForbidden, "cross-origin write blocked")
	return false
}

func originAllowed(r *http.Request) bool {
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return matchesRequestOrigin(origin, r)
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		return matchesRequestOrigin(referer, r)
	}
	return true
}

func requestFromLoopback(r *http.Request) bool {
	if !isLoopbackHost(remoteHost(r.RemoteAddr)) {
		return false
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		for _, item := range strings.Split(forwarded, ",") {
			if !isLoopbackHost(strings.TrimSpace(item)) {
				return false
			}
		}
	}
	return true
}

func remoteHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		return host
	}
	return raw
}

func isLoopbackHost(raw string) bool {
	raw = strings.Trim(strings.TrimSpace(strings.ToLower(raw)), "[]")
	if raw == "" {
		return false
	}
	if raw == "localhost" {
		return true
	}
	if ip := net.ParseIP(raw); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func matchesRequestOrigin(raw string, r *http.Request) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, requestHost(r)) && strings.EqualFold(parsed.Scheme, requestScheme(r))
}

func requestScheme(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if value := strings.TrimSpace(parts[0]); value != "" {
			return strings.ToLower(value)
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(r *http.Request) string {
	if host := strings.TrimSpace(r.Host); host != "" {
		return host
	}
	return strings.TrimSpace(r.URL.Host)
}

func applyResponseSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), microphone=(), payment=(), usb=()")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	if requestScheme(r) == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
}

func writeThrottledAuthError(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(retryAfter.Round(time.Second).Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeError(w, http.StatusTooManyRequests, "write auth temporarily throttled")
}

func repoRoot() string {
	if root := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT")); root != "" {
		return root
	}
	return "."
}

func stringRef(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

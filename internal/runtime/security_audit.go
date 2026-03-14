package runtime

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	securityReviewMode   = "security_review"
	securityReviewMaxAge = 7 * 24 * time.Hour
	securitySignalMaxAge = 24 * time.Hour
	securitySignalLimit  = 5
)

var requiredSecurityReviewChecks = []string{
	"write_auth_enabled",
	"secret_plaintext_surface_minimized",
	"edge_tls_required",
	"edge_access_control_required",
}

type SecurityAudit struct {
	Status                     string                  `json:"status"`
	WriteAuthEnabled           bool                    `json:"write_auth_enabled"`
	OpenSecurityItems          []string                `json:"open_security_items"`
	LastReviewAt               *time.Time              `json:"last_review_at,omitempty"`
	LastSecuritySignalAt       *time.Time              `json:"last_security_signal_at,omitempty"`
	RecentSecuritySignalCount  int                     `json:"recent_security_signal_count"`
	RecentEscalatedSignalCount int                     `json:"recent_escalated_security_signal_count"`
	RecentSecuritySignals      []SecuritySignalSummary `json:"recent_security_signals"`
	RecommendedCadence         []string                `json:"recommended_cadence"`
	SuggestedCommands          []string                `json:"suggested_commands"`
	RuntimeSecretFindings      []string                `json:"runtime_secret_findings"`
	Issues                     []string                `json:"issues"`
	Checks                     []string                `json:"checks"`
}

type SecuritySignalSummary struct {
	Kind       string    `json:"kind"`
	Summary    string    `json:"summary"`
	Severity   string    `json:"severity"`
	Source     string    `json:"source"`
	Actor      string    `json:"actor"`
	Promoted   bool      `json:"promoted"`
	DetectedAt time.Time `json:"detected_at"`
}

func AuditSecurity(_ string, dataDir string) SecurityAudit {
	audit := defaultSecurityAudit()
	disk, err := newDiskStore(dataDir)
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit failed: "+err.Error())
		return audit
	}
	authConfig, err := disk.loadAuthConfig()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit auth load failed: "+err.Error())
		return audit
	}
	room, err := disk.loadReviewRoom()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit review-room load failed: "+err.Error())
		return audit
	}
	preflights, err := disk.loadPreflights()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit preflight load failed: "+err.Error())
		return audit
	}
	releases, err := disk.loadReleases()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit release load failed: "+err.Error())
		return audit
	}
	targets, err := disk.loadTargets()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit deploy target load failed: "+err.Error())
		return audit
	}
	events, err := disk.loadEvents()
	if err != nil {
		audit.Status = "degraded"
		audit.Issues = append(audit.Issues, "security audit event load failed: "+err.Error())
		return audit
	}
	status := AuthStatus{
		WriteAuthEnabled: authEnabled(authConfig),
		UpdatedAt:        authConfig.UpdatedAt,
	}
	audit = evaluateSecurityAuditWithFindings(status, room, preflights, releases, targets, events, scanRuntimeSecretFiles(disk.baseDir), zeroSafeNow())
	if authStorageShouldExternalize(disk.baseDir) {
		if pathWithinDir(disk.authPath(), disk.baseDir) {
			audit.Issues = appendUniqueString(audit.Issues, "auth trust root is collocated with runtime data dir: move auth storage outside .layer-os or set LAYER_OS_AUTH_FILE")
			audit.Status = "degraded"
		} else {
			audit.Checks = appendUniqueString(audit.Checks, "auth trust root separated from runtime data dir")
		}
	}
	return audit
}

func EvaluateSecurityAudit(auth AuthStatus, room ReviewRoom, preflights []PreflightRecord, releases []ReleasePacket, targets []DeployTarget, events []EventEnvelope, now time.Time) SecurityAudit {
	return evaluateSecurityAuditWithFindings(auth, room, preflights, releases, targets, events, nil, now)
}

func evaluateSecurityAuditWithFindings(auth AuthStatus, room ReviewRoom, preflights []PreflightRecord, releases []ReleasePacket, targets []DeployTarget, events []EventEnvelope, runtimeSecretFindings []string, now time.Time) SecurityAudit {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	audit := defaultSecurityAudit()
	audit.WriteAuthEnabled = auth.WriteAuthEnabled
	audit.OpenSecurityItems = openSecurityReviewItems(room)
	audit.RuntimeSecretFindings = append([]string{}, runtimeSecretFindings...)
	audit.RecentSecuritySignals = recentSecuritySignals(events, now)
	audit.RecentSecuritySignalCount = len(audit.RecentSecuritySignals)
	audit.RecentEscalatedSignalCount = countEscalatedSecuritySignals(audit.RecentSecuritySignals)
	if latest := latestSecuritySignalAt(audit.RecentSecuritySignals); latest != nil {
		audit.LastSecuritySignalAt = latest
	}

	if auth.WriteAuthEnabled {
		audit.Checks = appendUniqueString(audit.Checks, "write auth enabled")
	} else {
		audit.Issues = appendUniqueString(audit.Issues, "write auth disabled: mutating CLI/API paths are not bearer-gated")
	}

	if len(audit.RuntimeSecretFindings) == 0 {
		audit.Checks = appendUniqueString(audit.Checks, "no plaintext-like runtime secrets detected")
	} else {
		audit.Issues = appendUniqueString(audit.Issues, fmt.Sprintf("runtime secret plaintext findings detected: %d", len(audit.RuntimeSecretFindings)))
	}

	if len(audit.OpenSecurityItems) == 0 {
		audit.Checks = appendUniqueString(audit.Checks, "no open security agenda items")
	} else {
		audit.Issues = appendUniqueString(audit.Issues, fmt.Sprintf("open security agenda items present: %d", len(audit.OpenSecurityItems)))
	}

	if audit.RecentSecuritySignalCount == 0 {
		audit.Checks = appendUniqueString(audit.Checks, "no recent suspicious security signals recorded")
	} else {
		audit.Checks = appendUniqueString(audit.Checks, fmt.Sprintf("recent suspicious security signals recorded: %d", audit.RecentSecuritySignalCount))
		if audit.RecentEscalatedSignalCount > 0 {
			audit.Issues = appendUniqueString(audit.Issues, fmt.Sprintf("recent escalated security signals detected: %d", audit.RecentEscalatedSignalCount))
		}
	}

	latestReview := latestSecurityReviewPreflight(preflights)
	latestReleaseAt := latestReleaseTime(releases)
	if latestReview == nil {
		audit.Issues = appendUniqueString(audit.Issues, "security review cadence missing: no security_review preflight recorded")
		if len(targets) > 0 {
			audit.Issues = appendUniqueString(audit.Issues, "security exposure gate missing: deploy target configured without recorded security review")
		}
	} else {
		reviewAt := latestReview.GeneratedAt.UTC()
		audit.LastReviewAt = &reviewAt
		audit.Checks = appendUniqueString(audit.Checks, "security review evidence recorded")
		missingChecks := missingSecurityReviewChecks(*latestReview)
		if len(missingChecks) > 0 {
			audit.Issues = appendUniqueString(audit.Issues, "security review gate incomplete: missing checks "+strings.Join(missingChecks, ", "))
		} else {
			audit.Checks = appendUniqueString(audit.Checks, "security review required checks recorded")
		}
		if latestReview.Status == "ready" && latestReview.Decision == "go" && len(missingChecks) == 0 {
			audit.Checks = appendUniqueString(audit.Checks, "latest security review cleared go/ready gate")
		} else {
			audit.Issues = appendUniqueString(audit.Issues, "latest security review still holding release/exposure gate")
		}
		if now.Sub(reviewAt) <= securityReviewMaxAge {
			audit.Checks = appendUniqueString(audit.Checks, "security review freshness within 7d")
		} else {
			audit.Issues = appendUniqueString(audit.Issues, "security review cadence stale: last review is older than 7d")
		}
		if latestReleaseAt != nil {
			if reviewAt.Before(*latestReleaseAt) {
				audit.Issues = appendUniqueString(audit.Issues, "security release gate stale: latest release is newer than latest security review")
			} else {
				audit.Checks = appendUniqueString(audit.Checks, "latest release covered by security review")
			}
		}
	}

	if len(targets) > 0 {
		audit.Checks = appendUniqueString(audit.Checks, "deploy target configured")
		if !auth.WriteAuthEnabled {
			audit.Issues = appendUniqueString(audit.Issues, "deploy target configured while write auth remains disabled")
		}
	}

	if len(audit.Issues) > 0 {
		audit.Status = "degraded"
	} else {
		audit.Status = "ok"
	}
	return audit
}

func defaultSecurityAudit() SecurityAudit {
	return SecurityAudit{
		Status:                "ok",
		OpenSecurityItems:     []string{},
		RecentSecuritySignals: []SecuritySignalSummary{},
		RecommendedCadence:    []string{"weekly", "before_release", "before_external_exposure"},
		SuggestedCommands: []string{
			"layer-osctl audit security --strict",
			"layer-osctl auth set [--token <token>]",
			"layer-osctl preflight create --id security_review_<ts> --task 'security review: weekly' --mode security_review --status ready|degraded-lite --decision go|hold --checks write_auth_enabled,secret_plaintext_surface_minimized,edge_tls_required,edge_access_control_required",
			"layer-osctl event list | rg '^security\\.'",
		},
		RuntimeSecretFindings: []string{},
		Issues:                []string{},
		Checks:                []string{},
	}
}

func recentSecuritySignals(events []EventEnvelope, now time.Time) []SecuritySignalSummary {
	if len(events) == 0 {
		return []SecuritySignalSummary{}
	}
	items := make([]SecuritySignalSummary, 0, securitySignalLimit)
	cutoff := now.Add(-securitySignalMaxAge)
	for i := len(events) - 1; i >= 0; i-- {
		item, ok := securitySignalSummaryFromEvent(events[i], cutoff)
		if !ok {
			continue
		}
		items = append(items, item)
		if len(items) >= securitySignalLimit {
			break
		}
	}
	return items
}

func securitySignalSummaryFromEvent(item EventEnvelope, cutoff time.Time) (SecuritySignalSummary, bool) {
	if !strings.HasPrefix(strings.TrimSpace(item.Kind), "security.") {
		return SecuritySignalSummary{}, false
	}
	if item.Timestamp.IsZero() || item.Timestamp.Before(cutoff) {
		return SecuritySignalSummary{}, false
	}
	summary := jsonString(item.Data["summary"])
	if summary == "" {
		summary = strings.TrimPrefix(strings.TrimSpace(item.Kind), "security.")
	}
	severity := strings.TrimSpace(strings.ToLower(jsonString(item.Data["severity"])))
	if severity == "" {
		severity = "medium"
	}
	source := jsonString(item.Data["source"])
	if source == "" {
		source = "runtime.security"
	}
	return SecuritySignalSummary{
		Kind:       strings.TrimSpace(item.Kind),
		Summary:    summary,
		Severity:   severity,
		Source:     source,
		Actor:      strings.TrimSpace(item.Actor),
		Promoted:   jsonBool(item.Data["review_promoted"]),
		DetectedAt: item.Timestamp.UTC(),
	}, true
}

func latestSecuritySignalAt(items []SecuritySignalSummary) *time.Time {
	if len(items) == 0 {
		return nil
	}
	latest := items[0].DetectedAt.UTC()
	for _, item := range items[1:] {
		if item.DetectedAt.After(latest) {
			latest = item.DetectedAt.UTC()
		}
	}
	return &latest
}

func countEscalatedSecuritySignals(items []SecuritySignalSummary) int {
	count := 0
	for _, item := range items {
		if item.Promoted {
			count++
		}
	}
	return count
}

func jsonString(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func jsonBool(value any) bool {
	flag, _ := value.(bool)
	return flag
}

func latestSecurityReviewPreflight(items []PreflightRecord) *PreflightRecord {
	var latest *PreflightRecord
	for _, item := range items {
		if !isSecurityReviewPreflight(item) {
			continue
		}
		copy := item
		if latest == nil || latest.GeneratedAt.Before(copy.GeneratedAt) {
			latest = &copy
		}
	}
	return latest
}

func isSecurityReviewPreflight(item PreflightRecord) bool {
	mode := strings.TrimSpace(strings.ToLower(item.Mode))
	if mode == securityReviewMode {
		return true
	}
	recordID := strings.TrimSpace(strings.ToLower(item.RecordID))
	if strings.HasPrefix(recordID, "security_review_") {
		return true
	}
	task := strings.TrimSpace(strings.ToLower(item.Task))
	return strings.Contains(task, "security review") || strings.Contains(task, "보안 점검")
}

func missingSecurityReviewChecks(item PreflightRecord) []string {
	seen := map[string]struct{}{}
	for _, check := range item.Checks {
		normalized := strings.TrimSpace(strings.ToLower(check))
		if normalized == "" {
			continue
		}
		seen[normalized] = struct{}{}
	}
	missing := []string{}
	for _, check := range requiredSecurityReviewChecks {
		if _, ok := seen[check]; ok {
			continue
		}
		missing = append(missing, check)
	}
	return missing
}

func openSecurityReviewItems(room ReviewRoom) []string {
	items := []string{}
	for _, item := range normalizeReviewRoom(room).Open {
		if !isSecurityReviewItem(item) {
			continue
		}
		items = append(items, strings.TrimSpace(item.Text))
	}
	return items
}

func isSecurityReviewItem(item ReviewRoomItem) bool {
	if item.Ref != nil {
		ref := strings.TrimSpace(strings.ToLower(*item.Ref))
		if strings.HasPrefix(ref, "security_") {
			return true
		}
	}
	text := strings.TrimSpace(strings.ToLower(item.Text))
	return strings.HasPrefix(text, "[security_")
}

func latestReleaseTime(items []ReleasePacket) *time.Time {
	var latest *time.Time
	for _, item := range items {
		if item.ReleasedAt == nil || item.ReleasedAt.IsZero() {
			continue
		}
		value := item.ReleasedAt.UTC()
		if latest == nil || latest.Before(value) {
			copy := value
			latest = &copy
		}
	}
	return latest
}

func scanRuntimeSecretFiles(dataDir string) []string {
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		return []string{}
	}
	findings := []string{}
	_ = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		var payload any
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil
		}
		rel := filepath.Base(path)
		findings = append(findings, scanSecretValue(rel, payload, nil)...)
		return nil
	})
	sort.Strings(findings)
	return findings
}

func scanSecretValue(file string, value any, trail []string) []string {
	findings := []string{}
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			findings = append(findings, scanSecretValue(file, nested, append(trail, key))...)
		}
	case []any:
		for index, nested := range typed {
			findings = append(findings, scanSecretValue(file, nested, append(trail, strconv.Itoa(index)))...)
		}
	case string:
		key := ""
		if len(trail) > 0 {
			key = trail[len(trail)-1]
		}
		if looksLikeRuntimeSecret(key, typed) {
			path := strings.Join(trail, ".")
			if path == "" {
				path = key
			}
			findings = append(findings, fmt.Sprintf("%s:%s", file, path))
		}
	}
	return findings
}

func looksLikeRuntimeSecret(key string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	key = strings.TrimSpace(strings.ToLower(key))
	if isSafeRuntimeKey(key) {
		return false
	}
	if isSensitiveRuntimeKey(key) {
		return !looksLikeHashValue(value)
	}
	lowerValue := strings.ToLower(value)
	if strings.HasPrefix(value, "sk-") || strings.HasPrefix(value, "AIza") || strings.HasPrefix(lowerValue, "bearer ") {
		return true
	}
	return false
}

func isSensitiveRuntimeKey(key string) bool {
	if key == "" {
		return false
	}
	switch key {
	case "token", "write_token", "api_key", "apikey", "access_token", "refresh_token", "secret", "client_secret", "bot_token":
		return true
	}
	return strings.HasSuffix(key, "_token") || strings.HasSuffix(key, "_secret") || strings.Contains(key, "api_key")
}

func isSafeRuntimeKey(key string) bool {
	if key == "" {
		return false
	}
	if strings.HasSuffix(key, "_hash") || strings.HasSuffix(key, "_id") || strings.HasSuffix(key, "_at") || strings.HasSuffix(key, "_class") || strings.HasSuffix(key, "_status") {
		return true
	}
	switch key {
	case "token_class", "sync_key", "writer_id", "model", "provider", "channel", "target":
		return true
	default:
		return false
	}
}

func looksLikeHashValue(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

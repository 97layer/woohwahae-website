package runtime

import (
	"fmt"
	"os"
	"strings"
)

const statusDashboardArchitectVerificationScope = "architect_loop"

func (s *Service) enrichedCompanyStateLocked(state CompanyState) CompanyState {
	dashboard := s.deriveProgressDashboardLocked(state)
	state.Progress = &dashboard
	return state
}

func (s *Service) deriveProgressDashboardLocked(state CompanyState) ProgressDashboard {
	room := s.reviewRoom.current()
	observations := s.observation.list()
	agentJobs := s.agentJob.list()
	verifications := s.verify.list()
	releases := s.release.list()
	deploys := s.deploy.list()
	targets := s.target.list()
	gatewayCalls := s.gateway.list()
	entries, err := s.disk.loadCapitalizationEntries()
	if err != nil {
		entries = []CapitalizationEntry{}
	}
	adapters := AdapterSummary{
		Gateway:                s.gatewayAdapter.Name(),
		GatewaySemantics:       s.gatewayAdapter.Semantics(),
		GatewayDispatchEnabled: s.gatewayAdapter.DispatchEnabled(),
		GatewayRequiredMode:    canonicalPolicyMode(s.gatewayAdapter.RequiredMode()),
		Verify:                 s.verifyAdapter.Name(),
		Deploy:                 s.deployAdapter.Name(),
		Rollback:               s.rollbackAdapter.Name(),
	}
	providers := providerSummaries(adapters)
	security := s.securityAuditLocked()
	axes := []ProgressAxis{
		deriveInfrastructureAxis(state, room, releases, deploys, targets),
		deriveSecurityPostureAxis(security),
		deriveKnowledgePipelineAxis(entries, observations),
		deriveAgentControlAxis(room, agentJobs, verifications, providers),
		deriveMultiAgentChainAxis(agentJobs, gatewayCalls),
		deriveAutomationTriggerAxis(verifications, agentJobs),
		deriveCommitStabilityAxis(state, room, verifications, agentJobs),
	}
	total := 0
	for _, axis := range axes {
		total += axis.Percent
	}
	overall := 0
	if len(axes) > 0 {
		overall = total / len(axes)
	}
	return ProgressDashboard{
		OverallPercent: overall,
		OverallBar:     progressBar(overall),
		OverallStatus:  progressStatus(overall),
		Axes:           axes,
	}
}

func deriveInfrastructureAxis(state CompanyState, room ReviewRoom, releases []ReleasePacket, deploys []DeployRun, targets []DeployTarget) ProgressAxis {
	score := 20
	signals := []string{"twin surface kernel 유지"}
	if strings.TrimSpace(state.DeployHealth) == "ready" {
		score += 20
		signals = append(signals, "deploy health ready")
	} else {
		signals = append(signals, "deploy health degraded")
	}
	if len(targets) > 0 {
		score += 20
		signals = append(signals, "deploy target configured")
	} else {
		signals = append(signals, "deploy target missing")
	}
	if len(releases) > 0 || len(deploys) > 0 {
		score += 20
		signals = append(signals, "release/deploy evidence recorded")
	} else {
		signals = append(signals, "release/deploy evidence still empty")
	}
	if !reviewRoomHasOpenText(room, "VM 배포 결정 대기") {
		score += 20
		signals = append(signals, "VM deployment decision closed")
	} else {
		signals = append(signals, "VM deployment decision still open")
	}
	return newProgressAxis("infra_core", "핵심 인프라", score, signals)
}

func deriveSecurityPostureAxis(audit SecurityAudit) ProgressAxis {
	score := 20
	signals := []string{"verification/egress/path hardening landed"}
	if audit.WriteAuthEnabled {
		score += 25
		signals = append(signals, "write auth enabled")
	} else {
		signals = append(signals, "write auth disabled")
	}
	if len(audit.OpenSecurityItems) == 0 {
		score += 15
		signals = append(signals, "no open security agenda")
	} else {
		signals = append(signals, "security agenda still open")
	}
	if audit.RecentSecuritySignalCount == 0 {
		signals = append(signals, "no recent suspicious security signals")
	} else {
		signals = append(signals, fmt.Sprintf("recent suspicious security signals: %d", audit.RecentSecuritySignalCount))
		if audit.RecentEscalatedSignalCount > 0 {
			signals = append(signals, "recent escalated signal requires review")
		}
	}
	if audit.LastReviewAt != nil {
		score += 20
		signals = append(signals, "security review recorded")
	} else {
		signals = append(signals, "security review missing")
	}
	if securityAuditHasIssueFragment(audit, "cadence") || securityAuditHasIssueFragment(audit, "release gate") || securityAuditHasIssueFragment(audit, "exposure gate") {
		signals = append(signals, "security review freshness/gate degraded")
	} else {
		score += 20
		signals = append(signals, "security review freshness/gate healthy")
	}
	if audit.Status == "ok" {
		score += 20
		signals = append(signals, "security posture operational")
	} else {
		signals = append(signals, "security posture degraded")
	}
	return newProgressAxis("security_posture", "보안 자세", score, signals)
}

func securityAuditHasIssueFragment(audit SecurityAudit, fragment string) bool {
	fragment = strings.TrimSpace(strings.ToLower(fragment))
	if fragment == "" {
		return false
	}
	for _, issue := range audit.Issues {
		if strings.Contains(strings.ToLower(strings.TrimSpace(issue)), fragment) {
			return true
		}
	}
	return false
}

func deriveKnowledgePipelineAxis(entries []CapitalizationEntry, observations []ObservationRecord) ProgressAxis {
	score := 40
	signals := []string{"observation ingest surface live", "corpus search surface live"}
	if len(entries) > 0 {
		score += 20
		signals = append(signals, "corpus entries accumulating")
	} else {
		signals = append(signals, "corpus entries not accumulated yet")
	}
	if len(observations) > 0 {
		score += 20
		signals = append(signals, "observations accumulating")
	} else {
		signals = append(signals, "observations still empty")
	}
	if distinctObservationChannels(observations) >= 2 {
		score += 20
		signals = append(signals, "cross-channel evidence available")
	} else {
		signals = append(signals, "cross-channel evidence still thin")
	}
	return newProgressAxis("knowledge_pipeline", "지식 파이프라인", score, signals)
}

func deriveAgentControlAxis(room ReviewRoom, agentJobs []AgentJob, verifications []VerificationRecord, providers []ProviderSummary) ProgressAxis {
	score := 40
	signals := []string{"review-room transitions enforced", "Gemini containment lane wired"}
	if len(agentJobs) > 0 || len(providers) > 0 {
		score += 20
		signals = append(signals, "agent lanes/provider roster declared")
	} else {
		signals = append(signals, "agent lanes/provider roster still thin")
	}
	if len(verifications) > 0 {
		score += 20
		signals = append(signals, "verification evidence recorded")
	} else {
		signals = append(signals, "verification evidence not recorded yet")
	}
	if !reviewRoomHasOpenSource(room, "audit.gemini") && !reviewRoomHasOpenRationaleRule(room, "review_room.auto.agent_job_failed") {
		score += 20
		signals = append(signals, "no open containment/job-failure drift")
	} else {
		signals = append(signals, "containment or agent-failure drift still open")
	}
	return newProgressAxis("agent_control", "에이전트 통제", score, signals)
}

func deriveMultiAgentChainAxis(agentJobs []AgentJob, gatewayCalls []GatewayCall) ProgressAxis {
	score := 60
	signals := []string{"chain_rules contract wired", "job.report follow-up engine live", "http_push transport supported"}
	if chainedJobCount(agentJobs) > 0 {
		score += 20
		signals = append(signals, "chained jobs observed")
	} else {
		signals = append(signals, "chained jobs not observed yet")
	}
	if chainRuntimeEvidence(agentJobs, gatewayCalls) {
		score += 20
		signals = append(signals, "chain transport evidence recorded")
	} else {
		signals = append(signals, "chain transport evidence not recorded yet")
	}
	return newProgressAxis("multi_agent_chain", "멀티에이전트 chain", score, signals)
}

func deriveAutomationTriggerAxis(verifications []VerificationRecord, agentJobs []AgentJob) ProgressAxis {
	score := 0
	signals := []string{}
	if statusDashboardEnvEnabled("LAYER_OS_ARCHITECT_LOOP") {
		score += 20
		signals = append(signals, "architect loop enabled")
	} else {
		signals = append(signals, "architect loop disabled")
	}
	if statusDashboardEnvEnabled("LAYER_OS_ARCHITECT_AUTOVERIFY") {
		score += 20
		signals = append(signals, "architect autoverify enabled")
	} else {
		signals = append(signals, "architect autoverify disabled")
	}
	if statusDashboardEnvEnabled("LAYER_OS_ARCHITECT_GEMINI_RECOVERY") {
		score += 20
		signals = append(signals, "Gemini recovery enabled")
	} else {
		signals = append(signals, "Gemini recovery disabled")
	}
	if latestVerificationByScope(verifications, statusDashboardArchitectVerificationScope) != nil {
		score += 20
		signals = append(signals, "architect verification evidence recorded")
	} else {
		signals = append(signals, "architect verification evidence not recorded yet")
	}
	if hasAutomationGeneratedJobs(agentJobs) {
		score += 20
		signals = append(signals, "automation-generated jobs observed")
	} else {
		signals = append(signals, "automation-generated jobs not observed yet")
	}
	return newProgressAxis("automation_triggers", "자동 트리거", score, signals)
}

func deriveCommitStabilityAxis(state CompanyState, room ReviewRoom, verifications []VerificationRecord, agentJobs []AgentJob) ProgressAxis {
	score := 0
	signals := []string{}
	latestArchitect := latestVerificationByScope(verifications, statusDashboardArchitectVerificationScope)
	if latestArchitect != nil {
		score += 20
		signals = append(signals, "recent architect verification exists")
	} else {
		signals = append(signals, "recent architect verification missing")
	}
	if latestArchitect != nil && strings.TrimSpace(latestArchitect.Status) == "passed" {
		score += 20
		signals = append(signals, "latest architect verification passed")
	} else {
		signals = append(signals, "latest architect verification not green")
	}
	if !reviewRoomHasOpenRationaleRule(room, "review_room.auto.verification_failed") {
		score += 20
		signals = append(signals, "no open verification failure review item")
	} else {
		signals = append(signals, "verification failure review item still open")
	}
	if !hasActiveVerificationRepairJob(agentJobs) {
		score += 20
		signals = append(signals, "no active verification repair job backlog")
	} else {
		signals = append(signals, "verification repair backlog still active")
	}
	if strings.TrimSpace(state.DeployHealth) == "ready" {
		score += 20
		signals = append(signals, "deploy health ready")
	} else {
		signals = append(signals, "deploy health degraded")
	}
	return newProgressAxis("commit_stability", "커밋 안정화", score, signals)
}

func newProgressAxis(axisID string, label string, score int, signals []string) ProgressAxis {
	percent := clampPercent(score)
	return ProgressAxis{
		AxisID:  strings.TrimSpace(axisID),
		Label:   strings.TrimSpace(label),
		Percent: percent,
		Bar:     progressBar(percent),
		Status:  progressStatus(percent),
		Signals: append([]string{}, signals...),
	}
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func progressBar(percent int) string {
	percent = clampPercent(percent)
	filled := percent / 10
	if filled < 0 {
		filled = 0
	}
	if filled > 10 {
		filled = 10
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", 10-filled) + "]"
}

func progressStatus(percent int) string {
	percent = clampPercent(percent)
	switch {
	case percent >= 90:
		return "compounding"
	case percent >= 70:
		return "operational"
	case percent >= 40:
		return "building"
	default:
		return "foundational"
	}
}

func distinctObservationChannels(items []ObservationRecord) int {
	seen := map[string]struct{}{}
	for _, item := range items {
		channel := strings.TrimSpace(item.SourceChannel)
		if channel == "" {
			continue
		}
		seen[channel] = struct{}{}
	}
	return len(seen)
}

func reviewRoomHasOpenText(room ReviewRoom, needle string) bool {
	needle = strings.TrimSpace(strings.ToLower(needle))
	if needle == "" {
		return false
	}
	for _, item := range room.Open {
		if strings.Contains(strings.ToLower(strings.TrimSpace(item.Text)), needle) {
			return true
		}
	}
	return false
}

func reviewRoomHasOpenSource(room ReviewRoom, source string) bool {
	source = strings.TrimSpace(source)
	if source == "" {
		return false
	}
	for _, item := range room.Open {
		if strings.TrimSpace(item.Source) == source {
			return true
		}
	}
	return false
}

func reviewRoomHasOpenRationaleRule(room ReviewRoom, rule string) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false
	}
	for _, item := range room.Open {
		if item.Rationale != nil && strings.TrimSpace(item.Rationale.Rule) == rule {
			return true
		}
	}
	return false
}

func chainedJobCount(items []AgentJob) int {
	count := 0
	for _, item := range items {
		if strings.TrimSpace(item.Source) == "agent_job.chain" {
			count++
		}
	}
	return count
}

func chainRuntimeEvidence(agentJobs []AgentJob, gatewayCalls []GatewayCall) bool {
	for _, item := range agentJobs {
		if item.Payload != nil {
			if _, ok := item.Payload["chain_rules"]; ok {
				return true
			}
		}
		if item.Result != nil {
			if transport, ok := item.Result["dispatch_transport"].(string); ok && strings.TrimSpace(transport) == "http_push" {
				return true
			}
		}
	}
	for _, item := range gatewayCalls {
		if strings.TrimSpace(item.Status) == "sent" {
			return true
		}
		for _, note := range item.Notes {
			if strings.Contains(strings.TrimSpace(note), "dispatch:http_status=") {
				return true
			}
		}
	}
	return false
}

func latestVerificationByScope(items []VerificationRecord, scope string) *VerificationRecord {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return nil
	}
	for index := len(items) - 1; index >= 0; index-- {
		item := items[index]
		if strings.TrimSpace(item.Scope) != scope {
			continue
		}
		copy := item
		return &copy
	}
	return nil
}

func hasAutomationGeneratedJobs(items []AgentJob) bool {
	for _, item := range items {
		source := strings.TrimSpace(item.Source)
		if source == "knowledge.parallel_candidate" || source == "knowledge.open_thread" || source == "architect.verification_failed" {
			return true
		}
	}
	return false
}

func hasActiveVerificationRepairJob(items []AgentJob) bool {
	for _, item := range items {
		if strings.TrimSpace(item.Source) != "architect.verification_failed" {
			continue
		}
		status := strings.TrimSpace(item.Status)
		if status == "queued" || status == "running" || status == "failed" {
			return true
		}
	}
	return false
}

func statusDashboardEnvEnabled(key string) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

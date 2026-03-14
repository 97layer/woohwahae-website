package api

import (
	"strings"
	"time"

	"layer-os/internal/runtime"
)

const cockpitBrandPrepTopic = "brand_publish_prep"

type cockpitTelegramSnapshot struct {
	Packet  runtime.TelegramPacket `json:"packet"`
	Enabled bool                   `json:"enabled"`
	Adapter string                 `json:"adapter"`
	Status  runtime.TelegramStatus `json:"status"`
}

type cockpitThreadsSnapshot struct {
	Enabled bool                  `json:"enabled"`
	Adapter string                `json:"adapter"`
	Status  runtime.ThreadsStatus `json:"status"`
}

type cockpitJobCounts struct {
	Open        int    `json:"open"`
	Queued      int    `json:"queued"`
	Running     int    `json:"running"`
	Terminal    int    `json:"terminal"`
	SummaryNote string `json:"summary_note"`
	SummaryMeta string `json:"summary_meta"`
}

type cockpitSourceIntakeAttention struct {
	Mode                string `json:"mode"`
	Summary             string `json:"summary"`
	Detail              string `json:"detail"`
	Ref                 string `json:"ref,omitempty"`
	SourceObservationID string `json:"source_observation_id,omitempty"`
	DraftObservationID  string `json:"draft_observation_id,omitempty"`
}

type cockpitSourceIntakeSnapshot struct {
	GeneratedAt     time.Time                    `json:"generated_at"`
	RecentCount     int                          `json:"recent_count"`
	ActionCount     int                          `json:"action_count"`
	QuietCount      int                          `json:"quiet_count"`
	FeedCount       int                          `json:"feed_count"`
	DraftSeedCount  int                          `json:"draft_seed_count"`
	PrepCount       int                          `json:"prep_count"`
	ReviewCount     int                          `json:"review_count"`
	PrepReadyCount  int                          `json:"prep_ready_count"`
	AutoRoutedCount int                          `json:"auto_routed_count"`
	SummaryNote     string                       `json:"summary_note"`
	SummaryMeta     string                       `json:"summary_meta"`
	QuietNote       string                       `json:"quiet_note"`
	Attention       cockpitSourceIntakeAttention `json:"attention"`
}

type cockpitFounderAttentionSnapshot struct {
	Mode            string   `json:"mode"`
	Summary         string   `json:"summary"`
	Detail          string   `json:"detail"`
	Ref             string   `json:"ref,omitempty"`
	Lane            string   `json:"lane,omitempty"`
	Source          string   `json:"source,omitempty"`
	Rule            string   `json:"rule,omitempty"`
	WaitingCount    int      `json:"waiting_count"`
	RiskCount       int      `json:"risk_count"`
	ReviewOpenCount int      `json:"review_open_count"`
	NowCount        int      `json:"now_count"`
	NextJobIDs      []string `json:"next_job_ids,omitempty"`
	LastReleaseRef  string   `json:"last_release_ref,omitempty"`
	LastRollbackRef string   `json:"last_rollback_ref,omitempty"`
}

type cockpitFounderFlowAttentionSnapshot struct {
	Summary      string `json:"summary"`
	Ref          string `json:"ref,omitempty"`
	WaitingCount int    `json:"waiting_count"`
	RiskCount    int    `json:"risk_count"`
	Detail       string `json:"detail"`
}

type cockpitFounderFlowDefaultsSnapshot struct {
	FlowID     string                              `json:"flow_id,omitempty"`
	ApprovalID string                              `json:"approval_id,omitempty"`
	Title      string                              `json:"title"`
	Intent     string                              `json:"intent"`
	Attention  cockpitFounderFlowAttentionSnapshot `json:"attention"`
}

type cockpitDashboardSnapshot struct {
	GeneratedAt      time.Time                            `json:"generated_at"`
	Status           string                               `json:"status"`
	Daemon           runtime.DaemonStatus                 `json:"daemon"`
	CompanyState     runtime.CompanyState                 `json:"company_state"`
	Security         runtime.SecurityAudit                `json:"security"`
	Auth             runtime.AuthStatus                   `json:"auth"`
	Adapters         runtime.AdapterSummary               `json:"adapters"`
	Providers        []runtime.ProviderSummary            `json:"providers"`
	Capabilities     runtime.CapabilityRegistry           `json:"capabilities"`
	WriteLease       runtime.WriteLease                   `json:"write_lease"`
	Telegram         cockpitTelegramSnapshot              `json:"telegram"`
	Threads          cockpitThreadsSnapshot               `json:"threads"`
	Conversation     runtime.ConversationAutomationStatus `json:"conversation"`
	SessionBootstrap *runtime.SessionBootstrapPacket      `json:"session_bootstrap,omitempty"`
	BootstrapError   *string                              `json:"bootstrap_error,omitempty"`
	ReviewRoom       runtime.ReviewRoomSummary            `json:"review_room"`
	ReviewNote       string                               `json:"review_note"`
	ReviewMeta       string                               `json:"review_meta"`
	Knowledge        runtime.KnowledgePacket              `json:"knowledge"`
	Handoff          runtime.HandoffSummary               `json:"handoff"`
	FounderSummary   runtime.FounderSummary               `json:"founder_summary"`
	FounderView      runtime.FounderView                  `json:"founder_view"`
	FounderAttention cockpitFounderAttentionSnapshot      `json:"founder_attention"`
	PrimaryAttention cockpitFounderAttentionSnapshot      `json:"primary_attention"`
	FounderFlow      cockpitFounderFlowDefaultsSnapshot   `json:"founder_flow_defaults"`
	SourceIntake     cockpitSourceIntakeSnapshot          `json:"source_intake"`
	OpenJobs         []runtime.AgentJob                   `json:"open_jobs"`
	JobCounts        cockpitJobCounts                     `json:"job_counts"`
	RecentEvents     []runtime.EventEnvelope              `json:"recent_events"`
}

func buildCockpitDashboardSnapshot(service *runtime.Service, daemonInfo DaemonRuntimeInfo) cockpitDashboardSnapshot {
	daemon := daemonStatus(service, daemonInfo)
	telegramStatus := service.TelegramStatus()
	threadsStatus := service.ThreadsStatus()
	reviewSummary := service.ReviewRoomSummary()
	openJobs, jobCounts := summarizeOpenJobs(service.ListAgentJobs())
	events := recentEvents(service.ListEvents(), 20)
	founderSummary := service.FounderSummary()
	founderView := service.FounderView()
	sourceIntake := summarizeSourceIntake(service)
	founderAttention := summarizeFounderAttention(founderSummary, founderView, reviewSummary, openJobs)
	primaryAttention := summarizePrimaryAttention(founderAttention, openJobs, sourceIntake)
	snapshot := cockpitDashboardSnapshot{
		GeneratedAt:  time.Now().UTC(),
		Status:       daemon.Status,
		Daemon:       daemon,
		CompanyState: service.Status(),
		Security:     service.SecurityAudit(),
		Auth:         service.AuthStatus(),
		Adapters:     service.Adapters(),
		Providers:    publicProviderSummaries(service.Providers()),
		Capabilities: service.Capabilities(),
		WriteLease:   service.WriteLease(),
		Telegram: cockpitTelegramSnapshot{
			Packet:  service.Telegram(),
			Enabled: service.TelegramEnabled(),
			Adapter: service.TelegramAdapterName(),
			Status:  telegramStatus,
		},
		Threads: cockpitThreadsSnapshot{
			Enabled: threadsStatus.PublishConfigured,
			Adapter: threadsStatus.Adapter,
			Status:  threadsStatus,
		},
		Conversation:     service.ConversationAutomationStatus(),
		ReviewRoom:       reviewSummary,
		ReviewNote:       summarizeReviewNote(reviewSummary),
		ReviewMeta:       summarizeReviewMeta(reviewSummary),
		Knowledge:        service.Knowledge(),
		Handoff:          service.HandoffSummary(),
		FounderSummary:   founderSummary,
		FounderView:      founderView,
		FounderAttention: founderAttention,
		PrimaryAttention: primaryAttention,
		FounderFlow:      summarizeFounderFlowDefaults(founderSummary, founderAttention, primaryAttention, openJobs),
		SourceIntake:     sourceIntake,
		OpenJobs:         openJobs,
		JobCounts:        jobCounts,
		RecentEvents:     events,
	}
	packet := service.SessionBootstrap("daemon", false, false)
	snapshot.SessionBootstrap = &packet
	return snapshot
}

func buildCockpitPayload(service *runtime.Service, daemonInfo DaemonRuntimeInfo, full bool) map[string]any {
	dashboard := buildCockpitDashboardSnapshot(service, daemonInfo)
	payload := map[string]any{
		"status":                dashboard.Status,
		"daemon":                dashboard.Daemon,
		"company_state":         dashboard.CompanyState,
		"knowledge":             dashboard.Knowledge,
		"handoff":               dashboard.Handoff,
		"conversation":          dashboard.Conversation,
		"founder_view":          dashboard.FounderView,
		"founder_summary":       dashboard.FounderSummary,
		"founder_attention":     dashboard.FounderAttention,
		"primary_attention":     dashboard.PrimaryAttention,
		"founder_flow_defaults": dashboard.FounderFlow,
		"source_intake":         dashboard.SourceIntake,
		"session_bootstrap":     dashboard.SessionBootstrap,
		"bootstrap_error":       dashboard.BootstrapError,
		"review_summary":        dashboard.ReviewRoom,
		"adapters":              dashboard.Adapters,
		"capabilities":          dashboard.Capabilities,
		"providers":             dashboard.Providers,
		"telegram":              dashboard.Telegram,
		"threads":               dashboard.Threads,
		"write_lease":           dashboard.WriteLease,
		"auth":                  dashboard.Auth,
		"dashboard":             dashboard,
		"releases":              service.ListReleases(),
		"deploys":               service.ListDeploys(),
		"rollbacks":             service.ListRollbacks(),
		"targets":               service.ListTargets(),
		"verifications":         service.ListVerifications(),
		"security":              dashboard.Security,
	}
	if !full {
		return payload
	}
	payload["review_room"] = service.ReviewRoom()
	payload["branches"] = service.ListBranches()
	payload["work_items"] = service.ListWorkItems()
	payload["flows"] = service.ListFlows()
	payload["approvals"] = service.ListApprovals()
	payload["events"] = service.ListEvents()
	payload["policies"] = service.ListPolicies()
	payload["gateway_calls"] = service.ListGatewayCalls()
	payload["executes"] = service.ListExecutes()
	payload["memory"] = service.Memory()
	payload["snapshot"] = service.Snapshot()
	payload["structure"] = runtime.AuditStructure(repoRoot())
	payload["contracts"] = runtime.AuditContracts(repoRoot())
	payload["residue"] = runtime.AuditResidue(repoRoot())
	payload["gemini"] = runtime.AuditGemini(repoRoot())
	payload["surface"] = runtime.AuditSurface(repoRoot())
	return payload
}

func publicProviderSummaries(items []runtime.ProviderSummary) []runtime.ProviderSummary {
	sanitized := make([]runtime.ProviderSummary, 0, len(items))
	for _, item := range items {
		copy := item
		copy.AuthReady = false
		copy.AuthSource = nil
		copy.AuthEnvKeys = []string{}
		copy.Notes = publicProviderNotes(copy.Notes)
		sanitized = append(sanitized, copy)
	}
	return sanitized
}

func publicProviderNotes(items []string) []string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "provider credentials ") {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func summarizeSourceIntake(service *runtime.Service) cockpitSourceIntakeSnapshot {
	view := buildSourceIntakeView(service)
	return cockpitSourceIntakeSnapshot{
		GeneratedAt:     view.GeneratedAt,
		RecentCount:     view.RecentCount,
		ActionCount:     view.ActionCount,
		QuietCount:      view.QuietCount,
		FeedCount:       view.FeedCount,
		DraftSeedCount:  view.DraftSeedCount,
		PrepCount:       view.PrepCount,
		ReviewCount:     view.ReviewCount,
		PrepReadyCount:  view.PrepReadyCount,
		AutoRoutedCount: view.AutoRoutedCount,
		SummaryNote:     nonEmptyString(view.SummaryNote, "링크와 메모를 먼저 쌓는 입구입니다."),
		SummaryMeta:     nonEmptyString(view.SummaryMeta, "action 0 · quiet 0 · drafts 0 · prep 0"),
		QuietNote:       nonEmptyString(view.QuietNote, "quiet candidate는 아직 없습니다."),
		Attention: cockpitSourceIntakeAttention{
			Mode:                view.Attention.Mode,
			Summary:             view.Attention.Summary,
			Detail:              view.Attention.Detail,
			Ref:                 view.Attention.Ref,
			SourceObservationID: view.Attention.SourceObservationID,
			DraftObservationID:  view.Attention.DraftObservationID,
		},
	}
}

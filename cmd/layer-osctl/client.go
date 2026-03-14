package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type cliService interface {
	Adapters() runtime.AdapterSummary
	Providers() []runtime.ProviderSummary
	ListCapitalizationEntries() []runtime.CapitalizationEntry
	SearchKnowledge(query string) runtime.KnowledgeSearchResponse
	ConversationStatus() runtime.ConversationAutomationStatus
	ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord
	ListOpenThreads(limit int) []runtime.OpenThread
	Capabilities() runtime.CapabilityRegistry
	Knowledge() runtime.KnowledgePacket
	Telegram() runtime.TelegramPacket
	SendTelegram() error
	TelegramEnabled() bool
	ThreadsStatus() runtime.ThreadsStatus
	PublishThreads(approvalID string) (runtime.ThreadsPublishReceipt, error)
	Status() runtime.CompanyState
	ListBranches() []runtime.Branch
	ListProposals() []runtime.ProposalItem
	ListAgentJobs() []runtime.AgentJob
	ListWorkItems() []runtime.WorkItem
	ListFlows() []runtime.FlowRun
	CreateBranch(item runtime.Branch) error
	CreateProposal(item runtime.ProposalItem) error
	CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error)
	CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error)
	RecoverCorpusMarkdown(limit int, cleanup bool, dryRun bool) (runtime.CorpusMarkdownRecoverResult, error)
	MergeBranch(branchID string, targetBranchID string, notes []string) (runtime.Branch, error)
	PromoteProposal(proposalID string, workID string) (runtime.ProposalItem, runtime.WorkItem, error)
	CreateAgentJob(item runtime.AgentJob) error
	PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error)
	UpdateAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJob, error)
	ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error)
	DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error)
	AgentDispatchProfiles() []runtime.AgentDispatchProfile
	AgentRunPacket(jobID string) (runtime.AgentRunPacket, error)
	CreateWorkItem(item runtime.WorkItem) error
	CreateFlow(item runtime.FlowRun) error
	SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (runtime.FlowRun, error)
	ListApprovals() []runtime.ApprovalItem
	ListEvents() []runtime.EventEnvelope
	ListReleases() []runtime.ReleasePacket
	ListDeploys() []runtime.DeployRun
	ListRollbacks() []runtime.RollbackRun
	ListTargets() []runtime.DeployTarget
	ListPreflights() []runtime.PreflightRecord
	ListPolicies() []runtime.PolicyDecision
	ListGatewayCalls() []runtime.GatewayCall
	ListExecutes() []runtime.ExecuteRun
	ListVerifications() []runtime.VerificationRecord
	CreateEvent(input runtime.EventCreateInput) (runtime.EventEnvelope, error)
	StreamEvents(out io.Writer) error
	Memory() runtime.SystemMemory
	ReviewRoom() runtime.ReviewRoom
	ReviewRoomSummary() runtime.ReviewRoomSummary
	AddReviewRoomItem(section string, item runtime.ReviewRoomItem) (runtime.ReviewRoom, error)
	TransitionReviewRoomItem(action string, item string) (runtime.ReviewRoom, error)
	TransitionStructuredReviewRoomItem(action string, item string, resolution *runtime.ReviewRoomResolution) (runtime.ReviewRoom, error)
	WriteLease() runtime.WriteLease
	Handoff() runtime.HandoffPacket
	SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error)
	CheckpointSession(input runtime.SessionCheckpointInput) (runtime.SessionCheckpoint, error)
	SessionNote(input runtime.SessionNoteInput) (runtime.SessionNoteResult, error)
	SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error)
	AuthStatus() runtime.AuthStatus
	SetWriteToken(token string) (runtime.AuthStatus, error)
	ClearWriteToken() (runtime.AuthStatus, error)
	ReplaceMemory(memory runtime.SystemMemory) error
	CreateApproval(item runtime.ApprovalItem) error
	CreatePreflight(item runtime.PreflightRecord) error
	EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (runtime.PolicyDecision, error)
	CreateGatewayCall(item runtime.GatewayCall) error
	RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (runtime.ExecuteRun, error)
	RunVerification(recordID string, scope string, workdir string, command []string, baseNotes []string) (runtime.VerificationRecord, error)
	ResolveApproval(approvalID string, status string) (runtime.ApprovalItem, error)
	CreateRelease(item runtime.ReleasePacket) error
	CreateDeploy(item runtime.DeployRun) error
	CreateRollback(item runtime.RollbackRun) error
	PutTarget(item runtime.DeployTarget) error
	ExecuteDeploy(deployID string, releaseID string, baseNotes []string) (runtime.DeployRun, error)
	ExecuteRollback(rollbackID string, releaseID string, deployID string, baseNotes []string) (runtime.RollbackRun, error)
	FounderSummary() runtime.FounderSummary
	Snapshot() runtime.SnapshotPacket
	ImportSnapshot(packet runtime.SnapshotPacket) error
	StartFounderFlow(flowID string, workID string, approvalID string, title string, intent string, notes []string) (runtime.FlowRun, error)
	ApproveFounderFlow(flowID string, notes []string) (runtime.FlowRun, error)
	ReleaseFounderFlow(flowID string, releaseID string, deployID string, target string, channel string, notes []string) (runtime.FlowRun, error)
	RollbackFounderFlow(flowID string, rollbackID string, notes []string) (runtime.FlowRun, error)
}

type authStatusProvider interface {
	AuthStatus() runtime.AuthStatus
}

type daemonClient struct {
	baseURL    string
	httpClient *http.Client
}

func newDaemonClient() *daemonClient {
	return &daemonClient{
		baseURL: daemonBaseURL(),
		httpClient: &http.Client{
			Timeout: defaultDaemonHTTPTimeout(),
		},
	}
}

func defaultDaemonHTTPTimeout() time.Duration {
	return daemonHTTPTimeoutFromEnv("LAYER_OS_HTTP_TIMEOUT", 30*time.Second)
}

func longRunningDaemonHTTPTimeout() time.Duration {
	return daemonHTTPTimeoutFromEnv("LAYER_OS_LONG_HTTP_TIMEOUT", 5*time.Minute)
}

func daemonHTTPTimeoutFromEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func daemonRequestTimeout(method string, path string) time.Duration {
	if strings.EqualFold(method, http.MethodPost) {
		switch path {
		case "/api/layer-os/deploys/execute",
			"/api/layer-os/rollbacks/execute",
			"/api/layer-os/founder-actions/release",
			"/api/layer-os/founder-actions/rollback":
			return longRunningDaemonHTTPTimeout()
		}
	}
	return defaultDaemonHTTPTimeout()
}

func daemonBaseURL() string {
	if raw := strings.TrimSpace(os.Getenv("LAYER_OS_BASE_URL")); raw != "" {
		return strings.TrimRight(raw, "/")
	}
	addr := strings.TrimSpace(os.Getenv("LAYER_OS_ADDR"))
	switch {
	case addr == "":
		return "http://" + runtime.DefaultDaemonAddr
	case strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://"):
		return strings.TrimRight(addr, "/")
	case strings.HasPrefix(addr, ":"):
		return "http://127.0.0.1" + addr
	default:
		return "http://" + strings.TrimRight(addr, "/")
	}
}

func requestActor() string {
	return runtime.ResolveActor(os.Getenv("LAYER_OS_ACTOR"), os.Getenv("LAYER_OS_WRITER_ID"))
}

func requestModels() []string {
	return mergeModelLists(splitCSV(os.Getenv("LAYER_OS_MODELS")), splitCSV(os.Getenv("LAYER_OS_MODEL")))
}

func mergeModelLists(groups ...[]string) []string {
	seen := map[string]bool{}
	items := []string{}
	for _, group := range groups {
		for _, raw := range group {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			if seen[value] {
				continue
			}
			seen[value] = true
			items = append(items, value)
		}
	}
	return items
}

type apiError struct {
	Error string `json:"error"`
}

func (c *daemonClient) Adapters() runtime.AdapterSummary {
	var response struct {
		Adapters runtime.AdapterSummary `json:"adapters"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/adapters", nil, &response)
	return response.Adapters
}

func (c *daemonClient) Providers() []runtime.ProviderSummary {
	var response struct {
		Providers []runtime.ProviderSummary `json:"providers"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/providers", nil, &response)
	return response.Providers
}

func (c *daemonClient) ListCapitalizationEntries() []runtime.CapitalizationEntry {
	var response struct {
		Entries []runtime.CapitalizationEntry `json:"entries"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/corpus/entries", nil, &response)
	return response.Entries
}

func (c *daemonClient) SearchKnowledge(query string) runtime.KnowledgeSearchResponse {
	path := "/api/layer-os/knowledge/search?q=" + url.QueryEscape(strings.TrimSpace(query))
	var response runtime.KnowledgeSearchResponse
	c.mustRequest(http.MethodGet, path, nil, &response)
	return response
}

func (c *daemonClient) ListObservations(query runtime.ObservationQuery) []runtime.ObservationRecord {
	values := url.Values{}
	if value := strings.TrimSpace(query.SourceChannel); value != "" {
		values.Set("source_channel", value)
	}
	if value := strings.TrimSpace(query.Topic); value != "" {
		values.Set("topic", value)
	}
	if value := strings.TrimSpace(query.Actor); value != "" {
		values.Set("actor", value)
	}
	if value := strings.TrimSpace(query.Ref); value != "" {
		values.Set("ref", value)
	}
	if value := strings.TrimSpace(query.Text); value != "" {
		values.Set("text", value)
	}
	if query.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	path := "/api/layer-os/observations"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var response struct {
		Items []runtime.ObservationRecord `json:"items"`
	}
	c.mustRequest(http.MethodGet, path, nil, &response)
	return response.Items
}

func (c *daemonClient) ConversationStatus() runtime.ConversationAutomationStatus {
	var response struct {
		Conversation runtime.ConversationAutomationStatus `json:"conversation"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/conversation", nil, &response)
	return response.Conversation
}

func (c *daemonClient) ListOpenThreads(limit int) []runtime.OpenThread {
	path := "/api/layer-os/threads"
	if limit > 0 {
		path += "?limit=" + url.QueryEscape(fmt.Sprintf("%d", limit))
	}
	var response struct {
		Items []runtime.OpenThread `json:"items"`
	}
	c.mustRequest(http.MethodGet, path, nil, &response)
	return response.Items
}

func (c *daemonClient) Capabilities() runtime.CapabilityRegistry {
	var response struct {
		Capabilities runtime.CapabilityRegistry `json:"capabilities"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/capabilities", nil, &response)
	return response.Capabilities
}

func (c *daemonClient) Knowledge() runtime.KnowledgePacket {
	var response struct {
		Knowledge runtime.KnowledgePacket `json:"knowledge"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/knowledge", nil, &response)
	return response.Knowledge
}

func (c *daemonClient) Telegram() runtime.TelegramPacket {
	var response struct {
		Telegram runtime.TelegramPacket `json:"telegram"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/telegram", nil, &response)
	return response.Telegram
}

func (c *daemonClient) SendTelegram() error {
	var response struct {
		OK bool `json:"ok"`
	}
	c.mustRequest(http.MethodPost, "/api/layer-os/telegram", map[string]any{}, &response)
	return nil
}

func (c *daemonClient) TelegramEnabled() bool {
	var response struct {
		Enabled bool `json:"enabled"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/telegram", nil, &response)
	return response.Enabled
}

func (c *daemonClient) CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error) {
	var created runtime.ConversationNoteResult
	err := c.request(http.MethodPost, "/api/layer-os/conversation", input, &created)
	return created, err
}

func (c *daemonClient) ThreadsStatus() runtime.ThreadsStatus {
	var response struct {
		Status runtime.ThreadsStatus `json:"status"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/social/threads", nil, &response)
	return response.Status
}

func (c *daemonClient) PublishThreads(approvalID string) (runtime.ThreadsPublishReceipt, error) {
	var item runtime.ThreadsPublishReceipt
	err := c.request(http.MethodPost, "/api/layer-os/social/threads", map[string]any{
		"approval_id": strings.TrimSpace(approvalID),
	}, &item)
	return item, err
}

func (c *daemonClient) Status() runtime.CompanyState {
	var response struct {
		CompanyState runtime.CompanyState `json:"company_state"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/status", nil, &response)
	return response.CompanyState
}

func (c *daemonClient) DaemonStatus() runtime.DaemonStatus {
	item, err := c.TryDaemonStatus()
	if err != nil {
		log.Fatal(err)
	}
	return item
}

func (c *daemonClient) TryDaemonStatus() (runtime.DaemonStatus, error) {
	var response struct {
		Daemon runtime.DaemonStatus `json:"daemon"`
	}
	err := c.request(http.MethodGet, "/api/layer-os/daemon", nil, &response)
	return response.Daemon, err
}

func (c *daemonClient) ListProposals() []runtime.ProposalItem {
	var response struct {
		Items []runtime.ProposalItem `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/proposals", nil, &response)
	return response.Items
}

func (c *daemonClient) ListBranches() []runtime.Branch {
	var response struct {
		Items []runtime.Branch `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/branches", nil, &response)
	return response.Items
}

func (c *daemonClient) ListAgentJobs() []runtime.AgentJob {
	return c.ListAgentJobsByStatus("", 0)
}

func (c *daemonClient) ListAgentJobsByStatus(status string, limit int) []runtime.AgentJob {
	var response struct {
		Items []runtime.AgentJob `json:"items"`
	}
	query := url.Values{}
	if strings.TrimSpace(status) != "" {
		query.Set("status", strings.TrimSpace(status))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/api/layer-os/jobs"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	c.mustRequest(http.MethodGet, path, nil, &response)
	return response.Items
}

func (c *daemonClient) ListWorkItems() []runtime.WorkItem {
	var response struct {
		Items []runtime.WorkItem `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/work-items", nil, &response)
	return response.Items
}

func (c *daemonClient) ListFlows() []runtime.FlowRun {
	var response struct {
		Items []runtime.FlowRun `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/flows", nil, &response)
	return response.Items
}

func (c *daemonClient) CreateProposal(item runtime.ProposalItem) error {
	return c.request(http.MethodPost, "/api/layer-os/proposals", item, nil)
}

func (c *daemonClient) CreateBranch(item runtime.Branch) error {
	return c.request(http.MethodPost, "/api/layer-os/branches", item, nil)
}

func (c *daemonClient) CreateObservation(item runtime.ObservationRecord) (runtime.ObservationRecord, error) {
	var created runtime.ObservationRecord
	err := c.request(http.MethodPost, "/api/layer-os/observations", item, &created)
	return created, err
}

func (c *daemonClient) RecoverCorpusMarkdown(limit int, cleanup bool, dryRun bool) (runtime.CorpusMarkdownRecoverResult, error) {
	var result runtime.CorpusMarkdownRecoverResult
	err := c.request(http.MethodPost, "/api/layer-os/corpus/recover", map[string]any{"limit": limit, "cleanup": cleanup, "dry_run": dryRun}, &result)
	return result, err
}

func (c *daemonClient) PromoteProposal(proposalID string, workID string) (runtime.ProposalItem, runtime.WorkItem, error) {
	var response struct {
		Proposal runtime.ProposalItem `json:"proposal"`
		WorkItem runtime.WorkItem     `json:"work_item"`
	}
	err := c.request(http.MethodPost, "/api/layer-os/proposals/promote", map[string]any{
		"proposal_id":  proposalID,
		"work_item_id": workID,
	}, &response)
	return response.Proposal, response.WorkItem, err
}

func (c *daemonClient) MergeBranch(branchID string, targetBranchID string, notes []string) (runtime.Branch, error) {
	var item runtime.Branch
	err := c.request(http.MethodPost, "/api/layer-os/branches/merge", map[string]any{
		"branch_id":        branchID,
		"target_branch_id": targetBranchID,
		"notes":            notes,
	}, &item)
	return item, err
}

func (c *daemonClient) CreateAgentJob(item runtime.AgentJob) error {
	return c.request(http.MethodPost, "/api/layer-os/jobs", item, nil)
}

func (c *daemonClient) PromoteContextJobs(limit int, dispatch bool) (runtime.AgentJobPromotionResult, error) {
	var result runtime.AgentJobPromotionResult
	err := c.request(http.MethodPost, "/api/layer-os/jobs/promote", map[string]any{"limit": limit, "dispatch": dispatch}, &result)
	return result, err
}

func (c *daemonClient) UpdateAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJob, error) {
	var item runtime.AgentJob
	err := c.request(http.MethodPost, "/api/layer-os/jobs/update", map[string]any{
		"job_id": jobID,
		"status": status,
		"notes":  notes,
		"result": result,
	}, &item)
	return item, err
}

func (c *daemonClient) ReportAgentJob(jobID string, status string, notes []string, result map[string]any) (runtime.AgentJobReportResult, error) {
	var item runtime.AgentJobReportResult
	err := c.request(http.MethodPost, "/api/layer-os/jobs/report", map[string]any{"job_id": jobID, "status": status, "notes": notes, "result": result}, &item)
	return item, err
}

func (c *daemonClient) DispatchAgentJob(jobID string) (runtime.AgentDispatchResult, error) {
	var result runtime.AgentDispatchResult
	err := c.request(http.MethodPost, "/api/layer-os/jobs/dispatch", map[string]any{"job_id": jobID}, &result)
	return result, err
}

func (c *daemonClient) AgentDispatchProfiles() []runtime.AgentDispatchProfile {
	var response struct {
		Items []runtime.AgentDispatchProfile `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/jobs/profiles", nil, &response)
	return response.Items
}

func (c *daemonClient) AgentRunPacket(jobID string) (runtime.AgentRunPacket, error) {
	var packet runtime.AgentRunPacket
	err := c.request(http.MethodGet, "/api/layer-os/jobs/packet?job_id="+url.QueryEscape(jobID), nil, &packet)
	return packet, err
}

func (c *daemonClient) CreateWorkItem(item runtime.WorkItem) error {
	return c.request(http.MethodPost, "/api/layer-os/work-items", item, nil)
}

func (c *daemonClient) CreateFlow(item runtime.FlowRun) error {
	return c.request(http.MethodPost, "/api/layer-os/flows", item, nil)
}

func (c *daemonClient) SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (runtime.FlowRun, error) {
	payload := map[string]any{
		"flow_id":      flowID,
		"work_item_id": workItemID,
		"notes":        notes,
	}
	if approvalID != nil {
		payload["approval_id"] = *approvalID
	}
	if policyDecisionID != nil {
		payload["policy_decision_id"] = *policyDecisionID
	}
	if executeID != nil {
		payload["execute_id"] = *executeID
	}
	if verificationID != nil {
		payload["verification_id"] = *verificationID
	}
	if releaseID != nil {
		payload["release_id"] = *releaseID
	}
	if deployID != nil {
		payload["deploy_id"] = *deployID
	}
	if rollbackID != nil {
		payload["rollback_id"] = *rollbackID
	}
	var item runtime.FlowRun
	err := c.request(http.MethodPost, "/api/layer-os/flows/sync", payload, &item)
	return item, err
}

func (c *daemonClient) ListApprovals() []runtime.ApprovalItem {
	var response struct {
		Items []runtime.ApprovalItem `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/approval-inbox", nil, &response)
	return response.Items
}

func (c *daemonClient) ListEvents() []runtime.EventEnvelope {
	var response struct {
		Items []runtime.EventEnvelope `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/events", nil, &response)
	return response.Items
}

func (c *daemonClient) CreateEvent(input runtime.EventCreateInput) (runtime.EventEnvelope, error) {
	var event runtime.EventEnvelope
	err := c.request(http.MethodPost, "/api/layer-os/events", input, &event)
	return event, err
}

func (c *daemonClient) StreamEvents(out io.Writer) error {
	attempts := streamConnectAttempts()
	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := c.doRequest(http.MethodGet, "/api/layer-os/events/stream", nil)
		if err != nil {
			return err
		}
		err = copyDaemonStream(out, response)
		if err == nil {
			return nil
		}
		if !isDaemonUnavailable(err) || attempt == attempts {
			return err
		}
		daemonRetrySleep(daemonReadRetryDelay)
	}
	return nil
}

func copyDaemonStream(out io.Writer, response *http.Response) error {
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return decodeAPIError(response)
	}
	if _, err := io.Copy(out, response.Body); err != nil {
		return wrapDaemonError(err)
	}
	return nil
}

func (c *daemonClient) ListReleases() []runtime.ReleasePacket {
	var response struct {
		Items []runtime.ReleasePacket `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/releases", nil, &response)
	return response.Items
}

func (c *daemonClient) ListDeploys() []runtime.DeployRun {
	var response struct {
		Items []runtime.DeployRun `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/deploys", nil, &response)
	return response.Items
}

func (c *daemonClient) ListRollbacks() []runtime.RollbackRun {
	var response struct {
		Items []runtime.RollbackRun `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/rollbacks", nil, &response)
	return response.Items
}

func (c *daemonClient) ListTargets() []runtime.DeployTarget {
	var response struct {
		Items []runtime.DeployTarget `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/deploy-targets", nil, &response)
	return response.Items
}

func (c *daemonClient) ListPreflights() []runtime.PreflightRecord {
	var response struct {
		Items []runtime.PreflightRecord `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/preflights", nil, &response)
	return response.Items
}

func (c *daemonClient) ListPolicies() []runtime.PolicyDecision {
	var response struct {
		Items []runtime.PolicyDecision `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/policies", nil, &response)
	return response.Items
}

func (c *daemonClient) ListGatewayCalls() []runtime.GatewayCall {
	var response struct {
		Items []runtime.GatewayCall `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/gateway-calls", nil, &response)
	return response.Items
}

func (c *daemonClient) ListExecutes() []runtime.ExecuteRun {
	var response struct {
		Items []runtime.ExecuteRun `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/execute-runs", nil, &response)
	return response.Items
}

func (c *daemonClient) ListVerifications() []runtime.VerificationRecord {
	var response struct {
		Items []runtime.VerificationRecord `json:"items"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/verifications", nil, &response)
	return response.Items
}

func (c *daemonClient) Memory() runtime.SystemMemory {
	var response struct {
		Memory runtime.SystemMemory `json:"memory"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/memory", nil, &response)
	return response.Memory
}

func (c *daemonClient) ReviewRoom() runtime.ReviewRoom {
	var response struct {
		ReviewRoom runtime.ReviewRoom `json:"review_room"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/review-room", nil, &response)
	return response.ReviewRoom
}

func (c *daemonClient) ReviewRoomSummary() runtime.ReviewRoomSummary {
	var response struct {
		Summary runtime.ReviewRoomSummary `json:"summary"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/review-room", nil, &response)
	return response.Summary
}

func (c *daemonClient) AddReviewRoomItem(section string, item runtime.ReviewRoomItem) (runtime.ReviewRoom, error) {
	var response struct {
		ReviewRoom runtime.ReviewRoom `json:"review_room"`
	}
	payload := map[string]any{
		"section":  section,
		"item":     item.Text,
		"kind":     item.Kind,
		"severity": item.Severity,
		"source":   item.Source,
	}
	if item.Ref != nil {
		payload["ref"] = *item.Ref
	}
	if strings.TrimSpace(item.Why) != "" {
		payload["why"] = item.Why
	}
	if item.WhyUnresolved != nil {
		payload["why_unresolved"] = *item.WhyUnresolved
	}
	if item.Contradiction != nil {
		payload["contradiction"] = *item.Contradiction
	}
	if len(item.Contradictions) > 0 {
		payload["contradictions"] = append([]string{}, item.Contradictions...)
	}
	if len(item.PatternRefs) > 0 {
		payload["pattern_refs"] = append([]string{}, item.PatternRefs...)
	}
	err := c.request(http.MethodPost, "/api/layer-os/review-room", payload, &response)
	return response.ReviewRoom, err
}

func (c *daemonClient) TransitionReviewRoomItem(action string, item string) (runtime.ReviewRoom, error) {
	return c.TransitionStructuredReviewRoomItem(action, item, nil)
}

func (c *daemonClient) TransitionStructuredReviewRoomItem(action string, item string, resolution *runtime.ReviewRoomResolution) (runtime.ReviewRoom, error) {
	var response struct {
		ReviewRoom runtime.ReviewRoom `json:"review_room"`
	}
	payload := map[string]any{
		"action": action,
		"item":   item,
	}
	if resolution != nil {
		payload["reason"] = resolution.Reason
		payload["rule"] = resolution.Rule
		if len(resolution.Evidence) > 0 {
			payload["evidence"] = append([]string{}, resolution.Evidence...)
		}
	}
	err := c.request(http.MethodPost, "/api/layer-os/review-room", payload, &response)
	return response.ReviewRoom, err
}

func (c *daemonClient) WriteLease() runtime.WriteLease {
	var response struct {
		WriteLease runtime.WriteLease `json:"write_lease"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/writer", nil, &response)
	return response.WriteLease
}

func (c *daemonClient) Handoff() runtime.HandoffPacket {
	var response struct {
		Handoff runtime.HandoffPacket `json:"handoff"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/handoff", nil, &response)
	return response.Handoff
}

func (c *daemonClient) SessionBootstrap(full bool) (runtime.SessionBootstrapPacket, error) {
	path := "/api/layer-os/session/bootstrap"
	if full {
		path += "?full=1"
	}
	var response struct {
		Session runtime.SessionBootstrapPacket `json:"session"`
	}
	err := c.request(http.MethodGet, path, nil, &response)
	return response.Session, err
}

func (c *daemonClient) CheckpointSession(input runtime.SessionCheckpointInput) (runtime.SessionCheckpoint, error) {
	var result runtime.SessionCheckpoint
	err := c.request(http.MethodPost, "/api/layer-os/session/checkpoint", input, &result)
	return result, err
}

func (c *daemonClient) SessionNote(input runtime.SessionNoteInput) (runtime.SessionNoteResult, error) {
	var result runtime.SessionNoteResult
	err := c.request(http.MethodPost, "/api/layer-os/session/note", input, &result)
	return result, err
}

func (c *daemonClient) SessionFinish(input runtime.SessionFinishInput) (runtime.SessionFinishResult, error) {
	var result runtime.SessionFinishResult
	err := c.request(http.MethodPost, "/api/layer-os/session/finish", input, &result)
	return result, err
}

func (c *daemonClient) AuthStatus() runtime.AuthStatus {
	var response struct {
		Auth runtime.AuthStatus `json:"auth"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/auth", nil, &response)
	return response.Auth
}

func (c *daemonClient) SetWriteToken(token string) (runtime.AuthStatus, error) {
	var response struct {
		Auth runtime.AuthStatus `json:"auth"`
	}
	err := c.request(http.MethodPost, "/api/layer-os/auth", map[string]string{"token": token}, &response)
	return response.Auth, err
}

func (c *daemonClient) ClearWriteToken() (runtime.AuthStatus, error) {
	var response struct {
		Auth runtime.AuthStatus `json:"auth"`
	}
	err := c.request(http.MethodDelete, "/api/layer-os/auth", nil, &response)
	return response.Auth, err
}

func (c *daemonClient) ReplaceMemory(memory runtime.SystemMemory) error {
	return c.request(http.MethodPost, "/api/layer-os/memory", memory, nil)
}

func (c *daemonClient) CreateApproval(item runtime.ApprovalItem) error {
	return c.request(http.MethodPost, "/api/layer-os/approval-inbox", item, nil)
}

func (c *daemonClient) CreatePreflight(item runtime.PreflightRecord) error {
	return c.request(http.MethodPost, "/api/layer-os/preflights", item, nil)
}

func (c *daemonClient) EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (runtime.PolicyDecision, error) {
	var item runtime.PolicyDecision
	err := c.request(http.MethodPost, "/api/layer-os/policies/evaluate", map[string]any{
		"decision_id":       decisionID,
		"intent":            intent,
		"scope":             scope,
		"risk":              risk,
		"novelty":           novelty,
		"token_class":       tokenClass,
		"requires_approval": requiresApproval,
	}, &item)
	return item, err
}

func (c *daemonClient) CreateGatewayCall(item runtime.GatewayCall) error {
	return c.request(http.MethodPost, "/api/layer-os/gateway-calls", item, nil)
}

func (c *daemonClient) RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (runtime.ExecuteRun, error) {
	var item runtime.ExecuteRun
	err := c.requestWithPartial(http.MethodPost, "/api/layer-os/execute-runs/run", map[string]any{
		"execute_id":         executeID,
		"work_item_id":       workItemID,
		"policy_decision_id": policyDecisionID,
		"notes":              baseNotes,
	}, "execute", &item)
	return item, err
}

func (c *daemonClient) RunVerification(recordID string, scope string, _ string, _ []string, baseNotes []string) (runtime.VerificationRecord, error) {
	var item runtime.VerificationRecord
	err := c.requestWithPartial(http.MethodPost, "/api/layer-os/verifications/run", map[string]any{
		"record_id": recordID,
		"scope":     scope,
		"notes":     baseNotes,
	}, "verification", &item)
	return item, err
}

func (c *daemonClient) ResolveApproval(approvalID string, status string) (runtime.ApprovalItem, error) {
	var item runtime.ApprovalItem
	err := c.request(http.MethodPost, "/api/layer-os/approval-inbox/resolve", map[string]any{
		"approval_id": approvalID,
		"status":      status,
	}, &item)
	return item, err
}

func (c *daemonClient) CreateRelease(item runtime.ReleasePacket) error {
	return c.request(http.MethodPost, "/api/layer-os/releases", item, nil)
}

func (c *daemonClient) CreateDeploy(item runtime.DeployRun) error {
	return c.request(http.MethodPost, "/api/layer-os/deploys", item, nil)
}

func (c *daemonClient) CreateRollback(item runtime.RollbackRun) error {
	return c.request(http.MethodPost, "/api/layer-os/rollbacks", item, nil)
}

func (c *daemonClient) PutTarget(item runtime.DeployTarget) error {
	return c.request(http.MethodPost, "/api/layer-os/deploy-targets", item, nil)
}

func (c *daemonClient) ExecuteDeploy(deployID string, releaseID string, baseNotes []string) (runtime.DeployRun, error) {
	var item runtime.DeployRun
	err := c.requestWithPartial(http.MethodPost, "/api/layer-os/deploys/execute", map[string]any{
		"deploy_id":  deployID,
		"release_id": releaseID,
		"notes":      baseNotes,
	}, "deploy", &item)
	return item, err
}

func (c *daemonClient) ExecuteRollback(rollbackID string, releaseID string, deployID string, baseNotes []string) (runtime.RollbackRun, error) {
	var item runtime.RollbackRun
	err := c.requestWithPartial(http.MethodPost, "/api/layer-os/rollbacks/execute", map[string]any{
		"rollback_id": rollbackID,
		"release_id":  releaseID,
		"deploy_id":   deployID,
		"notes":       baseNotes,
	}, "rollback", &item)
	return item, err
}

func (c *daemonClient) FounderSummary() runtime.FounderSummary {
	var response struct {
		FounderSummary runtime.FounderSummary `json:"founder_summary"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/founder-summary", nil, &response)
	return response.FounderSummary
}

func (c *daemonClient) Snapshot() runtime.SnapshotPacket {
	var response struct {
		Snapshot runtime.SnapshotPacket `json:"snapshot"`
	}
	c.mustRequest(http.MethodGet, "/api/layer-os/snapshot", nil, &response)
	return response.Snapshot
}

func (c *daemonClient) ImportSnapshot(packet runtime.SnapshotPacket) error {
	return c.request(http.MethodPost, "/api/layer-os/snapshot", packet, nil)
}

func (c *daemonClient) StartFounderFlow(flowID string, workID string, approvalID string, title string, intent string, notes []string) (runtime.FlowRun, error) {
	var item runtime.FlowRun
	err := c.request(http.MethodPost, "/api/layer-os/founder-actions/start", map[string]any{
		"flow_id":      flowID,
		"work_item_id": workID,
		"approval_id":  approvalID,
		"title":        title,
		"intent":       intent,
		"notes":        notes,
	}, &item)
	return item, err
}

func (c *daemonClient) ApproveFounderFlow(flowID string, notes []string) (runtime.FlowRun, error) {
	var item runtime.FlowRun
	err := c.request(http.MethodPost, "/api/layer-os/founder-actions/approve", map[string]any{
		"flow_id": flowID,
		"notes":   notes,
	}, &item)
	return item, err
}

func (c *daemonClient) ReleaseFounderFlow(flowID string, releaseID string, deployID string, target string, channel string, notes []string) (runtime.FlowRun, error) {
	var item runtime.FlowRun
	err := c.request(http.MethodPost, "/api/layer-os/founder-actions/release", map[string]any{
		"flow_id":    flowID,
		"release_id": releaseID,
		"deploy_id":  deployID,
		"target":     target,
		"channel":    channel,
		"notes":      notes,
	}, &item)
	return item, err
}

func (c *daemonClient) RollbackFounderFlow(flowID string, rollbackID string, notes []string) (runtime.FlowRun, error) {
	var item runtime.FlowRun
	err := c.request(http.MethodPost, "/api/layer-os/founder-actions/rollback", map[string]any{
		"flow_id":     flowID,
		"rollback_id": rollbackID,
		"notes":       notes,
	}, &item)
	return item, err
}

func (c *daemonClient) mustRequest(method string, path string, payload any, out any) {
	if err := c.request(method, path, payload, out); err != nil {
		log.Fatal(err)
	}
}

func (c *daemonClient) request(method string, path string, payload any, out any) error {
	response, err := c.doRequest(method, path, payload)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return decodeAPIError(response)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, response.Body)
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("decode daemon response: %w", err)
	}
	return nil
}

func (c *daemonClient) requestWithPartial(method string, path string, payload any, field string, out any) error {
	response, err := c.doRequest(method, path, payload)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusBadRequest {
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			return fmt.Errorf("decode daemon response: %w", err)
		}
		return nil
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read daemon error response: %w", err)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err == nil {
		if encoded, ok := body[field]; ok {
			_ = json.Unmarshal(encoded, out)
		}
		if encoded, ok := body["error"]; ok {
			var message string
			if err := json.Unmarshal(encoded, &message); err == nil && strings.TrimSpace(message) != "" {
				return errors.New(message)
			}
		}
	}
	return decodeRawAPIError(response.Status, raw)
}

func (c *daemonClient) newRequest(method string, path string, payload any) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode daemon payload: %w", err)
		}
		body = bytes.NewReader(raw)
	}
	request, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build daemon request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token := strings.TrimSpace(os.Getenv("LAYER_OS_WRITE_TOKEN")); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if actor := requestActor(); actor != "" {
		request.Header.Set("X-Layer-Actor", actor)
	}
	if models := requestModels(); len(models) > 0 {
		request.Header.Set("X-Layer-Models", strings.Join(models, ","))
	}
	return request, nil
}

var errDaemonUnavailable = errors.New("layer-osd unavailable")

type daemonUnavailableError struct {
	baseURL string
	cause   error
}

func (e *daemonUnavailableError) Error() string {
	return fmt.Sprintf("cannot reach layer-osd at %s: %v", e.baseURL, e.cause)
}

func (e *daemonUnavailableError) Unwrap() error {
	return e.cause
}

func (e *daemonUnavailableError) Is(target error) bool {
	return target == errDaemonUnavailable
}

func wrapDaemonError(err error) error {
	baseURL := daemonBaseURL()
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		return &daemonUnavailableError{baseURL: baseURL, cause: urlErr.Err}
	}
	return &daemonUnavailableError{baseURL: baseURL, cause: err}
}

func decodeAPIError(response *http.Response) error {
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read daemon error response: %w", err)
	}
	return decodeRawAPIError(response.Status, raw)
}

func decodeRawAPIError(status string, raw []byte) error {
	var apiErr apiError
	if err := json.Unmarshal(raw, &apiErr); err == nil && strings.TrimSpace(apiErr.Error) != "" {
		return errors.New(apiErr.Error)
	}
	message := strings.TrimSpace(string(raw))
	if message == "" {
		return fmt.Errorf("daemon request failed: %s", status)
	}
	return fmt.Errorf("daemon request failed: %s: %s", status, message)
}

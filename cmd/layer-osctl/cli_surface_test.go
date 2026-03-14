package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"layer-os/internal/runtime"
)

type providerServiceStub struct{ providers []runtime.ProviderSummary }

type telegramServiceStub struct{ packet runtime.TelegramPacket }
type conversationServiceStub struct {
	auth      runtime.AuthStatus
	status    runtime.ConversationAutomationStatus
	noteInput runtime.ConversationNoteInput
	result    runtime.ConversationNoteResult
}
type threadsServiceStub struct {
	auth       runtime.AuthStatus
	status     runtime.ThreadsStatus
	receipt    runtime.ThreadsPublishReceipt
	approvalID string
}
type corpusServiceStub struct {
	entries     []runtime.CapitalizationEntry
	search      runtime.KnowledgeSearchResponse
	searchQuery string
}
type gatewayServiceStub struct {
	auth      runtime.AuthStatus
	calls     []runtime.GatewayCall
	created   runtime.GatewayCall
	preflight runtime.PreflightRecord
	policy    runtime.PolicyDecision
	event     runtime.EventEnvelope
	execute   runtime.ExecuteRun
}

func (s *providerServiceStub) Providers() []runtime.ProviderSummary { return s.providers }
func (s *telegramServiceStub) Telegram() runtime.TelegramPacket     { return s.packet }
func (s *telegramServiceStub) SendTelegram() error                  { return nil }
func (s *telegramServiceStub) TelegramEnabled() bool                { return false }
func (s *conversationServiceStub) AuthStatus() runtime.AuthStatus   { return s.auth }
func (s *conversationServiceStub) ConversationStatus() runtime.ConversationAutomationStatus {
	return s.status
}
func (s *conversationServiceStub) CreateConversationNote(input runtime.ConversationNoteInput) (runtime.ConversationNoteResult, error) {
	s.noteInput = input
	if s.result.ConversationID == "" {
		s.result = runtime.ConversationNoteResult{
			ConversationID: "conversation_001",
			Observation: runtime.ObservationRecord{
				ObservationID:     "observation_001",
				SourceChannel:     "conversation:" + strings.TrimSpace(input.SourceChannel),
				Confidence:        strings.TrimSpace(input.Confidence),
				Refs:              append([]string{}, input.Refs...),
				NormalizedSummary: strings.TrimSpace(input.Summary),
			},
		}
	}
	return s.result, nil
}
func (s *threadsServiceStub) AuthStatus() runtime.AuthStatus       { return s.auth }
func (s *threadsServiceStub) ThreadsStatus() runtime.ThreadsStatus { return s.status }
func (s *threadsServiceStub) PublishThreads(approvalID string) (runtime.ThreadsPublishReceipt, error) {
	s.approvalID = approvalID
	return s.receipt, nil
}
func (s *corpusServiceStub) ListCapitalizationEntries() []runtime.CapitalizationEntry {
	return s.entries
}
func (s *corpusServiceStub) SearchKnowledge(query string) runtime.KnowledgeSearchResponse {
	s.searchQuery = query
	return s.search
}
func (s *gatewayServiceStub) AuthStatus() runtime.AuthStatus          { return s.auth }
func (s *gatewayServiceStub) ListGatewayCalls() []runtime.GatewayCall { return s.calls }
func (s *gatewayServiceStub) CreateGatewayCall(item runtime.GatewayCall) error {
	s.created = item
	return nil
}
func (s *gatewayServiceStub) ListPreflights() []runtime.PreflightRecord { return nil }
func (s *gatewayServiceStub) CreatePreflight(item runtime.PreflightRecord) error {
	s.preflight = item
	return nil
}
func (s *gatewayServiceStub) ListVerifications() []runtime.VerificationRecord { return nil }
func (s *gatewayServiceStub) RunVerification(recordID string, scope string, workdir string, command []string, baseNotes []string) (runtime.VerificationRecord, error) {
	return runtime.VerificationRecord{}, nil
}
func (s *gatewayServiceStub) ListPolicies() []runtime.PolicyDecision { return nil }
func (s *gatewayServiceStub) EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (runtime.PolicyDecision, error) {
	s.policy = runtime.PolicyDecision{DecisionID: decisionID, Intent: intent, Scope: scope, Risk: risk, Novelty: novelty, TokenClass: tokenClass, RequiresApproval: requiresApproval}
	return s.policy, nil
}
func (s *gatewayServiceStub) ListEvents() []runtime.EventEnvelope { return nil }
func (s *gatewayServiceStub) CreateEvent(input runtime.EventCreateInput) (runtime.EventEnvelope, error) {
	s.event = runtime.EventEnvelope{Kind: input.Kind, Surface: input.Surface, WorkItemID: input.WorkItemID, Stage: input.Stage, Data: input.Data}
	return s.event, nil
}
func (s *gatewayServiceStub) StreamEvents(out io.Writer) error { return nil }
func (s *gatewayServiceStub) ListExecutes() []runtime.ExecuteRun {
	return nil
}
func (s *gatewayServiceStub) RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (runtime.ExecuteRun, error) {
	s.execute = runtime.ExecuteRun{ExecuteID: executeID, WorkItemID: workItemID, PolicyDecisionID: policyDecisionID, Notes: append([]string{}, baseNotes...)}
	return s.execute, nil
}

type reviewRoomServiceStub struct {
	auth                 runtime.AuthStatus
	room                 runtime.ReviewRoom
	summary              runtime.ReviewRoomSummary
	addSection           string
	addItem              runtime.ReviewRoomItem
	transitionAction     string
	transitionItem       string
	transitionResolution *runtime.ReviewRoomResolution
}

func (s *reviewRoomServiceStub) AuthStatus() runtime.AuthStatus               { return s.auth }
func (s *reviewRoomServiceStub) ReviewRoom() runtime.ReviewRoom               { return s.room }
func (s *reviewRoomServiceStub) ReviewRoomSummary() runtime.ReviewRoomSummary { return s.summary }
func (s *reviewRoomServiceStub) AddReviewRoomItem(section string, item runtime.ReviewRoomItem) (runtime.ReviewRoom, error) {
	s.addSection = section
	s.addItem = item
	return s.room, nil
}
func (s *reviewRoomServiceStub) TransitionStructuredReviewRoomItem(action string, item string, resolution *runtime.ReviewRoomResolution) (runtime.ReviewRoom, error) {
	s.transitionAction = action
	s.transitionItem = item
	s.transitionResolution = resolution
	return s.room, nil
}

type releaseStageServiceStub struct {
	auth     runtime.AuthStatus
	release  runtime.ReleasePacket
	deploy   runtime.DeployRun
	rollback runtime.RollbackRun
	target   runtime.DeployTarget
}

func (s *releaseStageServiceStub) AuthStatus() runtime.AuthStatus        { return s.auth }
func (s *releaseStageServiceStub) ListReleases() []runtime.ReleasePacket { return nil }
func (s *releaseStageServiceStub) ListDeploys() []runtime.DeployRun      { return nil }
func (s *releaseStageServiceStub) ListRollbacks() []runtime.RollbackRun  { return nil }
func (s *releaseStageServiceStub) ListTargets() []runtime.DeployTarget {
	if s.target.TargetID == "" {
		return nil
	}
	return []runtime.DeployTarget{s.target}
}
func (s *releaseStageServiceStub) CreateRelease(item runtime.ReleasePacket) error {
	s.release = item
	return nil
}
func (s *releaseStageServiceStub) CreateDeploy(item runtime.DeployRun) error {
	s.deploy = item
	return nil
}
func (s *releaseStageServiceStub) CreateRollback(item runtime.RollbackRun) error {
	s.rollback = item
	return nil
}
func (s *releaseStageServiceStub) PutTarget(item runtime.DeployTarget) error {
	s.target = item
	return nil
}
func (s *releaseStageServiceStub) ExecuteDeploy(deployID string, releaseID string, baseNotes []string) (runtime.DeployRun, error) {
	s.deploy = runtime.DeployRun{DeployID: deployID, ReleaseID: releaseID, Notes: append([]string{}, baseNotes...)}
	return s.deploy, nil
}
func (s *releaseStageServiceStub) ExecuteRollback(rollbackID string, releaseID string, deployID string, baseNotes []string) (runtime.RollbackRun, error) {
	s.rollback = runtime.RollbackRun{RollbackID: rollbackID, ReleaseID: releaseID, Notes: append([]string{}, baseNotes...)}
	return s.rollback, nil
}

type workFlowApprovalServiceStub struct {
	auth     runtime.AuthStatus
	work     runtime.WorkItem
	flow     runtime.FlowRun
	syncArgs struct {
		flowID     string
		workID     string
		approvalID *string
		policyID   *string
		executeID  *string
		verifyID   *string
		releaseID  *string
		deployID   *string
		rollbackID *string
		notes      []string
	}
	approval           runtime.ApprovalItem
	resolvedApprovalID string
	resolvedStatus     string
}

func (s *workFlowApprovalServiceStub) AuthStatus() runtime.AuthStatus    { return s.auth }
func (s *workFlowApprovalServiceStub) ListWorkItems() []runtime.WorkItem { return nil }
func (s *workFlowApprovalServiceStub) CreateWorkItem(item runtime.WorkItem) error {
	s.work = item
	return nil
}
func (s *workFlowApprovalServiceStub) ListFlows() []runtime.FlowRun { return nil }
func (s *workFlowApprovalServiceStub) CreateFlow(item runtime.FlowRun) error {
	s.flow = item
	return nil
}
func (s *workFlowApprovalServiceStub) SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (runtime.FlowRun, error) {
	s.syncArgs.flowID = flowID
	s.syncArgs.workID = workItemID
	s.syncArgs.approvalID = approvalID
	s.syncArgs.policyID = policyDecisionID
	s.syncArgs.executeID = executeID
	s.syncArgs.verifyID = verificationID
	s.syncArgs.releaseID = releaseID
	s.syncArgs.deployID = deployID
	s.syncArgs.rollbackID = rollbackID
	s.syncArgs.notes = append([]string{}, notes...)
	return runtime.FlowRun{FlowID: flowID, WorkItemID: workItemID}, nil
}
func (s *workFlowApprovalServiceStub) ListApprovals() []runtime.ApprovalItem { return nil }
func (s *workFlowApprovalServiceStub) CreateApproval(item runtime.ApprovalItem) error {
	s.approval = item
	return nil
}
func (s *workFlowApprovalServiceStub) ResolveApproval(approvalID string, status string) (runtime.ApprovalItem, error) {
	s.resolvedApprovalID = approvalID
	s.resolvedStatus = status
	return runtime.ApprovalItem{ApprovalID: approvalID, Status: status}, nil
}

type systemStateServiceStub struct {
	auth     runtime.AuthStatus
	memory   runtime.SystemMemory
	setToken string
	cleared  bool
	replaced runtime.SystemMemory
	snapshot runtime.SnapshotPacket
	imported runtime.SnapshotPacket
}

func (s *systemStateServiceStub) AuthStatus() runtime.AuthStatus { return s.auth }
func (s *systemStateServiceStub) Memory() runtime.SystemMemory   { return s.memory }
func (s *systemStateServiceStub) SetWriteToken(token string) (runtime.AuthStatus, error) {
	s.setToken = token
	return runtime.AuthStatus{WriteAuthEnabled: true}, nil
}
func (s *systemStateServiceStub) ClearWriteToken() (runtime.AuthStatus, error) {
	s.cleared = true
	return runtime.AuthStatus{}, nil
}
func (s *systemStateServiceStub) ReplaceMemory(memory runtime.SystemMemory) error {
	s.replaced = memory
	s.memory = memory
	return nil
}
func (s *systemStateServiceStub) Snapshot() runtime.SnapshotPacket { return s.snapshot }
func (s *systemStateServiceStub) ImportSnapshot(packet runtime.SnapshotPacket) error {
	s.imported = packet
	return nil
}

type controlPlaneServiceStub struct {
	auth         runtime.AuthStatus
	preflight    runtime.PreflightRecord
	policy       runtime.PolicyDecision
	gateway      runtime.GatewayCall
	event        runtime.EventEnvelope
	stream       string
	execute      runtime.ExecuteRun
	verification runtime.VerificationRecord
}

func (s *controlPlaneServiceStub) AuthStatus() runtime.AuthStatus            { return s.auth }
func (s *controlPlaneServiceStub) ListPreflights() []runtime.PreflightRecord { return nil }
func (s *controlPlaneServiceStub) CreatePreflight(item runtime.PreflightRecord) error {
	s.preflight = item
	return nil
}
func (s *controlPlaneServiceStub) ListVerifications() []runtime.VerificationRecord { return nil }
func (s *controlPlaneServiceStub) RunVerification(recordID string, scope string, workdir string, command []string, baseNotes []string) (runtime.VerificationRecord, error) {
	s.verification = runtime.VerificationRecord{RecordID: recordID, Scope: scope, Command: append([]string{}, command...), Notes: append([]string{}, baseNotes...)}
	return s.verification, nil
}
func (s *controlPlaneServiceStub) ListPolicies() []runtime.PolicyDecision { return nil }
func (s *controlPlaneServiceStub) EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (runtime.PolicyDecision, error) {
	s.policy = runtime.PolicyDecision{DecisionID: decisionID, Intent: intent, Scope: scope, Risk: risk, Novelty: novelty, TokenClass: tokenClass, RequiresApproval: requiresApproval}
	return s.policy, nil
}
func (s *controlPlaneServiceStub) ListGatewayCalls() []runtime.GatewayCall { return nil }
func (s *controlPlaneServiceStub) CreateGatewayCall(item runtime.GatewayCall) error {
	s.gateway = item
	return nil
}
func (s *controlPlaneServiceStub) ListEvents() []runtime.EventEnvelope { return nil }
func (s *controlPlaneServiceStub) CreateEvent(input runtime.EventCreateInput) (runtime.EventEnvelope, error) {
	s.event = runtime.EventEnvelope{Kind: input.Kind, Surface: input.Surface, WorkItemID: input.WorkItemID, Stage: input.Stage, Data: input.Data}
	return s.event, nil
}
func (s *controlPlaneServiceStub) StreamEvents(out io.Writer) error {
	_, err := io.WriteString(out, s.stream)
	return err
}
func (s *controlPlaneServiceStub) ListExecutes() []runtime.ExecuteRun { return nil }
func (s *controlPlaneServiceStub) RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (runtime.ExecuteRun, error) {
	s.execute = runtime.ExecuteRun{ExecuteID: executeID, WorkItemID: workItemID, PolicyDecisionID: policyDecisionID, Notes: append([]string{}, baseNotes...)}
	return s.execute, nil
}

type proposalServiceStub struct {
	auth     runtime.AuthStatus
	proposal runtime.ProposalItem
	work     runtime.WorkItem
}

func (s *proposalServiceStub) AuthStatus() runtime.AuthStatus        { return s.auth }
func (s *proposalServiceStub) ListProposals() []runtime.ProposalItem { return nil }
func (s *proposalServiceStub) CreateProposal(item runtime.ProposalItem) error {
	s.proposal = item
	return nil
}
func (s *proposalServiceStub) PromoteProposal(proposalID string, workID string) (runtime.ProposalItem, runtime.WorkItem, error) {
	s.proposal = runtime.ProposalItem{ProposalID: proposalID, Title: "Plan queue", Intent: "close planning gap", Summary: "Plan queue", Surface: runtime.SurfaceAPI, Priority: "high", Risk: "medium", Status: "promoted", Notes: []string{}, PromotedWorkItemID: &workID, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	s.work = runtime.WorkItem{ID: workID, Title: "Plan queue", Intent: "close planning gap", Stage: runtime.StageDiscover, Surface: runtime.SurfaceAPI, Pack: "proposal", Priority: "high", Risk: "medium", RequiresApproval: false, Payload: map[string]any{"proposal_id": proposalID}, CorrelationID: proposalID, CreatedAt: time.Now().UTC()}
	return s.proposal, s.work, nil
}

type founderServiceStub struct {
	auth    runtime.AuthStatus
	summary runtime.FounderSummary
}

func (s *founderServiceStub) AuthStatus() runtime.AuthStatus         { return s.auth }
func (s *founderServiceStub) FounderSummary() runtime.FounderSummary { return s.summary }
func (s *founderServiceStub) StartFounderFlow(flowID string, workID string, approvalID string, title string, intent string, notes []string) (runtime.FlowRun, error) {
	return runtime.FlowRun{}, nil
}
func (s *founderServiceStub) ApproveFounderFlow(flowID string, notes []string) (runtime.FlowRun, error) {
	return runtime.FlowRun{}, nil
}
func (s *founderServiceStub) ReleaseFounderFlow(flowID string, releaseID string, deployID string, target string, channel string, notes []string) (runtime.FlowRun, error) {
	return runtime.FlowRun{}, nil
}
func (s *founderServiceStub) RollbackFounderFlow(flowID string, rollbackID string, notes []string) (runtime.FlowRun, error) {
	return runtime.FlowRun{}, nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(raw)
}

func assertContains(t *testing.T, raw string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(raw, want) {
			t.Fatalf("missing %q in output: %s", want, raw)
		}
	}
}

func assertNotContains(t *testing.T, raw string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if strings.Contains(raw, want) {
			t.Fatalf("unexpected %q in output: %s", want, raw)
		}
	}
}

func TestRunCorpusListWritesCapitalizationEntries(t *testing.T) {
	service := &corpusServiceStub{entries: []runtime.CapitalizationEntry{{EntryID: "cap_event_001", SourceKind: "session.finished", Situation: runtime.CapitalizationFacet{Summary: "Close drift"}}}}
	raw := captureStdout(t, func() {
		runCorpus(service, []string{"list"})
	})
	assertContains(t, raw, "cap_event_001", "session.finished")
}

func TestRunCorpusSearchWritesKnowledgeSearchResults(t *testing.T) {
	service := &corpusServiceStub{search: runtime.KnowledgeSearchResponse{
		Query: "queue drift",
		Results: []runtime.KnowledgeSearchResult{{
			Entry:       runtime.CapitalizationEntry{EntryID: "cap_event_002", SourceKind: "job.report"},
			MatchFields: []string{"situation", "outcome"},
			MatchCount:  3,
		}},
	}}
	raw := captureStdout(t, func() {
		runCorpus(service, []string{"search", "--query", "queue drift"})
	})
	if service.searchQuery != "queue drift" {
		t.Fatalf("expected search query, got %q", service.searchQuery)
	}
	assertContains(t, raw, "cap_event_002", "queue drift")
}

func TestRunTelegramPreviewWritesTelegramPacket(t *testing.T) {
	service := &telegramServiceStub{packet: runtime.TelegramPacket{Headline: "Review-room agenda needs founder attention", ReviewOpenCount: 2, RecommendedMode: "conserve"}}
	raw := captureStdout(t, func() {
		runTelegram(service, []string{"preview"})
	})
	assertContains(t, raw, "Review-room agenda needs founder attention", "conserve")
}

func TestRunConversationStatusWritesAutomationStatus(t *testing.T) {
	service := &conversationServiceStub{status: runtime.ConversationAutomationStatus{
		Toggles: runtime.ConversationRuntimeToggles{
			AutoPlanEnabled:     true,
			AutoRiskEnabled:     true,
			AutoDispatchEnabled: false,
			LLMTagEnabled:       false,
		},
	}}
	raw := captureStdout(t, func() {
		runConversation(service, []string{"status"})
	})
	assertContains(t, raw, "\"auto_plan_enabled\": true", "\"auto_dispatch_enabled\": false")
}

func TestRunConversationNoteUsesSingleThreadInputs(t *testing.T) {
	service := &conversationServiceStub{}
	raw := captureStdout(t, func() {
		runConversation(service, []string{
			"note",
			"--text", "Need founder review on this thread",
			"--source", "telegram",
			"--summary", "Founder review needed",
			"--tags", "todo,link",
			"--refs", "thread:001,url:https://example.com/story",
			"--confidence", "high",
			"--auto-plan", "true",
			"--auto-risk", "false",
			"--auto-dispatch", "false",
		})
	})
	if service.noteInput.SourceChannel != "telegram" {
		t.Fatalf("expected telegram source, got %+v", service.noteInput)
	}
	if len(service.noteInput.Tags) != 2 || service.noteInput.Tags[1] != "link" {
		t.Fatalf("unexpected tags: %+v", service.noteInput.Tags)
	}
	if service.noteInput.AutoPlan == nil || !*service.noteInput.AutoPlan {
		t.Fatalf("expected auto-plan override, got %+v", service.noteInput.AutoPlan)
	}
	if service.noteInput.AutoRisk == nil || *service.noteInput.AutoRisk {
		t.Fatalf("expected auto-risk false override, got %+v", service.noteInput.AutoRisk)
	}
	if service.noteInput.AutoDispatch == nil || *service.noteInput.AutoDispatch {
		t.Fatalf("expected auto-dispatch false override, got %+v", service.noteInput.AutoDispatch)
	}
	assertContains(t, raw, "conversation_001", "conversation:telegram")
}

func TestRunThreadsStatusWritesThreadsStatus(t *testing.T) {
	service := &threadsServiceStub{status: runtime.ThreadsStatus{
		Adapter:           "threads_api",
		PublishConfigured: true,
		Notes:             []string{"Approved Threads drafts can publish from the canonical daemon route."},
	}}
	raw := captureStdout(t, func() {
		runThreads(service, []string{"status"})
	})
	assertContains(t, raw, "threads_api", "publish_configured")
}

func TestRunThreadsPublishUsesApprovalID(t *testing.T) {
	service := &threadsServiceStub{
		receipt: runtime.ThreadsPublishReceipt{
			ApprovalID:    "approval_123",
			ObservationID: "obs_123",
			CreationID:    "creation_123",
			ThreadID:      "thread_123",
			Title:         "Launch note",
		},
	}
	raw := captureStdout(t, func() {
		runThreads(service, []string{"publish", "--approval", "approval_123"})
	})
	if service.approvalID != "approval_123" {
		t.Fatalf("expected publish approval id, got %q", service.approvalID)
	}
	assertContains(t, raw, "approval_123", "thread_123")
}

func TestRunProviderListWritesProviders(t *testing.T) {
	service := &providerServiceStub{providers: []runtime.ProviderSummary{{Provider: "openai", Declared: true, DispatchEnabled: true, GatewayAdapter: "api", Semantics: "dispatch", Notes: []string{"api endpoint configured"}}}}
	raw := captureStdout(t, func() {
		runProvider(service, []string{"list"})
	})
	assertContains(t, raw, "openai", "dispatch")
}

func TestRunReviewRoomSupportsAddWithStructuredFields(t *testing.T) {
	service := &reviewRoomServiceStub{}
	captureStdout(t, func() {
		runReviewRoom(service, []string{"add", "--section", "open", "--item", "Fix data race.", "--kind", "bug", "--severity", "high", "--source", "agent.codex", "--ref", "issue_001", "--why", "reduce deploy-risk before retry", "--why-unresolved", "race root cause still unknown", "--contradictions", "accepted mitigation exists but failure persists,keep root budget tight", "--pattern-refs", "issue_001,session_003"})
	})
	if service.addSection != "open" {
		t.Fatalf("expected open section, got %q", service.addSection)
	}
	if service.addItem.Kind != "bug" || service.addItem.Severity != "high" || service.addItem.Source != "agent.codex" {
		t.Fatalf("unexpected structured add item: %+v", service.addItem)
	}
	if service.addItem.Ref == nil || *service.addItem.Ref != "issue_001" {
		t.Fatalf("unexpected ref: %+v", service.addItem.Ref)
	}
	if service.addItem.Why != "reduce deploy-risk before retry" {
		t.Fatalf("unexpected why: %+v", service.addItem.Why)
	}
	if service.addItem.WhyUnresolved == nil || *service.addItem.WhyUnresolved != "race root cause still unknown" {
		t.Fatalf("unexpected why_unresolved: %+v", service.addItem.WhyUnresolved)
	}
	if len(service.addItem.Contradictions) != 2 || service.addItem.Contradictions[1] != "keep root budget tight" {
		t.Fatalf("unexpected contradictions: %+v", service.addItem.Contradictions)
	}
	if len(service.addItem.PatternRefs) != 2 || service.addItem.PatternRefs[1] != "session_003" {
		t.Fatalf("unexpected pattern refs: %+v", service.addItem.PatternRefs)
	}
}

func TestRunReviewRoomSupportsTransitions(t *testing.T) {
	ref := "issue_001"
	service := &reviewRoomServiceStub{room: runtime.ReviewRoom{Open: []runtime.ReviewRoomItem{{Text: "Fix data race.", Ref: &ref}}}}
	captureStdout(t, func() {
		runReviewRoom(service, []string{"accept", "--id", "issue_001"})
	})
	if service.transitionAction != "accept" || service.transitionItem != "Fix data race." {
		t.Fatalf("unexpected transition call: action=%q item=%q", service.transitionAction, service.transitionItem)
	}
}

func TestRunReviewRoomSupportsTransitionsByItem(t *testing.T) {
	service := &reviewRoomServiceStub{}
	captureStdout(t, func() {
		runReviewRoom(service, []string{"resolve", "--item", "Close corpus backlog.", "--reason", "implemented", "--rule", "review_room.transition.resolve.implemented", "--evidence", "file:cmd/layer-osctl/corpus.go"})
	})
	if service.transitionAction != "resolve" || service.transitionItem != "Close corpus backlog." {
		t.Fatalf("unexpected transition-by-item call: action=%q item=%q", service.transitionAction, service.transitionItem)
	}
	if service.transitionResolution == nil || service.transitionResolution.Rule != "review_room.transition.resolve.implemented" {
		t.Fatalf("unexpected transition resolution: %+v", service.transitionResolution)
	}
}

func TestRunFounderSummaryWritesFounderSummary(t *testing.T) {
	service := &founderServiceStub{summary: runtime.FounderSummary{PrimaryAction: "review_room", ReviewOpenCount: 1}}
	raw := captureStdout(t, func() {
		runFounder(service, []string{"summary"})
	})
	assertContains(t, raw, "\"primary_action\": \"review_room\"")
	var decoded runtime.FounderSummary
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("decode founder summary: %v", err)
	}
	if decoded.ReviewOpenCount != 1 {
		t.Fatalf("expected review_open_count 1, got %d", decoded.ReviewOpenCount)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return string(raw)
}

func TestRunReleaseParsesStructuredFlags(t *testing.T) {
	service := &releaseStageServiceStub{}
	captureStdout(t, func() {
		runRelease(service, []string{"create", "--id", "release_001", "--work", "work_001", "--target", "vm", "--channel", "cockpit", "--artifacts", "a,b", "--approvals", "approval_001", "--metrics", "latency_ms=120,target=vm"})
	})
	if service.release.ReleaseID != "release_001" || service.release.WorkItemID != "work_001" {
		t.Fatalf("unexpected release item: %+v", service.release)
	}
	if service.release.Metrics["latency_ms"] != "120" || service.release.Metrics["target"] != "vm" {
		t.Fatalf("unexpected release metrics: %+v", service.release.Metrics)
	}
}

func TestRunDeployExecuteRoutesExecute(t *testing.T) {
	service := &releaseStageServiceStub{}
	raw := captureStdout(t, func() {
		runDeploy(service, []string{"execute", "--id", "deploy_001", "--release", "release_001", "--notes", "smoke,prod"})
	})
	if service.deploy.DeployID != "deploy_001" || service.deploy.ReleaseID != "release_001" {
		t.Fatalf("unexpected deploy execute routing: %+v", service.deploy)
	}
	assertContains(t, raw, "deploy_001")
}

func TestRunTargetPutParsesFields(t *testing.T) {
	service := &releaseStageServiceStub{}
	raw := captureStdout(t, func() {
		runTarget(service, []string{"put", "--id", "target_vm", "--command", "/usr/bin/true", "--workdir", "/tmp"})
	})
	if service.target.TargetID != "target_vm" || len(service.target.Command) != 1 || service.target.Command[0] != "/usr/bin/true" {
		t.Fatalf("unexpected target: %+v", service.target)
	}
	if service.target.Workdir == nil || *service.target.Workdir != "/tmp" {
		t.Fatalf("unexpected workdir: %+v", service.target.Workdir)
	}
	assertContains(t, raw, "target_vm")
}

func TestRunTargetListWritesTargets(t *testing.T) {
	service := &releaseStageServiceStub{}
	service.target = runtime.DeployTarget{TargetID: "vm_list", Command: []string{"/usr/bin/true"}}
	raw := captureStdout(t, func() {
		runTarget(service, []string{"list"})
	})
	assertContains(t, raw, "vm_list")
}

type gatewayCreateServiceStub struct {
	controlPlaneServiceStub
	created runtime.GatewayCall
}

func (s *gatewayCreateServiceStub) CreateGatewayCall(item runtime.GatewayCall) error {
	s.created = item
	return nil
}

func TestRunGatewayCreateParsesFields(t *testing.T) {
	service := &gatewayCreateServiceStub{}
	raw := captureStdout(t, func() {
		runGateway(service, []string{"create", "--id", "gateway_1", "--decision", "policy_1", "--provider", "openai", "--model", "gpt", "--kind", "chat", "--status", "recorded", "--budget", "120", "--notes", "one,two"})
	})
	if service.created.CallID != "gateway_1" || service.created.Provider != "openai" || service.created.TokenBudget != 120 {
		t.Fatalf("unexpected gateway call: %+v", service.created)
	}
	if len(service.created.Notes) != 2 || service.created.Notes[1] != "two" {
		t.Fatalf("unexpected gateway notes: %+v", service.created.Notes)
	}
	assertContains(t, raw, "gateway_1")
}

func TestUsageMentionsCurrentOperatorSurface(t *testing.T) {
	raw := captureStderr(t, usage)
	for _, snippet := range runtime.CanonicalSurfaceSpec().CLIUsage {
		assertContains(t, raw, snippet)
	}
}

func TestRunProposalCreateParsesFields(t *testing.T) {
	service := &proposalServiceStub{}
	raw := captureStdout(t, func() {
		runProposal(service, []string{"create", "--id", "proposal_001", "--branch", "branch_001", "--title", "Plan queue", "--intent", "close planning gap", "--summary", "Queue gap", "--surface", "api", "--priority", "high", "--risk", "medium", "--notes", "one,two"})
	})
	if service.proposal.ProposalID != "proposal_001" || service.proposal.Intent != "close planning gap" {
		t.Fatalf("unexpected proposal: %+v", service.proposal)
	}
	if service.proposal.BranchID == nil || *service.proposal.BranchID != "branch_001" {
		t.Fatalf("unexpected proposal branch: %+v", service.proposal.BranchID)
	}
	if len(service.proposal.Notes) != 2 || service.proposal.Notes[0] != "one" {
		t.Fatalf("unexpected proposal notes: %+v", service.proposal.Notes)
	}
	assertContains(t, raw, "proposal_001")
}

func TestRunProposalPromoteRoutes(t *testing.T) {
	service := &proposalServiceStub{}
	raw := captureStdout(t, func() {
		runProposal(service, []string{"promote", "--id", "proposal_001", "--work", "work_001"})
	})
	if service.work.ID != "work_001" || service.proposal.PromotedWorkItemID == nil || *service.proposal.PromotedWorkItemID != "work_001" {
		t.Fatalf("unexpected proposal promote output: proposal=%+v work=%+v", service.proposal, service.work)
	}
	assertContains(t, raw, "work_001")
}

func TestRunWorkCreateParsesFields(t *testing.T) {
	service := &workFlowApprovalServiceStub{}
	captureStdout(t, func() {
		runWork(service, []string{"create", "--id", "work_001", "--branch", "branch_001", "--title", "Ship kernel", "--intent", "hardening", "--stage", "discover", "--surface", "api", "--pack", "core", "--priority", "high", "--risk", "medium", "--correlation", "corr_001", "--requires-approval"})
	})
	if service.work.ID != "work_001" || service.work.Intent != "hardening" || !service.work.RequiresApproval {
		t.Fatalf("unexpected work item: %+v", service.work)
	}
	if service.work.BranchID == nil || *service.work.BranchID != "branch_001" {
		t.Fatalf("unexpected work branch: %+v", service.work.BranchID)
	}
}

func TestRunFlowSyncParsesRefs(t *testing.T) {
	service := &workFlowApprovalServiceStub{}
	raw := captureStdout(t, func() {
		runFlow(service, []string{"sync", "--id", "flow_001", "--work", "work_001", "--approval", "approval_001", "--policy", "policy_001", "--execute", "execute_001", "--verify", "verify_001", "--release", "release_001", "--deploy", "deploy_001", "--rollback", "rollback_001", "--notes", "one,two"})
	})
	if service.syncArgs.flowID != "flow_001" || service.syncArgs.workID != "work_001" {
		t.Fatalf("unexpected sync args: %+v", service.syncArgs)
	}
	if service.syncArgs.approvalID == nil || *service.syncArgs.approvalID != "approval_001" || service.syncArgs.rollbackID == nil || *service.syncArgs.rollbackID != "rollback_001" {
		t.Fatalf("unexpected ref args: %+v", service.syncArgs)
	}
	assertContains(t, raw, "flow_001")
}

func TestRunApprovalResolveRoutes(t *testing.T) {
	service := &workFlowApprovalServiceStub{}
	raw := captureStdout(t, func() {
		runApproval(service, []string{"resolve", "--id", "approval_001", "--status", "approved"})
	})
	if service.resolvedApprovalID != "approval_001" || service.resolvedStatus != "approved" {
		t.Fatalf("unexpected approval resolve call: id=%q status=%q", service.resolvedApprovalID, service.resolvedStatus)
	}
	assertContains(t, raw, "approved")
}

func TestRunAuthSetUsesProvidedToken(t *testing.T) {
	service := &systemStateServiceStub{}
	raw := captureStdout(t, func() {
		runAuth(service, []string{"set", "--token", "secret-token"})
	})
	if service.setToken != "secret-token" {
		t.Fatalf("unexpected token set: %q", service.setToken)
	}
	assertNotContains(t, raw, "secret-token")
}

func TestRunMemorySetParsesFocus(t *testing.T) {
	service := &systemStateServiceStub{}
	captureStdout(t, func() {
		runMemory(service, []string{"set", "--focus", "Lock kernel", "--goal", "Backend integrity", "--steps", "one,two", "--risks", "drift", "--handoff", "continue", "--operator", "codex"})
	})
	if service.replaced.CurrentFocus != "Lock kernel" || service.replaced.CurrentGoal == nil || *service.replaced.CurrentGoal != "Backend integrity" {
		t.Fatalf("unexpected memory replacement: %+v", service.replaced)
	}
}

func TestRunSnapshotExportWritesSnapshot(t *testing.T) {
	service := &systemStateServiceStub{snapshot: runtime.SnapshotPacket{GeneratedAt: time.Now().UTC()}}
	raw := captureStdout(t, func() {
		runSnapshot(service, []string{"export"})
	})
	assertContains(t, raw, "generated_at")
}

func TestRunEventCreateParsesFields(t *testing.T) {
	service := &controlPlaneServiceStub{}
	raw := captureStdout(t, func() {
		runEvent(service, []string{"create", "--kind", "session.finished", "--surface", "cockpit", "--work", "work_001", "--stage", "verify", "--data", "lane=review,source=cockpit"})
	})
	if service.event.Kind != "session.finished" || service.event.WorkItemID != "work_001" {
		t.Fatalf("unexpected event create payload: %+v", service.event)
	}
	if service.event.Data["lane"] != "review" || service.event.Data["source"] != "cockpit" {
		t.Fatalf("unexpected event data: %+v", service.event.Data)
	}
	assertContains(t, raw, "session.finished")
}

func TestRunEventStreamCopiesDaemonStream(t *testing.T) {
	service := &controlPlaneServiceStub{stream: "id: e1\nevent: session.finished\ndata: {}\n\n"}
	raw := captureStdout(t, func() {
		runEvent(service, []string{"stream"})
	})
	assertContains(t, raw, "event: session.finished")
}

func TestRunPolicyEvaluateRoutes(t *testing.T) {
	service := &controlPlaneServiceStub{}
	raw := captureStdout(t, func() {
		runPolicy(service, []string{"evaluate", "--id", "policy_001", "--intent", "ship", "--scope", "kernel", "--risk", "high", "--novelty", "medium", "--token-class", "small", "--requires-approval"})
	})
	if service.policy.DecisionID != "policy_001" || !service.policy.RequiresApproval {
		t.Fatalf("unexpected policy evaluation: %+v", service.policy)
	}
	assertContains(t, raw, "policy_001")
}

func TestRunExecuteRunRoutes(t *testing.T) {
	service := &controlPlaneServiceStub{}
	raw := captureStdout(t, func() {
		runExecute(service, []string{"run", "--id", "execute_001", "--work", "work_001", "--decision", "policy_001", "--notes", "one,two"})
	})
	if service.execute.ExecuteID != "execute_001" || service.execute.PolicyDecisionID != "policy_001" {
		t.Fatalf("unexpected execute routing: %+v", service.execute)
	}
	assertContains(t, raw, "execute_001")
}

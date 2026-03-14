package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type dirtyFlags struct {
	state       bool
	branch      bool
	proposal    bool
	agentJob    bool
	observation bool
	continuity  bool
	work        bool
	flow        bool
	approval    bool
	event       bool
	release     bool
	deploy      bool
	rollback    bool
	target      bool
	preflight   bool
	policy      bool
	gateway     bool
	execute     bool
	verify      bool
	reviewRoom  bool
	memory      bool
	auth        bool
}

type persistSignatures struct {
	state       string
	branch      string
	proposal    string
	agentJob    string
	observation string
	continuity  string
	work        string
	flow        string
	approval    string
	event       string
	release     string
	deploy      string
	rollback    string
	target      string
	preflight   string
	policy      string
	gateway     string
	execute     string
	verify      string
	reviewRoom  string
	memory      string
	auth        string
}

type persistSnapshot struct {
	state               CompanyState
	branch              []Branch
	proposal            []ProposalItem
	agentJob            []AgentJob
	observation         []ObservationRecord
	continuity          []ContinuityRecord
	work                []WorkItem
	flow                []FlowRun
	approval            []ApprovalItem
	event               []EventEnvelope
	pendingEventArchive []EventEnvelope
	release             []ReleasePacket
	deploy              []DeployRun
	rollback            []RollbackRun
	target              []DeployTarget
	preflight           []PreflightRecord
	policy              []PolicyDecision
	gateway             []GatewayCall
	execute             []ExecuteRun
	verify              []VerificationRecord
	reviewRoom          ReviewRoom
	memory              SystemMemory
	auth                authConfig
	signatures          persistSignatures
}

type persistDomain struct {
	dirtyFlag          func(*dirtyFlags) *bool
	persistedSignature func(*persistSignatures) *string
	snapshotSignature  func(*persistSnapshot) *string
	snapshotValue      func(*persistSnapshot) any
	save               func(*Service, persistSnapshot) error
	isDirty            func(*persistSnapshot, *persistSignatures) bool
}

func (d dirtyFlags) any() bool {
	for _, domain := range persistDomains {
		if *domain.dirtyFlag(&d) {
			return true
		}
	}
	return false
}

func signatureFor(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func standardPersistDomain[T any](
	dirtyFlag func(*dirtyFlags) *bool,
	persistedSignature func(*persistSignatures) *string,
	snapshotSignature func(*persistSnapshot) *string,
	snapshotValue func(*persistSnapshot) T,
	save func(*diskStore, T) error,
) persistDomain {
	return persistDomain{
		dirtyFlag:          dirtyFlag,
		persistedSignature: persistedSignature,
		snapshotSignature:  snapshotSignature,
		snapshotValue: func(snapshot *persistSnapshot) any {
			return snapshotValue(snapshot)
		},
		save: func(service *Service, snapshot persistSnapshot) error {
			return save(service.disk, snapshotValue(&snapshot))
		},
	}
}

func (d persistDomain) dirty(snapshot *persistSnapshot, persisted *persistSignatures) bool {
	if d.isDirty != nil {
		return d.isDirty(snapshot, persisted)
	}
	return *d.snapshotSignature(snapshot) != *d.persistedSignature(persisted)
}

var persistDomains = []persistDomain{
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.branch },
		func(signatures *persistSignatures) *string { return &signatures.branch },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.branch },
		func(snapshot *persistSnapshot) []Branch { return snapshot.branch },
		func(disk *diskStore, value []Branch) error { return disk.saveBranches(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.proposal },
		func(signatures *persistSignatures) *string { return &signatures.proposal },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.proposal },
		func(snapshot *persistSnapshot) []ProposalItem { return snapshot.proposal },
		func(disk *diskStore, value []ProposalItem) error { return disk.saveProposals(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.agentJob },
		func(signatures *persistSignatures) *string { return &signatures.agentJob },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.agentJob },
		func(snapshot *persistSnapshot) []AgentJob { return snapshot.agentJob },
		func(disk *diskStore, value []AgentJob) error { return disk.saveAgentJobs(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.observation },
		func(signatures *persistSignatures) *string { return &signatures.observation },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.observation },
		func(snapshot *persistSnapshot) []ObservationRecord { return snapshot.observation },
		func(disk *diskStore, value []ObservationRecord) error { return disk.saveObservations(value) },
	),
	{
		dirtyFlag:          func(dirty *dirtyFlags) *bool { return &dirty.continuity },
		persistedSignature: func(signatures *persistSignatures) *string { return &signatures.continuity },
		snapshotSignature:  func(snapshot *persistSnapshot) *string { return &snapshot.signatures.continuity },
		snapshotValue:      func(snapshot *persistSnapshot) any { return snapshot.continuity },
		save: func(service *Service, snapshot persistSnapshot) error {
			if err := service.disk.saveContinuityRecords(snapshot.continuity); err != nil {
				return err
			}
			checkpoint := continuityRecordForCompatibility(snapshot.continuity)
			if checkpoint == nil {
				return service.disk.clearSessionCheckpoint()
			}
			return service.disk.saveSessionCheckpoint(*checkpoint)
		},
	},
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.work },
		func(signatures *persistSignatures) *string { return &signatures.work },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.work },
		func(snapshot *persistSnapshot) []WorkItem { return snapshot.work },
		func(disk *diskStore, value []WorkItem) error { return disk.saveWorkItems(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.flow },
		func(signatures *persistSignatures) *string { return &signatures.flow },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.flow },
		func(snapshot *persistSnapshot) []FlowRun { return snapshot.flow },
		func(disk *diskStore, value []FlowRun) error { return disk.saveFlows(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.approval },
		func(signatures *persistSignatures) *string { return &signatures.approval },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.approval },
		func(snapshot *persistSnapshot) []ApprovalItem { return snapshot.approval },
		func(disk *diskStore, value []ApprovalItem) error { return disk.saveApprovalItems(value) },
	),
	{
		dirtyFlag:          func(dirty *dirtyFlags) *bool { return &dirty.event },
		persistedSignature: func(signatures *persistSignatures) *string { return &signatures.event },
		snapshotSignature:  func(snapshot *persistSnapshot) *string { return &snapshot.signatures.event },
		snapshotValue:      func(snapshot *persistSnapshot) any { return snapshot.event },
		save: func(service *Service, snapshot persistSnapshot) error {
			if err := service.disk.appendEventArchive(snapshot.pendingEventArchive); err != nil {
				return err
			}
			if err := service.disk.saveEvents(snapshot.event); err != nil {
				return err
			}
			service.event.markArchivePersisted()
			return nil
		},
		isDirty: func(snapshot *persistSnapshot, persisted *persistSignatures) bool {
			return len(snapshot.pendingEventArchive) > 0 || snapshot.signatures.event != persisted.event
		},
	},
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.release },
		func(signatures *persistSignatures) *string { return &signatures.release },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.release },
		func(snapshot *persistSnapshot) []ReleasePacket { return snapshot.release },
		func(disk *diskStore, value []ReleasePacket) error { return disk.saveReleases(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.deploy },
		func(signatures *persistSignatures) *string { return &signatures.deploy },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.deploy },
		func(snapshot *persistSnapshot) []DeployRun { return snapshot.deploy },
		func(disk *diskStore, value []DeployRun) error { return disk.saveDeploys(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.rollback },
		func(signatures *persistSignatures) *string { return &signatures.rollback },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.rollback },
		func(snapshot *persistSnapshot) []RollbackRun { return snapshot.rollback },
		func(disk *diskStore, value []RollbackRun) error { return disk.saveRollbacks(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.target },
		func(signatures *persistSignatures) *string { return &signatures.target },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.target },
		func(snapshot *persistSnapshot) []DeployTarget { return snapshot.target },
		func(disk *diskStore, value []DeployTarget) error { return disk.saveTargets(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.preflight },
		func(signatures *persistSignatures) *string { return &signatures.preflight },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.preflight },
		func(snapshot *persistSnapshot) []PreflightRecord { return snapshot.preflight },
		func(disk *diskStore, value []PreflightRecord) error { return disk.savePreflights(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.policy },
		func(signatures *persistSignatures) *string { return &signatures.policy },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.policy },
		func(snapshot *persistSnapshot) []PolicyDecision { return snapshot.policy },
		func(disk *diskStore, value []PolicyDecision) error { return disk.savePolicies(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.gateway },
		func(signatures *persistSignatures) *string { return &signatures.gateway },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.gateway },
		func(snapshot *persistSnapshot) []GatewayCall { return snapshot.gateway },
		func(disk *diskStore, value []GatewayCall) error { return disk.saveGatewayCalls(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.execute },
		func(signatures *persistSignatures) *string { return &signatures.execute },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.execute },
		func(snapshot *persistSnapshot) []ExecuteRun { return snapshot.execute },
		func(disk *diskStore, value []ExecuteRun) error { return disk.saveExecutes(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.verify },
		func(signatures *persistSignatures) *string { return &signatures.verify },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.verify },
		func(snapshot *persistSnapshot) []VerificationRecord { return snapshot.verify },
		func(disk *diskStore, value []VerificationRecord) error { return disk.saveVerifications(value) },
	),
	{
		dirtyFlag:          func(dirty *dirtyFlags) *bool { return &dirty.reviewRoom },
		persistedSignature: func(signatures *persistSignatures) *string { return &signatures.reviewRoom },
		snapshotSignature:  func(snapshot *persistSnapshot) *string { return &snapshot.signatures.reviewRoom },
		snapshotValue:      func(snapshot *persistSnapshot) any { return snapshot.reviewRoom },
		save: func(service *Service, snapshot persistSnapshot) error {
			if err := service.disk.saveReviewRoom(snapshot.reviewRoom); err != nil {
				return err
			}
			return service.disk.saveReviewRoomSeal(snapshot.reviewRoom)
		},
	},
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.memory },
		func(signatures *persistSignatures) *string { return &signatures.memory },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.memory },
		func(snapshot *persistSnapshot) SystemMemory { return snapshot.memory },
		func(disk *diskStore, value SystemMemory) error { return disk.saveSystemMemory(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.auth },
		func(signatures *persistSignatures) *string { return &signatures.auth },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.auth },
		func(snapshot *persistSnapshot) authConfig { return snapshot.auth },
		func(disk *diskStore, value authConfig) error { return disk.saveAuthConfig(value) },
	),
	standardPersistDomain(
		func(dirty *dirtyFlags) *bool { return &dirty.state },
		func(signatures *persistSignatures) *string { return &signatures.state },
		func(snapshot *persistSnapshot) *string { return &snapshot.signatures.state },
		func(snapshot *persistSnapshot) CompanyState { return snapshot.state },
		func(disk *diskStore, value CompanyState) error { return disk.saveState(value) },
	),
}

func populatePersistSignatures(snapshot *persistSnapshot) error {
	for _, domain := range persistDomains {
		signature, err := signatureFor(domain.snapshotValue(snapshot))
		if err != nil {
			return err
		}
		*domain.snapshotSignature(snapshot) = signature
	}
	return nil
}

func persistedSnapshotFromDisk(disk *diskStore) (persistSnapshot, error) {
	state, err := disk.loadState()
	if err != nil {
		return persistSnapshot{}, err
	}
	branches, err := disk.loadBranches()
	if err != nil {
		return persistSnapshot{}, err
	}
	proposals, err := disk.loadProposals()
	if err != nil {
		return persistSnapshot{}, err
	}
	agentJobs, err := disk.loadAgentJobs()
	if err != nil {
		return persistSnapshot{}, err
	}
	observations, err := disk.loadObservations()
	if err != nil {
		return persistSnapshot{}, err
	}
	continuity, err := disk.loadContinuityRecords()
	if err != nil {
		return persistSnapshot{}, err
	}
	workItems, err := disk.loadWorkItems()
	if err != nil {
		return persistSnapshot{}, err
	}
	flows, err := disk.loadFlows()
	if err != nil {
		return persistSnapshot{}, err
	}
	approvalItems, err := disk.loadApprovalItems()
	if err != nil {
		return persistSnapshot{}, err
	}
	events, err := disk.loadEvents()
	if err != nil {
		return persistSnapshot{}, err
	}
	releases, err := disk.loadReleases()
	if err != nil {
		return persistSnapshot{}, err
	}
	deploys, err := disk.loadDeploys()
	if err != nil {
		return persistSnapshot{}, err
	}
	rollbacks, err := disk.loadRollbacks()
	if err != nil {
		return persistSnapshot{}, err
	}
	targets, err := disk.loadTargets()
	if err != nil {
		return persistSnapshot{}, err
	}
	preflights, err := disk.loadPreflights()
	if err != nil {
		return persistSnapshot{}, err
	}
	policies, err := disk.loadPolicies()
	if err != nil {
		return persistSnapshot{}, err
	}
	gatewayCalls, err := disk.loadGatewayCalls()
	if err != nil {
		return persistSnapshot{}, err
	}
	executes, err := disk.loadExecutes()
	if err != nil {
		return persistSnapshot{}, err
	}
	verifications, err := disk.loadVerifications()
	if err != nil {
		return persistSnapshot{}, err
	}
	reviewRoom, err := disk.loadReviewRoom()
	if err != nil {
		return persistSnapshot{}, err
	}
	memory, err := disk.loadSystemMemory()
	if err != nil {
		return persistSnapshot{}, err
	}
	auth, err := disk.loadAuthConfig()
	if err != nil {
		return persistSnapshot{}, err
	}
	snapshot := persistSnapshot{
		state:               state,
		branch:              branches,
		proposal:            proposals,
		agentJob:            agentJobs,
		observation:         observations,
		continuity:          continuity,
		work:                workItems,
		flow:                flows,
		approval:            approvalItems,
		event:               events,
		pendingEventArchive: []EventEnvelope{},
		release:             releases,
		deploy:              deploys,
		rollback:            rollbacks,
		target:              targets,
		preflight:           preflights,
		policy:              policies,
		gateway:             gatewayCalls,
		execute:             executes,
		verify:              verifications,
		reviewRoom:          reviewRoom,
		memory:              memory,
		auth:                auth,
	}
	if err := populatePersistSignatures(&snapshot); err != nil {
		return persistSnapshot{}, err
	}
	return snapshot, nil
}

type Service struct {
	mu                    sync.Mutex
	actorScope            sync.Mutex
	actorOverride         atomic.Value
	founderTelegramNotice founderTelegramNoticeState
	dirty                 dirtyFlags
	persisted             persistSignatures
	repoRoot              string
	daemonStartedAt       time.Time
	state                 *stateStore
	branch                *branchStore
	proposal              *proposalStore
	agentJob              *agentJobStore
	observation           *observationStore
	continuity            *continuityStore
	work                  *workStore
	flow                  *flowStore
	approval              *approvalStore
	event                 *eventStore
	release               *releaseStore
	deploy                *deployStore
	rollback              *rollbackStore
	target                *targetStore
	preflight             *preflightStore
	policy                *policyStore
	gateway               *gatewayStore
	execute               *executeStore
	verify                *verificationStore
	reviewRoom            *reviewRoomStore
	memory                *memoryStore
	auth                  *authStore
	disk                  *diskStore
	gatewayAdapter        GatewayAdapter
	verifyAdapter         VerifyAdapter
	deployAdapter         DeployAdapter
	rollbackAdapter       DeployAdapter
	telegramAdapter       TelegramAdapter
}

func NewService(dataDir string) (*Service, error) {
	disk, err := newDiskStore(dataDir)
	if err != nil {
		return nil, err
	}

	state, err := disk.loadState()
	if err != nil {
		return nil, err
	}
	branches, err := disk.loadBranches()
	if err != nil {
		return nil, err
	}
	proposals, err := disk.loadProposals()
	if err != nil {
		return nil, err
	}
	agentJobs, err := disk.loadAgentJobs()
	if err != nil {
		return nil, err
	}
	observations, err := disk.loadObservations()
	if err != nil {
		return nil, err
	}
	continuity, err := disk.loadContinuityRecords()
	if err != nil {
		return nil, err
	}
	workItems, err := disk.loadWorkItems()
	if err != nil {
		return nil, err
	}
	flows, err := disk.loadFlows()
	if err != nil {
		return nil, err
	}
	approvalItems, err := disk.loadApprovalItems()
	if err != nil {
		return nil, err
	}
	events, err := disk.loadEvents()
	if err != nil {
		return nil, err
	}
	eventArchiveCount, err := disk.countEventArchive()
	if err != nil {
		return nil, err
	}
	releases, err := disk.loadReleases()
	if err != nil {
		return nil, err
	}
	deploys, err := disk.loadDeploys()
	if err != nil {
		return nil, err
	}
	rollbacks, err := disk.loadRollbacks()
	if err != nil {
		return nil, err
	}
	targets, err := disk.loadTargets()
	if err != nil {
		return nil, err
	}
	preflights, err := disk.loadPreflights()
	if err != nil {
		return nil, err
	}
	policies, err := disk.loadPolicies()
	if err != nil {
		return nil, err
	}
	gatewayCalls, err := disk.loadGatewayCalls()
	if err != nil {
		return nil, err
	}
	executes, err := disk.loadExecutes()
	if err != nil {
		return nil, err
	}
	verifications, err := disk.loadVerifications()
	if err != nil {
		return nil, err
	}
	reviewRoom, err := disk.loadReviewRoom()
	if err != nil {
		return nil, err
	}
	repoRoot := repoRootFromEnv()
	if normalizedRoom, changed := normalizeRuntimeReviewRoomSource(reviewRoom, repoRoot, disk.reviewRoomPath()); changed {
		reviewRoom = normalizedRoom
		if err := disk.saveReviewRoom(reviewRoom); err != nil {
			return nil, err
		}
		if err := disk.saveReviewRoomSeal(reviewRoom); err != nil {
			return nil, err
		}
	}
	reviewRoomSeal, err := disk.loadReviewRoomSeal()
	if err != nil {
		return nil, err
	}
	integrityAudit := verifyReviewRoomSeal(reviewRoom, reviewRoomSeal)
	if !integrityAudit.Sealed {
		if err := disk.saveReviewRoomSeal(reviewRoom); err != nil {
			return nil, err
		}
	} else if len(integrityAudit.Issues) > 0 {
		reviewRoom = applyReviewRoomIntegrityGuardrail(reviewRoom, integrityAudit)
	}
	memory, err := disk.loadSystemMemory()
	if err != nil {
		return nil, err
	}
	if len(continuity) == 0 {
		if checkpoint, checkpointErr := disk.loadSessionCheckpoint(); checkpointErr != nil {
			return nil, checkpointErr
		} else if checkpoint != nil {
			continuity = []ContinuityRecord{continuityRecordFromSessionCheckpoint(*checkpoint)}
		}
	}
	auth, err := disk.loadAuthConfig()
	if err != nil {
		return nil, err
	}

	state.WorkItemsActive = len(workItems)
	state.ApprovalsPending = len(approvalItems)
	if len(releases) > 0 {
		state.LastReleaseAt = releases[len(releases)-1].ReleasedAt
	}
	if len(deploys) > 0 {
		if deploys[len(deploys)-1].Status == "failed" {
			state.DeployHealth = "degraded"
		} else {
			state.DeployHealth = "ready"
		}
	}

	service := &Service{
		repoRoot:        repoRoot,
		state:           newStateStore(state),
		branch:          newBranchStore(branches),
		proposal:        newProposalStore(proposals),
		agentJob:        newAgentJobStore(agentJobs),
		observation:     newObservationStore(observations),
		continuity:      newContinuityStore(continuity),
		work:            newWorkStore(workItems),
		flow:            newFlowStore(flows),
		approval:        newApprovalStore(approvalItems),
		event:           newEventStoreWithArchiveCount(events, eventArchiveCount),
		release:         newReleaseStore(releases),
		deploy:          newDeployStore(deploys),
		rollback:        newRollbackStore(rollbacks),
		target:          newTargetStore(targets),
		preflight:       newPreflightStore(preflights),
		policy:          newPolicyStore(policies),
		gateway:         newGatewayStore(gatewayCalls),
		execute:         newExecuteStore(executes),
		verify:          newVerificationStore(verifications),
		reviewRoom:      newReviewRoomStore(reviewRoom),
		memory:          newMemoryStore(memory),
		auth:            newAuthStore(auth),
		disk:            disk,
		gatewayAdapter:  gatewayAdapterFromEnv(),
		verifyAdapter:   commandVerifyAdapter{},
		deployAdapter:   commandDeployAdapter{},
		rollbackAdapter: commandDeployAdapter{},
		telegramAdapter: telegramAdapterFromEnv(),
	}
	persistedSnapshot, err := persistedSnapshotFromDisk(disk)
	if err != nil {
		return nil, err
	}
	service.persisted = persistedSnapshot.signatures
	service.reclaimStaleJobsOnStartup()
	return service, nil
}

func (s *Service) currentPersistSnapshot() (persistSnapshot, error) {
	snapshot := persistSnapshot{
		state:               s.state.current(),
		branch:              s.branch.list(),
		proposal:            s.proposal.list(),
		agentJob:            s.agentJob.list(),
		observation:         s.observation.list(),
		continuity:          s.continuity.list(),
		work:                s.work.list(),
		flow:                s.flow.list(),
		approval:            s.approval.list(),
		event:               s.event.recent(),
		pendingEventArchive: s.event.pendingArchiveList(),
		release:             s.release.list(),
		deploy:              s.deploy.list(),
		rollback:            s.rollback.list(),
		target:              s.target.list(),
		preflight:           s.preflight.list(),
		policy:              s.policy.list(),
		gateway:             s.gateway.list(),
		execute:             s.execute.list(),
		verify:              s.verify.list(),
		reviewRoom:          s.reviewRoom.current(),
		memory:              s.memory.current(),
		auth:                s.auth.current(),
	}
	if err := populatePersistSignatures(&snapshot); err != nil {
		return persistSnapshot{}, err
	}
	return snapshot, nil
}

func (s *Service) syncDirtyLocked(snapshot persistSnapshot) {
	var dirty dirtyFlags
	for _, domain := range persistDomains {
		*domain.dirtyFlag(&dirty) = domain.dirty(&snapshot, &s.persisted)
	}
	s.dirty = dirty
}

func gatewayAdapterFromEnv() GatewayAdapter {
	base := singleGatewayAdapterFromEnv()
	// If role-based providers are configured, wrap in a multi-provider router
	// so each job is dispatched to the adapter matching its assigned provider.
	if len(splitRoleBindingEnv(os.Getenv("LAYER_OS_AGENT_ROLE_PROVIDERS"))) > 0 {
		return newMultiProviderAdapter(base)
	}
	return base
}

func singleGatewayAdapterFromEnv() GatewayAdapter {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LAYER_OS_GATEWAY_ADAPTER"))) {
	case "api":
		return apiGatewayAdapter{}
	case "gemini":
		return geminiDirectAdapter{}
	case "openai":
		return openaiDirectAdapter{}
	case "claude":
		return claudeDirectAdapter{}
	}
	return recordGatewayAdapter{}
}

func (s *Service) Adapters() AdapterSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AdapterSummary{
		Gateway:                s.gatewayAdapter.Name(),
		GatewaySemantics:       s.gatewayAdapter.Semantics(),
		GatewayDispatchEnabled: s.gatewayAdapter.DispatchEnabled(),
		GatewayRequiredMode:    canonicalPolicyMode(s.gatewayAdapter.RequiredMode()),
		Verify:                 s.verifyAdapter.Name(),
		Deploy:                 s.deployAdapter.Name(),
		Rollback:               s.rollbackAdapter.Name(),
	}
}

func (s *Service) Providers() []ProviderSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return providerSummaries(AdapterSummary{
		Gateway:                s.gatewayAdapter.Name(),
		GatewaySemantics:       s.gatewayAdapter.Semantics(),
		GatewayDispatchEnabled: s.gatewayAdapter.DispatchEnabled(),
		GatewayRequiredMode:    canonicalPolicyMode(s.gatewayAdapter.RequiredMode()),
		Verify:                 s.verifyAdapter.Name(),
		Deploy:                 s.deployAdapter.Name(),
		Rollback:               s.rollbackAdapter.Name(),
	})
}

func (s *Service) Status() CompanyState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.updateWorkItemsActive(s.work.count())
	s.state.updateApprovalsPending(s.approval.count())
	return s.enrichedCompanyStateLocked(s.state.current())
}

func (s *Service) ListProposals() []ProposalItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.proposal.list()
}

func (s *Service) ListAgentJobs() []AgentJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agentJob.list()
}

func (s *Service) ListWorkItems() []WorkItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.work.list()
}

func (s *Service) ListFlows() []FlowRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flow.list()
}

func (s *Service) CreateWorkItem(item WorkItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureActiveBranchLocked(item.BranchID); err != nil {
		return err
	}
	oldItems := s.work.list()
	oldState := s.state.current()
	oldEventState := s.event.state()

	if err := s.work.create(item); err != nil {
		return err
	}
	s.state.updateWorkItemsActive(s.work.count())

	if err := s.event.append(newEvent("work_item.created", s.currentActor(), item.Surface, item.ID, item.Stage, map[string]any{
		"title": item.Title,
	})); err != nil {
		s.work = newWorkStore(oldItems)
		s.state = newStateStore(oldState)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.work = newWorkStore(oldItems)
		s.state = newStateStore(oldState)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) CreateFlow(item FlowRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	work, ok := s.work.get(item.WorkItemID)
	if !ok {
		return errors.New("flow work_item_id not found")
	}
	if item.Status == "" {
		item.Status = "active"
	}
	if item.Notes == nil {
		item.Notes = []string{}
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = zeroSafeNow()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}

	oldItems := s.flow.list()
	oldEventState := s.event.state()
	if err := s.flow.create(item); err != nil {
		return err
	}
	if err := s.event.append(newEvent("flow.created", s.currentActor(), work.Surface, work.ID, work.Stage, map[string]any{
		"flow_id": item.FlowID,
		"status":  item.Status,
	})); err != nil {
		s.flow = newFlowStore(oldItems)
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.flow = newFlowStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return err
	}
	return nil
}

func (s *Service) SyncFlow(flowID string, workItemID string, approvalID *string, policyDecisionID *string, executeID *string, verificationID *string, releaseID *string, deployID *string, rollbackID *string, notes []string) (FlowRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	work, ok := s.work.get(workItemID)
	if !ok {
		return FlowRun{}, errors.New("flow work_item_id not found")
	}
	if approvalID != nil && strings.TrimSpace(*approvalID) != "" {
		if _, ok := s.approval.get(strings.TrimSpace(*approvalID)); !ok {
			return FlowRun{}, errors.New("flow approval_id not found")
		}
	}
	if policyDecisionID != nil && strings.TrimSpace(*policyDecisionID) != "" {
		if _, ok := s.policy.get(strings.TrimSpace(*policyDecisionID)); !ok {
			return FlowRun{}, errors.New("flow policy_decision_id not found")
		}
	}
	if executeID != nil && strings.TrimSpace(*executeID) != "" {
		found := false
		for _, item := range s.execute.list() {
			if item.ExecuteID == strings.TrimSpace(*executeID) {
				found = true
				break
			}
		}
		if !found {
			return FlowRun{}, errors.New("flow execute_id not found")
		}
	}
	if verificationID != nil && strings.TrimSpace(*verificationID) != "" {
		found := false
		for _, item := range s.verify.list() {
			if item.RecordID == strings.TrimSpace(*verificationID) {
				found = true
				break
			}
		}
		if !found {
			return FlowRun{}, errors.New("flow verification_id not found")
		}
	}
	if releaseID != nil && strings.TrimSpace(*releaseID) != "" {
		if _, ok := s.release.get(strings.TrimSpace(*releaseID)); !ok {
			return FlowRun{}, errors.New("flow release_id not found")
		}
	}
	if deployID != nil && strings.TrimSpace(*deployID) != "" {
		if _, ok := s.deploy.get(strings.TrimSpace(*deployID)); !ok {
			return FlowRun{}, errors.New("flow deploy_id not found")
		}
	}
	if rollbackID != nil && strings.TrimSpace(*rollbackID) != "" {
		found := false
		for _, item := range s.rollback.list() {
			if item.RollbackID == strings.TrimSpace(*rollbackID) {
				found = true
				break
			}
		}
		if !found {
			return FlowRun{}, errors.New("flow rollback_id not found")
		}
	}

	now := zeroSafeNow()
	existing, exists := s.flow.get(flowID)
	item := FlowRun{
		FlowID:           flowID,
		WorkItemID:       work.ID,
		ApprovalID:       normalizeRef(approvalID),
		PolicyDecisionID: normalizeRef(policyDecisionID),
		ExecuteID:        normalizeRef(executeID),
		VerificationID:   normalizeRef(verificationID),
		ReleaseID:        normalizeRef(releaseID),
		DeployID:         normalizeRef(deployID),
		RollbackID:       normalizeRef(rollbackID),
		Notes:            append([]string{}, notes...),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if exists {
		item.CreatedAt = existing.CreatedAt
	}
	item.Status = s.deriveFlowStatusLocked(item)

	oldItems := s.flow.list()
	oldEventState := s.event.state()
	if err := s.flow.update(item); err != nil {
		return FlowRun{}, err
	}
	if err := s.event.append(newEvent("flow.synced", s.currentActor(), work.Surface, work.ID, work.Stage, map[string]any{
		"flow_id": item.FlowID,
		"status":  item.Status,
	})); err != nil {
		s.flow = newFlowStore(oldItems)
		return FlowRun{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.flow = newFlowStore(oldItems)
		s.event = newEventStoreFromState(oldEventState)
		return FlowRun{}, err
	}
	return item, nil
}

func (s *Service) ListApprovals() []ApprovalItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.approval.list()
}

func (s *Service) ListEvents() []EventEnvelope {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.event.recent()
}

func (s *Service) ListPreflights() []PreflightRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.preflight.list()
}

func (s *Service) ListPolicies() []PolicyDecision {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.policy.list()
}

func (s *Service) ListGatewayCalls() []GatewayCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gateway.list()
}

func (s *Service) ListExecutes() []ExecuteRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.execute.list()
}

func (s *Service) ListVerifications() []VerificationRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.verify.list()
}

func (s *Service) Memory() SystemMemory {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memory.current()
}

func (s *Service) ReviewRoom() ReviewRoom {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentReviewRoomLocked()
}

func (s *Service) ReviewRoomSummary() ReviewRoomSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SummarizeReviewRoom(s.currentReviewRoomLocked())
}

func (s *Service) autoOpenReviewRoomItemLocked(item ReviewRoomItem) error {
	updated, added, err := ensureReviewRoomItem(s.reviewRoom.current(), "open", item)
	if err != nil || !added {
		return err
	}
	s.reviewRoom = newReviewRoomStore(updated)
	item = reviewRoomItemFromSection(updated, "open", item.Text)
	data := map[string]any{
		"section":  "open",
		"item":     strings.TrimSpace(item.Text),
		"kind":     item.Kind,
		"severity": item.Severity,
		"source":   item.Source,
	}
	if item.Ref != nil {
		data["ref"] = strings.TrimSpace(*item.Ref)
	}
	if strings.TrimSpace(item.Why) != "" {
		data["why"] = item.Why
	}
	if item.WhyUnresolved != nil {
		data["why_unresolved"] = strings.TrimSpace(*item.WhyUnresolved)
	}
	if item.Contradiction != nil {
		data["contradiction"] = strings.TrimSpace(*item.Contradiction)
	}
	if len(item.Contradictions) > 0 {
		data["contradictions"] = append([]string{}, item.Contradictions...)
	}
	if len(item.PatternRefs) > 0 {
		data["pattern_refs"] = append([]string{}, item.PatternRefs...)
	}
	if item.Rationale != nil {
		data["rationale"] = map[string]any{
			"trigger": item.Rationale.Trigger,
			"reason":  item.Rationale.Reason,
			"rule":    item.Rationale.Rule,
		}
	}
	if len(item.Evidence) > 0 {
		data["evidence"] = append([]string{}, item.Evidence...)
	}
	return s.event.append(s.newSystemEvent("review_room.item_auto_added", data))
}

func (s *Service) AddStructuredReviewRoomItem(section string, item ReviewRoomItem) (ReviewRoom, error) {
	s.mu.Lock()

	item = normalizeReviewRoomItem(item)
	if err := item.Validate(); err != nil {
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()
	updated, err := s.reviewRoom.add(section, item)
	if err != nil {
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	item = reviewRoomItemFromSection(updated, section, item.Text)
	data := map[string]any{
		"section":  normalizeReviewSection(section),
		"item":     strings.TrimSpace(item.Text),
		"kind":     item.Kind,
		"severity": item.Severity,
		"source":   item.Source,
	}
	if item.Ref != nil {
		data["ref"] = strings.TrimSpace(*item.Ref)
	}
	if strings.TrimSpace(item.Why) != "" {
		data["why"] = item.Why
	}
	if item.WhyUnresolved != nil {
		data["why_unresolved"] = strings.TrimSpace(*item.WhyUnresolved)
	}
	if item.Contradiction != nil {
		data["contradiction"] = strings.TrimSpace(*item.Contradiction)
	}
	if len(item.Contradictions) > 0 {
		data["contradictions"] = append([]string{}, item.Contradictions...)
	}
	if len(item.PatternRefs) > 0 {
		data["pattern_refs"] = append([]string{}, item.PatternRefs...)
	}
	if item.Rationale != nil {
		data["rationale"] = map[string]any{
			"trigger": item.Rationale.Trigger,
			"reason":  item.Rationale.Reason,
			"rule":    item.Rationale.Rule,
		}
	}
	if len(item.Evidence) > 0 {
		data["evidence"] = append([]string{}, item.Evidence...)
	}
	if err := s.event.append(s.newSystemEvent("review_room.item_added", data)); err != nil {
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	s.mu.Unlock()
	s.maybeNotifyFounderAttention()
	return updated, nil
}

func reviewRoomItemFromSection(room ReviewRoom, section string, text string) ReviewRoomItem {
	room = normalizeReviewRoom(room)
	items := reviewRoomSection(&room, normalizeReviewSection(section))
	if items == nil {
		return ReviewRoomItem{Text: strings.TrimSpace(text)}
	}
	for _, item := range *items {
		if item.Text == strings.TrimSpace(text) {
			return item
		}
	}
	return ReviewRoomItem{Text: strings.TrimSpace(text)}
}

func (s *Service) AddReviewRoomItem(section string, item string) (ReviewRoom, error) {
	return s.AddStructuredReviewRoomItem(section, newManualReviewRoomItem(item, "agenda", "medium", "manual", nil))
}

func reviewRoomRationaleData(rationale *ReviewRoomRationale) map[string]any {
	if rationale == nil {
		return nil
	}
	return map[string]any{
		"trigger": rationale.Trigger,
		"reason":  rationale.Reason,
		"rule":    rationale.Rule,
	}
}

func reviewRoomResolutionData(resolution *ReviewRoomResolution) map[string]any {
	if resolution == nil {
		return nil
	}
	payload := map[string]any{
		"action":      resolution.Action,
		"reason":      resolution.Reason,
		"rule":        resolution.Rule,
		"resolved_at": resolution.ResolvedAt.Format(time.RFC3339Nano),
	}
	if len(resolution.Evidence) > 0 {
		payload["evidence"] = append([]string{}, resolution.Evidence...)
	}
	return payload
}

func defaultReviewRoomResolution(action string, resolution *ReviewRoomResolution, resolvedAt time.Time) *ReviewRoomResolution {
	action = normalizeReviewAction(action)
	if action == "" {
		return normalizeReviewRoomResolution(resolution)
	}
	base := &ReviewRoomResolution{
		Action:     action,
		ResolvedAt: resolvedAt,
	}
	switch action {
	case "accept":
		base.Reason = "operator accepted the agenda item into the active decision set"
		base.Rule = "review_room.transition.accept"
	case "defer":
		base.Reason = "operator deferred the agenda item for later review"
		base.Rule = "review_room.transition.defer"
	case "resolve":
		base.Reason = "operator resolved the agenda item and closed the loop"
		base.Rule = "review_room.transition.resolve"
	}
	if resolution != nil {
		if value := normalizeReviewRoomResolution(resolution); value != nil {
			if value.Action != "" {
				base.Action = value.Action
			}
			if value.Reason != "" {
				base.Reason = value.Reason
			}
			if value.Rule != "" {
				base.Rule = value.Rule
			}
			if len(value.Evidence) > 0 {
				base.Evidence = append([]string{}, value.Evidence...)
			}
			if !value.ResolvedAt.IsZero() {
				base.ResolvedAt = value.ResolvedAt
			}
		}
	}
	return normalizeReviewRoomResolution(base)
}

func (s *Service) TransitionReviewRoomItem(action string, item string) (ReviewRoom, error) {
	return s.TransitionStructuredReviewRoomItem(action, item, nil)
}

func (s *Service) TransitionStructuredReviewRoomItem(action string, item string, resolution *ReviewRoomResolution) (ReviewRoom, error) {
	s.mu.Lock()

	oldRoom := s.reviewRoom.current()
	oldEventState := s.event.state()
	updated, transitioned, err := s.reviewRoom.transitionStructured(action, item, resolution)
	if err != nil {
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	data := map[string]any{
		"action": normalizeReviewAction(action),
		"item":   strings.TrimSpace(item),
	}
	if transitioned != nil {
		data["kind"] = transitioned.Kind
		data["severity"] = transitioned.Severity
		data["source"] = transitioned.Source
		if transitioned.Ref != nil {
			data["ref"] = strings.TrimSpace(*transitioned.Ref)
		}
		if strings.TrimSpace(transitioned.Why) != "" {
			data["why"] = transitioned.Why
		}
		if transitioned.WhyUnresolved != nil {
			data["why_unresolved"] = strings.TrimSpace(*transitioned.WhyUnresolved)
		}
		if transitioned.Contradiction != nil {
			data["contradiction"] = strings.TrimSpace(*transitioned.Contradiction)
		}
		if len(transitioned.Contradictions) > 0 {
			data["contradictions"] = append([]string{}, transitioned.Contradictions...)
		}
		if len(transitioned.PatternRefs) > 0 {
			data["pattern_refs"] = append([]string{}, transitioned.PatternRefs...)
		}
		if transitioned.Rationale != nil {
			data["rationale"] = reviewRoomRationaleData(transitioned.Rationale)
		}
		if transitioned.Resolution != nil {
			data["resolution"] = reviewRoomResolutionData(transitioned.Resolution)
		}
		if len(transitioned.Evidence) > 0 {
			data["evidence"] = append([]string{}, transitioned.Evidence...)
		}
	}
	if err := s.event.append(s.newSystemEvent("review_room.item_transitioned", data)); err != nil {
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.reviewRoom = newReviewRoomStore(oldRoom)
		s.event = newEventStoreFromState(oldEventState)
		s.mu.Unlock()
		return ReviewRoom{}, err
	}
	s.mu.Unlock()
	s.maybeNotifyFounderAttention()
	return updated, nil
}

func (s *Service) WriteLease() WriteLease {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disk.currentLease()
}

func (s *Service) Handoff() HandoffPacket {
	s.mu.Lock()
	s.state.updateWorkItemsActive(s.work.count())
	branches := s.branch.list()
	s.state.updateApprovalsPending(s.approval.count())
	proposals := s.proposal.list()
	agentJobs := s.agentJob.list()
	observations := s.observation.list()
	workItems := s.work.list()
	flows := s.flow.list()
	approvals := s.approval.list()
	releases := s.release.list()
	deploys := s.deploy.list()
	targets := s.target.list()

	founderView := s.deriveFounderViewLocked()
	fullReviewRoom := s.currentReviewRoomLocked()
	reviewRoom := SummarizeReviewRoom(fullReviewRoom)
	founderSummary := s.deriveFounderSummaryLocked(founderView, reviewRoom)
	continuityRecords := s.continuity.list()
	disk := s.disk
	companyState := s.state.current()
	systemMemory := s.memory.current()
	auth := s.auth.status()
	writeLease := s.disk.currentLease()
	eventCount := s.event.count()
	rollbackCount := s.rollback.count()
	preflightCount := s.preflight.count()
	policyCount := s.policy.count()
	gatewayCount := s.gateway.count()
	executeCount := s.execute.count()
	verifyCount := s.verify.count()
	s.mu.Unlock()

	_, _, surprising, _, proposalCandidates, openThreads := loadKnowledgeContext(disk, observations, systemMemory, founderSummary, reviewRoom, fullReviewRoom)
	actionRoutes := deriveKnowledgeActionRoutes(systemMemory, reviewRoom, surprising, proposalCandidates, openThreads)
	continuity := continuityViewForBootstrap(continuityRecords, continuityRecordCompatibilityAny, continuitySourceTerminal, KnowledgePacket{
		OpenThreads:  limitOpenThreads(openThreads, 3),
		ActionRoutes: actionRoutes,
	})
	parallelCandidates := deriveParallelCandidates(founderView, founderSummary, fullReviewRoom)
	return HandoffPacket{
		GeneratedAt:        zeroSafeNow(),
		CompanyState:       companyState,
		SystemMemory:       systemMemory,
		Auth:               auth,
		WriteLease:         writeLease,
		FounderView:        founderView,
		FounderSummary:     founderSummary,
		ReviewRoom:         reviewRoom,
		Continuity:         continuity,
		ActiveBranches:     activeBranches(branches, 3),
		ParallelCandidates: parallelCandidates,
		Counts: HandoffCounts{
			Branches:      len(branches),
			Proposals:     len(proposals),
			AgentJobs:     len(agentJobs),
			Observations:  len(observations),
			WorkItems:     len(workItems),
			Flows:         len(flows),
			Approvals:     len(approvals),
			Releases:      len(releases),
			Events:        eventCount,
			Deploys:       len(deploys),
			Rollbacks:     rollbackCount,
			Targets:       len(targets),
			Preflights:    preflightCount,
			Policies:      policyCount,
			GatewayCalls:  gatewayCount,
			Executes:      executeCount,
			Verifications: verifyCount,
		},
	}
}

func repoRootFromEnv() string {
	root := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT"))
	if root == "" {
		return "."
	}
	return root
}

func (s *Service) AuthStatus() AuthStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.auth.status()
}

func (s *Service) SetWriteToken(token string) (AuthStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.auth.current()
	oldEventState := s.event.state()
	actor := s.currentActor()
	if err := s.auth.setToken(token, actor, zeroSafeNow()); err != nil {
		return AuthStatus{}, err
	}
	status := s.auth.status()
	if err := s.event.append(s.newSystemEvent("auth.enabled", map[string]any{
		"write_auth_enabled": status.WriteAuthEnabled,
		"actor":              actor,
	})); err != nil {
		s.auth = newAuthStore(old)
		return AuthStatus{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.auth = newAuthStore(old)
		s.event = newEventStoreFromState(oldEventState)
		return AuthStatus{}, err
	}
	return status, nil
}

func (s *Service) ClearWriteToken() (AuthStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.auth.current()
	oldEventState := s.event.state()
	actor := s.currentActor()
	s.auth.clear(actor, zeroSafeNow())
	status := s.auth.status()
	if err := s.event.append(s.newSystemEvent("auth.cleared", map[string]any{
		"write_auth_enabled": status.WriteAuthEnabled,
		"actor":              actor,
	})); err != nil {
		s.auth = newAuthStore(old)
		return AuthStatus{}, err
	}
	if err := s.persistLocked(); err != nil {
		s.auth = newAuthStore(old)
		s.event = newEventStoreFromState(oldEventState)
		return AuthStatus{}, err
	}
	return status, nil
}

func (s *Service) AuthorizeWriteToken(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.auth.authorize(token, s.currentActor())
}

func (s *Service) AuthorizeWriteTokenForActor(token string, actor string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.auth.authorize(token, actor)
}

func (s *Service) ReplaceMemory(memory SystemMemory) error {
	s.mu.Lock()

	oldMemory := s.memory.current()
	oldEventState := s.event.state()
	if err := s.memory.replace(memory); err != nil {
		s.mu.Unlock()
		return err
	}
	if err := s.event.append(s.newSystemEvent("memory.updated", map[string]any{
		"current_focus": memory.CurrentFocus,
	})); err != nil {
		s.memory = newMemoryStore(oldMemory)
		s.mu.Unlock()
		return err
	}
	if err := s.persistLocked(); err != nil {
		s.memory = newMemoryStore(oldMemory)
		s.event = newEventStoreFromState(oldEventState)
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()
	s.maybeNotifyFounderAttention()
	return nil
}

func (s *Service) persistLocked() error {
	if err := s.disk.ensureWriterLease(); err != nil {
		return err
	}
	if room, changed := syncAuditReviewRoom(s.reviewRoom.current(), s.repoRoot, s.daemonStartedAt); changed {
		s.reviewRoom = newReviewRoomStore(room)
	}
	snapshot, err := s.currentPersistSnapshot()
	if err != nil {
		return err
	}
	s.syncDirtyLocked(snapshot)
	if !s.dirty.any() {
		return nil
	}
	for _, domain := range persistDomains {
		if !*domain.dirtyFlag(&s.dirty) {
			continue
		}
		if err := domain.save(s, snapshot); err != nil {
			return err
		}
	}
	s.persisted = snapshot.signatures
	s.dirty = dirtyFlags{}
	return nil
}

func (s *Service) newSystemEvent(kind string, data map[string]any) EventEnvelope {
	return newEvent(kind, s.currentActor(), SurfaceAPI, "system", StageDiscover, data)
}

func (s *Service) WithActor(actor string, fn func(*Service) error) error {
	actor = normalizeActor(actor)
	if actor == "" {
		return fn(s)
	}
	prev := s.currentActorOverride()
	s.actorScope.Lock()
	s.actorOverride.Store(actor)
	defer func() {
		if prev == "" {
			s.actorOverride.Store("")
		} else {
			s.actorOverride.Store(prev)
		}
		s.actorScope.Unlock()
	}()
	return fn(s)
}

func (s *Service) currentActor() string {
	if actor := s.currentActorOverride(); actor != "" {
		return actor
	}
	return ResolveActor(os.Getenv("LAYER_OS_ACTOR"), os.Getenv("LAYER_OS_WRITER_ID"))
}

func (s *Service) currentActorOverride() string {
	value := s.actorOverride.Load()
	actor, _ := value.(string)
	return normalizeActor(actor)
}

func runtimeModelsFromEnv() []string {
	return mergeRuntimeModels(splitEnvModels("LAYER_OS_MODELS"), splitEnvModels("LAYER_OS_MODEL"))
}

func splitEnvModels(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return []string{}
	}
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

func mergeRuntimeModels(groups ...[]string) []string {
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

func normalizeRef(ref *string) *string {
	if ref == nil {
		return nil
	}
	value := strings.TrimSpace(*ref)
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) deriveFlowStatusLocked(item FlowRun) string {
	if item.RollbackID != nil {
		for _, rollback := range s.rollback.list() {
			if rollback.RollbackID != *item.RollbackID {
				continue
			}
			if rollback.Status == "failed" {
				return "blocked"
			}
			if rollback.Status == "succeeded" || rollback.Status == "recorded" {
				return "rolled_back"
			}
		}
	}
	if item.DeployID != nil {
		deploy, ok := s.deploy.get(*item.DeployID)
		if ok {
			if deploy.Status == "failed" {
				return "blocked"
			}
			if deploy.Status == "succeeded" {
				return "released"
			}
		}
	}
	if item.ReleaseID != nil {
		return "released"
	}
	if item.ApprovalID != nil {
		approval, ok := s.approval.get(*item.ApprovalID)
		if ok && approval.Status == "rejected" {
			return "blocked"
		}
	}
	if item.VerificationID != nil {
		for _, verification := range s.verify.list() {
			if verification.RecordID != *item.VerificationID {
				continue
			}
			if verification.Status == "failed" {
				return "blocked"
			}
			return "waiting"
		}
	}
	if item.ExecuteID != nil {
		for _, execute := range s.execute.list() {
			if execute.ExecuteID != *item.ExecuteID {
				continue
			}
			if execute.Status == "failed" {
				return "blocked"
			}
			return "waiting"
		}
	}
	if item.ApprovalID != nil {
		approval, ok := s.approval.get(*item.ApprovalID)
		if ok && approval.Status == "pending" {
			return "waiting"
		}
	}
	return "active"
}

func (s *Service) reclaimStaleJobsOnStartup() {
	jobs := s.agentJob.list()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, job := range jobs {
		if job.Status != "running" {
			continue
		}
		notes := appendUniqueString(append([]string{}, job.Notes...), "startup_reclaim")
		result := cloneJSONObject(job.Result)
		if result == nil {
			result = map[string]any{}
		}
		result["reclaim_reason"] = "daemon_restart"
		_, _, _ = s.updateAgentJobLocked(job.JobID, "failed", notes, result, "agent_job.failed")
	}
}

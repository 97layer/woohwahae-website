package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func (d *diskStore) loadState() (CompanyState, error) {
	return readJSON(d.statePath(), defaultCompanyState())
}

func (d *diskStore) saveState(state CompanyState) error {
	return writeJSONAtomic(d.statePath(), state)
}

func (d *diskStore) loadBranches() ([]Branch, error) {
	return readJSON(d.branchesPath(), []Branch{})
}

func (d *diskStore) saveBranches(items []Branch) error {
	return writeJSONAtomic(d.branchesPath(), items)
}

func (d *diskStore) loadProposals() ([]ProposalItem, error) {
	return readJSON(d.proposalsPath(), []ProposalItem{})
}

func (d *diskStore) saveProposals(items []ProposalItem) error {
	return writeJSONAtomic(d.proposalsPath(), items)
}

func (d *diskStore) loadAgentJobs() ([]AgentJob, error) {
	return readJSON(d.agentJobsPath(), []AgentJob{})
}

func (d *diskStore) saveAgentJobs(items []AgentJob) error {
	return writeJSONAtomic(d.agentJobsPath(), items)
}

func (d *diskStore) loadObservations() ([]ObservationRecord, error) {
	return readJSON(d.observationsPath(), []ObservationRecord{})
}

func (d *diskStore) saveObservations(items []ObservationRecord) error {
	return writeJSONAtomic(d.observationsPath(), items)
}

func (d *diskStore) loadWorkItems() ([]WorkItem, error) {
	return readJSON(d.workItemsPath(), []WorkItem{})
}

func (d *diskStore) saveWorkItems(items []WorkItem) error {
	return writeJSONAtomic(d.workItemsPath(), items)
}

func (d *diskStore) loadFlows() ([]FlowRun, error) {
	return readJSON(d.flowsPath(), []FlowRun{})
}

func (d *diskStore) saveFlows(items []FlowRun) error {
	return writeJSONAtomic(d.flowsPath(), items)
}

func (d *diskStore) loadApprovalItems() ([]ApprovalItem, error) {
	return readJSON(d.approvalItemsPath(), []ApprovalItem{})
}

func (d *diskStore) saveApprovalItems(items []ApprovalItem) error {
	return writeJSONAtomic(d.approvalItemsPath(), items)
}

func (d *diskStore) loadReleases() ([]ReleasePacket, error) {
	return readJSON(d.releasesPath(), []ReleasePacket{})
}

func (d *diskStore) saveReleases(items []ReleasePacket) error {
	return writeJSONAtomic(d.releasesPath(), items)
}

func (d *diskStore) loadDeploys() ([]DeployRun, error) {
	return readJSON(d.deploysPath(), []DeployRun{})
}

func (d *diskStore) saveDeploys(items []DeployRun) error {
	return writeJSONAtomic(d.deploysPath(), items)
}

func (d *diskStore) loadRollbacks() ([]RollbackRun, error) {
	return readJSON(d.rollbacksPath(), []RollbackRun{})
}

func (d *diskStore) saveRollbacks(items []RollbackRun) error {
	return writeJSONAtomic(d.rollbacksPath(), items)
}

func (d *diskStore) loadTargets() ([]DeployTarget, error) {
	return readJSON(d.targetsPath(), []DeployTarget{})
}

func (d *diskStore) saveTargets(items []DeployTarget) error {
	return writeJSONAtomic(d.targetsPath(), items)
}

func (d *diskStore) loadPreflights() ([]PreflightRecord, error) {
	return readJSON(d.preflightsPath(), []PreflightRecord{})
}

func (d *diskStore) savePreflights(items []PreflightRecord) error {
	return writeJSONAtomic(d.preflightsPath(), items)
}

func (d *diskStore) loadPolicies() ([]PolicyDecision, error) {
	return readJSON(d.policiesPath(), []PolicyDecision{})
}

func (d *diskStore) savePolicies(items []PolicyDecision) error {
	return writeJSONAtomic(d.policiesPath(), items)
}

func (d *diskStore) loadGatewayCalls() ([]GatewayCall, error) {
	return readJSON(d.gatewayCallsPath(), []GatewayCall{})
}

func (d *diskStore) saveGatewayCalls(items []GatewayCall) error {
	return writeJSONAtomic(d.gatewayCallsPath(), items)
}

func (d *diskStore) loadExecutes() ([]ExecuteRun, error) {
	return readJSON(d.executesPath(), []ExecuteRun{})
}

func (d *diskStore) saveExecutes(items []ExecuteRun) error {
	return writeJSONAtomic(d.executesPath(), items)
}

func (d *diskStore) loadVerifications() ([]VerificationRecord, error) {
	return readJSON(d.verificationsPath(), []VerificationRecord{})
}

func (d *diskStore) saveVerifications(items []VerificationRecord) error {
	return writeJSONAtomic(d.verificationsPath(), items)
}

func (d *diskStore) loadSystemMemory() (SystemMemory, error) {
	return readJSON(d.systemMemoryPath(), defaultSystemMemory())
}

func (d *diskStore) saveSystemMemory(memory SystemMemory) error {
	return writeJSONAtomic(d.systemMemoryPath(), memory)
}

func (d *diskStore) loadContinuityRecords() ([]ContinuityRecord, error) {
	return readJSON(d.continuityRecordsPath(), []ContinuityRecord{})
}

func (d *diskStore) saveContinuityRecords(items []ContinuityRecord) error {
	return writeJSONAtomic(d.continuityRecordsPath(), normalizeContinuityRecordsAt(items, zeroSafeNow()))
}

func (d *diskStore) loadSessionCheckpoint() (*SessionCheckpoint, error) {
	raw, err := os.ReadFile(d.sessionCheckpointPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var item SessionCheckpoint
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	item = normalizeSessionCheckpoint(item)
	if err := item.Validate(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *diskStore) saveSessionCheckpoint(item SessionCheckpoint) error {
	item = normalizeSessionCheckpoint(item)
	if err := item.Validate(); err != nil {
		return err
	}
	return writeJSONAtomic(d.sessionCheckpointPath(), item)
}

func (d *diskStore) clearSessionCheckpoint() error {
	err := os.Remove(d.sessionCheckpointPath())
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (d *diskStore) statePath() string {
	return filepath.Join(d.baseDir, "company_state.json")
}

func (d *diskStore) branchesPath() string {
	return filepath.Join(d.baseDir, "branches.json")
}

func (d *diskStore) proposalsPath() string {
	return filepath.Join(d.baseDir, "proposals.json")
}

func (d *diskStore) agentJobsPath() string {
	return filepath.Join(d.baseDir, "agent_jobs.json")
}

func (d *diskStore) observationsPath() string {
	return filepath.Join(d.baseDir, "observations.json")
}

func (d *diskStore) workItemsPath() string {
	return filepath.Join(d.baseDir, "work_items.json")
}

func (d *diskStore) flowsPath() string {
	return filepath.Join(d.baseDir, "flows.json")
}

func (d *diskStore) approvalItemsPath() string {
	return filepath.Join(d.baseDir, "approval_inbox.json")
}

func (d *diskStore) systemMemoryPath() string {
	return filepath.Join(d.baseDir, "system_memory.json")
}

func (d *diskStore) continuityRecordsPath() string {
	return filepath.Join(d.baseDir, "continuity_records.json")
}

func (d *diskStore) sessionCheckpointPath() string {
	return filepath.Join(d.baseDir, "session_checkpoint.json")
}

func (d *diskStore) releasesPath() string {
	return filepath.Join(d.baseDir, "releases.json")
}

func (d *diskStore) deploysPath() string {
	return filepath.Join(d.baseDir, "deploys.json")
}

func (d *diskStore) rollbacksPath() string {
	return filepath.Join(d.baseDir, "rollbacks.json")
}

func (d *diskStore) targetsPath() string {
	return filepath.Join(d.baseDir, "deploy_targets.json")
}

func (d *diskStore) preflightsPath() string {
	return filepath.Join(d.baseDir, "preflights.json")
}

func (d *diskStore) policiesPath() string {
	return filepath.Join(d.baseDir, "policies.json")
}

func (d *diskStore) gatewayCallsPath() string {
	return filepath.Join(d.baseDir, "gateway_calls.json")
}

func (d *diskStore) executesPath() string {
	return filepath.Join(d.baseDir, "executes.json")
}

func (d *diskStore) verificationsPath() string {
	return filepath.Join(d.baseDir, "verifications.json")
}

func (d *diskStore) openThreadsPath() string {
	return filepath.Join(d.baseDir, "open_threads.json")
}

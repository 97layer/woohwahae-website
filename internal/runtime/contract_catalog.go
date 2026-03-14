package runtime

import "sort"

type ContractTier string

const (
	ContractTierSpine   ContractTier = "spine"
	ContractTierHarness ContractTier = "harness"
)

type ContractSpec struct {
	Schema string       `json:"schema"`
	Title  string       `json:"title"`
	Tier   ContractTier `json:"tier"`
}

var canonicalContractCatalog = []ContractSpec{
	{Schema: "company_state.schema.json", Title: "CompanyState", Tier: ContractTierSpine},
	{Schema: "proposal_item.schema.json", Title: "ProposalItem", Tier: ContractTierSpine},
	{Schema: "work_item.schema.json", Title: "WorkItem", Tier: ContractTierSpine},
	{Schema: "approval_item.schema.json", Title: "ApprovalItem", Tier: ContractTierSpine},
	{Schema: "release_packet.schema.json", Title: "ReleasePacket", Tier: ContractTierSpine},
	{Schema: "deploy_run.schema.json", Title: "DeployRun", Tier: ContractTierSpine},
	{Schema: "deploy_target.schema.json", Title: "DeployTarget", Tier: ContractTierSpine},
	{Schema: "event_envelope.schema.json", Title: "EventEnvelope", Tier: ContractTierSpine},
	{Schema: "system_memory.schema.json", Title: "SystemMemory", Tier: ContractTierSpine},
	{Schema: "auth_status.schema.json", Title: "AuthStatus", Tier: ContractTierSpine},
	{Schema: "preflight_record.schema.json", Title: "PreflightRecord", Tier: ContractTierSpine},
	{Schema: "policy_decision.schema.json", Title: "PolicyDecision", Tier: ContractTierSpine},
	{Schema: "gateway_call.schema.json", Title: "GatewayCall", Tier: ContractTierSpine},
	{Schema: "execute_run.schema.json", Title: "ExecuteRun", Tier: ContractTierSpine},
	{Schema: "verification_record.schema.json", Title: "VerificationRecord", Tier: ContractTierSpine},
	{Schema: "rollback_run.schema.json", Title: "RollbackRun", Tier: ContractTierSpine},
	{Schema: "flow_run.schema.json", Title: "FlowRun", Tier: ContractTierSpine},
	{Schema: "founder_view.schema.json", Title: "FounderView", Tier: ContractTierSpine},
	{Schema: "founder_summary.schema.json", Title: "FounderSummary", Tier: ContractTierSpine},
	{Schema: "review_room.schema.json", Title: "ReviewRoom", Tier: ContractTierSpine},
	{Schema: "review_room_summary.schema.json", Title: "ReviewRoomSummary", Tier: ContractTierSpine},
	{Schema: "adapter_summary.schema.json", Title: "AdapterSummary", Tier: ContractTierSpine},
	{Schema: "capability_registry.schema.json", Title: "CapabilityRegistry", Tier: ContractTierSpine},
	{Schema: "knowledge_packet.schema.json", Title: "KnowledgePacket", Tier: ContractTierSpine},
	{Schema: "session_bootstrap_packet.schema.json", Title: "SessionBootstrapPacket", Tier: ContractTierSpine},
	{Schema: "snapshot_packet.schema.json", Title: "SnapshotPacket", Tier: ContractTierSpine},
	{Schema: "write_lease.schema.json", Title: "WriteLease", Tier: ContractTierSpine},
	{Schema: "handoff_packet.schema.json", Title: "HandoffPacket", Tier: ContractTierSpine},
	{Schema: "daemon_status.schema.json", Title: "DaemonStatus", Tier: ContractTierSpine},
	{Schema: "chain_rules.schema.json", Title: "ChainRules", Tier: ContractTierSpine},
	{Schema: "branch.schema.json", Title: "Branch", Tier: ContractTierHarness},
	{Schema: "agent_job.schema.json", Title: "AgentJob", Tier: ContractTierHarness},
	{Schema: "agent_dispatch_profile.schema.json", Title: "AgentDispatchProfile", Tier: ContractTierHarness},
	{Schema: "agent_dispatch_result.schema.json", Title: "AgentDispatchResult", Tier: ContractTierHarness},
	{Schema: "agent_job_report_result.schema.json", Title: "AgentJobReportResult", Tier: ContractTierHarness},
	{Schema: "agent_run_packet.schema.json", Title: "AgentRunPacket", Tier: ContractTierHarness},
	{Schema: "observation_record.schema.json", Title: "ObservationRecord", Tier: ContractTierHarness},
	{Schema: "open_thread.schema.json", Title: "OpenThread", Tier: ContractTierHarness},
	{Schema: "provider_summary.schema.json", Title: "ProviderSummary", Tier: ContractTierHarness},
	{Schema: "telegram_packet.schema.json", Title: "TelegramPacket", Tier: ContractTierHarness},
	{Schema: "authority_boundary.schema.json", Title: "AuthorityBoundary", Tier: ContractTierHarness},
	{Schema: "capitalization_entry.schema.json", Title: "CapitalizationEntry", Tier: ContractTierHarness},
}

func CanonicalContractCatalog() []ContractSpec {
	catalog := make([]ContractSpec, len(canonicalContractCatalog))
	copy(catalog, canonicalContractCatalog)
	return catalog
}

func CanonicalContractFilenames() []string {
	files := make([]string, 0, len(canonicalContractCatalog))
	for _, spec := range canonicalContractCatalog {
		files = append(files, spec.Schema)
	}
	sort.Strings(files)
	return files
}

func ContractTitlesByTier(tier ContractTier) []string {
	titles := []string{}
	for _, spec := range canonicalContractCatalog {
		if spec.Tier == tier {
			titles = append(titles, spec.Title)
		}
	}
	return titles
}

func contractTitleBySchema(schema string) (string, bool) {
	for _, spec := range canonicalContractCatalog {
		if spec.Schema == schema {
			return spec.Title, true
		}
	}
	return "", false
}

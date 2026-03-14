package runtime

import (
	"errors"
	"strings"
	"time"
)

type WriteLease struct {
	Owner      string     `json:"owner"`
	PID        int        `json:"pid"`
	AcquiredAt *time.Time `json:"acquired_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	Status     string     `json:"status"`
}

type HandoffCounts struct {
	Branches      int `json:"branches"`
	Proposals     int `json:"proposals"`
	AgentJobs     int `json:"agent_jobs"`
	WorkItems     int `json:"work_items"`
	Flows         int `json:"flows"`
	Approvals     int `json:"approvals"`
	Releases      int `json:"releases"`
	Events        int `json:"events"`
	Deploys       int `json:"deploys"`
	Rollbacks     int `json:"rollbacks"`
	Targets       int `json:"targets"`
	Preflights    int `json:"preflights"`
	Policies      int `json:"policies"`
	GatewayCalls  int `json:"gateway_calls"`
	Executes      int `json:"executes"`
	Verifications int `json:"verifications"`
	Observations  int `json:"observations"`
}

type ParallelCandidate struct {
	Text     string  `json:"text"`
	Kind     string  `json:"kind"`
	Source   string  `json:"source"`
	Reason   string  `json:"reason"`
	Priority string  `json:"priority"`
	Ref      *string `json:"ref,omitempty"`
}

type PromptingContract struct {
	Directive          string   `json:"directive"`
	CognitionMode      string   `json:"cognition_mode"`
	DecisionScope      string   `json:"decision_scope"`
	AutonomyBudget     string   `json:"autonomy_budget"`
	MutationPolicy     string   `json:"mutation_policy"`
	EscalationTriggers []string `json:"escalation_triggers"`
	OpenQuestions      []string `json:"open_questions"`
}

type HandoffSummary struct {
	GeneratedAt        time.Time           `json:"generated_at"`
	CurrentFocus       string              `json:"current_focus"`
	NextSteps          []string            `json:"next_steps"`
	OpenRisks          []string            `json:"open_risks"`
	ReviewTopOpen      []ReviewRoomItem    `json:"review_top_open"`
	ActiveBranches     []Branch            `json:"active_branches"`
	ParallelCandidates []ParallelCandidate `json:"parallel_candidates"`
	Auth               AuthStatus          `json:"auth"`
	WriteLease         WriteLease          `json:"write_lease"`
	Counts             HandoffCounts       `json:"counts"`
}

type AgentRunPacket struct {
	GeneratedAt    time.Time            `json:"generated_at"`
	Source         string               `json:"source"`
	Job            AgentJob             `json:"job"`
	Runtime        AgentRuntimeContract `json:"runtime"`
	Knowledge      KnowledgePacket      `json:"knowledge"`
	Prompting      *PromptingContract   `json:"prompting,omitempty"`
	Handoff        *HandoffPacket       `json:"handoff,omitempty"`
	HandoffSummary *HandoffSummary      `json:"handoff_summary,omitempty"`
	Proposal       *ProposalItem        `json:"proposal,omitempty"`
}

type AgentRuntimeContract struct {
	SourceOfTruth     string   `json:"source_of_truth"`
	DispatchTransport string   `json:"dispatch_transport"`
	ReportPath        string   `json:"report_path"`
	ReportCommand     string   `json:"report_command"`
	TerminalStatuses  []string `json:"terminal_statuses"`
	WriteAuthRequired bool     `json:"write_auth_required"`
	WriteTokenEnv     *string  `json:"write_token_env,omitempty"`
}

type HandoffPacket struct {
	GeneratedAt        time.Time           `json:"generated_at"`
	CompanyState       CompanyState        `json:"company_state"`
	SystemMemory       SystemMemory        `json:"system_memory"`
	Auth               AuthStatus          `json:"auth"`
	WriteLease         WriteLease          `json:"write_lease"`
	FounderView        FounderView         `json:"founder_view"`
	FounderSummary     FounderSummary      `json:"founder_summary"`
	ReviewRoom         ReviewRoomSummary   `json:"review_room"`
	Continuity         *ContinuityView     `json:"continuity,omitempty"`
	ActiveBranches     []Branch            `json:"active_branches"`
	ParallelCandidates []ParallelCandidate `json:"parallel_candidates"`
	Counts             HandoffCounts       `json:"counts"`
}

type SessionBootstrapPacket struct {
	GeneratedAt  time.Time           `json:"generated_at"`
	Source       string              `json:"source"`
	ReadOnly     bool                `json:"read_only"`
	Degraded     bool                `json:"degraded"`
	Tooling      *ToolingHealth      `json:"tooling,omitempty"`
	Knowledge    KnowledgePacket     `json:"knowledge"`
	Prompting    *PromptingContract  `json:"prompting,omitempty"`
	Continuity   *ContinuityView     `json:"continuity,omitempty"`
	Resume       *SessionCheckpoint  `json:"resume,omitempty"`
	Handoff      *HandoffPacket      `json:"handoff,omitempty"`
	ReviewRoom   *ReviewRoomSummary  `json:"review_room,omitempty"`
	Capabilities *CapabilityRegistry `json:"capabilities,omitempty"`
}

type SnapshotPacket struct {
	GeneratedAt       time.Time            `json:"generated_at"`
	CompanyState      CompanyState         `json:"company_state"`
	SystemMemory      SystemMemory         `json:"system_memory"`
	ContinuityRecords []ContinuityRecord   `json:"continuity_records"`
	SessionCheckpoint *SessionCheckpoint   `json:"session_checkpoint,omitempty"`
	Auth              AuthStatus           `json:"auth"`
	Branches          []Branch             `json:"branches"`
	Proposals         []ProposalItem       `json:"proposals"`
	AgentJobs         []AgentJob           `json:"agent_jobs"`
	WorkItems         []WorkItem           `json:"work_items"`
	Flows             []FlowRun            `json:"flows"`
	Approvals         []ApprovalItem       `json:"approvals"`
	Releases          []ReleasePacket      `json:"releases"`
	Deploys           []DeployRun          `json:"deploys"`
	Rollbacks         []RollbackRun        `json:"rollbacks"`
	Targets           []DeployTarget       `json:"targets"`
	Events            []EventEnvelope      `json:"events"`
	Observations      []ObservationRecord  `json:"observations"`
	ReviewRoom        ReviewRoom           `json:"review_room"`
	Preflights        []PreflightRecord    `json:"preflights"`
	Policies          []PolicyDecision     `json:"policies"`
	GatewayCalls      []GatewayCall        `json:"gateway_calls"`
	Executes          []ExecuteRun         `json:"executes"`
	Verifications     []VerificationRecord `json:"verifications"`
}

func (p AgentRunPacket) Validate() error {
	if p.GeneratedAt.IsZero() {
		return errors.New("agent run packet generated_at is required")
	}
	if strings.TrimSpace(p.Source) == "" {
		return errors.New("agent run packet source is required")
	}
	if err := p.Job.Validate(); err != nil {
		return err
	}
	if err := p.Runtime.Validate(); err != nil {
		return err
	}
	if err := p.Knowledge.Validate(); err != nil {
		return err
	}
	if p.Prompting != nil {
		if err := p.Prompting.Validate(); err != nil {
			return err
		}
	}
	if p.Handoff == nil && p.HandoffSummary == nil {
		return errors.New("agent run packet handoff or handoff_summary is required")
	}
	if p.Handoff != nil {
		if err := p.Handoff.Validate(); err != nil {
			return err
		}
	}
	if p.HandoffSummary != nil {
		if err := p.HandoffSummary.Validate(); err != nil {
			return err
		}
	}
	if p.Proposal != nil {
		if err := p.Proposal.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p PromptingContract) Validate() error {
	if strings.TrimSpace(p.Directive) == "" {
		return errors.New("prompting directive is required")
	}
	if !validPromptingCognitionMode(p.CognitionMode) {
		return errors.New("prompting cognition_mode is invalid")
	}
	if !validPromptingDecisionScope(p.DecisionScope) {
		return errors.New("prompting decision_scope is invalid")
	}
	if !validPromptingAutonomyBudget(p.AutonomyBudget) {
		return errors.New("prompting autonomy_budget is invalid")
	}
	if !validPromptingMutationPolicy(p.MutationPolicy) {
		return errors.New("prompting mutation_policy is invalid")
	}
	if p.EscalationTriggers == nil {
		return errors.New("prompting escalation_triggers are required")
	}
	for _, item := range p.EscalationTriggers {
		if strings.TrimSpace(item) == "" {
			return errors.New("prompting escalation_triggers must not contain empty items")
		}
	}
	if p.OpenQuestions == nil {
		return errors.New("prompting open_questions are required")
	}
	for _, item := range p.OpenQuestions {
		if strings.TrimSpace(item) == "" {
			return errors.New("prompting open_questions must not contain empty items")
		}
	}
	return nil
}

func (h HandoffSummary) Validate() error {
	if h.GeneratedAt.IsZero() {
		return errors.New("handoff summary generated_at is required")
	}
	if strings.TrimSpace(h.CurrentFocus) == "" {
		return errors.New("handoff summary current_focus is required")
	}
	if h.NextSteps == nil {
		return errors.New("handoff summary next_steps is required")
	}
	for _, item := range h.NextSteps {
		if strings.TrimSpace(item) == "" {
			return errors.New("handoff summary next_steps must not contain empty items")
		}
	}
	if h.OpenRisks == nil {
		return errors.New("handoff summary open_risks is required")
	}
	for _, item := range h.OpenRisks {
		if strings.TrimSpace(item) == "" {
			return errors.New("handoff summary open_risks must not contain empty items")
		}
	}
	if h.ReviewTopOpen == nil {
		return errors.New("handoff summary review_top_open is required")
	}
	for _, item := range h.ReviewTopOpen {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if h.ActiveBranches == nil {
		return errors.New("handoff summary active_branches is required")
	}
	for _, item := range h.ActiveBranches {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if h.ParallelCandidates == nil {
		return errors.New("handoff summary parallel_candidates is required")
	}
	for _, item := range h.ParallelCandidates {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p ParallelCandidate) Validate() error {
	if strings.TrimSpace(p.Text) == "" {
		return errors.New("parallel candidate text is required")
	}
	if strings.TrimSpace(p.Kind) == "" {
		return errors.New("parallel candidate kind is required")
	}
	if strings.TrimSpace(p.Source) == "" {
		return errors.New("parallel candidate source is required")
	}
	if strings.TrimSpace(p.Reason) == "" {
		return errors.New("parallel candidate reason is required")
	}
	if strings.TrimSpace(p.Priority) == "" {
		return errors.New("parallel candidate priority is required")
	}
	return nil
}

func validPromptingCognitionMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case "staff_advisor", "executor":
		return true
	default:
		return false
	}
}

func validPromptingDecisionScope(scope string) bool {
	switch strings.TrimSpace(scope) {
	case "bounded", "full":
		return true
	default:
		return false
	}
}

func validPromptingAutonomyBudget(budget string) bool {
	switch strings.TrimSpace(budget) {
	case "manual_only", "single_step", "multi_step":
		return true
	default:
		return false
	}
}

func validPromptingMutationPolicy(policy string) bool {
	switch strings.TrimSpace(policy) {
	case "read_only", "scoped_write", "full_write":
		return true
	default:
		return false
	}
}

func (c HandoffCounts) Validate() error {
	for _, item := range []struct {
		name  string
		value int
	}{
		{name: "branches", value: c.Branches},
		{name: "proposals", value: c.Proposals},
		{name: "agent_jobs", value: c.AgentJobs},
		{name: "work_items", value: c.WorkItems},
		{name: "flows", value: c.Flows},
		{name: "approvals", value: c.Approvals},
		{name: "releases", value: c.Releases},
		{name: "events", value: c.Events},
		{name: "deploys", value: c.Deploys},
		{name: "rollbacks", value: c.Rollbacks},
		{name: "targets", value: c.Targets},
		{name: "preflights", value: c.Preflights},
		{name: "policies", value: c.Policies},
		{name: "gateway_calls", value: c.GatewayCalls},
		{name: "executes", value: c.Executes},
		{name: "verifications", value: c.Verifications},
		{name: "observations", value: c.Observations},
	} {
		if item.value < 0 {
			return errors.New("handoff counts " + item.name + " must be non-negative")
		}
	}
	return nil
}

func (h HandoffPacket) Validate() error {
	if h.GeneratedAt.IsZero() {
		return errors.New("handoff packet generated_at is required")
	}
	if err := h.CompanyState.Validate(); err != nil {
		return err
	}
	if err := h.SystemMemory.Validate(); err != nil {
		return err
	}
	if h.FounderView.Now == nil || h.FounderView.Waiting == nil || h.FounderView.Risk == nil {
		return errors.New("handoff packet founder_view lanes are required")
	}
	if err := h.ReviewRoom.Validate(); err != nil {
		return err
	}
	if h.Continuity != nil {
		if err := h.Continuity.Validate(); err != nil {
			return err
		}
	}
	if h.ActiveBranches == nil {
		return errors.New("handoff packet active_branches is required")
	}
	for _, item := range h.ActiveBranches {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if h.ParallelCandidates == nil {
		return errors.New("handoff packet parallel_candidates is required")
	}
	for _, item := range h.ParallelCandidates {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return h.Counts.Validate()
}

func (r AgentRuntimeContract) Validate() error {
	if strings.TrimSpace(r.SourceOfTruth) == "" {
		return errors.New("agent runtime contract source_of_truth is required")
	}
	if strings.TrimSpace(r.DispatchTransport) == "" {
		return errors.New("agent runtime contract dispatch_transport is required")
	}
	if strings.TrimSpace(r.ReportPath) == "" {
		return errors.New("agent runtime contract report_path is required")
	}
	if strings.TrimSpace(r.ReportCommand) == "" {
		return errors.New("agent runtime contract report_command is required")
	}
	if len(r.TerminalStatuses) == 0 {
		return errors.New("agent runtime contract terminal_statuses is required")
	}
	for _, status := range r.TerminalStatuses {
		if !validAgentJobTerminalStatus(status) {
			return errors.New("agent runtime contract terminal_status is invalid")
		}
	}
	if r.WriteTokenEnv != nil && strings.TrimSpace(*r.WriteTokenEnv) == "" {
		return errors.New("agent runtime contract write_token_env is invalid")
	}
	return nil
}

func (p SessionBootstrapPacket) Validate() error {
	if p.GeneratedAt.IsZero() {
		return errors.New("session bootstrap generated_at is required")
	}
	if strings.TrimSpace(p.Source) == "" {
		return errors.New("session bootstrap source is required")
	}
	if err := p.Knowledge.Validate(); err != nil {
		return err
	}
	if p.Prompting != nil {
		if err := p.Prompting.Validate(); err != nil {
			return err
		}
	}
	if p.Tooling != nil {
		if err := p.Tooling.Validate(); err != nil {
			return err
		}
	}
	if p.Continuity != nil {
		if err := p.Continuity.Validate(); err != nil {
			return err
		}
	}
	if p.Resume != nil {
		if err := p.Resume.Validate(); err != nil {
			return err
		}
	}
	if p.Handoff != nil {
		if err := p.Handoff.Validate(); err != nil {
			return err
		}
	}
	if p.ReviewRoom != nil {
		if err := p.ReviewRoom.Validate(); err != nil {
			return err
		}
	}
	if p.Capabilities != nil {
		if err := p.Capabilities.Validate(); err != nil {
			return err
		}
	}
	return nil
}

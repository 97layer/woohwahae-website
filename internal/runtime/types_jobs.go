package runtime

import (
	"errors"
	"strings"
	"time"
)

type ProposalItem struct {
	ProposalID         string    `json:"proposal_id"`
	BranchID           *string   `json:"branch_id,omitempty"`
	Title              string    `json:"title"`
	Intent             string    `json:"intent"`
	Summary            string    `json:"summary"`
	Surface            Surface   `json:"surface"`
	Priority           string    `json:"priority"`
	Risk               string    `json:"risk"`
	Status             string    `json:"status"`
	Notes              []string  `json:"notes"`
	PromotedWorkItemID *string   `json:"promoted_work_item_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AgentJob struct {
	JobID     string         `json:"job_id"`
	BranchID  *string        `json:"branch_id,omitempty"`
	Kind      string         `json:"kind"`
	Role      string         `json:"role"`
	Summary   string         `json:"summary"`
	Status    string         `json:"status"`
	Source    string         `json:"source"`
	Surface   Surface        `json:"surface"`
	Stage     Stage          `json:"stage"`
	Ref       *string        `json:"ref,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Result    map[string]any `json:"result,omitempty"`
	Notes     []string       `json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type AgentDispatchResult struct {
	Job     AgentJob       `json:"job"`
	Policy  PolicyDecision `json:"policy"`
	Gateway GatewayCall    `json:"gateway"`
}

type AgentJobFollowUp struct {
	Mode    string   `json:"mode"`
	Summary string   `json:"summary"`
	JobIDs  []string `json:"job_ids,omitempty"`
}

type AgentJobReportResult struct {
	Job            AgentJob             `json:"job"`
	Event          EventEnvelope        `json:"event"`
	FollowUp       AgentJobFollowUp     `json:"follow_up"`
	Capitalization *CapitalizationEntry `json:"capitalization,omitempty"`
	Chain          *AgentJobChainResult `json:"chain,omitempty"`
	Warnings       []string             `json:"warnings,omitempty"`
}

type AgentJobPromotionItem struct {
	SourceKind string               `json:"source_kind"`
	SourceID   string               `json:"source_id"`
	Summary    string               `json:"summary"`
	Status     string               `json:"status"`
	Reason     string               `json:"reason,omitempty"`
	Job        *AgentJob            `json:"job,omitempty"`
	Dispatch   *AgentDispatchResult `json:"dispatch,omitempty"`
}

type AgentJobPromotionResult struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Considered  int                     `json:"considered"`
	Created     int                     `json:"created"`
	Existing    int                     `json:"existing"`
	Dispatched  int                     `json:"dispatched"`
	Failed      int                     `json:"failed"`
	Items       []AgentJobPromotionItem `json:"items"`
}

type AgentDispatchProfile struct {
	Role          string   `json:"role"`
	Provider      string   `json:"provider"`
	Model         string   `json:"model"`
	Risk          string   `json:"risk"`
	Novelty       string   `json:"novelty"`
	TokenClass    string   `json:"token_class"`
	TokenBudget   int      `json:"token_budget"`
	DispatchReady bool     `json:"dispatch_ready"`
	Notes         []string `json:"notes"`
}

type ChainRules struct {
	Rules []ChainRule `json:"rules"`
}

type ChainRule struct {
	RuleID         string         `json:"rule_id"`
	OnStatus       string         `json:"on_status"`
	NextKind       string         `json:"next_kind"`
	NextRole       string         `json:"next_role"`
	NextStage      Stage          `json:"next_stage"`
	Surface        Surface        `json:"surface"`
	Summary        string         `json:"summary"`
	AutoDispatch   bool           `json:"auto_dispatch"`
	NotifyTelegram bool           `json:"notify_telegram"`
	Payload        map[string]any `json:"payload,omitempty"`
	Notes          []string       `json:"notes"`
}

func validProposalStatus(status string) bool {
	switch status {
	case "proposed", "promoted", "deferred":
		return true
	default:
		return false
	}
}

func validAgentJobStatus(status string) bool {
	switch status {
	case "queued", "running", "succeeded", "failed", "canceled":
		return true
	default:
		return false
	}
}

func validAgentJobTerminalStatus(status string) bool {
	switch status {
	case "succeeded", "failed", "canceled":
		return true
	default:
		return false
	}
}

func (p ProposalItem) Validate() error {
	if strings.TrimSpace(p.ProposalID) == "" {
		return errors.New("proposal item proposal_id is required")
	}
	if p.BranchID != nil && strings.TrimSpace(*p.BranchID) == "" {
		return errors.New("proposal item branch_id must not be empty")
	}
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("proposal item title is required")
	}
	if strings.TrimSpace(p.Intent) == "" {
		return errors.New("proposal item intent is required")
	}
	if strings.TrimSpace(p.Summary) == "" {
		return errors.New("proposal item summary is required")
	}
	if !validSurface(p.Surface) {
		return errors.New("proposal item surface is invalid")
	}
	if !validProposalStatus(p.Status) {
		return errors.New("proposal item status is invalid")
	}
	if p.Notes == nil {
		return errors.New("proposal item notes is required")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("proposal item created_at is required")
	}
	if p.UpdatedAt.IsZero() {
		return errors.New("proposal item updated_at is required")
	}
	return nil
}

func (r AgentJobReportResult) Validate() error {
	if err := r.Job.Validate(); err != nil {
		return err
	}
	if err := r.Event.Validate(); err != nil {
		return err
	}
	if err := r.FollowUp.Validate(); err != nil {
		return err
	}
	if r.Capitalization != nil {
		if strings.TrimSpace(r.Capitalization.EntryID) == "" {
			return errors.New("agent job report capitalization entry_id is required")
		}
		if r.Capitalization.CreatedAt.IsZero() {
			return errors.New("agent job report capitalization created_at is required")
		}
	}
	if r.Chain != nil {
		if err := r.Chain.Validate(); err != nil {
			return err
		}
	}
	for _, item := range r.Warnings {
		if strings.TrimSpace(item) == "" {
			return errors.New("agent job report warnings must not contain empty items")
		}
	}
	return nil
}

func (f AgentJobFollowUp) Validate() error {
	if strings.TrimSpace(f.Mode) == "" {
		return errors.New("agent job follow_up mode is required")
	}
	if strings.TrimSpace(f.Summary) == "" {
		return errors.New("agent job follow_up summary is required")
	}
	for _, jobID := range f.JobIDs {
		if strings.TrimSpace(jobID) == "" {
			return errors.New("agent job follow_up job_ids must not contain empty items")
		}
	}
	return nil
}

func (r AgentJobChainResult) Validate() error {
	if strings.TrimSpace(r.ParentJobID) == "" {
		return errors.New("agent job chain parent_job_id is required")
	}
	if r.Considered < 0 {
		return errors.New("agent job chain considered must be non-negative")
	}
	if r.Created < 0 {
		return errors.New("agent job chain created must be non-negative")
	}
	if r.Existing < 0 {
		return errors.New("agent job chain existing must be non-negative")
	}
	if r.Dispatched < 0 {
		return errors.New("agent job chain dispatched must be non-negative")
	}
	if r.TelegramRequested < 0 {
		return errors.New("agent job chain telegram_requested must be non-negative")
	}
	for _, item := range r.CreatedJobIDs {
		if strings.TrimSpace(item) == "" {
			return errors.New("agent job chain created_job_ids must not contain empty items")
		}
	}
	for _, item := range r.ExistingJobIDs {
		if strings.TrimSpace(item) == "" {
			return errors.New("agent job chain existing_job_ids must not contain empty items")
		}
	}
	return nil
}

func (j AgentJob) Validate() error {
	if strings.TrimSpace(j.JobID) == "" {
		return errors.New("agent job job_id is required")
	}
	if j.BranchID != nil && strings.TrimSpace(*j.BranchID) == "" {
		return errors.New("agent job branch_id must not be empty")
	}
	if strings.TrimSpace(j.Kind) == "" {
		return errors.New("agent job kind is required")
	}
	if strings.TrimSpace(j.Role) == "" {
		return errors.New("agent job role is required")
	}
	if strings.TrimSpace(j.Summary) == "" {
		return errors.New("agent job summary is required")
	}
	if !validAgentJobStatus(j.Status) {
		return errors.New("agent job status is invalid")
	}
	if strings.TrimSpace(j.Source) == "" {
		return errors.New("agent job source is required")
	}
	if !validSurface(j.Surface) {
		return errors.New("agent job surface is invalid")
	}
	if !validStage(j.Stage) {
		return errors.New("agent job stage is invalid")
	}
	if j.Ref != nil && strings.TrimSpace(*j.Ref) == "" {
		return errors.New("agent job ref must not be empty")
	}
	if j.Payload != nil {
		if err := validateJSONObject(j.Payload, "agent job payload"); err != nil {
			return err
		}
	}
	if j.Result != nil {
		if err := validateJSONObject(j.Result, "agent job result"); err != nil {
			return err
		}
	}
	if j.Notes == nil {
		return errors.New("agent job notes are required")
	}
	if j.CreatedAt.IsZero() {
		return errors.New("agent job created_at is required")
	}
	if j.UpdatedAt.IsZero() {
		return errors.New("agent job updated_at is required")
	}
	return nil
}

func (c ChainRules) Validate() error {
	if len(c.Rules) == 0 {
		return errors.New("chain rules are required")
	}
	for _, item := range c.Rules {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c ChainRule) Validate() error {
	if strings.TrimSpace(c.RuleID) == "" {
		return errors.New("chain rule rule_id is required")
	}
	if !validAgentJobTerminalStatus(c.OnStatus) {
		return errors.New("chain rule on_status is invalid")
	}
	if strings.TrimSpace(c.NextKind) == "" {
		return errors.New("chain rule next_kind is required")
	}
	if strings.TrimSpace(c.NextRole) == "" {
		return errors.New("chain rule next_role is required")
	}
	if c.NextStage != "" && !validStage(c.NextStage) {
		return errors.New("chain rule next_stage is invalid")
	}
	if c.Surface != "" && !validSurface(c.Surface) {
		return errors.New("chain rule surface is invalid")
	}
	if strings.TrimSpace(c.Summary) == "" {
		return errors.New("chain rule summary is required")
	}
	if c.Payload != nil {
		if err := validateJSONObject(c.Payload, "chain rule payload"); err != nil {
			return err
		}
	}
	return nil
}

package runtime

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

type WorkItem struct {
	ID               string         `json:"id"`
	BranchID         *string        `json:"branch_id,omitempty"`
	Title            string         `json:"title"`
	Intent           string         `json:"intent"`
	Stage            Stage          `json:"stage"`
	Surface          Surface        `json:"surface"`
	Pack             string         `json:"pack"`
	Priority         string         `json:"priority"`
	Risk             string         `json:"risk"`
	RequiresApproval bool           `json:"requires_approval"`
	Payload          map[string]any `json:"payload"`
	CorrelationID    string         `json:"correlation_id"`
	CreatedAt        time.Time      `json:"created_at"`
}

type Branch struct {
	BranchID            string     `json:"branch_id"`
	ParentBranchID      *string    `json:"parent_branch_id,omitempty"`
	RootBranchID        string     `json:"root_branch_id"`
	Title               string     `json:"title"`
	Intent              string     `json:"intent"`
	Summary             string     `json:"summary"`
	Stage               Stage      `json:"stage"`
	Surface             Surface    `json:"surface"`
	Status              string     `json:"status"`
	BasisRef            *string    `json:"basis_ref,omitempty"`
	MergeTargetBranchID *string    `json:"merge_target_branch_id,omitempty"`
	MergedAt            *time.Time `json:"merged_at,omitempty"`
	Notes               []string   `json:"notes"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type ApprovalItem struct {
	ApprovalID      string     `json:"approval_id"`
	WorkItemID      string     `json:"work_item_id"`
	Stage           Stage      `json:"stage"`
	Summary         string     `json:"summary"`
	Risks           []string   `json:"risks"`
	RollbackPlan    string     `json:"rollback_plan"`
	DecisionSurface Surface    `json:"decision_surface"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requested_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

type ReleasePacket struct {
	ReleaseID    string         `json:"release_id"`
	WorkItemID   string         `json:"work_item_id"`
	Target       string         `json:"target"`
	Channel      string         `json:"channel"`
	Artifacts    []string       `json:"artifacts"`
	Metrics      map[string]any `json:"metrics"`
	RollbackPlan string         `json:"rollback_plan"`
	ApprovalRefs []string       `json:"approval_refs"`
	ReleasedAt   *time.Time     `json:"released_at,omitempty"`
}

type DeployRun struct {
	DeployID   string     `json:"deploy_id"`
	ReleaseID  string     `json:"release_id"`
	Target     string     `json:"target"`
	Status     string     `json:"status"`
	Notes      []string   `json:"notes"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type RollbackRun struct {
	RollbackID string     `json:"rollback_id"`
	ReleaseID  string     `json:"release_id"`
	DeployID   *string    `json:"deploy_id,omitempty"`
	Target     string     `json:"target"`
	Status     string     `json:"status"`
	Notes      []string   `json:"notes"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type FlowRun struct {
	FlowID           string    `json:"flow_id"`
	WorkItemID       string    `json:"work_item_id"`
	ApprovalID       *string   `json:"approval_id,omitempty"`
	PolicyDecisionID *string   `json:"policy_decision_id,omitempty"`
	ExecuteID        *string   `json:"execute_id,omitempty"`
	VerificationID   *string   `json:"verification_id,omitempty"`
	ReleaseID        *string   `json:"release_id,omitempty"`
	DeployID         *string   `json:"deploy_id,omitempty"`
	RollbackID       *string   `json:"rollback_id,omitempty"`
	Status           string    `json:"status"`
	Notes            []string  `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DeployTarget struct {
	TargetID string   `json:"target_id"`
	Command  []string `json:"command"`
	Workdir  *string  `json:"workdir,omitempty"`
}

type PreflightRecord struct {
	RecordID    string    `json:"record_id"`
	Task        string    `json:"task"`
	Mode        string    `json:"mode"`
	Status      string    `json:"status"`
	Decision    string    `json:"decision"`
	ModelsUsed  []string  `json:"models_used"`
	Steps       []string  `json:"steps"`
	Risks       []string  `json:"risks"`
	Checks      []string  `json:"checks"`
	GeneratedAt time.Time `json:"generated_at"`
}

type PolicyDecision struct {
	DecisionID       string    `json:"decision_id"`
	Intent           string    `json:"intent"`
	Scope            string    `json:"scope"`
	Risk             string    `json:"risk"`
	Novelty          string    `json:"novelty"`
	TokenClass       string    `json:"token_class"`
	RequiresApproval bool      `json:"requires_approval"`
	Mode             string    `json:"mode"`
	Decision         string    `json:"decision"`
	Reasons          []string  `json:"reasons"`
	CreatedAt        time.Time `json:"created_at"`
}

type GatewayCall struct {
	CallID          string     `json:"call_id"`
	DecisionID      string     `json:"decision_id"`
	Provider        string     `json:"provider"`
	Model           string     `json:"model"`
	RequestKind     string     `json:"request_kind"`
	Status          string     `json:"status"`
	AttemptCount    int        `json:"attempt_count"`
	LastHTTPStatus  *int       `json:"last_http_status,omitempty"`
	LastError       *string    `json:"last_error,omitempty"`
	ResponsePreview *string    `json:"response_preview,omitempty"`
	DispatchedAt    *time.Time `json:"dispatched_at,omitempty"`
	TokenBudget     int        `json:"token_budget"`
	Notes           []string   `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ExecuteRun struct {
	ExecuteID        string     `json:"execute_id"`
	WorkItemID       string     `json:"work_item_id"`
	PolicyDecisionID string     `json:"policy_decision_id"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	Notes            []string   `json:"notes"`
	StartedAt        time.Time  `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
}

type VerificationRecord struct {
	RecordID   string     `json:"record_id"`
	Scope      string     `json:"scope"`
	Command    []string   `json:"command"`
	Status     string     `json:"status"`
	Notes      []string   `json:"notes"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func validBranchStatus(status string) bool {
	switch status {
	case "active", "merged", "archived":
		return true
	default:
		return false
	}
}

func validPreflightStatus(status string) bool {
	switch status {
	case "ready", "degraded-lite", "needs_clarification":
		return true
	default:
		return false
	}
}

func validPreflightDecision(decision string) bool {
	switch decision {
	case "go", "hold":
		return true
	default:
		return false
	}
}

func validPolicyScale(value string) bool {
	switch value {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func validTokenClass(value string) bool {
	switch value {
	case "tiny", "small", "medium", "large":
		return true
	default:
		return false
	}
}

func validPolicyMode(value string) bool {
	switch canonicalPolicyMode(value) {
	case "local", "single", "blocked":
		return true
	default:
		return false
	}
}

func validGatewayStatus(value string) bool {
	switch value {
	case "recorded", "sent", "failed":
		return true
	default:
		return false
	}
}

func validExecuteStatus(value string) bool {
	switch value {
	case "recorded", "succeeded", "failed":
		return true
	default:
		return false
	}
}

func validVerificationStatus(status string) bool {
	switch status {
	case "passed", "failed":
		return true
	default:
		return false
	}
}

func validRollbackStatus(status string) bool {
	switch status {
	case "recorded", "succeeded", "failed":
		return true
	default:
		return false
	}
}

func validFlowStatus(status string) bool {
	switch status {
	case "active", "waiting", "released", "rolled_back", "blocked":
		return true
	default:
		return false
	}
}

func validApprovalStatus(status string) bool {
	switch status {
	case "pending", "approved", "rejected":
		return true
	default:
		return false
	}
}

func validDeployStatus(status string) bool {
	switch status {
	case "succeeded", "failed":
		return true
	default:
		return false
	}
}

func (w WorkItem) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.New("work item id is required")
	}
	if w.BranchID != nil && strings.TrimSpace(*w.BranchID) == "" {
		return errors.New("work item branch_id must not be empty")
	}
	if strings.TrimSpace(w.Title) == "" {
		return errors.New("work item title is required")
	}
	if strings.TrimSpace(w.Intent) == "" {
		return errors.New("work item intent is required")
	}
	if !validStage(w.Stage) {
		return errors.New("work item stage is invalid")
	}
	if !validSurface(w.Surface) {
		return errors.New("work item surface is invalid")
	}
	if w.CreatedAt.IsZero() {
		return errors.New("work item created_at is required")
	}
	if w.Payload == nil {
		return errors.New("work item payload is required")
	}
	if err := validateJSONObject(w.Payload, "work item payload"); err != nil {
		return err
	}
	return nil
}

func (b Branch) Validate() error {
	if strings.TrimSpace(b.BranchID) == "" {
		return errors.New("branch branch_id is required")
	}
	if b.ParentBranchID != nil && strings.TrimSpace(*b.ParentBranchID) == "" {
		return errors.New("branch parent_branch_id must not be empty")
	}
	if strings.TrimSpace(b.RootBranchID) == "" {
		return errors.New("branch root_branch_id is required")
	}
	if strings.TrimSpace(b.Title) == "" {
		return errors.New("branch title is required")
	}
	if strings.TrimSpace(b.Intent) == "" {
		return errors.New("branch intent is required")
	}
	if strings.TrimSpace(b.Summary) == "" {
		return errors.New("branch summary is required")
	}
	if !validStage(b.Stage) {
		return errors.New("branch stage is invalid")
	}
	if !validSurface(b.Surface) {
		return errors.New("branch surface is invalid")
	}
	if !validBranchStatus(b.Status) {
		return errors.New("branch status is invalid")
	}
	if b.BasisRef != nil && strings.TrimSpace(*b.BasisRef) == "" {
		return errors.New("branch basis_ref must not be empty")
	}
	if b.MergeTargetBranchID != nil && strings.TrimSpace(*b.MergeTargetBranchID) == "" {
		return errors.New("branch merge_target_branch_id must not be empty")
	}
	if b.Status == "merged" {
		if b.MergeTargetBranchID == nil || strings.TrimSpace(*b.MergeTargetBranchID) == "" {
			return errors.New("branch merge_target_branch_id is required for merged branch")
		}
		if b.MergedAt == nil || b.MergedAt.IsZero() {
			return errors.New("branch merged_at is required for merged branch")
		}
	}
	if b.Notes == nil {
		return errors.New("branch notes are required")
	}
	if b.CreatedAt.IsZero() {
		return errors.New("branch created_at is required")
	}
	if b.UpdatedAt.IsZero() {
		return errors.New("branch updated_at is required")
	}
	return nil
}

func (a ApprovalItem) Validate() error {
	if strings.TrimSpace(a.ApprovalID) == "" {
		return errors.New("approval_id is required")
	}
	if strings.TrimSpace(a.WorkItemID) == "" {
		return errors.New("work_item_id is required")
	}
	if !validStage(a.Stage) {
		return errors.New("approval stage is invalid")
	}
	if !validSurface(a.DecisionSurface) {
		return errors.New("decision_surface is invalid")
	}
	if strings.TrimSpace(a.Summary) == "" {
		return errors.New("approval summary is required")
	}
	if !validApprovalStatus(a.Status) {
		return errors.New("approval status is invalid")
	}
	if a.RequestedAt.IsZero() {
		return errors.New("requested_at is required")
	}
	if (a.Status == "approved" || a.Status == "rejected") && (a.ResolvedAt == nil || a.ResolvedAt.IsZero()) {
		return errors.New("resolved_at is required for resolved approval")
	}
	return nil
}

func (r ReleasePacket) Validate() error {
	if strings.TrimSpace(r.ReleaseID) == "" {
		return errors.New("release_id is required")
	}
	if strings.TrimSpace(r.WorkItemID) == "" {
		return errors.New("work_item_id is required")
	}
	if strings.TrimSpace(r.Target) == "" {
		return errors.New("release target is required")
	}
	if strings.TrimSpace(r.Channel) == "" {
		return errors.New("release channel is required")
	}
	if r.Artifacts == nil {
		return errors.New("release artifacts are required")
	}
	if r.Metrics == nil {
		return errors.New("release metrics are required")
	}
	if err := validateJSONObject(r.Metrics, "release metrics"); err != nil {
		return err
	}
	if r.ApprovalRefs == nil {
		return errors.New("release approval_refs are required")
	}
	if r.ReleasedAt == nil || r.ReleasedAt.IsZero() {
		return errors.New("released_at is required")
	}
	return nil
}

func (d DeployRun) Validate() error {
	if strings.TrimSpace(d.DeployID) == "" {
		return errors.New("deploy_id is required")
	}
	if strings.TrimSpace(d.ReleaseID) == "" {
		return errors.New("release_id is required")
	}
	if strings.TrimSpace(d.Target) == "" {
		return errors.New("deploy target is required")
	}
	if !validDeployStatus(d.Status) {
		return errors.New("deploy status is invalid")
	}
	if d.Notes == nil {
		return errors.New("deploy notes are required")
	}
	if d.StartedAt.IsZero() {
		return errors.New("deploy started_at is required")
	}
	if (d.Status == "succeeded" || d.Status == "failed") && (d.FinishedAt == nil || d.FinishedAt.IsZero()) {
		return errors.New("deploy finished_at is required for completed deploy")
	}
	return nil
}

func (d DeployTarget) Validate() error {
	if strings.TrimSpace(d.TargetID) == "" {
		return errors.New("target_id is required")
	}
	if len(d.Command) == 0 {
		return errors.New("deploy target command is required")
	}
	for _, part := range d.Command {
		if strings.TrimSpace(part) == "" {
			return errors.New("deploy target command parts must not be empty")
		}
	}
	if !filepath.IsAbs(strings.TrimSpace(d.Command[0])) {
		return errors.New("deploy target command must start with an absolute binary path")
	}
	if d.Workdir != nil && strings.TrimSpace(*d.Workdir) != "" && !filepath.IsAbs(strings.TrimSpace(*d.Workdir)) {
		return errors.New("deploy target workdir must be an absolute path")
	}
	return nil
}

func (f FlowRun) Validate() error {
	if strings.TrimSpace(f.FlowID) == "" {
		return errors.New("flow_id is required")
	}
	if strings.TrimSpace(f.WorkItemID) == "" {
		return errors.New("flow work_item_id is required")
	}
	if !validFlowStatus(f.Status) {
		return errors.New("flow status is invalid")
	}
	if f.Notes == nil {
		return errors.New("flow notes are required")
	}
	if f.CreatedAt.IsZero() {
		return errors.New("flow created_at is required")
	}
	if f.UpdatedAt.IsZero() {
		return errors.New("flow updated_at is required")
	}
	return nil
}

func (r RollbackRun) Validate() error {
	if strings.TrimSpace(r.RollbackID) == "" {
		return errors.New("rollback_id is required")
	}
	if strings.TrimSpace(r.ReleaseID) == "" {
		return errors.New("release_id is required")
	}
	if strings.TrimSpace(r.Target) == "" {
		return errors.New("rollback target is required")
	}
	if !validRollbackStatus(r.Status) {
		return errors.New("rollback status is invalid")
	}
	if r.Notes == nil {
		return errors.New("rollback notes are required")
	}
	if r.StartedAt.IsZero() {
		return errors.New("rollback started_at is required")
	}
	if (r.Status == "succeeded" || r.Status == "failed") && (r.FinishedAt == nil || r.FinishedAt.IsZero()) {
		return errors.New("rollback finished_at is required for completed rollback")
	}
	return nil
}

func (p PreflightRecord) Validate() error {
	if strings.TrimSpace(p.RecordID) == "" {
		return errors.New("record_id is required")
	}
	if strings.TrimSpace(p.Task) == "" {
		return errors.New("preflight task is required")
	}
	if strings.TrimSpace(p.Mode) == "" {
		return errors.New("preflight mode is required")
	}
	if !validPreflightStatus(p.Status) {
		return errors.New("preflight status is invalid")
	}
	if !validPreflightDecision(p.Decision) {
		return errors.New("preflight decision is invalid")
	}
	if p.ModelsUsed == nil {
		return errors.New("preflight models_used is required")
	}
	if p.Steps == nil {
		return errors.New("preflight steps are required")
	}
	if p.Risks == nil {
		return errors.New("preflight risks are required")
	}
	if p.Checks == nil {
		return errors.New("preflight checks are required")
	}
	if p.GeneratedAt.IsZero() {
		return errors.New("preflight generated_at is required")
	}
	return nil
}

func (p PolicyDecision) Validate() error {
	if strings.TrimSpace(p.DecisionID) == "" {
		return errors.New("policy decision_id is required")
	}
	if strings.TrimSpace(p.Intent) == "" {
		return errors.New("policy intent is required")
	}
	if strings.TrimSpace(p.Scope) == "" {
		return errors.New("policy scope is required")
	}
	if !validPolicyScale(p.Risk) {
		return errors.New("policy risk is invalid")
	}
	if !validPolicyScale(p.Novelty) {
		return errors.New("policy novelty is invalid")
	}
	if !validTokenClass(p.TokenClass) {
		return errors.New("policy token_class is invalid")
	}
	if !validPolicyMode(p.Mode) {
		return errors.New("policy mode is invalid")
	}
	if !validPreflightDecision(p.Decision) {
		return errors.New("policy decision is invalid")
	}
	if p.Reasons == nil {
		return errors.New("policy reasons are required")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("policy created_at is required")
	}
	return nil
}

func (g GatewayCall) Validate() error {
	if strings.TrimSpace(g.CallID) == "" {
		return errors.New("gateway call_id is required")
	}
	if strings.TrimSpace(g.DecisionID) == "" {
		return errors.New("gateway decision_id is required")
	}
	if strings.TrimSpace(g.Provider) == "" {
		return errors.New("gateway provider is required")
	}
	if strings.TrimSpace(g.Model) == "" {
		return errors.New("gateway model is required")
	}
	if strings.TrimSpace(g.RequestKind) == "" {
		return errors.New("gateway request_kind is required")
	}
	if !validGatewayStatus(g.Status) {
		return errors.New("gateway status is invalid")
	}
	if g.AttemptCount < 0 {
		return errors.New("gateway attempt_count must not be negative")
	}
	if g.TokenBudget < 0 {
		return errors.New("gateway token_budget must not be negative")
	}
	if g.Notes == nil {
		return errors.New("gateway notes are required")
	}
	if g.CreatedAt.IsZero() {
		return errors.New("gateway created_at is required")
	}
	return nil
}

func (e ExecuteRun) Validate() error {
	if strings.TrimSpace(e.ExecuteID) == "" {
		return errors.New("execute_id is required")
	}
	if strings.TrimSpace(e.WorkItemID) == "" {
		return errors.New("execute work_item_id is required")
	}
	if strings.TrimSpace(e.PolicyDecisionID) == "" {
		return errors.New("execute policy_decision_id is required")
	}
	if !validPolicyMode(e.Mode) {
		return errors.New("execute mode is invalid")
	}
	if !validExecuteStatus(e.Status) {
		return errors.New("execute status is invalid")
	}
	if e.Notes == nil {
		return errors.New("execute notes are required")
	}
	if e.StartedAt.IsZero() {
		return errors.New("execute started_at is required")
	}
	if (e.Status == "succeeded" || e.Status == "failed") && (e.FinishedAt == nil || e.FinishedAt.IsZero()) {
		return errors.New("execute finished_at is required for completed execute")
	}
	return nil
}

func (v VerificationRecord) Validate() error {
	if strings.TrimSpace(v.RecordID) == "" {
		return errors.New("verification record_id is required")
	}
	if strings.TrimSpace(v.Scope) == "" {
		return errors.New("verification scope is required")
	}
	if len(v.Command) == 0 {
		return errors.New("verification command is required")
	}
	for _, part := range v.Command {
		if strings.TrimSpace(part) == "" {
			return errors.New("verification command parts must not be empty")
		}
	}
	if !filepath.IsAbs(strings.TrimSpace(v.Command[0])) {
		return errors.New("verification executable must be an absolute path")
	}
	if !validVerificationStatus(v.Status) {
		return errors.New("verification status is invalid")
	}
	if v.Notes == nil {
		return errors.New("verification notes are required")
	}
	if v.StartedAt.IsZero() {
		return errors.New("verification started_at is required")
	}
	if (v.Status == "passed" || v.Status == "failed") && (v.FinishedAt == nil || v.FinishedAt.IsZero()) {
		return errors.New("verification finished_at is required for completed verification")
	}
	return nil
}

package runtime

import (
	"errors"
	"strings"
	"time"
)

type FounderItem struct {
	Kind       string    `json:"kind"`
	Ref        string    `json:"ref"`
	WorkItemID string    `json:"work_item_id"`
	Summary    string    `json:"summary"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FounderView struct {
	Now          []FounderItem `json:"now"`
	Waiting      []FounderItem `json:"waiting"`
	Risk         []FounderItem `json:"risk"`
	LastRelease  *FounderItem  `json:"last_release,omitempty"`
	LastRollback *FounderItem  `json:"last_rollback,omitempty"`
}

type PriorityRationale struct {
	Lane   string `json:"lane"`
	Source string `json:"source"`
	Reason string `json:"reason"`
	Rule   string `json:"rule"`
}

type FounderSummary struct {
	EnvironmentAdvisory EnvironmentAdvisory `json:"environment_advisory"`
	PrimaryAction       string              `json:"primary_action"`
	PrimaryRef          string              `json:"primary_ref"`
	PriorityRationale   *PriorityRationale  `json:"priority_rationale,omitempty"`
	NowCount            int                 `json:"now_count"`
	WaitingCount        int                 `json:"waiting_count"`
	RiskCount           int                 `json:"risk_count"`
	ReviewOpenCount     int                 `json:"review_open_count"`
	ReviewTopOpen       []string            `json:"review_top_open"`
	LastChange          *FounderItem        `json:"last_change,omitempty"`
	LastRelease         *FounderItem        `json:"last_release,omitempty"`
	LastRollback        *FounderItem        `json:"last_rollback,omitempty"`
}

type CompanyState struct {
	ShellMode        string             `json:"shell_mode"`
	ApprovalsPending int                `json:"approvals_pending"`
	WorkItemsActive  int                `json:"work_items_active"`
	LastReleaseAt    *time.Time         `json:"last_release_at,omitempty"`
	MemoryHealth     string             `json:"memory_health"`
	DeployHealth     string             `json:"deploy_health"`
	PrimarySurface   Surface            `json:"primary_surface"`
	ActiveSurfaces   []Surface          `json:"active_surfaces"`
	Progress         *ProgressDashboard `json:"progress,omitempty"`
}

type ProgressDashboard struct {
	OverallPercent int            `json:"overall_percent"`
	OverallBar     string         `json:"overall_bar"`
	OverallStatus  string         `json:"overall_status"`
	Axes           []ProgressAxis `json:"axes"`
}

type ProgressAxis struct {
	AxisID  string   `json:"axis_id"`
	Label   string   `json:"label"`
	Percent int      `json:"percent"`
	Bar     string   `json:"bar"`
	Status  string   `json:"status"`
	Signals []string `json:"signals"`
}

type DaemonArchitectStatus struct {
	Enabled        bool       `json:"enabled"`
	AutoDispatch   bool       `json:"auto_dispatch"`
	Interval       string     `json:"interval"`
	PromoteLimit   int        `json:"promote_limit"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	LastError      *string    `json:"last_error,omitempty"`
	LastIdle       bool       `json:"last_idle"`
	LastConsidered int        `json:"last_considered"`
	LastCreated    int        `json:"last_created"`
	LastExisting   int        `json:"last_existing"`
	LastDispatched int        `json:"last_dispatched"`
	LastFailed     int        `json:"last_failed"`
}

type DaemonStatus struct {
	Service         string                 `json:"service"`
	Status          string                 `json:"status"`
	Address         string                 `json:"address"`
	StartedAt       time.Time              `json:"started_at"`
	UptimeSeconds   int64                  `json:"uptime_seconds"`
	MemoryHealth    string                 `json:"memory_health"`
	DeployHealth    string                 `json:"deploy_health"`
	DegradedReasons []string               `json:"degraded_reasons"`
	Architect       *DaemonArchitectStatus `json:"architect,omitempty"`
}

func validDaemonStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "ok", "degraded":
		return true
	default:
		return false
	}
}

func (c CompanyState) Validate() error {
	if strings.TrimSpace(c.ShellMode) == "" {
		return errors.New("company state shell_mode is required")
	}
	if c.ApprovalsPending < 0 {
		return errors.New("company state approvals_pending must be non-negative")
	}
	if c.WorkItemsActive < 0 {
		return errors.New("company state work_items_active must be non-negative")
	}
	if strings.TrimSpace(c.MemoryHealth) == "" {
		return errors.New("company state memory_health is required")
	}
	if strings.TrimSpace(c.DeployHealth) == "" {
		return errors.New("company state deploy_health is required")
	}
	if !validSurface(c.PrimarySurface) {
		return errors.New("company state primary_surface is invalid")
	}
	for _, surface := range c.ActiveSurfaces {
		if !validSurface(surface) {
			return errors.New("company state active_surfaces contains invalid surface")
		}
	}
	if c.Progress != nil {
		if err := c.Progress.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d ProgressDashboard) Validate() error {
	if d.OverallPercent < 0 || d.OverallPercent > 100 {
		return errors.New("progress dashboard overall_percent must be between 0 and 100")
	}
	if strings.TrimSpace(d.OverallBar) == "" {
		return errors.New("progress dashboard overall_bar is required")
	}
	if strings.TrimSpace(d.OverallStatus) == "" {
		return errors.New("progress dashboard overall_status is required")
	}
	if d.Axes == nil {
		return errors.New("progress dashboard axes is required")
	}
	for _, axis := range d.Axes {
		if err := axis.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ProgressAxis) Validate() error {
	if strings.TrimSpace(a.AxisID) == "" {
		return errors.New("progress axis axis_id is required")
	}
	if strings.TrimSpace(a.Label) == "" {
		return errors.New("progress axis label is required")
	}
	if a.Percent < 0 || a.Percent > 100 {
		return errors.New("progress axis percent must be between 0 and 100")
	}
	if strings.TrimSpace(a.Bar) == "" {
		return errors.New("progress axis bar is required")
	}
	if strings.TrimSpace(a.Status) == "" {
		return errors.New("progress axis status is required")
	}
	if a.Signals == nil {
		return errors.New("progress axis signals are required")
	}
	return nil
}

func (d DaemonArchitectStatus) Validate() error {
	if strings.TrimSpace(d.Interval) == "" {
		return errors.New("daemon architect status interval is required")
	}
	if d.PromoteLimit < 0 {
		return errors.New("daemon architect status promote_limit must be non-negative")
	}
	if d.LastError != nil && strings.TrimSpace(*d.LastError) == "" {
		return errors.New("daemon architect status last_error must not be empty")
	}
	if d.LastConsidered < 0 {
		return errors.New("daemon architect status last_considered must be non-negative")
	}
	if d.LastCreated < 0 {
		return errors.New("daemon architect status last_created must be non-negative")
	}
	if d.LastExisting < 0 {
		return errors.New("daemon architect status last_existing must be non-negative")
	}
	if d.LastDispatched < 0 {
		return errors.New("daemon architect status last_dispatched must be non-negative")
	}
	if d.LastFailed < 0 {
		return errors.New("daemon architect status last_failed must be non-negative")
	}
	return nil
}

func (d DaemonStatus) Validate() error {
	if strings.TrimSpace(d.Service) == "" {
		return errors.New("daemon status service is required")
	}
	if !validDaemonStatus(d.Status) {
		return errors.New("daemon status status is invalid")
	}
	if strings.TrimSpace(d.Address) == "" {
		return errors.New("daemon status address is required")
	}
	if d.StartedAt.IsZero() {
		return errors.New("daemon status started_at is required")
	}
	if d.UptimeSeconds < 0 {
		return errors.New("daemon status uptime_seconds must be non-negative")
	}
	if strings.TrimSpace(d.MemoryHealth) == "" {
		return errors.New("daemon status memory_health is required")
	}
	if strings.TrimSpace(d.DeployHealth) == "" {
		return errors.New("daemon status deploy_health is required")
	}
	if d.DegradedReasons == nil {
		return errors.New("daemon status degraded_reasons is required")
	}
	if d.Architect != nil {
		if err := d.Architect.Validate(); err != nil {
			return err
		}
	}
	return nil
}

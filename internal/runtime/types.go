package runtime

import (
	"errors"
	"strings"
	"time"
)

type Stage string

const (
	StageDiscover   Stage = "discover"
	StageCompose    Stage = "compose"
	StageExperience Stage = "experience"
	StageVerify     Stage = "verify"
	StageRelease    Stage = "release"
)

type Surface string

const (
	SurfaceCockpit  Surface = "cockpit"
	SurfaceTelegram Surface = "telegram"
	SurfaceAPI      Surface = "api"
)

type SystemMemory struct {
	CurrentFocus string    `json:"current_focus"`
	CurrentGoal  *string   `json:"current_goal,omitempty"`
	NextSteps    []string  `json:"next_steps"`
	OpenRisks    []string  `json:"open_risks"`
	HandoffNote  *string   `json:"handoff_note,omitempty"`
	LastOperator *string   `json:"last_operator,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthStatus struct {
	WriteAuthEnabled bool       `json:"write_auth_enabled"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type EventEnvelope struct {
	EventID    string         `json:"event_id"`
	Kind       string         `json:"kind"`
	Actor      string         `json:"actor"`
	Surface    Surface        `json:"surface"`
	WorkItemID string         `json:"work_item_id"`
	Stage      Stage          `json:"stage"`
	Timestamp  time.Time      `json:"timestamp"`
	Data       map[string]any `json:"data"`
}

type EventCreateInput struct {
	Kind       string         `json:"kind"`
	Surface    Surface        `json:"surface"`
	WorkItemID string         `json:"work_item_id"`
	Stage      Stage          `json:"stage"`
	Data       map[string]any `json:"data"`
}

type SessionFinishInput struct {
	CurrentFocus string   `json:"current_focus"`
	CurrentGoal  *string  `json:"current_goal,omitempty"`
	NextSteps    []string `json:"next_steps"`
	OpenRisks    []string `json:"open_risks"`
	HandoffNote  *string  `json:"handoff_note,omitempty"`
	Note         *string  `json:"note,omitempty"`
}

type SessionFinishResult struct {
	Memory SystemMemory  `json:"memory"`
	Event  EventEnvelope `json:"event"`
}

func validStage(stage Stage) bool {
	switch stage {
	case StageDiscover, StageCompose, StageExperience, StageVerify, StageRelease:
		return true
	default:
		return false
	}
}

func validSurface(surface Surface) bool {
	switch surface {
	case SurfaceCockpit, SurfaceTelegram, SurfaceAPI:
		return true
	default:
		return false
	}
}

func (m SystemMemory) Validate() error {
	if strings.TrimSpace(m.CurrentFocus) == "" {
		return errors.New("system memory current_focus is required")
	}
	if m.NextSteps == nil {
		return errors.New("system memory next_steps is required")
	}
	if m.OpenRisks == nil {
		return errors.New("system memory open_risks is required")
	}
	if m.UpdatedAt.IsZero() {
		return errors.New("system memory updated_at is required")
	}
	return nil
}

func (e EventCreateInput) Validate() error {
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("event create input kind is required")
	}
	if !validSurface(e.Surface) {
		return errors.New("event create input surface is invalid")
	}
	if strings.TrimSpace(e.WorkItemID) == "" {
		return errors.New("event create input work_item_id is required")
	}
	if !validStage(e.Stage) {
		return errors.New("event create input stage is invalid")
	}
	if e.Data == nil {
		return errors.New("event create input data is required")
	}
	if err := validateJSONObject(e.Data, "event create input data"); err != nil {
		return err
	}
	return nil
}

func (i SessionFinishInput) Validate() error {
	if strings.TrimSpace(i.CurrentFocus) == "" {
		return errors.New("session finish current_focus is required")
	}
	if i.NextSteps == nil {
		return errors.New("session finish next_steps is required")
	}
	if i.OpenRisks == nil {
		return errors.New("session finish open_risks is required")
	}
	return nil
}

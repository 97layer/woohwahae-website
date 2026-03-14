package runtime

import (
	"errors"
	"strings"
	"time"
)

type CapitalizationFacet struct {
	Summary string   `json:"summary"`
	Items   []string `json:"items"`
}

type CapitalizationEntry struct {
	EntryID       string              `json:"entry_id"`
	CreatedAt     time.Time           `json:"created_at"`
	Actor         string              `json:"actor"`
	SourceEventID string              `json:"source_event_id"`
	SourceKind    string              `json:"source_kind"`
	Situation     CapitalizationFacet `json:"situation"`
	Decision      CapitalizationFacet `json:"decision"`
	Cost          CapitalizationFacet `json:"cost"`
	Result        CapitalizationFacet `json:"result"`
}

type ContinuityNote struct {
	NoteID    string    `json:"note_id"`
	Kind      string    `json:"kind"`
	Text      string    `json:"text"`
	Refs      []string  `json:"refs"`
	CreatedAt time.Time `json:"created_at"`
}

type ContinuityRecord struct {
	RecordID     string           `json:"record_id"`
	Source       string           `json:"source"`
	Actor        string           `json:"actor"`
	Status       string           `json:"status"`
	CurrentFocus string           `json:"current_focus"`
	CurrentGoal  *string          `json:"current_goal,omitempty"`
	NextSteps    []string         `json:"next_steps"`
	OpenRisks    []string         `json:"open_risks"`
	HandoffNote  *string          `json:"handoff_note,omitempty"`
	Refs         []string         `json:"refs"`
	Notes        []ContinuityNote `json:"notes"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	ResolvedAt   *time.Time       `json:"resolved_at,omitempty"`
}

type ContinuitySuggestion struct {
	SuggestionID string  `json:"suggestion_id"`
	Kind         string  `json:"kind"`
	Summary      string  `json:"summary"`
	Ref          *string `json:"ref,omitempty"`
	Command      *string `json:"command,omitempty"`
	Source       string  `json:"source"`
}

type ContinuityView struct {
	Record         *ContinuityRecord      `json:"record,omitempty"`
	RelatedThreads []OpenThread           `json:"related_threads"`
	Suggestions    []ContinuitySuggestion `json:"suggestions"`
}

type SessionCheckpoint struct {
	CheckpointID string    `json:"checkpoint_id"`
	Source       string    `json:"source"`
	Actor        string    `json:"actor"`
	CurrentFocus string    `json:"current_focus"`
	CurrentGoal  *string   `json:"current_goal,omitempty"`
	NextSteps    []string  `json:"next_steps"`
	OpenRisks    []string  `json:"open_risks"`
	HandoffNote  *string   `json:"handoff_note,omitempty"`
	Note         *string   `json:"note,omitempty"`
	Refs         []string  `json:"refs"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SessionCheckpointInput struct {
	Source       string   `json:"source"`
	CurrentFocus string   `json:"current_focus"`
	CurrentGoal  *string  `json:"current_goal,omitempty"`
	NextSteps    []string `json:"next_steps"`
	OpenRisks    []string `json:"open_risks"`
	HandoffNote  *string  `json:"handoff_note,omitempty"`
	Note         *string  `json:"note,omitempty"`
	Refs         []string `json:"refs"`
}

type SessionNoteInput struct {
	Source string   `json:"source"`
	Kind   string   `json:"kind"`
	Text   string   `json:"text"`
	Refs   []string `json:"refs"`
}

type SessionNoteResult struct {
	Record      ContinuityRecord  `json:"record"`
	Note        ContinuityNote    `json:"note"`
	Observation ObservationRecord `json:"observation"`
	Event       EventEnvelope     `json:"event"`
}

func (n ContinuityNote) Validate() error {
	if strings.TrimSpace(n.NoteID) == "" {
		return errors.New("continuity note note_id is required")
	}
	if !validContinuityNoteKind(n.Kind) {
		return errors.New("continuity note kind is invalid")
	}
	if strings.TrimSpace(n.Text) == "" {
		return errors.New("continuity note text is required")
	}
	if n.Refs == nil {
		return errors.New("continuity note refs are required")
	}
	if n.CreatedAt.IsZero() {
		return errors.New("continuity note created_at is required")
	}
	return nil
}

func (c ContinuityRecord) Validate() error {
	if strings.TrimSpace(c.RecordID) == "" {
		return errors.New("continuity record record_id is required")
	}
	if strings.TrimSpace(c.Source) == "" {
		return errors.New("continuity record source is required")
	}
	if strings.TrimSpace(c.Actor) == "" {
		return errors.New("continuity record actor is required")
	}
	if !validContinuityStatus(c.Status) {
		return errors.New("continuity record status is invalid")
	}
	if strings.TrimSpace(c.CurrentFocus) == "" {
		return errors.New("continuity record current_focus is required")
	}
	if c.NextSteps == nil {
		return errors.New("continuity record next_steps is required")
	}
	if c.OpenRisks == nil {
		return errors.New("continuity record open_risks is required")
	}
	if c.Refs == nil {
		return errors.New("continuity record refs are required")
	}
	if c.Notes == nil {
		return errors.New("continuity record notes are required")
	}
	for _, note := range c.Notes {
		if err := note.Validate(); err != nil {
			return err
		}
	}
	if c.CreatedAt.IsZero() {
		return errors.New("continuity record created_at is required")
	}
	if c.UpdatedAt.IsZero() {
		return errors.New("continuity record updated_at is required")
	}
	if c.Status == continuityStatusResolved && (c.ResolvedAt == nil || c.ResolvedAt.IsZero()) {
		return errors.New("continuity record resolved_at is required when resolved")
	}
	return nil
}

func (s ContinuitySuggestion) Validate() error {
	if strings.TrimSpace(s.SuggestionID) == "" {
		return errors.New("continuity suggestion suggestion_id is required")
	}
	if !validContinuitySuggestionKind(s.Kind) {
		return errors.New("continuity suggestion kind is invalid")
	}
	if strings.TrimSpace(s.Summary) == "" {
		return errors.New("continuity suggestion summary is required")
	}
	if strings.TrimSpace(s.Source) == "" {
		return errors.New("continuity suggestion source is required")
	}
	return nil
}

func (c ContinuityView) Validate() error {
	if c.Record != nil {
		if err := c.Record.Validate(); err != nil {
			return err
		}
	}
	if c.RelatedThreads == nil {
		return errors.New("continuity view related_threads is required")
	}
	for _, item := range c.RelatedThreads {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if c.Suggestions == nil {
		return errors.New("continuity view suggestions is required")
	}
	for _, item := range c.Suggestions {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c SessionCheckpoint) Validate() error {
	if strings.TrimSpace(c.CheckpointID) == "" {
		return errors.New("session checkpoint checkpoint_id is required")
	}
	if strings.TrimSpace(c.Source) == "" {
		return errors.New("session checkpoint source is required")
	}
	if strings.TrimSpace(c.Actor) == "" {
		return errors.New("session checkpoint actor is required")
	}
	if strings.TrimSpace(c.CurrentFocus) == "" {
		return errors.New("session checkpoint current_focus is required")
	}
	if c.NextSteps == nil {
		return errors.New("session checkpoint next_steps is required")
	}
	if c.OpenRisks == nil {
		return errors.New("session checkpoint open_risks is required")
	}
	if c.Refs == nil {
		return errors.New("session checkpoint refs are required")
	}
	if c.UpdatedAt.IsZero() {
		return errors.New("session checkpoint updated_at is required")
	}
	return nil
}

func (i SessionCheckpointInput) Validate() error {
	if strings.TrimSpace(i.CurrentFocus) == "" {
		return errors.New("session checkpoint current_focus is required")
	}
	if i.NextSteps == nil {
		return errors.New("session checkpoint next_steps is required")
	}
	if i.OpenRisks == nil {
		return errors.New("session checkpoint open_risks is required")
	}
	if i.Refs == nil {
		return errors.New("session checkpoint refs are required")
	}
	return nil
}

func (i SessionNoteInput) Validate() error {
	if strings.TrimSpace(i.Source) == "" {
		return errors.New("session note source is required")
	}
	if !validContinuityNoteKind(i.Kind) {
		return errors.New("session note kind is invalid")
	}
	if strings.TrimSpace(i.Text) == "" {
		return errors.New("session note text is required")
	}
	if i.Refs == nil {
		return errors.New("session note refs are required")
	}
	return nil
}

func (r SessionNoteResult) Validate() error {
	if err := r.Record.Validate(); err != nil {
		return err
	}
	if err := r.Note.Validate(); err != nil {
		return err
	}
	return r.Observation.Validate()
}

package runtime

import (
	"errors"
	"strings"
	"time"
)

type ReviewRoomRationale struct {
	Trigger string `json:"trigger"`
	Reason  string `json:"reason"`
	Rule    string `json:"rule"`
}

type ReviewRoomResolution struct {
	Action     string    `json:"action"`
	Reason     string    `json:"reason"`
	Rule       string    `json:"rule"`
	Evidence   []string  `json:"evidence,omitempty"`
	ResolvedAt time.Time `json:"resolved_at"`
}

type ReviewRoomItem struct {
	Text           string                `json:"text"`
	Kind           string                `json:"kind"`
	Severity       string                `json:"severity"`
	Source         string                `json:"source"`
	Ref            *string               `json:"ref,omitempty"`
	Why            string                `json:"why,omitempty"`
	WhyUnresolved  *string               `json:"why_unresolved,omitempty"`
	Contradiction  *string               `json:"contradiction,omitempty"`
	Contradictions []string              `json:"contradictions,omitempty"`
	PatternRefs    []string              `json:"pattern_refs,omitempty"`
	Rationale      *ReviewRoomRationale  `json:"rationale,omitempty"`
	Resolution     *ReviewRoomResolution `json:"resolution,omitempty"`
	Evidence       []string              `json:"evidence,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type ReviewRoom struct {
	Source    string           `json:"source"`
	Accepted  []ReviewRoomItem `json:"accepted"`
	Open      []ReviewRoomItem `json:"open"`
	Deferred  []ReviewRoomItem `json:"deferred"`
	Issues    []string         `json:"issues"`
	UpdatedAt *time.Time       `json:"updated_at,omitempty"`
}

type ReviewRoomSummary struct {
	Source        string           `json:"source"`
	AcceptedCount int              `json:"accepted_count"`
	OpenCount     int              `json:"open_count"`
	DeferredCount int              `json:"deferred_count"`
	TopAccepted   []ReviewRoomItem `json:"top_accepted"`
	TopOpen       []ReviewRoomItem `json:"top_open"`
	TopDeferred   []ReviewRoomItem `json:"top_deferred"`
	Issues        []string         `json:"issues"`
	UpdatedAt     *time.Time       `json:"updated_at,omitempty"`
}

func (r ReviewRoomRationale) Validate() error {
	if strings.TrimSpace(r.Trigger) == "" {
		return errors.New("review room rationale trigger is required")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("review room rationale reason is required")
	}
	if strings.TrimSpace(r.Rule) == "" {
		return errors.New("review room rationale rule is required")
	}
	return nil
}

func (r ReviewRoomResolution) Validate() error {
	if normalizeReviewAction(r.Action) == "" {
		return errors.New("review room resolution action is invalid")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("review room resolution reason is required")
	}
	if strings.TrimSpace(r.Rule) == "" {
		return errors.New("review room resolution rule is required")
	}
	if r.ResolvedAt.IsZero() {
		return errors.New("review room resolution resolved_at is required")
	}
	for _, item := range r.Evidence {
		if strings.TrimSpace(item) == "" {
			return errors.New("review room resolution evidence must not contain empty items")
		}
	}
	return nil
}

func (r ReviewRoomItem) Validate() error {
	if strings.TrimSpace(r.Text) == "" {
		return errors.New("review room item text is required")
	}
	if strings.TrimSpace(r.Kind) == "" {
		return errors.New("review room item kind is required")
	}
	if strings.TrimSpace(r.Severity) == "" {
		return errors.New("review room item severity is required")
	}
	if strings.TrimSpace(r.Source) == "" {
		return errors.New("review room item source is required")
	}
	if r.Ref != nil && strings.TrimSpace(*r.Ref) == "" {
		return errors.New("review room item ref must not be empty")
	}
	if r.WhyUnresolved != nil && strings.TrimSpace(*r.WhyUnresolved) == "" {
		return errors.New("review room item why_unresolved must not be empty")
	}
	if r.Contradiction != nil && strings.TrimSpace(*r.Contradiction) == "" {
		return errors.New("review room item contradiction must not be empty")
	}
	for _, item := range r.Contradictions {
		if strings.TrimSpace(item) == "" {
			return errors.New("review room item contradictions must not contain empty items")
		}
	}
	for _, item := range r.PatternRefs {
		if strings.TrimSpace(item) == "" {
			return errors.New("review room item pattern_refs must not contain empty items")
		}
	}
	if r.Rationale != nil {
		if err := r.Rationale.Validate(); err != nil {
			return err
		}
	}
	if r.Resolution != nil {
		if err := r.Resolution.Validate(); err != nil {
			return err
		}
	}
	for _, item := range r.Evidence {
		if strings.TrimSpace(item) == "" {
			return errors.New("review room item evidence must not contain empty items")
		}
	}
	if r.CreatedAt.IsZero() {
		return errors.New("review room item created_at is required")
	}
	if r.UpdatedAt.IsZero() {
		return errors.New("review room item updated_at is required")
	}
	return nil
}

func (r ReviewRoom) Validate() error {
	if strings.TrimSpace(r.Source) == "" {
		return errors.New("review room source is required")
	}
	if r.Accepted == nil {
		return errors.New("review room accepted is required")
	}
	if r.Open == nil {
		return errors.New("review room open is required")
	}
	if r.Deferred == nil {
		return errors.New("review room deferred is required")
	}
	if r.Issues == nil {
		return errors.New("review room issues is required")
	}
	for _, item := range append(append(append([]ReviewRoomItem{}, r.Accepted...), r.Open...), r.Deferred...) {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r ReviewRoomSummary) Validate() error {
	if strings.TrimSpace(r.Source) == "" {
		return errors.New("review room summary source is required")
	}
	if r.AcceptedCount < 0 || r.OpenCount < 0 || r.DeferredCount < 0 {
		return errors.New("review room summary counts must not be negative")
	}
	if r.TopAccepted == nil {
		return errors.New("review room summary top_accepted is required")
	}
	if r.TopOpen == nil {
		return errors.New("review room summary top_open is required")
	}
	if r.TopDeferred == nil {
		return errors.New("review room summary top_deferred is required")
	}
	if r.Issues == nil {
		return errors.New("review room summary issues is required")
	}
	for _, item := range append(append(append([]ReviewRoomItem{}, r.TopAccepted...), r.TopOpen...), r.TopDeferred...) {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

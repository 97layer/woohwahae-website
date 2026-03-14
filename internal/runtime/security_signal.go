package runtime

import (
	"errors"
	"strings"
)

type SecuritySignalInput struct {
	Signal       string
	Summary      string
	Severity     string
	Source       string
	Ref          string
	Evidence     []string
	Data         map[string]any
	Promote      bool
	ReviewReason string
	ReviewRule   string
}

func (s *Service) RecordSecuritySignal(input SecuritySignalInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	input, err := normalizeSecuritySignalInput(input)
	if err != nil {
		return err
	}

	oldEventState := s.event.state()
	oldRoom := s.reviewRoom.current()

	data := cloneJSONObject(input.Data)
	if data == nil {
		data = map[string]any{}
	}
	data["signal"] = input.Signal
	data["summary"] = input.Summary
	data["severity"] = input.Severity
	data["source"] = input.Source
	if input.Ref != "" {
		data["ref"] = input.Ref
	}
	if len(input.Evidence) > 0 {
		data["evidence"] = append([]string{}, input.Evidence...)
	}
	if input.Promote {
		data["review_promoted"] = true
	}
	if err := validateJSONObject(data, "security signal data"); err != nil {
		return err
	}
	if err := s.event.append(s.newSystemEvent("security."+input.Signal, data)); err != nil {
		return err
	}
	if input.Promote {
		ref := input.Ref
		item := newSignalReviewRoomItem(
			input.Summary,
			"security."+input.Signal,
			&ref,
			input.ReviewReason,
			input.ReviewRule,
			input.Evidence,
		)
		updated, _, err := ensureReviewRoomItem(s.reviewRoom.current(), "open", item)
		if err != nil {
			s.event = newEventStoreFromState(oldEventState)
			return err
		}
		s.reviewRoom = newReviewRoomStore(updated)
	}
	if err := s.persistLocked(); err != nil {
		s.event = newEventStoreFromState(oldEventState)
		s.reviewRoom = newReviewRoomStore(oldRoom)
		return err
	}
	return nil
}

func normalizeSecuritySignalInput(input SecuritySignalInput) (SecuritySignalInput, error) {
	input.Signal = normalizeSecuritySignalToken(input.Signal)
	if input.Signal == "" {
		return SecuritySignalInput{}, errors.New("security signal is required")
	}
	input.Summary = strings.TrimSpace(input.Summary)
	if input.Summary == "" {
		return SecuritySignalInput{}, errors.New("security signal summary is required")
	}
	input.Severity = strings.TrimSpace(strings.ToLower(input.Severity))
	if input.Severity == "" {
		input.Severity = "medium"
	}
	input.Source = strings.TrimSpace(input.Source)
	if input.Source == "" {
		input.Source = "runtime.security"
	}
	input.Ref = strings.TrimSpace(input.Ref)
	if input.Ref == "" && input.Promote {
		input.Ref = "security_" + strings.ReplaceAll(input.Signal, ".", "_")
	}
	input.Evidence = normalizeReviewRoomEvidence(input.Evidence)
	input.ReviewReason = strings.TrimSpace(input.ReviewReason)
	if input.ReviewReason == "" && input.Promote {
		input.ReviewReason = "security signal escalation requires founder review before trusting the caller or lane"
	}
	input.ReviewRule = strings.TrimSpace(input.ReviewRule)
	if input.ReviewRule == "" && input.Promote {
		input.ReviewRule = "review_room.auto.security_signal"
	}
	if input.Data == nil {
		input.Data = map[string]any{}
	}
	return input, nil
}

func normalizeSecuritySignalToken(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "_", "-", "_", "/", "_", ":", "_")
	raw = replacer.Replace(raw)
	parts := strings.Split(raw, ".")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "_.")
		if part == "" {
			continue
		}
		normalized = append(normalized, part)
	}
	return strings.Join(normalized, ".")
}

package runtime

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const founderTelegramNoticeDedupWindow = 30 * time.Second

type founderTelegramNoticeState struct {
	mu     sync.Mutex
	key    string
	sentAt time.Time
}

func (s *Service) maybeNotifyFounderAttention() {
	if !telegramFounderAttentionEnabled() {
		return
	}
	if !s.TelegramRouteEnabled(TelegramRouteFounder) {
		return
	}
	packet, key, ok := s.founderAttentionTelegramPacket()
	if !ok || !s.shouldSendFounderTelegramNotice(key) {
		return
	}
	if adapter, ok := s.telegramAdapter.(telegramRouteSender); ok {
		if err := adapter.SendRoute(TelegramRouteFounder, packet); err == nil {
			s.recordFounderTelegramNotice(key)
		}
		return
	}
	if err := s.telegramAdapter.Send(packet); err == nil {
		s.recordFounderTelegramNotice(key)
	}
}

func (s *Service) founderAttentionTelegramPacket() (TelegramPacket, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	founderView := s.deriveFounderViewLocked()
	review := SummarizeReviewRoom(s.currentReviewRoomLocked())
	summary := s.deriveFounderSummaryLocked(founderView, review)
	memory := s.memory.current()

	action := strings.TrimSpace(summary.PrimaryAction)
	if action == "" || action == "idle" {
		return TelegramPacket{}, "", false
	}

	lines := []string{"다음 액션: " + telegramActionLabel(action)}
	if ref := strings.TrimSpace(summary.PrimaryRef); ref != "" {
		lines = append(lines, "기준: "+ref)
	}
	if focus := strings.TrimSpace(memory.CurrentFocus); focus != "" {
		lines = append(lines, "포커스: "+focus)
	}
	for _, item := range limitStrings(summary.ReviewTopOpen, 2) {
		lines = append(lines, "안건: "+item)
	}
	for _, risk := range limitStrings(memory.OpenRisks, 2) {
		lines = append(lines, "리스크: "+risk)
	}
	if notice := strings.TrimSpace(summary.EnvironmentAdvisory.FounderNotice); notice != "" {
		lines = append(lines, "환경: "+notice)
	}

	packet := TelegramPacket{
		GeneratedAt:     zeroSafeNow(),
		Headline:        "지금 먼저 볼 것",
		BodyLines:       lines,
		PrimaryAction:   action,
		PrimaryRef:      strings.TrimSpace(summary.PrimaryRef),
		CurrentFocus:    strings.TrimSpace(memory.CurrentFocus),
		CurrentGoal:     copyFounderGoal(memory.CurrentGoal),
		NextSteps:       limitStrings(memory.NextSteps, 3),
		OpenRisks:       limitStrings(memory.OpenRisks, 3),
		ReviewOpenCount: summary.ReviewOpenCount,
		ReviewTopOpen:   limitStrings(summary.ReviewTopOpen, 3),
		FounderNotice:   strings.TrimSpace(summary.EnvironmentAdvisory.FounderNotice),
		RecommendedMode: strings.TrimSpace(summary.EnvironmentAdvisory.AgentMode),
	}

	key := strings.Join([]string{
		action,
		strings.TrimSpace(summary.PrimaryRef),
		strconv.Itoa(summary.ReviewOpenCount),
		strconv.Itoa(summary.RiskCount),
		strings.Join(limitStrings(memory.OpenRisks, 2), "|"),
	}, "||")
	return packet, key, true
}

func (s *Service) shouldSendFounderTelegramNotice(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	s.founderTelegramNotice.mu.Lock()
	defer s.founderTelegramNotice.mu.Unlock()

	if s.founderTelegramNotice.key == key && !s.founderTelegramNotice.sentAt.IsZero() && zeroSafeNow().Sub(s.founderTelegramNotice.sentAt) < founderTelegramNoticeDedupWindow {
		return false
	}
	return true
}

func (s *Service) recordFounderTelegramNotice(key string) {
	s.founderTelegramNotice.mu.Lock()
	defer s.founderTelegramNotice.mu.Unlock()
	s.founderTelegramNotice.key = strings.TrimSpace(key)
	s.founderTelegramNotice.sentAt = zeroSafeNow()
}

func telegramFounderAttentionEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("LAYER_OS_TELEGRAM_FOUNDER_ALERTS")))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func copyFounderGoal(value *string) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(*value)
	if text == "" {
		return nil
	}
	return &text
}

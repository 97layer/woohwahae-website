package runtime

import (
	"fmt"
	"strings"
)

func telegramHeadline(summary FounderSummary) string {
	if summary.ReviewOpenCount > 0 {
		return "리뷰룸 안건 확인이 필요해"
	}
	if summary.PrimaryAction != "" {
		return "다음 액션: " + telegramActionLabel(summary.PrimaryAction)
	}
	return "Layer OS 텔레그램 요약"
}

func telegramBodyLines(memory SystemMemory, summary FounderSummary, review ReviewRoomSummary) []string {
	lines := []string{}
	if memory.CurrentFocus != "" {
		lines = append(lines, "포커스: "+memory.CurrentFocus)
	}
	for _, item := range topReviewTexts(review.TopOpen, 2) {
		lines = append(lines, "안건: "+item)
	}
	for _, risk := range limitStrings(memory.OpenRisks, 2) {
		lines = append(lines, "리스크: "+risk)
	}
	if notice := summary.EnvironmentAdvisory.FounderNotice; notice != "" {
		lines = append(lines, "환경: "+notice)
	}
	return lines
}

func telegramActionLabel(action string) string {
	switch strings.TrimSpace(action) {
	case "review_room":
		return "리뷰룸 확인"
	case "rollback_or_fix":
		return "롤백 또는 수정 판단"
	case "shape_or_promote":
		return "기획 정리 또는 승격"
	case "dispatch_job":
		return "작업 투입"
	case "approve":
		return "승인 판단"
	case "continue":
		return "진행 이어가기"
	case "job":
		return "작업 확인"
	case "idle":
		return "대기"
	default:
		if action == "" {
			return "대기"
		}
		return action
	}
}

func (s *Service) SendTelegram() error {
	packet := s.Telegram()
	return s.telegramAdapter.Send(packet)
}

func (s *Service) SendTelegramRoute(routeID string, packet TelegramPacket) error {
	if adapter, ok := s.telegramAdapter.(telegramRouteSender); ok {
		return adapter.SendRoute(routeID, packet)
	}
	if routeID == TelegramRouteFounder {
		return s.telegramAdapter.Send(packet)
	}
	return fmt.Errorf("telegram route %s not supported by adapter %s", routeID, s.telegramAdapter.Name())
}

func (s *Service) TelegramRouteEnabled(routeID string) bool {
	if adapter, ok := s.telegramAdapter.(telegramRouteSender); ok {
		return adapter.RouteEnabled(routeID)
	}
	return routeID == TelegramRouteFounder && s.telegramAdapter.Enabled()
}

func (s *Service) TelegramAdapterName() string {
	return s.telegramAdapter.Name()
}

func (s *Service) TelegramEnabled() bool {
	return s.telegramAdapter.Enabled()
}

func (s *Service) Telegram() TelegramPacket {
	s.mu.Lock()
	defer s.mu.Unlock()

	founderView := s.deriveFounderViewLocked()
	fullReviewRoom := s.currentReviewRoomLocked()
	reviewRoom := SummarizeReviewRoom(fullReviewRoom)
	founderSummary := s.deriveFounderSummaryLocked(founderView, reviewRoom)
	memory := s.memory.current()
	return TelegramPacket{
		GeneratedAt:     zeroSafeNow(),
		Headline:        telegramHeadline(founderSummary),
		BodyLines:       telegramBodyLines(memory, founderSummary, reviewRoom),
		PrimaryAction:   founderSummary.PrimaryAction,
		PrimaryRef:      founderSummary.PrimaryRef,
		CurrentFocus:    memory.CurrentFocus,
		CurrentGoal:     memory.CurrentGoal,
		NextSteps:       limitStrings(memory.NextSteps, 3),
		OpenRisks:       limitStrings(memory.OpenRisks, 3),
		ReviewOpenCount: founderSummary.ReviewOpenCount,
		ReviewTopOpen:   topReviewTexts(reviewRoom.TopOpen, 3),
		FounderNotice:   founderSummary.EnvironmentAdvisory.FounderNotice,
		RecommendedMode: founderSummary.EnvironmentAdvisory.AgentMode,
	}
}

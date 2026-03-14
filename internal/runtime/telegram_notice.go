package runtime

import (
	"fmt"
	"strings"
)

func (s *Service) SendTelegramPacket(packet TelegramPacket) error {
	if adapter, ok := s.telegramAdapter.(telegramRouteSender); ok {
		return adapter.SendRoute(TelegramRouteOps, packet)
	}
	return s.telegramAdapter.Send(packet)
}

func jobRequestsDispatchTelegram(job AgentJob) bool {
	if job.Payload == nil {
		return false
	}
	value, ok := job.Payload["notify_telegram"]
	if !ok {
		return false
	}
	switch current := value.(type) {
	case bool:
		return current
	case string:
		switch strings.TrimSpace(strings.ToLower(current)) {
		case "1", "true", "yes", "on":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func jobDispatchTelegramPacket(job AgentJob, gateway GatewayCall) TelegramPacket {
	lines := []string{
		"작업: " + job.JobID,
		"역할: " + job.Role,
		"요약: " + job.Summary,
		"공급자: " + gateway.Provider,
		"모델: " + gateway.Model,
	}
	if job.Result != nil {
		if transport, ok := job.Result["dispatch_transport"].(string); ok && strings.TrimSpace(transport) != "" {
			lines = append(lines, "전송: "+strings.TrimSpace(transport))
		}
		if state, ok := job.Result["dispatch_state"].(string); ok && strings.TrimSpace(state) != "" {
			lines = append(lines, "상태: "+strings.TrimSpace(state))
		}
		if reason, ok := job.Result["dispatch_reason"].(string); ok && strings.TrimSpace(reason) != "" {
			lines = append(lines, "사유: "+strings.TrimSpace(reason))
		}
	}
	if job.Payload != nil {
		if parent, ok := job.Payload["chain_parent_job_id"].(string); ok && strings.TrimSpace(parent) != "" {
			lines = append(lines, "부모 작업: "+strings.TrimSpace(parent))
		}
	}
	return TelegramPacket{
		GeneratedAt:   zeroSafeNow(),
		Headline:      fmt.Sprintf("에이전트 투입: %s", strings.TrimSpace(job.Role)),
		BodyLines:     lines,
		PrimaryAction: "job",
		PrimaryRef:    job.JobID,
		CurrentFocus:  job.Summary,
	}
}

func dispatchTelegramFailureReviewItem(job AgentJob, gateway GatewayCall, err error) ReviewRoomItem {
	ref := job.JobID
	evidence := []string{"job:" + job.JobID, "provider:" + gateway.Provider, "gateway:" + gateway.CallID}
	if err != nil {
		evidence = append(evidence, "error:"+strings.TrimSpace(err.Error()))
	}
	return newSignalReviewRoomItem(
		"에이전트 작업 `"+job.JobID+"`의 텔레그램 알림 전송이 실패했어. 이 알림 레인을 믿기 전에 전송 상태를 확인해야 해.",
		"telegram.dispatch_failed",
		&ref,
		"failed telegram dispatch notice requires review before the alert lane can be trusted",
		"review_room.auto.telegram_dispatch_failed",
		evidence,
	)
}

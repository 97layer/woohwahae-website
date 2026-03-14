package runtime

import "testing"

func TestTelegramPacketReflectsReviewAndMemory(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "Ship official API", NextSteps: []string{"verify daemon"}, OpenRisks: []string{"session drift"}, UpdatedAt: zeroSafeNow()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if _, err := service.AddReviewRoomItem("open", "Triage latest API drift."); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	packet := service.Telegram()
	if packet.ReviewOpenCount != 1 || packet.CurrentFocus != "Ship official API" {
		t.Fatalf("unexpected telegram packet: %+v", packet)
	}
	if len(packet.BodyLines) == 0 {
		t.Fatalf("expected telegram body lines, got %+v", packet)
	}
}

func TestTelegramStatusDefaultsToOffWithoutSecrets(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "")
	status := service.TelegramStatus()
	if status.InboundMode != "off" || status.FounderDelivery != "disabled" {
		t.Fatalf("unexpected telegram status: %+v", status)
	}
	if status.SendConfigured || status.PollingConfigured || status.ChatConfigured || status.GeminiConfigured {
		t.Fatalf("expected all telegram readiness flags to be false, got %+v", status)
	}
	if len(status.Routes) != 3 {
		t.Fatalf("expected three telegram routes, got %+v", status.Routes)
	}
}

func TestTelegramStatusSupportsCommandOnlyAndAssistantModes(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "")
	commandOnly := service.TelegramStatus()
	if commandOnly.InboundMode != "command_only" || commandOnly.FounderDelivery != "chat_missing" {
		t.Fatalf("unexpected command-only telegram status: %+v", commandOnly)
	}

	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "founder-chat")
	t.Setenv("GOOGLE_API_KEY", "gemini")
	assistant := service.TelegramStatus()
	if assistant.InboundMode != "assistant" || assistant.FounderDelivery != "ready" {
		t.Fatalf("unexpected assistant telegram status: %+v", assistant)
	}
	if !assistant.SendConfigured || !assistant.PollingConfigured || !assistant.ChatConfigured || !assistant.GeminiConfigured {
		t.Fatalf("expected telegram assistant readiness flags, got %+v", assistant)
	}
	if got := assistant.Routes[1].Delivery; got != "chat_missing" {
		t.Fatalf("expected ops route to require its own chat, got %+v", assistant.Routes)
	}
}

func TestTelegramStatusSupportsLegacyFounderAliasAndRouteSplit(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "legacy-founder-chat")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "ops-chat")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "gemini")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	status := service.TelegramStatus()
	if !status.SendConfigured || status.FounderDelivery != "ready" {
		t.Fatalf("expected founder route to be ready through legacy alias, got %+v", status)
	}
	if status.Routes[0].Delivery != "ready" || status.Routes[1].Delivery != "ready" || status.Routes[2].Delivery != "disabled" {
		t.Fatalf("unexpected route split: %+v", status.Routes)
	}
	if len(status.Routes[0].Notes) == 0 {
		t.Fatalf("expected founder route note about legacy alias, got %+v", status.Routes[0])
	}
	if !service.TelegramRouteEnabled(TelegramRouteFounder) {
		t.Fatalf("expected founder route to be enabled")
	}
	if !service.TelegramRouteEnabled(TelegramRouteOps) {
		t.Fatalf("expected ops route to be enabled")
	}
	if service.TelegramRouteEnabled(TelegramRouteBrand) {
		t.Fatalf("expected brand route to stay disabled without a dedicated chat")
	}
}

func TestTelegramStatusRecognizesGeminiAPIKeyAlias(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "founder-chat")
	t.Setenv("TELEGRAM_FOUNDER_DM_CHAT_ID", "")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "gemini")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	status := service.TelegramStatus()
	if status.InboundMode != "assistant" || !status.GeminiConfigured {
		t.Fatalf("expected GEMINI_API_KEY alias to enable assistant mode, got %+v", status)
	}
}

func TestTelegramStatusKeepsFounderRoomReadyButConversationCommandOnlyUntilDMExists(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "-5060692298")
	t.Setenv("TELEGRAM_FOUNDER_DM_CHAT_ID", "")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "gemini")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	status := service.TelegramStatus()
	if status.FounderDelivery != "ready" {
		t.Fatalf("expected founder room send path to stay ready, got %+v", status)
	}
	if status.InboundMode != "command_only" {
		t.Fatalf("expected founder conversation to stay command_only without DM, got %+v", status)
	}
	if len(status.Routes[0].Notes) == 0 {
		t.Fatalf("expected founder route note about DM split, got %+v", status.Routes[0])
	}
}

func TestTelegramStatusBlocksFounderAlertDeliveryWhenRoomAndDMAreSame(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", "7565534667")
	t.Setenv("TELEGRAM_FOUNDER_DM_CHAT_ID", "7565534667")
	t.Setenv("TELEGRAM_OPS_CHAT_ID", "")
	t.Setenv("TELEGRAM_BRAND_CHAT_ID", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GOOGLE_API_KEY", "gemini")

	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	status := service.TelegramStatus()
	if status.FounderDelivery != "split_required" || status.SendConfigured {
		t.Fatalf("expected founder split conflict to block send path, got %+v", status)
	}
	if service.TelegramRouteEnabled(TelegramRouteFounder) {
		t.Fatalf("expected founder route to stay disabled while room and dm overlap")
	}
	if status.InboundMode != "assistant" {
		t.Fatalf("expected founder DM conversation to stay assistant-ready, got %+v", status)
	}
	if len(status.Routes[0].Notes) == 0 {
		t.Fatalf("expected founder route conflict note, got %+v", status.Routes[0])
	}
}

func TestAddReviewRoomItemPushesFounderAttentionTelegram(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Setenv("LAYER_OS_TELEGRAM_FOUNDER_ALERTS", "1")
	telegram := &captureTelegramAdapter{}
	service.telegramAdapter = telegram

	if _, err := service.AddReviewRoomItem("open", "Triage latest API drift."); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	if len(telegram.packets) != 1 {
		t.Fatalf("expected 1 founder telegram notice, got %+v", telegram.packets)
	}
	if telegram.routes[0] != TelegramRouteFounder {
		t.Fatalf("expected founder route, got %+v", telegram.routes)
	}
	if telegram.packets[0].PrimaryAction != "review_room" {
		t.Fatalf("expected review_room attention, got %+v", telegram.packets[0])
	}
	if telegram.packets[0].Headline != "지금 먼저 볼 것" {
		t.Fatalf("expected korean founder attention headline, got %+v", telegram.packets[0])
	}
	if len(telegram.packets[0].BodyLines) == 0 || telegram.packets[0].BodyLines[0] != "다음 액션: 리뷰룸 확인" {
		t.Fatalf("expected localized founder action line, got %+v", telegram.packets[0].BodyLines)
	}
}

func TestFounderAttentionTelegramDedupesSameState(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Setenv("LAYER_OS_TELEGRAM_FOUNDER_ALERTS", "1")
	telegram := &captureTelegramAdapter{}
	service.telegramAdapter = telegram

	if _, err := service.AddReviewRoomItem("open", "Triage latest API drift."); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	if err := service.ReplaceMemory(SystemMemory{CurrentFocus: "bootstrap runtime", CurrentGoal: nil, NextSteps: []string{}, OpenRisks: []string{}, UpdatedAt: zeroSafeNow()}); err != nil {
		t.Fatalf("replace memory: %v", err)
	}
	if len(telegram.packets) != 1 {
		t.Fatalf("expected duplicate founder attention to be suppressed, got %+v", telegram.packets)
	}
}

func TestFounderAttentionTelegramStaysOffByDefault(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	telegram := &captureTelegramAdapter{}
	service.telegramAdapter = telegram

	if _, err := service.AddReviewRoomItem("open", "Triage latest API drift."); err != nil {
		t.Fatalf("add review room item: %v", err)
	}
	if len(telegram.packets) != 0 {
		t.Fatalf("expected founder attention push to stay off by default, got %+v", telegram.packets)
	}
}

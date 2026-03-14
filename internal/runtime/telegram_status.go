package runtime

import (
	"os"
	"strings"
)

func telegramEnvPresent(name string) bool {
	return strings.TrimSpace(os.Getenv(name)) != ""
}

func (s *Service) TelegramStatus() TelegramStatus {
	tokenReady := telegramEnvPresent("TELEGRAM_BOT_TOKEN")
	founderRoomConfigured := telegramFounderRoomChatID() != ""
	founderAlertReady := telegramFounderAlertChatID() != ""
	founderConversationReady := telegramFounderConversationChatID() != ""
	founderSplitConflict := telegramFounderRoomDMConflict()
	legacyFounderAlias := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")) != "" && strings.TrimSpace(os.Getenv("TELEGRAM_FOUNDER_CHAT_ID")) == ""
	geminiReady := ProviderCredentialReady("gemini")
	geminiEnvLabel := strings.Join(ProviderCredentialEnvKeys("gemini"), " or ")
	sendAdapter := telegramAdapterFromEnv().Name()
	routes := []TelegramRouteStatus{
		telegramFounderRouteStatus(tokenReady, founderConversationReady),
		telegramOpsRouteStatus(tokenReady, founderAlertReady),
		telegramBrandRouteStatus(tokenReady),
	}

	status := TelegramStatus{
		Adapter:           s.TelegramAdapterName(),
		SendAdapter:       sendAdapter,
		SendConfigured:    tokenReady && founderAlertReady,
		PollingConfigured: tokenReady,
		ChatConfigured:    founderAlertReady,
		GeminiConfigured:  geminiReady,
		InboundMode:       "off",
		FounderDelivery:   routes[0].Delivery,
		Routes:            routes,
		Notes:             []string{},
	}

	switch {
	case tokenReady && geminiReady && founderConversationReady:
		status.InboundMode = "assistant"
	case tokenReady:
		status.InboundMode = "command_only"
	}

	if !tokenReady {
		status.Notes = append(status.Notes, "Telegram bot token is missing, so inbound polling is off.")
	}
	if tokenReady && !founderRoomConfigured {
		status.Notes = append(status.Notes, "Telegram can receive commands, but founder packet delivery is blocked until TELEGRAM_FOUNDER_CHAT_ID (or legacy TELEGRAM_CHAT_ID) is set.")
	}
	if tokenReady && founderAlertReady {
		status.Notes = append(status.Notes, "Founder alert delivery is ready on the canonical send path.")
	}
	if tokenReady && founderAlertReady && !telegramFounderAttentionEnabled() {
		status.Notes = append(status.Notes, "Automatic founder alert pushes stay off by default; use admin as the control tower and keep Telegram focused on direct conversation and brand work.")
	}
	if tokenReady && founderSplitConflict {
		status.Notes = append(status.Notes, "Founder room and founder DM currently share the same chat id, so founder alerts are paused until those routes are split.")
	}
	if tokenReady && geminiReady && founderRoomConfigured && !founderConversationReady {
		status.Notes = append(status.Notes, "Founder room is live, but free-text assistant replies still need TELEGRAM_FOUNDER_DM_CHAT_ID because the current founder route is a room, not a 1:1 DM.")
	}
	if tokenReady && !geminiReady {
		status.Notes = append(status.Notes, "Inbound Telegram stays in command-only mode until "+geminiEnvLabel+" is configured.")
	}
	if tokenReady && geminiReady && founderConversationReady {
		status.Notes = append(status.Notes, "Inbound Telegram can answer founder messages with the Gemini-backed assistant path.")
	}
	if legacyFounderAlias {
		status.Notes = append(status.Notes, "Founder route is currently using legacy TELEGRAM_CHAT_ID. Move it to TELEGRAM_FOUNDER_CHAT_ID when convenient.")
	}

	return status
}

func telegramFounderRouteStatus(tokenReady bool, founderConversationReady bool) TelegramRouteStatus {
	roomConfigured := telegramFounderRoomChatID() != ""
	conflict := telegramFounderRoomDMConflict()
	sendReady := telegramFounderAlertChatID() != ""
	item := TelegramRouteStatus{
		RouteID:        TelegramRouteFounder,
		Label:          "Founder",
		ChatConfigured: sendReady,
		Delivery:       "disabled",
		Notes:          []string{},
	}
	switch {
	case tokenReady && sendReady:
		item.Delivery = "ready"
	case tokenReady && conflict:
		item.Delivery = "split_required"
	case tokenReady:
		item.Delivery = "chat_missing"
	case roomConfigured:
		item.Delivery = "token_missing"
	}
	if strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")) != "" && strings.TrimSpace(os.Getenv("TELEGRAM_FOUNDER_CHAT_ID")) == "" {
		item.Notes = append(item.Notes, "Using legacy TELEGRAM_CHAT_ID as the founder route.")
	}
	if conflict {
		item.Notes = append(item.Notes, "Founder room and founder DM are using the same chat id, so founder alerts stay paused until the routes are split.")
	}
	if roomConfigured && !conflict && !founderConversationReady {
		item.Notes = append(item.Notes, "Founder room is live for alerts and decisions, but 1:1 assistant replies still need TELEGRAM_FOUNDER_DM_CHAT_ID.")
	}
	if founderConversationReady && telegramFounderDMChatID() != "" && !conflict {
		item.Notes = append(item.Notes, "Founder 1:1 DM is split from the founder room for free-text assistant replies.")
	}
	return item
}

func telegramOpsRouteStatus(tokenReady bool, _ bool) TelegramRouteStatus {
	chatReady := telegramOpsChatID() != ""
	item := TelegramRouteStatus{
		RouteID:        TelegramRouteOps,
		Label:          "Ops",
		ChatConfigured: chatReady,
		Delivery:       "disabled",
		Notes:          []string{},
	}
	switch {
	case tokenReady && chatReady:
		item.Delivery = "ready"
	case tokenReady:
		item.Delivery = "chat_missing"
	case chatReady:
		item.Delivery = "token_missing"
	}
	return item
}

func telegramBrandRouteStatus(tokenReady bool) TelegramRouteStatus {
	chatReady := telegramBrandChatID() != ""
	item := TelegramRouteStatus{
		RouteID:        TelegramRouteBrand,
		Label:          "Brand",
		ChatConfigured: chatReady,
		Delivery:       "disabled",
		Notes:          []string{},
	}
	switch {
	case tokenReady && chatReady:
		item.Delivery = "ready"
	case chatReady:
		item.Delivery = "token_missing"
	}
	return item
}

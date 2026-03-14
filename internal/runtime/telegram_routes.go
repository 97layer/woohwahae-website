package runtime

import (
	"os"
	"strconv"
	"strings"
)

const (
	TelegramRouteFounder = "founder"
	TelegramRouteOps     = "ops"
	TelegramRouteBrand   = "brand"
)

type telegramRouteMeta struct {
	RouteID string
	Label   string
}

type TelegramInboundContext struct {
	ChatID    int64
	RouteID   string
	ChatType  string
	ChatLabel string
}

var telegramRouteCatalog = []telegramRouteMeta{
	{RouteID: TelegramRouteFounder, Label: "Founder"},
	{RouteID: TelegramRouteOps, Label: "Ops"},
	{RouteID: TelegramRouteBrand, Label: "Brand"},
}

func telegramBotToken() string {
	return strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
}

func telegramFounderRoomChatID() string {
	if value := strings.TrimSpace(os.Getenv("TELEGRAM_FOUNDER_CHAT_ID")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
}

func telegramFounderDMChatID() string {
	return strings.TrimSpace(os.Getenv("TELEGRAM_FOUNDER_DM_CHAT_ID"))
}

func telegramFounderRoomDMConflict() bool {
	room := strings.TrimSpace(telegramFounderRoomChatID())
	dm := strings.TrimSpace(telegramFounderDMChatID())
	return room != "" && dm != "" && room == dm
}

func telegramFounderAlertChatID() string {
	room := strings.TrimSpace(telegramFounderRoomChatID())
	if room == "" {
		return ""
	}
	if telegramFounderRoomDMConflict() {
		return ""
	}
	return room
}

func telegramFounderChatID() string {
	return telegramFounderRoomChatID()
}

func telegramFounderConversationChatID() string {
	if value := telegramFounderDMChatID(); value != "" {
		return value
	}
	room := telegramFounderRoomChatID()
	if strings.HasPrefix(room, "-") {
		return ""
	}
	return room
}

func telegramOpsChatID() string {
	return strings.TrimSpace(os.Getenv("TELEGRAM_OPS_CHAT_ID"))
}

func telegramBrandChatID() string {
	return strings.TrimSpace(os.Getenv("TELEGRAM_BRAND_CHAT_ID"))
}

func telegramRouteChatID(routeID string) string {
	switch routeID {
	case TelegramRouteFounder:
		return telegramFounderAlertChatID()
	case TelegramRouteOps:
		return telegramOpsChatID()
	case TelegramRouteBrand:
		return telegramBrandChatID()
	default:
		return ""
	}
}

func telegramRouteConfigured(routeID string) bool {
	return telegramRouteChatID(routeID) != ""
}

func TelegramRouteForChatID(chatID int64) string {
	return telegramRouteForChatValue(strconv.FormatInt(chatID, 10))
}

func telegramRouteForChatValue(chatID string) string {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return ""
	}
	if chatID == telegramFounderConversationChatID() {
		return TelegramRouteFounder
	}
	for _, route := range telegramRouteCatalog {
		if strings.TrimSpace(telegramRouteChatID(route.RouteID)) == chatID {
			return route.RouteID
		}
	}
	return ""
}

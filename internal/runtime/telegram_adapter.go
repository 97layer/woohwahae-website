package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TelegramAdapter sends a TelegramPacket to a Telegram chat.
type TelegramAdapter interface {
	Name() string
	Enabled() bool
	Send(packet TelegramPacket) error
}

type telegramRouteSender interface {
	SendRoute(routeID string, packet TelegramPacket) error
	RouteEnabled(routeID string) bool
}

// telegramAdapterFromEnv returns a bot adapter when a bot token plus at least
// one route chat are configured, otherwise a noop adapter.
func telegramAdapterFromEnv() TelegramAdapter {
	token := telegramBotToken()
	routes := map[string]string{}
	for _, route := range telegramRouteCatalog {
		if chatID := telegramRouteChatID(route.RouteID); chatID != "" {
			routes[route.RouteID] = chatID
		}
	}
	if token != "" && len(routes) > 0 {
		return botTelegramAdapter{token: token, routes: routes}
	}
	return noopTelegramAdapter{}
}

// botTelegramAdapter calls the Telegram Bot API.
type botTelegramAdapter struct {
	token  string
	routes map[string]string
}

func (a botTelegramAdapter) Name() string  { return "telegram_bot" }
func (a botTelegramAdapter) Enabled() bool { return a.RouteEnabled(TelegramRouteFounder) }

func (a botTelegramAdapter) Send(packet TelegramPacket) error {
	return a.SendRoute(TelegramRouteFounder, packet)
}

func (a botTelegramAdapter) RouteEnabled(routeID string) bool {
	return strings.TrimSpace(a.routes[routeID]) != ""
}

func (a botTelegramAdapter) SendRoute(routeID string, packet TelegramPacket) error {
	text := buildTelegramText(packet)
	chatID := strings.TrimSpace(a.routes[routeID])
	if chatID == "" {
		return fmt.Errorf("telegram route %s not configured", routeID)
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	if err != nil {
		return err
	}
	url := "https://api.telegram.org/bot" + a.token + "/sendMessage"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned %d", resp.StatusCode)
	}
	return nil
}

// noopTelegramAdapter does nothing (env vars not set).
type noopTelegramAdapter struct{}

func (a noopTelegramAdapter) Name() string  { return "noop" }
func (a noopTelegramAdapter) Enabled() bool { return false }
func (a noopTelegramAdapter) Send(_ TelegramPacket) error {
	return fmt.Errorf("telegram adapter not configured: set TELEGRAM_BOT_TOKEN plus TELEGRAM_FOUNDER_CHAT_ID (or TELEGRAM_CHAT_ID)")
}
func (a noopTelegramAdapter) SendRoute(routeID string, _ TelegramPacket) error {
	return fmt.Errorf("telegram route %s not configured", routeID)
}
func (a noopTelegramAdapter) RouteEnabled(_ string) bool { return false }

func buildTelegramText(packet TelegramPacket) string {
	var b strings.Builder
	b.WriteString("<b>" + escapeHTML(packet.Headline) + "</b>\n")
	for _, line := range packet.BodyLines {
		b.WriteString(escapeHTML(line) + "\n")
	}
	if packet.ReviewOpenCount > 0 {
		b.WriteString(fmt.Sprintf("\n📋 열린 안건: %d", packet.ReviewOpenCount))
	}
	return strings.TrimSpace(b.String())
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type tgUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		Chat      struct {
			ID        int64  `json:"id"`
			Type      string `json:"type"`
			Title     string `json:"title"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

func startTelegramLoop(ctx context.Context, service *runtime.Service) {
	if strings.TrimSpace(os.Getenv("LAYER_OS_DISABLE_TELEGRAM_POLLING")) == "1" {
		log.Println("telegram bot: polling disabled (LAYER_OS_DISABLE_TELEGRAM_POLLING=1)")
		return
	}
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return
	}
	handler := runtime.NewTelegramBotHandler(service)
	if !handler.Enabled() {
		log.Printf("telegram bot: %s not set, running in command-only mode", strings.Join(runtime.ProviderCredentialEnvKeys("gemini"), " or "))
	}

	log.Println("telegram bot: polling started")
	offset := 0
	client := &http.Client{Timeout: 40 * time.Second}
	baseURL := "https://api.telegram.org/bot" + token

	for {
		select {
		case <-ctx.Done():
			log.Println("telegram bot: stopped")
			return
		default:
		}

		updates, err := fetchUpdates(client, baseURL, offset)
		if err != nil {
			log.Printf("telegram bot: getUpdates error: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			text := strings.TrimSpace(u.Message.Text)
			chatID := u.Message.Chat.ID
			if text == "" || chatID == 0 {
				continue
			}
			routeID := runtime.TelegramRouteForChatID(chatID)
			if routeID == "" {
				routeID = "unmapped"
			}
			log.Printf(
				"telegram bot: inbound chat_id=%d route=%s chat_type=%s chat_label=%q message_id=%d text_len=%d update_id=%d",
				chatID,
				routeID,
				strings.TrimSpace(u.Message.Chat.Type),
				telegramChatLabel(u.Message.Chat.Title, u.Message.Chat.Username, u.Message.Chat.FirstName, u.Message.Chat.LastName),
				u.Message.MessageID,
				len(text),
				u.UpdateID,
			)
			reply := handler.HandleMessageWithContext(runtime.TelegramInboundContext{
				ChatID:    chatID,
				RouteID:   routeID,
				ChatType:  strings.TrimSpace(u.Message.Chat.Type),
				ChatLabel: telegramChatLabel(u.Message.Chat.Title, u.Message.Chat.Username, u.Message.Chat.FirstName, u.Message.Chat.LastName),
			}, text)
			if reply == "" {
				continue
			}
			if err := sendMessage(client, baseURL, chatID, reply); err != nil {
				log.Printf("telegram bot: sendMessage error: %v", err)
			}
		}

		if len(updates) == 0 {
			// long poll — minimal sleep before next call
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}
}

func telegramChatLabel(title, username, firstName, lastName string) string {
	if value := strings.TrimSpace(title); value != "" {
		return value
	}
	if value := strings.TrimSpace(username); value != "" {
		return "@" + strings.TrimPrefix(value, "@")
	}
	label := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(firstName), strings.TrimSpace(lastName)}, " "))
	return label
}

func fetchUpdates(client *http.Client, baseURL string, offset int) ([]tgUpdate, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", baseURL, offset)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result struct {
		OK          bool       `json:"ok"`
		Result      []tgUpdate `json:"result"`
		ErrorCode   int        `json:"error_code"`
		Description string     `json:"description"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("getUpdates decode failed status=%d body=%q: %w", resp.StatusCode, telegramBodySnippet(raw), err)
	}
	if resp.StatusCode != http.StatusOK || !result.OK {
		return nil, telegramAPIError("getUpdates", resp.StatusCode, result.ErrorCode, result.Description, raw)
	}
	return result.Result, nil
}

func sendMessage(client *http.Client, baseURL string, chatID int64, text string) error {
	// Telegram HTML 4096자 제한
	if len(text) > 4000 {
		text = text[:4000] + "\n…[잘림]"
	}
	payload, _ := json.Marshal(map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	resp, err := client.Post(baseURL+"/sendMessage", "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sendMessage read failed: %w", err)
	}
	var result struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("sendMessage failed status=%d body=%q", resp.StatusCode, telegramBodySnippet(raw))
		}
	}
	if resp.StatusCode != http.StatusOK || !result.OK {
		return telegramAPIError("sendMessage", resp.StatusCode, result.ErrorCode, result.Description, raw)
	}
	return nil
}

func telegramAPIError(method string, httpStatus int, telegramCode int, description string, raw []byte) error {
	parts := []string{method + " failed", fmt.Sprintf("status=%d", httpStatus)}
	if telegramCode != 0 {
		parts = append(parts, fmt.Sprintf("telegram_code=%d", telegramCode))
	}
	if text := strings.TrimSpace(description); text != "" {
		parts = append(parts, fmt.Sprintf("description=%q", text))
	} else if snippet := telegramBodySnippet(raw); snippet != "" {
		parts = append(parts, fmt.Sprintf("body=%q", snippet))
	}
	return fmt.Errorf(strings.Join(parts, " "))
}

func telegramBodySnippet(raw []byte) string {
	snippet := strings.TrimSpace(string(raw))
	if snippet == "" {
		return ""
	}
	runes := []rune(snippet)
	if len(runes) > 240 {
		return string(runes[:240]) + "…"
	}
	return snippet
}

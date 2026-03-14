package runtime

import (
	"os"
	"testing"
)

func TestTelegramSendManual(t *testing.T) {
	if os.Getenv("LAYER_OS_RUN_TELEGRAM_MANUAL_TEST") != "1" {
		t.Skip("manual telegram test disabled; set LAYER_OS_RUN_TELEGRAM_MANUAL_TEST=1 to enable")
	}
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_FOUNDER_CHAT_ID")
	if chatID == "" {
		chatID = os.Getenv("TELEGRAM_CHAT_ID")
	}
	if token == "" || chatID == "" {
		t.Skip("manual telegram test requires TELEGRAM_BOT_TOKEN and TELEGRAM_FOUNDER_CHAT_ID (or legacy TELEGRAM_CHAT_ID)")
	}

	t.Setenv("TELEGRAM_BOT_TOKEN", token)
	t.Setenv("TELEGRAM_FOUNDER_CHAT_ID", chatID)
	t.Setenv("TELEGRAM_CHAT_ID", "")
	adapter := telegramAdapterFromEnv()
	if !adapter.Enabled() {
		t.Fatal("adapter not enabled")
	}

	packet := TelegramPacket{
		Headline: "🛡️ Layer OS Integration Test",
		BodyLines: []string{
			"Status: External Port Active",
			"Actor: Antigravity",
			"Timestamp: " + zeroSafeNow().Format("2006-01-02 15:04:05"),
		},
		ReviewOpenCount: 1,
	}

	err := adapter.Send(packet)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	t.Log("Telegram message sent successfully")
}

package main

import (
	"log"

	"layer-os/internal/runtime"
)

type telegramService interface {
	Telegram() runtime.TelegramPacket
	SendTelegram() error
	TelegramEnabled() bool
}

func runTelegram(service telegramService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl telegram <preview|send>")
	}
	switch args[0] {
	case "preview":
		writeJSON(service.Telegram())
	case "send":
		if !service.TelegramEnabled() {
			log.Fatal("telegram adapter not configured: set TELEGRAM_BOT_TOKEN and TELEGRAM_FOUNDER_CHAT_ID")
		}
		if err := service.SendTelegram(); err != nil {
			log.Fatalf("telegram send failed: %v", err)
		}
		writeJSON(map[string]any{"ok": true, "sent": true})
	default:
		log.Fatal("usage: layer-osctl telegram <preview|send>")
	}
}

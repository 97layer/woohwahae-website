package main

import (
	"fmt"
	"os"
	"layer-os/internal/runtime"
)

func main() {
	dataDir := ".layer-os"
	s, err := runtime.NewService(dataDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	packet := s.Telegram()
	packet.Headline = "Brand 채널 연결 확인 (내부 엔진)"
	packet.BodyLines = []string{
		"Layer OS Brand 채널이 성공적으로 대몬에 연결되었습니다.",
		"ID: " + os.Getenv("TELEGRAM_BRAND_CHAT_ID"),
	}
	
	err = s.SendTelegramRoute(runtime.TelegramRouteBrand, packet)
	if err != nil {
		fmt.Printf("Error sending to Brand: %v\n", err)
		return
	}
	fmt.Println("Success: Sent test notification to Brand channel through Service")
}

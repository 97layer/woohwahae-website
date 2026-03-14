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
	packet.Headline = "Ops 채널 연결 확인 (내부 엔진)"
	packet.BodyLines = []string{
		"Layer OS Ops 채널이 성공적으로 대몬에 연결되었습니다.",
		"ID: " + os.Getenv("TELEGRAM_OPS_CHAT_ID"),
	}
	
	// SendTelegramRoute 메서드가 정상적으로 구현되었는지 확인
	err = s.SendTelegramRoute(runtime.TelegramRouteOps, packet)
	if err != nil {
		fmt.Printf("Error sending to Ops: %v\n", err)
		return
	}
	fmt.Println("Success: Sent test notification to Ops channel through Service")
}

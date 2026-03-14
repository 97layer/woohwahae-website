package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"layer-os/internal/api"
	"layer-os/internal/runtime"
)

func main() {
	if os.Getenv("LAYER_OS_WRITER_ID") == "" {
		os.Setenv("LAYER_OS_WRITER_ID", "layer-osd")
	}
	address := os.Getenv("LAYER_OS_ADDR")
	if address == "" {
		address = runtime.DefaultDaemonAddr
	}
	dataDir := os.Getenv("LAYER_OS_DATA_DIR")
	if dataDir == "" {
		dataDir = ".layer-os"
	}

	startedAt := time.Now().UTC()
	service, err := runtime.NewService(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	service.BindDaemonRuntime(startedAt)
	architectStatus := newArchitectLoopStatus(loadArchitectLoopConfig())
	router := api.NewRouterWithRuntime(service, api.DaemonRuntimeInfo{Address: address, StartedAt: startedAt, ArchitectStatus: architectStatus.Snapshot})
	server := newDaemonServer(address, router)
	listener, err := listenDaemon(address)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go startArchitectLoop(ctx, service, architectStatus)
	go startFeedSensorLoop(ctx, service)
	go startTelegramLoop(ctx, service)

	log.Printf("layer-osd listening on %s", address)
	if err := runDaemon(ctx, server, listener); err != nil {
		log.Fatal(err)
	}
}

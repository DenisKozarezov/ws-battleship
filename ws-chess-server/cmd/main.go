package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"ws-chess-server/internal/application"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/http/routers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	logger := application.NewDefaultLogger()

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatalf("failed to parse a config: %s", err)
	}
	logger.SetDebugMode(cfg.App.IsDebugMode)

	app := application.NewApp(&cfg.App, logger)

	router := routers.NewDefaultRouter(logger)
	app.Run(ctx, router)
}

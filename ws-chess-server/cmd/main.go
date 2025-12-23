package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"ws-chess-server/internal/application"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/http/routers"
	"ws-chess-server/pkg/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Sprintln("failed to parse a config", err))
	}

	logger, err := logger.NewLogger(&cfg.App, "[SERVER]")
	if err != nil {
		panic(fmt.Sprintln("failed to create a logger", err))
	}

	app := application.NewApp(ctx, &cfg.App, logger)

	router := routers.NewDefaultRouter(logger)
	app.Run(ctx, router)
}

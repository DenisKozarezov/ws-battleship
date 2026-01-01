package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"ws-battleship-server/internal/application"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/delivery/http/routers"
	"ws-battleship-shared/pkg/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Sprintln("failed to parse a config", err))
	}

	logger, err := logger.NewLogger(cfg.App.IsDebugMode, "[SERVER]")
	if err != nil {
		panic(fmt.Sprintln("failed to create a logger", err))
	}

	app := application.NewApp(cfg, logger)
	app.Run(ctx, routers.NewDefaultRouter(logger))
}

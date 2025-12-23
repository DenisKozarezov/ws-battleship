package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"ws-battleship-client/internal/application"
	"ws-battleship-client/internal/config"
	"ws-battleship-client/pkg/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Sprintln("failed to parse a config", err))
	}

	logger, err := logger.NewLogger(&cfg.App, "[CLIENT]")
	if err != nil {
		panic(fmt.Sprintln("failed to create a logger", err))
	}

	app := application.NewApp(&cfg.App, logger)
	app.Run(ctx)
}

package application

import (
	"context"
	"ws-chess-client/internal/config"
	client "ws-chess-client/internal/delivery/http"
	"ws-chess-client/internal/delivery/http/middleware"
)

type App struct {
	cfg    *config.AppConfig
	client *client.Client
	logger middleware.Logger
}

func NewApp(cfg *config.AppConfig, logger middleware.Logger) *App {
	client := client.NewClient(cfg)

	return &App{
		cfg:    cfg,
		logger: logger,
		client: client,
	}
}

func (a *App) Run(ctx context.Context) {
	a.logger.Printf("connecting to server :%s", a.cfg.ServerPort)
	go func() {
		if err := a.client.Connect(ctx); err != nil {
			a.logger.Fatalf("failed to connect to server: %s", err)
		}
	}()

	<-ctx.Done()
	a.logger.Println("received a signal to shutdown the server")

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a client: %s", err)
	}
}

func (a *App) Shutdown() error {
	a.logger.Println("shutting the client down...")
	return a.client.Shutdown()
}

package application

import (
	"context"
	"ws-chess-client/internal/config"
	client "ws-chess-client/internal/delivery/websocket"
	"ws-chess-client/internal/delivery/websocket/response"
	"ws-chess-client/pkg/logger"
)

type App struct {
	cfg    *config.AppConfig
	client *client.WebsocketClient
	logger logger.Logger
}

func NewApp(cfg *config.AppConfig, logger logger.Logger) *App {
	client := client.NewClient(cfg, logger)

	return &App{
		cfg:    cfg,
		logger: logger,
		client: client,
	}
}

func (a *App) Run(ctx context.Context) {
	a.logger.Infof("connecting to server %s", a.cfg.ServerHost)
	go func() {
		if err := a.client.Connect(ctx); err != nil {
			a.logger.Fatalf("failed to connect to server: %s", err)
		}
	}()

	go a.handleConnection(ctx)

	<-ctx.Done()
	a.logger.Info("received a signal to shutdown the client")

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a client: %s", err)
	}
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting the client down...")
	return a.client.Shutdown()
}

func (a *App) handleConnection(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case msg := <-a.client.Messages():
			a.handleMessage(msg)
		}
	}
}

func (a *App) handleMessage(event response.Event) {

}

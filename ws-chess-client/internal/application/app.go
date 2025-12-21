package application

import (
	"context"
	"encoding/json"
	"ws-chess-client/internal/config"
	"ws-chess-client/internal/delivery/http/middleware"
	client "ws-chess-client/internal/delivery/websocket"
	"ws-chess-client/internal/delivery/websocket/response"
)

type App struct {
	cfg    *config.AppConfig
	client *client.WebsocketClient
	logger middleware.Logger
}

func NewApp(cfg *config.AppConfig, logger middleware.Logger) *App {
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
			var event response.Response
			if err := json.Unmarshal(msg, &event); err != nil {
				a.logger.Errorf("failed to unmarshal message, discarding it: %s", err)
				continue
			}

			a.handleMessage(event)
		}
	}
}

func (a *App) handleMessage(event response.Response) {
	a.logger.Debugf("type=%s, timestamp=%s, payload=%s", event.Type, event.Timestamp, string(event.Data))
}

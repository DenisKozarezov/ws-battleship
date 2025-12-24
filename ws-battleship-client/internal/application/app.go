package application

import (
	"context"
	"sync"
	"ws-battleship-client/internal/config"
	client "ws-battleship-client/internal/delivery/websocket"
	"ws-battleship-client/internal/domain"
	"ws-battleship-client/pkg/logger"
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

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := a.client.Connect(ctx); err != nil {
			a.logger.Fatalf("failed to connect to server: %s", err)
		}
	}()
	go func() {
		defer wg.Done()
		a.handleConnection(ctx)
	}()

	g := NewGame()
	g.RenderScreen()

	<-ctx.Done()
	a.logger.Info("received a signal to shutdown the client")
	wg.Wait()

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a client: %s", err)
	}
	a.logger.Info("client is gracefully shutdown")
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

func (a *App) handleMessage(event domain.Event) {
	a.logger.Debug("Event Type: %d; Timestamp: %s; Payload: %s", event.Type, event.Timestamp, string(event.Data))
}

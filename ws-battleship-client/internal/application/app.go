package application

import (
	"context"
	"sync"
	"ws-battleship-client/internal/config"
	client "ws-battleship-client/internal/delivery/websocket"
	"ws-battleship-client/internal/domain"
	"ws-battleship-client/internal/domain/models"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-client/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type Client interface {
	Messages() <-chan domain.Event
	Connect(ctx context.Context) error
	Shutdown() error
}

type App struct {
	cfg    *config.AppConfig
	client Client
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
	var wg sync.WaitGroup
	a.startClient(ctx, &wg)
	a.startGame()

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

func (a *App) startClient(ctx context.Context, wg *sync.WaitGroup) {
	a.logger.Infof("connecting to server %s", a.cfg.ServerHost)

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
}

func (a *App) handleConnection(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case msg, opened := <-a.client.Messages():
			if opened {
				a.handleMessage(msg)
			}
		}
	}
}

func (a *App) handleMessage(event domain.Event) {
	a.logger.Debug("Event Type: %d; Timestamp: %s; Payload: %s", event.Type, event.Timestamp, string(event.Data))
}

func (a *App) startGame() {
	gameModel := models.NewGameModel()
	gameView := views.NewGameView(gameModel)

	clearTerminal()
	if _, err := tea.NewProgram(gameView).Run(); err != nil {
		a.logger.Fatalf("failed to run a game view: %s", err)
	}
}

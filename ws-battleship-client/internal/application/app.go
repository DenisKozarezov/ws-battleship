package application

import (
	"context"
	"sync"
	"time"
	"ws-battleship-client/internal/config"
	client "ws-battleship-client/internal/delivery/websocket"
	clientEvents "ws-battleship-client/internal/domain/events"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type Client interface {
	Messages() <-chan events.Event
	Connect(ctx context.Context, metadata domain.ClientMetadata) error
	SendMessage(e events.Event) error
	Shutdown() error
}

type App struct {
	cfg      *config.Config
	client   Client
	logger   logger.Logger
	gameView *views.GameView
	eventBus *events.EventBus
	metadata domain.ClientMetadata
}

func NewApp(ctx context.Context, cfg *config.Config, logger logger.Logger) *App {
	client := client.NewClient(ctx, &cfg.App, logger)
	eventBus := events.NewEventBus()

	metadata := domain.ClientMetadata{Nickname: uuid.New().Domain().String()}

	a := &App{
		cfg:      cfg,
		logger:   logger,
		client:   client,
		eventBus: eventBus,
		gameView: views.NewGameView(eventBus, metadata),
		metadata: metadata,
	}

	eventBus.Subscribe(clientEvents.PlayerTypedMessageType, a.onPlayerTypedMessage)

	return a
}

func (a *App) Run(ctx context.Context) {
	var wg sync.WaitGroup
	a.startClient(ctx, &wg)
	a.runGameLoop(ctx, &wg)

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
	a.logger.Infof("connecting to server %s", a.cfg.App.ServerHost)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := a.client.Connect(ctx, a.metadata); err != nil {
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
			if !opened {
				return
			}

			if err := a.eventBus.Invoke(msg); err != nil {
				a.logger.Errorf("error while invoking event bus: %s", err)
			}
		}
	}
}

func (a *App) runGameLoop(ctx context.Context, wg *sync.WaitGroup) {
	const fps = 60
	const fixedTime = time.Second / fps

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(fixedTime)
		defer func() {
			wg.Done()
			ticker.Stop()
		}()

		for {
			if err := ctx.Err(); err != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.gameView.FixedUpdate()
			}
		}
	}()

	clearTerminal()
	if _, err := tea.NewProgram(a.gameView).Run(); err != nil {
		a.logger.Fatalf("failed to run a game view: %s", err)
	}
	clearTerminal()
}

func (a *App) onPlayerTypedMessage(e events.Event) error {
	e.Type = events.SendMessageType
	return a.client.SendMessage(e)
}

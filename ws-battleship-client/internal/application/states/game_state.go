package states

import (
	"context"
	"net"
	"sync"
	clientEvents "ws-battleship-client/internal/domain/events"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/domain"
	serverEvents "ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type Client interface {
	Metadata() domain.ClientMetadata
	Messages() <-chan serverEvents.Event
	Connect(ctx context.Context, ipv4 net.IP) error
	Shutdown() error
	SendMessage(e serverEvents.Event) error
}

type GameState struct {
	logger logger.Logger
	wg     sync.WaitGroup

	stateMachine StateMachine
	client       Client
	metadata     domain.ClientMetadata
	eventBus     *serverEvents.EventBus
	gameView     *views.GameView
}

func NewGameState(stateMachine StateMachine, client Client, logger logger.Logger) *GameState {
	eventBus := serverEvents.NewEventBus()

	return &GameState{
		logger:       logger,
		stateMachine: stateMachine,
		client:       client,
		metadata:     client.Metadata(),
		eventBus:     eventBus,
		gameView:     views.NewGameView(eventBus, client.Metadata()),
	}
}

func (s *GameState) OnExit() {
	s.eventBus.Unsubscribe(serverEvents.GameStartEventType, s.onGameStartedHandler)
	s.eventBus.Unsubscribe(serverEvents.GameEndEventType, s.onGameEndHandler)
	s.eventBus.Unsubscribe(serverEvents.PlayerUpdateStateEventType, s.onPlayerUpdateState)
	s.eventBus.Unsubscribe(serverEvents.PlayerTurnEventType, s.onPlayerTurnHandler)
	s.eventBus.Unsubscribe(serverEvents.SendMessageType, s.onPlayerSendMessageHandler)
	s.eventBus.Unsubscribe(clientEvents.PlayerTypedMessageType, s.onPlayerTypedMessage)
	s.gameView.SetPlayerFiredHandler(nil)

	_ = s.client.Shutdown()
	s.wg.Wait()
}

func (s *GameState) OnEnter() {
	s.gameView.Init()
	s.eventBus.Subscribe(serverEvents.GameStartEventType, s.onGameStartedHandler)
	s.eventBus.Subscribe(serverEvents.GameEndEventType, s.onGameEndHandler)
	s.eventBus.Subscribe(serverEvents.PlayerUpdateStateEventType, s.onPlayerUpdateState)
	s.eventBus.Subscribe(serverEvents.PlayerTurnEventType, s.onPlayerTurnHandler)
	s.eventBus.Subscribe(serverEvents.SendMessageType, s.onPlayerSendMessageHandler)
	s.eventBus.Subscribe(clientEvents.PlayerTypedMessageType, s.onPlayerTypedMessage)
	s.gameView.SetPlayerFiredHandler(s.onPlayerPressedFireHandler)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleConnection(s.stateMachine.Context())
	}()
}

func (s *GameState) FixedUpdate() {
	s.gameView.FixedUpdate()
}

func (s *GameState) View() tea.Model {
	return s.gameView
}

func (s *GameState) handleConnection(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case msg, opened := <-s.client.Messages():
			if !opened {
				return
			}

			if err := s.eventBus.Invoke(msg); err != nil {
				s.logger.Errorf("error while invoking event: %s", err)
			}
		}
	}
}

package states

import (
	"net"
	"time"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type MainMenuState struct {
	stateMachine StateMachine
	menuView     *views.MainMenuView
	logger       logger.Logger
}

func NewMainMenuState(stateMachine StateMachine, logger logger.Logger) *MainMenuState {
	return &MainMenuState{
		stateMachine: stateMachine,
		menuView:     views.NewMainMenuView(),
		logger:       logger,
	}
}

func (s *MainMenuState) OnExit() {
	s.menuView.ConnectFunc = nil
}

func (s *MainMenuState) OnEnter() {
	s.menuView.Init()
	s.menuView.ConnectFunc = s.onPlayerConnecting
}

func (s *MainMenuState) FixedUpdate() {
	s.menuView.FixedUpdate()
}

func (s *MainMenuState) View() tea.Model {
	return s.menuView
}

func (s *MainMenuState) onPlayerConnecting(ipv4 net.IP) {
	connectionState := NewConnectingState(s.stateMachine, ipv4, s.logger)

	// If connection to server succeeded, then switch to game boards
	connectionState.SetOnSuccess(func(client Client) {
		time.Sleep(time.Second)
		s.stateMachine.SwitchState(NewGameState(s.stateMachine, client, s.logger))
	})

	// If connection to server failed, then return to main menu
	connectionState.SetOnError(func(err error) {
		s.menuView.IPv4Error = err
		s.stateMachine.SwitchState(s)
	})

	s.stateMachine.SwitchState(connectionState)
}

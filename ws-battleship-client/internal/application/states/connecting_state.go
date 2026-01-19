package states

import (
	"context"
	"errors"
	"net"
	"time"
	client "ws-battleship-client/internal/delivery/websocket"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ConnectingState struct {
	stateMachine StateMachine
	cancel       context.CancelFunc

	logger logger.Logger
	client Client
	ipv4   net.IP

	connectServerView *views.ConnectServerView

	onSuccess func(client Client)
	onError   func(err error)
}

func NewConnectingState(stateMachine StateMachine, ipv4 net.IP, logger logger.Logger) *ConnectingState {
	metadata := domain.NewClientMetadata(uuid.New().Domain().String())

	return &ConnectingState{
		stateMachine:      stateMachine,
		logger:            logger,
		client:            client.NewClient(stateMachine.Context(), logger, metadata),
		ipv4:              ipv4,
		connectServerView: views.NewConnectServerView(),
	}
}

func (s *ConnectingState) OnExit() {

}

func (s *ConnectingState) OnEnter() {
	var ctx context.Context
	ctx, s.cancel = context.WithTimeout(s.stateMachine.Context(), 5*time.Second)
	go s.startClient(ctx)
}

func (s *ConnectingState) FixedUpdate() {
	s.connectServerView.FixedUpdate()
}

func (s *ConnectingState) View() tea.Model {
	return s.connectServerView
}

func (s *ConnectingState) SetOnSuccess(fn func(client Client)) {
	s.onSuccess = fn
}

func (s *ConnectingState) SetOnError(fn func(err error)) {
	s.onError = fn
}

var (
	ErrTimeout        = context.DeadlineExceeded
	ErrInvalidAddress = errors.New("invalid IP-address")
)

func (s *ConnectingState) startClient(ctx context.Context) {
	if err := s.client.Connect(ctx, s.ipv4); err != nil {
		if s.onError != nil {
			switch err {
			case context.DeadlineExceeded, context.Canceled:
				err = ErrTimeout
			case websocket.ErrBadHandshake:
				err = ErrInvalidAddress
			}

			s.onError(errors.Join(errors.New("unable to connect to the server:"), err))
			return
		}
	}

	if s.onSuccess != nil {
		s.onSuccess(s.client)
	}
}

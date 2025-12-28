package handlers

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/delivery/http/response"
	"ws-battleship-server/internal/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/gorilla/websocket"
)

type WebsocketListener struct {
	upgrader   *websocket.Upgrader
	once       sync.Once
	isShutdown atomic.Bool

	joinCh chan *domain.Player
	logger logger.Logger
}

func NewWebsocketListener(cfg *config.AppConfig, logger logger.Logger, joinCh chan *domain.Player) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  events.ReadBufferBytesMax,
		WriteBufferSize: events.WriteBufferBytesMax,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WebsocketListener{
		upgrader: &websocketUpgrader,
		logger:   logger,
		joinCh:   joinCh,
	}
}

func (l *WebsocketListener) Close() {
	l.isShutdown.Store(true)

	l.once.Do(func() {
		l.logger.Info("websocket listener is closed")
	})
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	if l.isShutdown.Load() {
		response.Error(w, errors.New("listener is closed"), 499)
		return nil
	}

	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	newClient := domain.NewClient(conn, l.logger, events.ParseClientMetadataFromHeaders(r))
	newPlayer := domain.NewPlayer(newClient)

	select {
	case l.joinCh <- newPlayer:
	default:
	}
	return nil
}

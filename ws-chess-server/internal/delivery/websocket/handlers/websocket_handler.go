package handlers

import (
	"context"
	"net/http"
	"sync"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/domain"
	"ws-chess-server/pkg/logger"

	"github.com/gorilla/websocket"
)

const (
	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type WebsocketListener struct {
	upgrader *websocket.Upgrader
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc

	logger logger.Logger
	joinCh chan *domain.Client
	readCh chan domain.Event
}

func NewWebsocketListener(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferBytesMax,
		WriteBufferSize: writeBufferBytesMax,
		CheckOrigin: func(r *http.Request) bool {
			return cfg.IsDebugMode
		},
	}

	ctx, cancel := context.WithCancel(ctx)

	return &WebsocketListener{
		upgrader: &websocketUpgrader,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		joinCh:   make(chan *domain.Client, cfg.ClientsConnectionsMax),
		readCh:   make(chan domain.Event, readBufferBytesMax),
	}
}

func (l *WebsocketListener) Close() {
	l.cancel()

	close(l.joinCh)
	close(l.readCh)

	l.logger.Info("websocket listener is closed")
}

func (l *WebsocketListener) WaitForAllConnections() {
	l.wg.Wait()
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	newClient := domain.NewClient(conn, l.logger, domain.ParseClientMetadata(r))
	l.joinCh <- newClient

	l.wg.Add(1)
	l.wg.Go(func() {
		defer l.wg.Done()
		newClient.ReadMessage(l.ctx, l.readCh)
	})

	return nil
}

func (l *WebsocketListener) JoinChan() <-chan *domain.Client {
	return l.joinCh
}

func (l *WebsocketListener) Messages() <-chan domain.Event {
	return l.readCh
}

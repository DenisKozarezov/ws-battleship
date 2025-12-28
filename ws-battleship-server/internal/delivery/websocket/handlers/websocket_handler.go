package handlers

import (
	"context"
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
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	once       sync.Once
	isShutdown atomic.Bool

	logger logger.Logger
	joinCh chan *domain.Client
	readCh chan events.Event
}

func NewWebsocketListener(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  events.ReadBufferBytesMax,
		WriteBufferSize: events.WriteBufferBytesMax,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ctx, cancel := context.WithCancel(ctx)

	return &WebsocketListener{
		upgrader: &websocketUpgrader,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		joinCh:   make(chan *domain.Client, cfg.ClientsConnectionsMax),
		readCh:   make(chan events.Event, events.ReadBufferBytesMax),
	}
}

func (l *WebsocketListener) Close() {
	l.isShutdown.Store(true)

	l.once.Do(func() {
		l.cancel()

		close(l.joinCh)
		close(l.readCh)

		l.logger.Info("websocket listener is closed")
	})
}

func (l *WebsocketListener) WaitForAllConnections() {
	l.wg.Wait()
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
	l.joinCh <- newClient

	l.wg.Add(2)
	go func(wg *sync.WaitGroup, client *domain.Client) {
		defer wg.Done()
		client.ReadMessages(l.ctx, l.readCh)
	}(&l.wg, newClient)

	go func(wg *sync.WaitGroup, client *domain.Client) {
		defer wg.Done()
		client.WriteMessages(l.ctx)
	}(&l.wg, newClient)

	return nil
}

func (l *WebsocketListener) JoinChan() <-chan *domain.Client {
	return l.joinCh
}

func (l *WebsocketListener) Messages() <-chan events.Event {
	return l.readCh
}

package handlers

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/domain"
	"ws-battleship-server/pkg/logger"

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
	readCh chan domain.Event
}

func NewWebsocketListener(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  domain.ReadBufferBytesMax,
		WriteBufferSize: domain.WriteBufferBytesMax,
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
		readCh:   make(chan domain.Event, domain.ReadBufferBytesMax),
	}
}

func (l *WebsocketListener) Close() {
	l.isShutdown.Store(true)

	l.once.Do(func() {
		l.cancel()

		// TODO: DANGER! if we close these channels and then someone connects to the server, the listener will write in already closed channel...
		close(l.joinCh)
		close(l.readCh)

		l.logger.Info("websocket listener is closed")
	})
}

func (l *WebsocketListener) WaitForAllConnections() {
	l.wg.Wait()
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	if l.isShutdown.Load() {
		return nil
	}

	newClient := domain.NewClient(conn, l.logger, domain.ParseClientMetadata(r))
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

func (l *WebsocketListener) Messages() <-chan domain.Event {
	return l.readCh
}

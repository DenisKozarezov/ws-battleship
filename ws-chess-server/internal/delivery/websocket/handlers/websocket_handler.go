package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/http/middleware"

	"github.com/gorilla/websocket"
)

const (
	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type WebsocketListener struct {
	upgrader        *websocket.Upgrader
	logger          middleware.Logger
	register        chan *websocket.Conn
	readCh, writeCh chan []byte
}

func NewWebsocketListener(cfg *config.AppConfig, logger middleware.Logger) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferBytesMax,
		WriteBufferSize: writeBufferBytesMax,
		CheckOrigin: func(r *http.Request) bool {
			return cfg.IsDebugMode
		},
	}

	return &WebsocketListener{
		upgrader: &websocketUpgrader,
		logger:   logger,
		register: make(chan *websocket.Conn, cfg.ClientsConnectionsMax),
		readCh:   make(chan []byte, readBufferBytesMax),
		writeCh:  make(chan []byte, writeBufferBytesMax),
	}
}

func (l *WebsocketListener) Close() {
	close(l.readCh)
	close(l.writeCh)
	close(l.register)
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	l.register <- conn

	ctx := r.Context()
	go l.handleReadConnection(ctx, conn)
	go l.handleWriteConnection(ctx, conn)

	return nil
}

func (l *WebsocketListener) RegisterChan() <-chan *websocket.Conn {
	return l.register
}

func (l *WebsocketListener) Messages() <-chan []byte {
	return l.readCh
}

func (l *WebsocketListener) handleReadConnection(ctx context.Context, conn *websocket.Conn) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			_, payload, err := conn.ReadMessage()
			if err != nil {
				switch {
				case websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure):
					l.logger.Errorf("failed to read a message: %s", err)
					return
				case websocket.IsUnexpectedCloseError(err, websocket.CloseMessage):
					l.logger.Info("received a close signal from the client")
					return
				default:
					l.logger.Errorf("unknown error while reading message: %s", err)
					return
				}
			}
			l.readCh <- payload
		}
	}
}

func (l *WebsocketListener) handleWriteConnection(ctx context.Context, conn *websocket.Conn) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case msg := <-l.writeCh:
			if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				l.logger.Errorf("failed to send a message: %s", err)
			}
		}
	}
}

func (l *WebsocketListener) Broadcast(obj any) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	l.writeCh <- payload
	return nil
}

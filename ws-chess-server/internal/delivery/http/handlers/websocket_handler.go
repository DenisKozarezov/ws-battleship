package handlers

import (
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
			return true
		},
	}

	return &WebsocketListener{
		upgrader: &websocketUpgrader,
		logger:   logger,
		register: make(chan *websocket.Conn, 10),
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

	go l.handleReadConnection(conn)
	go l.handleWriteConnection(conn)

	return nil
}

func (l *WebsocketListener) RegisterChan() chan *websocket.Conn {
	return l.register
}

func (l *WebsocketListener) handleReadConnection(conn *websocket.Conn) {
	for {
		messageType, payload, err := conn.ReadMessage()
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
		l.logger.Infof("message type = %d, payload = %s", messageType, string(payload))
		l.readCh <- payload
	}
}

func (l *WebsocketListener) handleWriteConnection(conn *websocket.Conn) {
	for msg := range l.writeCh {
		if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			l.logger.Errorf("failed to send a message: %s", err)
		}
	}
}

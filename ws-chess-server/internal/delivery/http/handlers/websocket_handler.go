package handlers

import (
	"log"
	"net/http"
	"ws-chess-server/internal/config"

	"github.com/gorilla/websocket"
)

const (
	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type WebsocketListener struct {
	upgrader   *websocket.Upgrader
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewWebsocketListener(cfg *config.AppConfig) *WebsocketListener {
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferBytesMax,
		WriteBufferSize: writeBufferBytesMax,
		CheckOrigin: func(r *http.Request) bool {
			return !cfg.IsDebugMode
		},
	}

	return &WebsocketListener{
		upgrader:   &websocketUpgrader,
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
	defer conn.Close()

	l.register <- conn

	go l.handleReadWS(conn)

	return nil
}

func (l *WebsocketListener) handleReadWS(conn *websocket.Conn) {
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		log.Printf("message type = %d, text = %s", messageType, string(payload))
		// обработка сообщения...
	}
}

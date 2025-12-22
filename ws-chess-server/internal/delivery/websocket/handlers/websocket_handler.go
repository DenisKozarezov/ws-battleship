package handlers

import (
	"encoding/json"
	"net/http"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/websocket/response"
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
	logger   logger.Logger
	joinCh   chan *domain.Client
	readCh   chan response.Event
	closeCh  chan struct{}
	writeCh  chan []byte
}

func NewWebsocketListener(cfg *config.AppConfig, logger logger.Logger) *WebsocketListener {
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
		joinCh:   make(chan *domain.Client, cfg.ClientsConnectionsMax),
		readCh:   make(chan response.Event, readBufferBytesMax),
		writeCh:  make(chan []byte, writeBufferBytesMax),
		closeCh:  make(chan struct{}),
	}
}

func (l *WebsocketListener) Close() {
	close(l.closeCh)
	close(l.joinCh)
}

func (l *WebsocketListener) HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	newClient := domain.NewClient(conn, domain.ParseClientMetadata(r))
	l.joinCh <- newClient

	go l.handleReadConnection(conn)
	go l.handleWriteConnection(conn)

	return nil
}

func (l *WebsocketListener) JoinChan() <-chan *domain.Client {
	return l.joinCh
}

func (l *WebsocketListener) Messages() <-chan response.Event {
	return l.readCh
}

func (l *WebsocketListener) handleReadConnection(conn *websocket.Conn) {
	defer close(l.readCh)

	for {
		select {
		case <-l.closeCh:
			l.logger.Info("listener received a closing signal, stopping reading messages...")
			return
		default:
			select {
			case <-l.closeCh:
				l.logger.Info("listener received a closing signal, stopping reading messages...")
				return
			default:
			}

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

			var event response.Event
			if err := json.Unmarshal(payload, &event); err != nil {
				l.logger.Errorf("failed to unmarshal message, discarding it: %s", err)
				continue
			}

			l.readCh <- event
		}
	}
}

func (l *WebsocketListener) handleWriteConnection(conn *websocket.Conn) {
	defer close(l.writeCh)

	for {
		select {
		case <-l.closeCh:
			l.logger.Info("listener received a closing signal, stopping writing messages...")
			return
		case msg := <-l.writeCh:
			select {
			case <-l.closeCh:
				l.logger.Info("listener received a closing signal, stopping reading messages...")
				return
			default:
			}

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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"ws-battleship-client/internal/config"
	"ws-battleship-client/internal/delivery/websocket/response"
	"ws-battleship-client/pkg/logger"

	"github.com/gorilla/websocket"
)

const (
	websocketProtocol = "ws"
	websocketEndpoint = "/ws"

	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type WebsocketClient struct {
	cfg    *config.AppConfig
	logger logger.Logger
	conn   *websocket.Conn
	readCh chan response.Event
}

func NewClient(cfg *config.AppConfig, logger logger.Logger) *WebsocketClient {
	return &WebsocketClient{
		cfg:    cfg,
		logger: logger,
		readCh: make(chan response.Event, readBufferBytesMax),
	}
}

func (c *WebsocketClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   readBufferBytesMax,
		WriteBufferSize:  writeBufferBytesMax,
	}

	serverUrl := fmt.Sprintf("%s://%s%s", websocketProtocol, c.cfg.ServerHost, websocketEndpoint)

	headers := make(http.Header)
	headers.Set("X-Nickname", "player123")

	conn, _, err := dialer.DialContext(ctx, serverUrl, headers)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	const pongTimeout = time.Second * 10
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(pongTimeout))
	})
	conn.SetPongHandler(func(appData string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	})
	c.conn = conn

	go c.handleReadConnection(ctx, conn)

	return nil
}

func (c *WebsocketClient) Shutdown() error {
	return c.conn.Close()
}

func (c *WebsocketClient) Messages() <-chan response.Event {
	return c.readCh
}

func (c *WebsocketClient) handleReadConnection(ctx context.Context, conn *websocket.Conn) {
	defer close(c.readCh)

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
					c.logger.Errorf("failed to read a message: %s", err)
					return
				case websocket.IsUnexpectedCloseError(err, websocket.CloseMessage):
					c.logger.Info("received a close signal from the server")
					return
				default:
					c.logger.Errorf("unknown error while reading message: %s", err)
					return
				}
			}

			var event response.Event
			if err := json.Unmarshal(payload, &event); err != nil {
				c.logger.Errorf("failed to unmarshal message, discarding it: %s", err)
				continue
			}

			c.readCh <- event
		}
	}
}

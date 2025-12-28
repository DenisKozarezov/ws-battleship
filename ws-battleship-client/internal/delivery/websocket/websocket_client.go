package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"ws-battleship-client/internal/config"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/gorilla/websocket"
)

const (
	websocketProtocol = "ws"
	websocketEndpoint = "/ws"
)

type WebsocketClient struct {
	cfg    *config.AppConfig
	logger logger.Logger
	conn   *websocket.Conn
	readCh chan events.Event
}

func NewClient(cfg *config.AppConfig, logger logger.Logger) *WebsocketClient {
	return &WebsocketClient{
		cfg:    cfg,
		logger: logger,
		readCh: make(chan events.Event, events.ReadBufferBytesMax),
	}
}

func (c *WebsocketClient) Connect(ctx context.Context, metadata events.ClientMetadata) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   events.ReadBufferBytesMax,
		WriteBufferSize:  events.WriteBufferBytesMax,
	}

	serverUrl := fmt.Sprintf("%s://%s%s", websocketProtocol, c.cfg.ServerHost, websocketEndpoint)

	conn, _, err := dialer.DialContext(ctx, serverUrl, events.ParseClientMetadataToHeaders(metadata))
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

func (c *WebsocketClient) Messages() <-chan events.Event {
	return c.readCh
}

func (c *WebsocketClient) handleReadConnection(ctx context.Context, conn *websocket.Conn) {
	defer close(c.readCh)

	for {
		if err := ctx.Err(); err != nil {
			c.logger.Info("client received a closing signal, stopping reading messages...")
			return
		}

		select {
		case <-ctx.Done():
			c.logger.Info("client received a closing signal, stopping reading messages...")
			return
		default:
			_, payload, err := conn.ReadMessage()
			if err != nil {
				if err := ctx.Err(); err != nil {
					c.logger.Info("client received a closing signal, stopping reading messages...")
					return
				}

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

			var event events.Event
			if err := json.Unmarshal(payload, &event); err != nil {
				c.logger.Errorf("failed to unmarshal message, discarding it: %s", err)
				continue
			}

			c.readCh <- event
		}
	}
}

package client

import (
	"context"
	"fmt"
	"time"
	"ws-chess-client/internal/config"
	"ws-chess-client/internal/delivery/http/middleware"

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
	logger middleware.Logger
	conn   *websocket.Conn
	readCh chan []byte
}

func NewClient(cfg *config.AppConfig, logger middleware.Logger) *WebsocketClient {
	return &WebsocketClient{
		cfg:    cfg,
		logger: logger,
		readCh: make(chan []byte, readBufferBytesMax),
	}
}

func (c *WebsocketClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   readBufferBytesMax,
		WriteBufferSize:  writeBufferBytesMax,
	}

	serverUrl := fmt.Sprintf("%s://%s%s", websocketProtocol, c.cfg.ServerHost, websocketEndpoint)

	conn, _, err := dialer.DialContext(ctx, serverUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	const pongTimeout = time.Millisecond * 100
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(pongTimeout))
	})
	c.conn = conn

	go c.handleReadConnection(ctx, conn)

	return nil
}

func (c *WebsocketClient) Shutdown() error {
	close(c.readCh)
	return c.conn.Close()
}

func (c *WebsocketClient) Messages() chan []byte {
	return c.readCh
}

func (c *WebsocketClient) handleReadConnection(ctx context.Context, conn *websocket.Conn) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			messageType, payload, err := conn.ReadMessage()
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
			c.logger.Infof("message type = %d, payload = %s", messageType, string(payload))
			c.readCh <- payload
		}
	}
}

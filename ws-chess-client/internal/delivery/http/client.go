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
	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type Client struct {
	cfg             *config.AppConfig
	logger          middleware.Logger
	conn            *websocket.Conn
	readCh, writeCh chan []byte
}

func NewClient(cfg *config.AppConfig, logger middleware.Logger) *Client {
	return &Client{
		cfg:     cfg,
		logger:  logger,
		readCh:  make(chan []byte, readBufferBytesMax),
		writeCh: make(chan []byte, writeBufferBytesMax),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   readBufferBytesMax,
		WriteBufferSize:  writeBufferBytesMax,
	}

	serverUrl := fmt.Sprintf("ws://%s/ws", c.cfg.ServerPort)

	conn, _, err := dialer.DialContext(ctx, serverUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	c.conn = conn

	go c.handleReadConnection(conn)

	return nil
}

func (c *Client) Shutdown() error {
	close(c.readCh)
	close(c.writeCh)
	return c.conn.Close()
}

func (c *Client) handleReadConnection(conn *websocket.Conn) {
	for {
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

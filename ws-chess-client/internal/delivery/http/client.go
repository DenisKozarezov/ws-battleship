package client

import (
	"context"
	"fmt"
	"time"
	"ws-chess-client/internal/config"

	"github.com/gorilla/websocket"
)

const (
	readBufferBytesMax  = 1024
	writeBufferBytesMax = 1024
)

type Client struct {
	conn *websocket.Conn
	cfg  *config.AppConfig
}

func NewClient(cfg *config.AppConfig) *Client {
	return &Client{
		cfg: cfg,
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

	return nil
}

func (c *Client) Shutdown() error {
	return c.conn.Close()
}

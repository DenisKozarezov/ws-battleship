package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	clientID uuid.UUID
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:     conn,
		clientID: uuid.New(),
	}
}

func (c *Client) ID() string {
	return c.clientID.String()
}

func (c *Client) Equal(rhs *Client) bool {
	if rhs == nil {
		return false
	}
	return c.ID() == rhs.ID()
}

func (c *Client) Compare(rhs *Client) int {
	if rhs == nil {
		return -1
	}
	return strings.Compare(c.ID(), rhs.ID())
}

func (c *Client) Ping() error {
	const pingTimeout = time.Millisecond * 100
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingTimeout))
}

func (c *Client) SendMessage(obj any) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, payload)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

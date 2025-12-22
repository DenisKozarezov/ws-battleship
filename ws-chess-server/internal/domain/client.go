package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	clientID uuid.UUID
	nickname string
}

func NewClient(conn *websocket.Conn, metadata ClientMetadata) *Client {
	return &Client{
		conn:     conn,
		clientID: uuid.New(),
		nickname: metadata.Nickname,
	}
}

func (c *Client) ID() string {
	return c.clientID.String()
}

func (c *Client) Nickname() string {
	return c.nickname
}

func (c *Client) String() string {
	return fmt.Sprintf(`'%s' [%s]`, c.Nickname(), c.ID())
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

package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ws-chess-server/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientID = string

type Client struct {
	conn    *websocket.Conn
	logger  logger.Logger
	closeCh chan struct{}

	clientID ClientID
	nickname string
}

func NewClient(conn *websocket.Conn, logger logger.Logger, metadata ClientMetadata) *Client {
	return &Client{
		conn:    conn,
		logger:  logger,
		closeCh: make(chan struct{}),

		clientID: uuid.New().String(),
		nickname: metadata.Nickname,
	}
}

func (c *Client) ID() ClientID {
	return c.clientID
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
	close(c.closeCh)
	return c.conn.Close()
}

func (c *Client) ReadMessage(ctx context.Context, messagesCh chan Event) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-c.closeCh:
			c.logger.Infof("client id=%s received a closing signal, stopping reading messages...", c.ID())
			return
		default:
			_, payload, err := c.conn.ReadMessage()
			if err != nil {
				select {
				case <-c.closeCh:
					c.logger.Infof("client id=%s received a closing signal, stopping reading messages...", c.ID())
					return
				default:
				}

				switch {
				case websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure):
					c.logger.Errorf("failed to read a message: %s", err)
					return
				case websocket.IsUnexpectedCloseError(err, websocket.CloseMessage):
					c.logger.Info("received a close signal from the client")
					return
				default:
					c.logger.Errorf("unknown error while reading message: %s", err)
					return
				}
			}

			var event Event
			if err := json.Unmarshal(payload, &event); err != nil {
				c.logger.Errorf("failed to unmarshal message, discarding it: %s", err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case <-c.closeCh:
				c.logger.Infof("client id=%s received a closing signal, stopping reading messages...", c.ID())
				return
			case messagesCh <- event:
			}
		}
	}
}

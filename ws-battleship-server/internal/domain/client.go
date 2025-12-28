package domain

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientID = string

type Client struct {
	conn *websocket.Conn

	logger  logger.Logger
	once    sync.Once
	closeCh chan struct{}
	writeCh chan []byte

	clientID ClientID
	metadata events.ClientMetadata
}

func NewClient(conn *websocket.Conn, logger logger.Logger, metadata events.ClientMetadata) *Client {
	return &Client{
		conn:    conn,
		logger:  logger,
		closeCh: make(chan struct{}),
		writeCh: make(chan []byte, events.WriteBufferBytesMax),

		clientID: uuid.New().String(),
		metadata: metadata,
	}
}

func (c *Client) ID() ClientID {
	return c.clientID
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
	const pingTimeout = time.Second * 5
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingTimeout))
}

func (c *Client) SendMessage(eventType events.EventType, obj any) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	payload, err = json.Marshal(events.Event{
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      payload,
	})
	if err != nil {
		return err
	}

	select {
	case <-c.closeCh:
		return nil
	case c.writeCh <- payload:
	}
	return nil
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.closeCh)
		if err := c.conn.Close(); err != nil {
			c.logger.Errorf("failed to close a client id=%s: %s", c.ID(), err)
		}
	})
}

func (c *Client) ReadMessages(ctx context.Context, messagesCh chan<- events.Event) {
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

			var event events.Event
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

func (c *Client) WriteMessages(ctx context.Context) {
	defer close(c.writeCh)

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-c.closeCh:
			c.logger.Infof("client id=%s received a closing signal, stopping writing messages...", c.ID())
			return
		case msg := <-c.writeCh:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				c.logger.Errorf("failed to send a message to client id=%s: %s", c.ID(), err)
			}
		}
	}
}

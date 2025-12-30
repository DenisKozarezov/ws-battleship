package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
	"ws-battleship-server/internal/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	conn *websocket.Conn

	logger  logger.Logger
	once    sync.Once
	closeCh chan struct{}
	writeCh chan []byte

	clientID domain.ClientID
}

func NewWebsocketClient(conn *websocket.Conn, logger logger.Logger) *WebsocketClient {
	return &WebsocketClient{
		conn:    conn,
		logger:  logger,
		closeCh: make(chan struct{}),
		writeCh: make(chan []byte, events.WriteBufferBytesMax),

		clientID: uuid.New().String(),
	}
}

func (c *WebsocketClient) ID() domain.ClientID {
	return c.clientID
}

func (c *WebsocketClient) Equal(rhs *WebsocketClient) bool {
	if rhs == nil {
		return false
	}
	return c.ID() == rhs.ID()
}

func (c *WebsocketClient) Compare(rhs *WebsocketClient) int {
	if rhs == nil {
		return -1
	}
	return strings.Compare(c.ID(), rhs.ID())
}

func (c *WebsocketClient) Ping() error {
	const pingTimeout = time.Second * 5
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingTimeout))
}

func (c *WebsocketClient) SendMessage(e events.Event) error {
	payload, err := json.Marshal(e)
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

func (c *WebsocketClient) Close() {
	c.once.Do(func() {
		close(c.closeCh)
		if err := c.conn.Close(); err != nil {
			c.logger.Errorf("failed to close a client id=%s: %s", c.ID(), err)
		}
	})
}

func (c *WebsocketClient) ReadMessages(ctx context.Context, messagesCh chan<- events.Event) {
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

func (c *WebsocketClient) WriteMessages(ctx context.Context) {
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

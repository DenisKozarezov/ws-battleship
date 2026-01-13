package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ws-battleship-client/internal/config"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/gorilla/websocket"
)

const (
	websocketProtocol = "ws"
	websocketEndpoint = "/ws"
)

type WebsocketClient struct {
	once sync.Once
	wg   sync.WaitGroup
	ctx  context.Context

	cfg     *config.AppConfig
	logger  logger.Logger
	conn    *websocket.Conn
	readCh  chan events.Event
	writeCh chan []byte
	closeCh chan struct{}
}

func NewClient(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *WebsocketClient {
	return &WebsocketClient{
		ctx:     ctx,
		cfg:     cfg,
		logger:  logger,
		readCh:  make(chan events.Event, events.ReadBufferBytesMax),
		writeCh: make(chan []byte, events.WriteBufferBytesMax),
		closeCh: make(chan struct{}),
	}
}

func (c *WebsocketClient) Connect(ctx context.Context, metadata domain.ClientMetadata) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   events.ReadBufferBytesMax,
		WriteBufferSize:  events.WriteBufferBytesMax,
	}

	serverUrl := fmt.Sprintf("%s://%s%s", websocketProtocol, c.cfg.ServerHost, websocketEndpoint)

	conn, _, err := dialer.DialContext(ctx, serverUrl, domain.ParseClientMetadataToHeaders(metadata))
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

	c.wg.Add(2)
	go func(wg *sync.WaitGroup, conn *websocket.Conn) {
		defer wg.Done()
		c.ReadMessages(c.ctx, conn)
	}(&c.wg, conn)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		c.WriteMessages(c.ctx)
	}(&c.wg)

	return nil
}

func (c *WebsocketClient) Shutdown() error {
	c.once.Do(func() {
		close(c.closeCh)
		if err := c.conn.Close(); err != nil {
			c.logger.Errorf("failed to close a websocket client: %s", err)
		}
	})
	c.wg.Wait()
	return nil
}

func (c *WebsocketClient) Messages() <-chan events.Event {
	return c.readCh
}

func (c *WebsocketClient) SendMessage(e events.Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}

	select {
	case <-c.closeCh:
		return nil
	default:
	}

	select {
	case <-c.closeCh:
		return nil
	case c.writeCh <- payload:
	}
	return nil
}

func (c *WebsocketClient) ReadMessages(ctx context.Context, conn *websocket.Conn) {
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
			c.logger.Info("client received a closing signal, stopping writing messages...")
			return
		case msg := <-c.writeCh:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				c.logger.Errorf("failed to send a message to client: %s", err)
			}
		}
	}
}

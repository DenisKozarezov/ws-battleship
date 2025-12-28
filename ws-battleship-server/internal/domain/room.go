package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/google/uuid"
)

type Room struct {
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	clients map[string]*Client
	readCh  chan events.Event

	id     string
	cfg    *config.AppConfig
	logger logger.Logger
}

func NewRoom(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *Room {
	ctx, cancel := context.WithCancel(ctx)

	r := &Room{
		ctx:     ctx,
		cancel:  cancel,
		clients: make(map[string]*Client, cfg.RoomCapacityMax),
		readCh:  make(chan events.Event, events.ReadBufferBytesMax),
		id:      uuid.New().String(),
		cfg:     cfg,
		logger:  logger,
	}

	r.wg.Add(2)
	r.wg.Go(func() {
		defer r.wg.Done()
		r.handleConnections(ctx)
	})
	r.wg.Go(func() {
		defer r.wg.Done()
		r.pingClients(ctx)
	})

	return r
}

func (r *Room) ID() string {
	return r.id
}

func (c *Room) Equal(rhs *Room) bool {
	if rhs == nil {
		return false
	}
	return c.ID() == rhs.ID()
}

func (c *Room) Compare(rhs *Room) int {
	if rhs == nil {
		return -1
	}
	return strings.Compare(c.ID(), rhs.ID())
}

func (r *Room) Close() error {
	r.logger.Infof("room id=%s is closing...", r.ID())

	r.once.Do(func() {
		r.cancel()
		close(r.readCh)
	})

	for _, client := range r.clients {
		if err := r.UnregisterClient(client); err != nil {
			return fmt.Errorf("failed to unregister a client: %w", err)
		}
	}
	r.wg.Wait()

	r.logger.Infof("all clients in room id=%s were unregistered", r.ID())
	r.logger.Infof("room id=%s is closed", r.ID())

	return nil
}

func (r *Room) IsFull() bool {
	return r.Capacity() == int(r.cfg.RoomCapacityMax)
}

func (r *Room) RegisterNewClient(newClient *Client) {
	if r.IsFull() {
		r.logger.Infof("room id=%s is full", r.ID())
		return
	}

	r.mu.Lock()
	r.clients[newClient.ID()] = newClient
	r.mu.Unlock()

	r.wg.Add(2)
	go func(wg *sync.WaitGroup, client *Client) {
		defer wg.Done()
		client.ReadMessages(r.ctx, r.readCh)
	}(&r.wg, newClient)

	go func(wg *sync.WaitGroup, client *Client) {
		defer wg.Done()
		client.WriteMessages(r.ctx)
	}(&r.wg, newClient)

	r.logger.Infof("client %s is connected to room id=%s [players: %d]", newClient.String(), r.ID(), r.Capacity())
}

func (r *Room) UnregisterClient(client *Client) error {
	client.Close()

	r.mu.Lock()
	delete(r.clients, client.ID())
	r.mu.Unlock()

	r.logger.Infof("client %s was unregistered from the room id=%s [players: %d]", client.String(), r.ID(), r.Capacity())

	return nil
}

func (r *Room) Broadcast(eventType events.EventType, obj any) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		if err := client.SendMessage(eventType, obj); err != nil {
			r.logger.Errorf("failed to send a broadcast message to client id=%s", client.ID())
		}
	}
}

func (r *Room) Capacity() (capacity int) {
	r.mu.RLock()
	capacity = len(r.clients)
	r.mu.RUnlock()
	return
}

func (r *Room) Messages() chan events.Event {
	return r.readCh
}

func (r *Room) handleConnections(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case msg, opened := <-r.readCh:
			if opened {
				r.handleMessage(msg)
			}
		}
	}
}

func (r *Room) handleMessage(event events.Event) {
	r.logger.Debug("Event Type: %d; Timestamp: %s; Payload: %s", event.Type, event.Timestamp, string(event.Data))
}

func (r *Room) pingClients(ctx context.Context) {
	pingTicker := time.NewTicker(r.cfg.KeepAlivePeriod)
	defer pingTicker.Stop()

	deadClients := make(chan *Client, r.cfg.RoomCapacityMax)
	defer close(deadClients)

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		// We should periodically send a ping-message to all clients just to be ensured, that the clients
		// are still alive. If no, then the server collects dead clients to a special queue for further
		// unregistering.
		case <-pingTicker.C:
			r.mu.RLock()
			for _, client := range r.clients {
				go r.pingClient(client, deadClients)
			}
			r.mu.RUnlock()

		// We must kick potentially dead clients who didn't response to our ping-message. There are literally zero
		// reasons to keep stalled connections alive, so the server deallocates them for other needs.
		case deadClient := <-deadClients:
			r.logger.Infof("client %s didn't response to ping and was declared as potentially dead by the server, unregistering it...", deadClient.String())
			if err := r.UnregisterClient(deadClient); err != nil {
				r.logger.Errorf("failed to disconnect a dead client: %s", err)
			}
		}
	}
}

func (r *Room) pingClient(client *Client, deadClients chan<- *Client) {
	if err := client.Ping(); err != nil {
		r.logger.Errorf("failed to ping a client id=%s: %s", client.ID(), err)
		deadClients <- client
	}
}

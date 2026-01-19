package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/delivery/websocket"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/google/uuid"
)

type Room struct {
	once sync.Once
	ctx  context.Context
	mu   sync.RWMutex
	wg   sync.WaitGroup

	clients    map[string]websocket.Client
	joinCh     chan websocket.Client
	leaveCh    chan websocket.Client
	messagesCh chan events.Event
	closeCh    chan struct{}

	id     string
	cfg    *config.AppConfig
	logger logger.Logger

	clientJoinedHandler func(websocket.Client)
	clientLeftHandler   func(websocket.Client)
}

func NewRoom(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *Room {
	r := &Room{
		ctx:        ctx,
		clients:    make(map[string]websocket.Client, cfg.RoomCapacityMax),
		joinCh:     make(chan websocket.Client, cfg.RoomCapacityMax),
		leaveCh:    make(chan websocket.Client, cfg.RoomCapacityMax),
		messagesCh: make(chan events.Event, events.ReadBufferBytesMax),
		closeCh:    make(chan struct{}),
		id:         uuid.New().String(),
		cfg:        cfg,
		logger:     logger,
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

func (r *Room) Equal(rhs *Room) bool {
	if rhs == nil {
		return false
	}
	return r.ID() == rhs.ID()
}

func (r *Room) Compare(rhs *Room) int {
	if rhs == nil {
		return -1
	}
	return strings.Compare(r.ID(), rhs.ID())
}

func (r *Room) Close() error {
	r.once.Do(func() {
		close(r.closeCh)
		r.logger.Infof("room id=%s [clients: %d] is closing...", r.ID(), r.Capacity())

		for _, client := range r.GetClients() {
			if err := r.unregisterClient(client); err != nil {
				r.logger.Errorf("failed to unregister a client_id=%s: %s", client.ID(), err)
			}
		}
		r.wg.Wait()

		close(r.messagesCh)
	})

	r.logger.Infof("all clients in room id=%s were unregistered", r.ID())
	r.logger.Infof("room id=%s is closed", r.ID())

	return nil
}

func (r *Room) IsFull() bool {
	return r.Capacity() >= int(r.cfg.RoomCapacityMax)
}

func (r *Room) JoinNewClient(joinedClient websocket.Client) error {
	select {
	case <-r.closeCh:
		return ErrRoomIsClosed
	case r.joinCh <- joinedClient:
	default:
	}
	return nil
}

func (r *Room) LeaveClient(client websocket.Client) {
	select {
	case <-r.closeCh:
	case r.leaveCh <- client:
	default:
	}
}

func (r *Room) Capacity() (capacity int) {
	r.mu.RLock()
	capacity = len(r.clients)
	r.mu.RUnlock()
	return
}

func (r *Room) SendMessageToClient(clientID string, msg events.Event) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, found := r.clients[clientID]; !found {
		return ErrPlayerNotExist
	}

	return r.clients[clientID].SendMessage(msg)
}

func (r *Room) Broadcast(e events.Event) error {
	for _, client := range r.GetClients() {
		if err := client.SendMessage(e); err != nil {
			return fmt.Errorf("failed to send a broadcast message to client id=%s", client.ID())
		}
	}
	return nil
}

func (r *Room) GetClients() []websocket.Client {
	r.mu.RLock()
	clients := make([]websocket.Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}
	r.mu.RUnlock()
	return clients
}

func (r *Room) Events() <-chan events.Event {
	return r.messagesCh
}

func (r *Room) SetClientJoinedHandler(fn func(websocket.Client)) {
	r.clientJoinedHandler = fn
}

func (r *Room) SetClientLeftHandler(fn func(websocket.Client)) {
	r.clientLeftHandler = fn
}

func (r *Room) handleConnections(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case <-r.closeCh:
			return

		case joinedClient := <-r.joinCh:
			if err := r.onClientJoinedHandler(joinedClient); err != nil {
				r.logger.Error(err)
			}

		case leftClient := <-r.leaveCh:
			if err := r.onClientLeftHandler(leftClient); err != nil {
				r.logger.Error(err)
			}
		}
	}
}

func (r *Room) pingClients(ctx context.Context) {
	pingTicker := time.NewTicker(r.cfg.KeepAlivePeriod)
	defer pingTicker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case <-r.closeCh:
			return

		// Health check: ping all connected clients at regular intervals.
		// Clients that fail to respond are considered dead and disconnected.
		// This prevents resource leaks from abandoned connections.
		case <-pingTicker.C:
			for _, client := range r.GetClients() {
				if err := client.Ping(); err != nil {
					r.logger.Errorf("failed to ping a client id=%s: %s", client.ID(), err)
					r.LeaveClient(client)
				}
			}
		}
	}
}

func (r *Room) registerNewClient(newClient websocket.Client) error {
	if r.IsFull() {
		return ErrRoomIsFull
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, found := r.clients[newClient.ID()]; found {
		return ErrPlayerAlreadyInRoom
	}

	r.clients[newClient.ID()] = newClient

	r.wg.Add(2)
	go func(wg *sync.WaitGroup, client websocket.Client) {
		defer wg.Done()
		client.ReadMessages(r.ctx, r.messagesCh)
	}(&r.wg, newClient)

	go func(wg *sync.WaitGroup, client websocket.Client) {
		defer wg.Done()
		client.WriteMessages(r.ctx)
	}(&r.wg, newClient)

	return nil
}

func (r *Room) unregisterClient(client websocket.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.clients[client.ID()]; !found {
		return ErrPlayerNotExist
	}

	client.Close()
	delete(r.clients, client.ID())

	return nil
}

func (r *Room) onClientJoinedHandler(joinedClient websocket.Client) error {
	if err := r.registerNewClient(joinedClient); err != nil {
		return fmt.Errorf("failed to register new client_id=%s: %s", joinedClient.ID(), err)
	}

	if r.clientJoinedHandler != nil {
		r.clientJoinedHandler(joinedClient)
	}
	return nil
}

func (r *Room) onClientLeftHandler(leftClient websocket.Client) error {
	if err := r.unregisterClient(leftClient); err != nil {
		return fmt.Errorf("failed to unregister client_id=%s: %s", leftClient.ID(), err)
	}

	if r.clientLeftHandler != nil {
		r.clientLeftHandler(leftClient)
	}

	return nil
}

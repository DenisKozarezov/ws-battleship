package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/domain"
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

	players map[string]*Player
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
		players: make(map[string]*Player, cfg.RoomCapacityMax),
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
		r.pingPlayers(ctx)
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
	r.logger.Infof("room id=%s [players: %d] is closing...", r.ID(), r.Capacity())

	r.once.Do(func() {
		r.cancel()
		close(r.readCh)
	})

	for _, player := range r.players {
		if err := r.UnregisterPlayer(player); err != nil {
			return fmt.Errorf("failed to unregister a player: %w", err)
		}
	}
	r.wg.Wait()

	r.logger.Infof("all players in room id=%s were unregistered", r.ID())
	r.logger.Infof("room id=%s is closed", r.ID())

	return nil
}

func (r *Room) IsFull() bool {
	return r.Capacity() >= int(r.cfg.RoomCapacityMax)
}

func (r *Room) RegisterNewPlayer(newPlayer *Player) {
	if r.IsFull() {
		r.logger.Infof("room id=%s is full", r.ID())
		return
	}

	r.mu.Lock()
	r.players[newPlayer.ID()] = newPlayer
	r.mu.Unlock()

	r.wg.Add(2)
	go func(wg *sync.WaitGroup, player *Player) {
		defer wg.Done()
		player.ReadMessages(r.ctx, r.readCh)
	}(&r.wg, newPlayer)

	go func(wg *sync.WaitGroup, player *Player) {
		defer wg.Done()
		player.WriteMessages(r.ctx)
	}(&r.wg, newPlayer)

	r.logger.Infof("player %s is connected to room id=%s [players: %d]", newPlayer.String(), r.ID(), r.Capacity())
}

func (r *Room) UnregisterPlayer(player *Player) error {
	player.Close()

	r.mu.Lock()
	delete(r.players, player.ID())
	r.mu.Unlock()

	r.logger.Infof("player %s was unregistered from the room id=%s [players: %d]", player.String(), r.ID(), r.Capacity())

	return nil
}

func (r *Room) Broadcast(eventType events.EventType, obj any) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, player := range r.players {
		if err := player.SendMessage(eventType, obj); err != nil {
			r.logger.Errorf("failed to send a broadcast message to player id=%s", player.ID())
		}
	}
}

func (r *Room) Capacity() (capacity int) {
	r.mu.RLock()
	capacity = len(r.players)
	r.mu.RUnlock()
	return
}

func (r *Room) StartMatch() {
	r.logger.Infof("room id=%s is starting a match [players: %d]", r.ID(), r.Capacity())

	playerModels := make(map[string]*domain.PlayerModel, len(r.players))
	for _, player := range r.players {
		playerModels[player.ID()] = player.Model
	}

	gameModel := domain.NewGameModel(playerModels)

	r.Broadcast(events.GameStartEvent, gameModel)
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

func (r *Room) pingPlayers(ctx context.Context) {
	pingTicker := time.NewTicker(r.cfg.KeepAlivePeriod)
	defer pingTicker.Stop()

	deadPlayers := make(chan *Player, r.cfg.RoomCapacityMax)
	defer close(deadPlayers)

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
			for _, player := range r.players {
				go r.pingPlayer(player, deadPlayers)
			}
			r.mu.RUnlock()

		// We must kick potentially dead clients who didn't response to our ping-message. There are literally zero
		// reasons to keep stalled connections alive, so the server deallocates them for other needs.
		case deadPlayer := <-deadPlayers:
			r.logger.Infof("player %s didn't response to ping and was declared as potentially dead by the server, unregistering it...", deadPlayer.String())
			if err := r.UnregisterPlayer(deadPlayer); err != nil {
				r.logger.Errorf("failed to disconnect a dead player: %s", err)
			}
		}
	}
}

func (r *Room) pingPlayer(player *Player, deadPlayer chan<- *Player) {
	if err := player.Ping(); err != nil {
		r.logger.Errorf("failed to ping a player id=%s: %s", player.ID(), err)
		deadPlayer <- player
	}
}

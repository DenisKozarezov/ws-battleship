package domain

import (
	"context"
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

	players    map[string]*Player
	joinCh     chan *Player
	leaveCh    chan *Player
	messagesCh chan events.Event

	id     string
	cfg    *config.AppConfig
	logger logger.Logger
}

func NewRoom(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *Room {
	ctx, cancel := context.WithCancel(ctx)

	r := &Room{
		ctx:        ctx,
		cancel:     cancel,
		players:    make(map[string]*Player, cfg.RoomCapacityMax),
		joinCh:     make(chan *Player, cfg.RoomCapacityMax),
		leaveCh:    make(chan *Player, cfg.RoomCapacityMax),
		messagesCh: make(chan events.Event, events.ReadBufferBytesMax),
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
		close(r.messagesCh)
		close(r.joinCh)
		close(r.leaveCh)
	})

	for _, player := range r.players {
		if err := r.unregisterPlayer(player); err != nil {
			return err
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

func (r *Room) JoinNewPlayer(newPlayer *Player) error {
	if r.IsFull() {
		return ErrRoomIsFull
	}

	r.joinCh <- newPlayer
	return nil
}

func (r *Room) LeavePlayer(player *Player) {
	r.leaveCh <- player
}

func (r *Room) Broadcast(e events.Event) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, player := range r.players {
		if err := player.SendMessage(e); err != nil {
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

	e, _ := events.NewGameStartEvent(gameModel)

	r.Broadcast(e)
}

func (r *Room) handleConnections(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case joinedPlayer, opened := <-r.joinCh:
			if opened {
				if err := r.registerNewPlayer(joinedPlayer); err == nil {
					r.logger.Errorf("failed to register new player: %s", err)
					return
				}

				e, _ := events.NewPlayerJoinedEvent(joinedPlayer.Model)
				r.Broadcast(e)

				r.logger.Infof("player %s is connected to room id=%s [players: %d]", joinedPlayer.String(), r.ID(), r.Capacity())
			}

		case leavedPlayer, opened := <-r.leaveCh:
			if opened {
				if err := r.unregisterPlayer(leavedPlayer); err != nil {
					r.logger.Errorf("failed to unregister player: %s", err)
					return
				}

				e, _ := events.NewPlayerLeavedEvent(leavedPlayer.Model)
				r.Broadcast(e)

				r.logger.Infof("player %s was unregistered from the room id=%s [players: %d]", leavedPlayer.String(), r.ID(), r.Capacity())
			}

		case msg, opened := <-r.messagesCh:
			if opened {
				r.handleEvent(msg)
			}
		}
	}
}

func (r *Room) pingPlayers(ctx context.Context) {
	pingTicker := time.NewTicker(r.cfg.KeepAlivePeriod)
	defer pingTicker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		// We should periodically send a ping-message to all clients just to be ensured, that the clients
		// are still alive. If no, then the server unregisters potentially dead clients. There are literally
		// zero reasons to keep stalled connections alive, so the server deallocates them for other needs.
		case <-pingTicker.C:
			r.mu.RLock()
			for _, player := range r.players {
				go r.pingPlayer(player)
			}
			r.mu.RUnlock()
		}
	}
}

func (r *Room) pingPlayer(player *Player) {
	if err := player.Ping(); err != nil {
		r.logger.Errorf("failed to ping a player id=%s: %s", player.ID(), err)
		r.LeavePlayer(player)
	}
}

func (r *Room) registerNewPlayer(newPlayer *Player) error {
	if r.IsFull() {
		return ErrRoomIsFull
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.players[newPlayer.ID()]; found {
		return ErrPlayerAlreadyInRoom
	}

	r.players[newPlayer.ID()] = newPlayer

	r.wg.Add(2)
	go func(wg *sync.WaitGroup, player *Player) {
		defer wg.Done()
		player.ReadMessages(r.ctx, r.messagesCh)
	}(&r.wg, newPlayer)

	go func(wg *sync.WaitGroup, player *Player) {
		defer wg.Done()
		player.WriteMessages(r.ctx)
	}(&r.wg, newPlayer)

	return nil
}

func (r *Room) unregisterPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.players[player.ID()]; !found {
		return ErrPlayerNotExist
	}

	player.Close()
	delete(r.players, player.ID())

	return nil
}

func (r *Room) handleEvent(e events.Event) {
	r.logger.Debug("[room: %s] type: %d; timestamp: %s", r.ID(), e.Type, e.Timestamp)

}

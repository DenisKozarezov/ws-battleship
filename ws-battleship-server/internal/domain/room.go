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
	once sync.Once
	ctx  context.Context
	mu   sync.RWMutex
	wg   sync.WaitGroup

	players    map[string]*Player
	joinCh     chan *Player
	leaveCh    chan *Player
	messagesCh chan events.Event
	closeCh    chan struct{}

	id     string
	cfg    *config.AppConfig
	logger logger.Logger

	playerJoinedHandler func(joinedPlayer *Player)
	playerLeftHandler   func(leftPlayer *Player)
}

func NewRoom(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *Room {
	r := &Room{
		ctx:        ctx,
		players:    make(map[string]*Player, cfg.RoomCapacityMax),
		joinCh:     make(chan *Player, cfg.RoomCapacityMax),
		leaveCh:    make(chan *Player, cfg.RoomCapacityMax),
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
	r.once.Do(func() {
		close(r.closeCh)
		r.logger.Infof("room id=%s [players: %d] is closing...", r.ID(), r.Capacity())
	})

	for _, player := range r.GetPlayers() {
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
	select {
	case <-r.closeCh:
		return ErrRoomIsClosed
	case r.joinCh <- newPlayer:
	default:
	}
	return nil
}

func (r *Room) LeavePlayer(player *Player) {
	select {
	case <-r.closeCh:
	case r.leaveCh <- player:
	default:
	}
}

func (r *Room) Capacity() (capacity int) {
	r.mu.RLock()
	capacity = len(r.players)
	r.mu.RUnlock()
	return
}

func (r *Room) SendNotification(msg string, notificationType events.ChatMessageType) error {
	event, err := events.NewChatNotificationEvent(msg, notificationType)
	if err != nil {
		return fmt.Errorf("couldn't send a chat notification: %w", err)
	}
	return r.Broadcast(event)
}

func (r *Room) SendMessageToClient(clientID string, msg events.Event) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, found := r.players[clientID]; !found {
		return ErrPlayerNotExist
	}

	return r.players[clientID].SendMessage(msg)
}

func (r *Room) Broadcast(e events.Event) error {
	for _, player := range r.GetPlayers() {
		if err := player.SendMessage(e); err != nil {
			return fmt.Errorf("failed to send a broadcast message to player id=%s", player.ID())
		}
	}
	return nil
}

func (r *Room) GetPlayers() []*Player {
	r.mu.RLock()
	players := make([]*Player, 0, len(r.players))
	for _, player := range r.players {
		players = append(players, player)
	}
	r.mu.RUnlock()
	return players
}

func (r *Room) Events() <-chan events.Event {
	return r.messagesCh
}

func (r *Room) SetPlayerJoinedHandler(fn func(joinedPlayer *Player)) {
	r.playerJoinedHandler = fn
}

func (r *Room) SetPlayerLeftHandler(fn func(leftPlayer *Player)) {
	r.playerLeftHandler = fn
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

		case joinedPlayer, opened := <-r.joinCh:
			if !opened {
				return
			}

			if err := r.onPlayerJoinedHandler(joinedPlayer); err != nil {
				r.logger.Error(err)
			}

		case leftPlayer, opened := <-r.leaveCh:
			if !opened {
				return
			}

			if err := r.onPlayerLeftHandler(leftPlayer); err != nil {
				r.logger.Error(err)
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

		case <-r.closeCh:
			return

		// We should periodically send a ping-message to all clients just to be ensured, that the clients
		// are still alive. If no, then the server unregisters potentially dead clients. There are literally
		// zero reasons to keep stalled connections alive, so the server deallocates them for other needs.
		case <-pingTicker.C:
			for _, player := range r.GetPlayers() {
				if err := player.Ping(); err != nil {
					r.logger.Errorf("failed to ping a player id=%s: %s", player.ID(), err)
					r.LeavePlayer(player)
				}
			}
		}
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

func (r *Room) onPlayerJoinedHandler(joinedPlayer *Player) error {
	if err := r.registerNewPlayer(joinedPlayer); err != nil {
		return fmt.Errorf("failed to register new player: %s", err)
	}

	event, err := events.NewPlayerJoinedEvent(joinedPlayer.Model)
	if err != nil {
		return err
	}

	if err = r.Broadcast(event); err != nil {
		return err
	}

	if r.playerJoinedHandler != nil {
		r.playerJoinedHandler(joinedPlayer)
	}
	return nil
}

func (r *Room) onPlayerLeftHandler(leftPlayer *Player) error {
	if err := r.unregisterPlayer(leftPlayer); err != nil {
		return fmt.Errorf("failed to unregister player: %s", err)
	}

	event, err := events.NewPlayerLeftEvent(leftPlayer.Model)
	if err != nil {
		return err
	}

	if err = r.Broadcast(event); err != nil {
		return err
	}

	if r.playerLeftHandler != nil {
		r.playerLeftHandler(leftPlayer)
	}

	return nil
}

package domain

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"
)

type Match struct {
	ctx     context.Context
	closeCh chan struct{}
	wg      sync.WaitGroup
	once    sync.Once

	room   *Room
	cfg    *config.Config
	logger logger.Logger

	isStarted     bool
	isClosed      bool
	turningPlayer *domain.PlayerModel
	gameModel     domain.GameModel

	eventBus *events.EventBus
}

func NewMatch(ctx context.Context, cfg *config.Config, logger logger.Logger) *Match {
	match := &Match{
		ctx:      ctx,
		closeCh:  make(chan struct{}),
		room:     NewRoom(ctx, &cfg.App, logger),
		cfg:      cfg,
		logger:   logger,
		eventBus: events.NewEventBus(),
	}

	match.room.SetPlayerJoinedHandler(match.onPlayerJoinedHandler)
	match.room.SetPlayerLeftHandler(match.onPlayerLeftHandler)
	match.eventBus.Subscribe(events.SendMessageType, match.onPlayerSentMessageHandler)
	match.eventBus.Subscribe(events.PlayerFireEventType, match.onPlayerFiredHandler)
	return match
}

func (m *Match) ID() string {
	return m.room.ID()
}

func (m *Match) Equal(rhs *Match) bool {
	if rhs == nil {
		return false
	}
	return m.room.Equal(rhs.room)
}

func (m *Match) Compare(rhs *Match) int {
	if rhs == nil {
		return -1
	}
	return m.room.Compare(rhs.room)
}

func (m *Match) Close() error {
	m.once.Do(func() {
		m.isClosed = true
		close(m.closeCh)
		m.logger.Infof("match id=%s is closing...", m.ID())
	})

	if err := m.room.Close(); err != nil {
		return err
	}

	m.wg.Wait()
	m.logger.Infof("match id=%s is closed", m.ID())
	return nil
}

func (m *Match) JoinNewPlayer(newPlayer *Player) error {
	if err := m.CheckIsAvailableForJoin(); err != nil {
		return err
	}
	return m.room.JoinNewPlayer(newPlayer)
}

func (m *Match) CheckIsAvailableForJoin() error {
	switch {
	case m.isClosed:
		return ErrRoomIsClosed
	case m.isStarted:
		return ErrAlreadyStarted
	case m.room.IsFull():
		return ErrRoomIsFull
	}
	return nil
}

func (m *Match) IsReadyToStart() bool {
	return !m.isClosed && !m.isStarted && m.room.IsFull()
}

func (m *Match) StartMatch() error {
	m.isStarted = true

	m.logger.Infof("match is starting in room id=%s [players: %d]", m.ID(), len(m.room.GetPlayers()))
	event, err := events.NewGameStartEvent(&m.gameModel)
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.gameLoop(m.ctx)
	}()

	return m.room.SendNotification("Game started!", events.RoomNotificationType)
}

func (m *Match) GiveTurnToNextPlayer() error {
	if m.turningPlayer == nil {
		return m.GiveTurnToRandomPlayer()
	}

	if m.turningPlayer.Equal(m.gameModel.LeftPlayer) {
		return m.GiveTurnToPlayer(m.gameModel.RightPlayer)
	} else {
		return m.GiveTurnToPlayer(m.gameModel.LeftPlayer)
	}
}

func (m *Match) GiveTurnToRandomPlayer() error {
	return m.GiveTurnToPlayer(m.getRandomPlayer())
}

func (m *Match) GiveTurnToPlayer(turningPlayer *domain.PlayerModel) error {
	m.turningPlayer = turningPlayer

	event, err := events.NewPlayerTurnEvent(m.turningPlayer, m.cfg.Game.GameTurnTime)
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	return m.room.SendNotification(fmt.Sprintf("Player '%s' turns now.", m.turningPlayer.Nickname), events.GameNotificationType)
}

func (m *Match) getRandomPlayer() *domain.PlayerModel {
	capacity := m.room.Capacity()
	if capacity == 0 {
		return nil
	}

	return m.room.GetPlayers()[rand.Intn(capacity)].Model
}

func (m *Match) gameLoop(ctx context.Context) {
	if err := m.GiveTurnToRandomPlayer(); err != nil {
		m.logger.Errorf("failed to give the first turn to a random player: %s", err)
		return
	}

	gameTurnTimer := time.NewTimer(m.cfg.Game.GameTurnTime)
	defer gameTurnTimer.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case <-m.closeCh:
			return

		case <-gameTurnTimer.C:
			if err := m.GiveTurnToNextPlayer(); err != nil {
				m.logger.Errorf("failed to give a turn to the next player: %s", err)
			}
			gameTurnTimer.Reset(m.cfg.Game.GameTurnTime)

		case msg, opened := <-m.room.Events():
			if !opened {
				continue
			}

			if err := m.eventBus.Invoke(msg); err != nil {
				m.logger.Errorf("error while invoking event: %s", err)
			}
		}
	}
}

func (m *Match) allPlayersUpdate() error {
	event, err := events.NewPlayerUpdateStateEvent(&m.gameModel)
	if err != nil {
		return err
	}
	return m.room.Broadcast(event)
}

func (m *Match) onPlayerJoinedHandler(joinedPlayer *Player) {
	if m.gameModel.LeftPlayer == nil {
		m.gameModel.LeftPlayer = joinedPlayer.Model
	} else {
		m.gameModel.RightPlayer = joinedPlayer.Model
	}

	if err := m.allPlayersUpdate(); err != nil {
		m.logger.Errorf("failed to update players: %s", err)
	}

	if err := m.room.SendNotification(fmt.Sprintf("Player '%s' joined the game.", joinedPlayer.Nickname()), events.RoomNotificationType); err != nil {
		m.logger.Error(err)
	}
	m.room.logger.Infof("player %s joined the match id=%s [players: %d]", joinedPlayer.String(), m.ID(), m.room.Capacity())

	if !m.IsReadyToStart() {
		return
	}

	if err := m.StartMatch(); err != nil {
		m.logger.Errorf("failed to start match id=%s: %s", m.ID(), err)
	}
}

func (m *Match) onPlayerLeftHandler(leftPlayer *Player) {
	m.room.logger.Infof("player %s left the match id=%s [players: %d]", leftPlayer.String(), m.ID(), m.room.Capacity())
	if err := m.room.SendNotification(fmt.Sprintf("Player '%s' left the game.", leftPlayer.Nickname()), events.RoomNotificationType); err != nil {
		m.logger.Error(err)
	}
}

func (m *Match) onPlayerSentMessageHandler(e events.Event) error {
	return m.room.Broadcast(e)
}

func (m *Match) onPlayerFiredHandler(e events.Event) error {
	playerFiredEvent, err := events.CastTo[events.PlayerFireEvent](e)
	if err != nil {
		return err
	}

	if m.turningPlayer.ID != playerFiredEvent.PlayerID {
		return ErrNotYourTurn
	}

	var targetBoard *domain.Board
	if m.gameModel.LeftPlayer.ID == m.turningPlayer.ID {
		targetBoard = &m.gameModel.RightPlayer.Board
	} else {
		targetBoard = &m.gameModel.LeftPlayer.Board
	}

	if err := m.fireAtCell(targetBoard, playerFiredEvent.CellX, playerFiredEvent.CellY); err != nil {
		return err
	}

	cellStr := strings.ToUpper(targetBoard.CellString(playerFiredEvent.CellX, playerFiredEvent.CellY))
	m.room.SendNotification(fmt.Sprintf("Player '%s' fired at (%s)!", playerFiredEvent.PlayerNickname, cellStr), events.GameNotificationType)
	m.logger.Infof("player id=%s fired at cell (%s)", playerFiredEvent.PlayerID, cellStr)

	return m.allPlayersUpdate()
}

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrNotYourTurn   = errors.New("this player doesn't have permission to fire")
)

func (m *Match) fireAtCell(board *domain.Board, cellX, cellY byte) error {
	switch {
	case board.IsCellEmpty(cellX, cellY):
		board.SetCell(cellX, cellY, domain.Miss)
		return nil

	case board.GetCellType(cellX, cellY) == domain.Alive:
		board.SetCell(cellX, cellY, domain.Dead)
		return nil

	default:
		return ErrInvalidTarget
	}
}

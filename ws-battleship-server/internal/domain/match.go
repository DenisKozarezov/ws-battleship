package domain

import (
	"context"
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
	room   *Room
	cfg    *config.Config
	logger logger.Logger

	isStarted     bool
	isClosed      bool
	turningPlayer *domain.PlayerModel
	gameModel     domain.GameModel

	ctx context.Context
	wg  sync.WaitGroup
}

func NewMatch(ctx context.Context, cfg *config.Config, logger logger.Logger) *Match {
	m := &Match{
		room:   NewRoom(ctx, &cfg.App, logger),
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
	}

	m.room.SetPlayerJoinedHandler(m.onPlayerJoined)

	return m
}

func (m *Match) ID() string {
	return m.room.ID()
}

func (m *Match) Equal(rhs *Match) bool {
	if rhs == nil {
		return false
	}
	return m.ID() == rhs.ID()
}

func (m *Match) Compare(rhs *Match) int {
	if rhs == nil {
		return -1
	}
	return strings.Compare(m.ID(), rhs.ID())
}

func (m *Match) Close() error {
	m.isClosed = true
	m.logger.Infof("match id=%s is closing...", m.ID())
	if err := m.room.Close(); err != nil {
		return err
	}

	m.wg.Wait()
	m.logger.Infof("match id=%s is closed", m.ID())
	return nil
}

func (m *Match) JoinNewPlayer(newPlayer *Player) error {
	return m.room.JoinNewPlayer(newPlayer)
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

	m.room.SendChatNotification("Game started!")

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.gameLoop()
	}()
	return nil
}

func (m *Match) CheckIsAvailable() error {
	if m.isClosed {
		return ErrRoomIsClosed
	}

	if m.isStarted {
		return ErrAlreadyStarted
	}

	if m.room.IsFull() {
		return ErrRoomIsFull
	}

	return nil
}

func (m *Match) IsReadyToStart() bool {
	return !m.isClosed && !m.isStarted && m.room.IsFull()
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

	m.room.SendChatNotification(fmt.Sprintf("Player '%s' turns now.", m.turningPlayer.Nickname))
	return nil
}

func (m *Match) getRandomPlayer() *domain.PlayerModel {
	randIdx := rand.Intn(m.room.Capacity())
	return m.room.GetPlayers()[randIdx].Model
}

func (m *Match) gameLoop() {
	gameTurnTimer := time.NewTimer(m.cfg.Game.GameTurnTime)
	defer gameTurnTimer.Stop()

	if err := m.GiveTurnToRandomPlayer(); err != nil {
		m.logger.Errorf("failed to get the first turn to a random player: %s", err)
		return
	}

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-gameTurnTimer.C:
			if err := m.GiveTurnToNextPlayer(); err != nil {
				m.logger.Errorf("failed to get turn to the next player: %s", err)
			}
			gameTurnTimer.Reset(m.cfg.Game.GameTurnTime)
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

func (m *Match) onPlayerJoined(joinedPlayer *Player) {
	if m.gameModel.LeftPlayer == nil {
		m.gameModel.LeftPlayer = joinedPlayer.Model
	} else {
		m.gameModel.RightPlayer = joinedPlayer.Model
	}

	m.allPlayersUpdate()

	if !m.IsReadyToStart() {
		return
	}

	if err := m.StartMatch(); err != nil {
		m.logger.Errorf("failed to start match id=%s: %s", m.ID(), err)
	}
}

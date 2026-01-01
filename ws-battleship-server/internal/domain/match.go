package domain

import (
	"context"
	"fmt"
	"math/rand"
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
	match := &Match{
		room:   NewRoom(ctx, &cfg.App, logger),
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
	}

	match.room.SetPlayerJoinedHandler(match.onPlayerJoinedHandler)
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

	m.room.SendChatNotification("Game started!")

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.gameLoop(m.ctx)
	}()
	return nil
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
		case <-gameTurnTimer.C:
			if err := m.GiveTurnToNextPlayer(); err != nil {
				m.logger.Errorf("failed to give a turn to the next player: %s", err)
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

func (m *Match) onPlayerJoinedHandler(joinedPlayer *Player) {
	if m.gameModel.LeftPlayer == nil {
		m.gameModel.LeftPlayer = joinedPlayer.Model
	} else {
		m.gameModel.RightPlayer = joinedPlayer.Model
	}

	if err := m.allPlayersUpdate(); err != nil {
		m.logger.Errorf("failed to update players: %s", err)
	}

	if !m.IsReadyToStart() {
		return
	}

	if err := m.StartMatch(); err != nil {
		m.logger.Errorf("failed to start match id=%s: %s", m.ID(), err)
	}
}

package domain

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
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
}

func NewMatch(ctx context.Context, cfg *config.Config, logger logger.Logger) *Match {
	m := &Match{
		room:   NewRoom(ctx, &cfg.App, logger),
		cfg:    cfg,
		logger: logger,
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
	return m.room.Close()
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

	var randTurningPlayer *domain.PlayerModel
	if rand.Intn(2) == 0 {
		randTurningPlayer = m.gameModel.LeftPlayer
	} else {
		randTurningPlayer = m.gameModel.RightPlayer
	}
	return m.GiveTurnToPlayer(randTurningPlayer)
}

func (m *Match) CheckAvailable() error {
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
func (m *Match) GiveTurnToPlayer(player *domain.PlayerModel) error {
	m.turningPlayer = player

	event, err := events.NewPlayerTurnEvent(player, m.cfg.Game.GameTurnTime)
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	m.room.SendChatNotification(fmt.Sprintf("Player '%s' turns now.", player.Nickname))
	return nil
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
		m.logger.Errorf("failed to start a match id=%s: %s", m.ID(), err)
	}
}

func (m *Match) allPlayersUpdate() error {
	event, err := events.NewPlayerUpdateStateEvent(&m.gameModel)
	if err != nil {
		return err
	}

	return m.room.Broadcast(event)
}

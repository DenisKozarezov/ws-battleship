package domain

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"
)

type Match struct {
	closeCh chan struct{}
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	mu      sync.RWMutex

	room   *Room
	cfg    *config.Config
	logger logger.Logger

	isStarted        atomic.Bool
	isClosed         atomic.Bool
	gameTurnTimer    *time.Timer
	players          map[string]*Player
	turningPlayer    *Player
	turningPlayerIdx int
	gameModel        domain.GameModel

	cmds     chan Command
	eventBus *events.EventBus
}

func NewMatch(ctx context.Context, cfg *config.Config, logger logger.Logger) *Match {
	matchCtx, cancel := context.WithCancel(ctx)

	match := &Match{
		closeCh:       make(chan struct{}),
		cancel:        cancel,
		room:          NewRoom(matchCtx, &cfg.App, logger),
		cfg:           cfg,
		logger:        logger,
		gameTurnTimer: time.NewTimer(0),
		players:       make(map[string]*Player, cfg.App.ClientsConnectionsMax),
		cmds:          make(chan Command, 10),
		eventBus:      events.NewEventBus(),
	}

	match.room.SetClientJoinedHandler(match.onPlayerJoinedHandler)
	match.room.SetClientLeftHandler(match.onPlayerLeftHandler)
	match.eventBus.Subscribe(events.SendMessageType, match.onPlayerSentMessageHandler)
	match.eventBus.Subscribe(events.PlayerFireEventType, match.onPlayerFiredHandler)

	<-match.gameTurnTimer.C
	match.wg.Add(1)
	go match.gameLoop(ctx)

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
		m.isClosed.Store(true)
		m.cancel()
		close(m.closeCh)
		close(m.cmds)
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

	m.players[newPlayer.ID()] = newPlayer

	return m.room.JoinNewClient(newPlayer)
}

func (m *Match) CheckIsAvailableForJoin() error {
	switch {
	case m.isClosed.Load():
		return ErrRoomIsClosed
	case m.isStarted.Load():
		return ErrAlreadyStarted
	case m.room.IsFull():
		return ErrRoomIsFull
	}
	return nil
}

func (m *Match) IsReadyToStart() bool {
	return !m.isClosed.Load() && !m.isStarted.Load() && m.room.IsFull()
}

func (m *Match) StartMatch() error {
	m.isStarted.Store(true)

	event, err := events.NewGameStartEvent()
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	m.Dispatch(NewGameTurnCommand())

	return m.SendNotification("Game started!", events.RoomNotificationType)
}

func (m *Match) EndMatch(winningPlayer *Player) error {
	event, err := events.NewGameEndEvent(winningPlayer.Model)
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	_ = m.SendNotification(fmt.Sprintf("Player '%s' has won!", winningPlayer.Nickname()), events.RoomNotificationType)

	m.Dispatch(NewCloseMatchCommand())
	return nil
}

func (m *Match) GiveTurnToNextPlayer() error {
	defer m.resetGameTurnTimer()

	m.gameModel.TurnCount++

	if m.turningPlayer == nil {
		return m.GiveTurnToPlayer(m.getRandomPlayer())
	} else {
		return m.GiveTurnToPlayer(m.getNextTarget())
	}
}

func (m *Match) GiveTurnToPlayer(turningPlayer *Player) error {
	m.turningPlayer = turningPlayer

	event, err := events.NewPlayerTurnEvent(m.gameModel.TurnCount, turningPlayer.ID(), m.cfg.Game.GameTurnTime)
	if err != nil {
		return err
	}

	if err := m.room.Broadcast(event); err != nil {
		return err
	}

	return m.SendNotification(fmt.Sprintf("Player '%s' turns now.", turningPlayer.Nickname()), events.GameNotificationType)
}

func (m *Match) SendNotification(msg string, notificationType events.ChatMessageType) error {
	event, err := events.NewChatNotificationEvent(msg, notificationType)
	if err != nil {
		return fmt.Errorf("failed to send a chat notification: %w", err)
	}
	return m.room.Broadcast(event)
}

func (m *Match) Fire(args events.FireCommandArgs) error {
	if m.turningPlayer.ID() != args.FiringPlayerID {
		return ErrNotYourTurn
	}

	firingPlayer := m.players[args.FiringPlayerID]
	targetPlayer := m.players[args.TargetPlayerID]

	if err := m.fireAtCell(targetPlayer.Model, args.CellX, args.CellY); err != nil {
		return err
	}
	firingPlayer.RevealCell(args.CellX, args.CellY)

	_ = m.SendNotification(fmt.Sprintf("Player '%s' fired at cell (%s).", firingPlayer.Nickname(), targetPlayer.Model.Board.CellString(args.CellX, args.CellY)), events.GameNotificationType)

	if err := m.allPlayersUpdate(); err != nil {
		return err
	}

	if targetPlayer.Model.IsDead() {
		m.Dispatch(NewGameEndCommand(m.logger, m.turningPlayer))
	} else {
		m.Dispatch(NewGameTurnCommand())
	}
	return nil
}

func (m *Match) Dispatch(cmd Command) {
	if m.isClosed.Load() {
		return
	}
	m.cmds <- cmd
}

func (m *Match) GetPlayers() []*Player {
	players := make([]*Player, 0, len(m.players))
	for _, player := range m.players {
		players = append(players, player)
	}

	slices.SortFunc(players, func(lhs, rhs *Player) int {
		return lhs.Compare(rhs)
	})

	return players
}

func (m *Match) getRandomPlayer() *Player {
	if len(m.players) == 0 {
		return nil
	}

	return m.GetPlayers()[rand.Intn(len(m.players))]
}

func (m *Match) getNextTarget() *Player {
	if m.turningPlayerIdx+1 >= len(m.players) {
		m.turningPlayerIdx = 0
	} else {
		m.turningPlayerIdx++
	}
	return m.GetPlayers()[m.turningPlayerIdx]
}

func (m *Match) gameLoop(ctx context.Context) {
	defer func() {
		m.gameTurnTimer.Stop()
		m.wg.Done()
	}()

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case <-m.closeCh:
			m.logger.Infof("stoping game loop in match id=%s", m.ID())
			return

		case <-m.gameTurnTimer.C:
			m.Dispatch(NewGameTurnCommand())

		case cmd, opened := <-m.cmds:
			if !opened {
				return
			}
			if err := cmd.Execute(m); err != nil {
				m.logger.Errorf("failed to execute a command: %s", err)
				m.Dispatch(NewCloseMatchCommand())
			}

		case msg, opened := <-m.room.Events():
			if !opened {
				return
			}

			if err := m.eventBus.Invoke(msg); err != nil {
				m.logger.Errorf("error while invoking event: %s", err)
			}
		}
	}
}

func (m *Match) resetGameTurnTimer() {
	m.gameTurnTimer.Reset(m.cfg.Game.GameTurnTime)
}

func (m *Match) allPlayersUpdate() error {
	for playerID, player := range m.players {
		gameModel := m.buildGameModelForPlayer(player)

		event, err := events.NewPlayerUpdateStateEvent(&gameModel)
		if err != nil {
			return err
		}

		if err := m.room.SendMessageToClient(playerID, event); err != nil {
			return err
		}
	}

	return nil
}

func (m *Match) buildGameModelForPlayer(targetPlayer *Player) domain.GameModel {
	maskedGameModel := domain.GameModel{
		TurnCount: m.gameModel.TurnCount,
		Players:   make(map[string]*domain.PlayerModel, len(m.players)),
	}

	for playerID, player := range m.players {
		playerModel := domain.NewPlayerModel(player.Model.Board, domain.ClientMetadata{
			ClientID: player.Model.ID,
			Nickname: player.Model.Nickname,
		})

		if !playerModel.Equal(targetPlayer.Model) {
			playerModel.Board = player.maskBoardForPlayer(targetPlayer)
		}

		maskedGameModel.Players[playerID] = playerModel
	}

	return maskedGameModel
}

func (m *Match) fireAtCell(targetPlayer *domain.PlayerModel, cellX, cellY byte) error {
	var newType domain.CellType
	switch {
	// First case: we missed.
	case targetPlayer.Board.IsCellEmpty(cellX, cellY):
		newType = domain.Miss

	// Second case: we hit a ship cell.
	case targetPlayer.Board.GetCellType(cellX, cellY) == domain.Ship:
		targetPlayer.Hit()
		newType = domain.Dead

	// Otherwise, we return an error.
	default:
		return ErrInvalidTarget
	}

	targetPlayer.Board.SetCell(cellX, cellY, newType)
	return nil
}

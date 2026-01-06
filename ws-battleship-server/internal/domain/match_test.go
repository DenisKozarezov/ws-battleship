package domain

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/pkg/logger"

	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckIsAvailableForJoin(t *testing.T) {
	t.Run("match is not available for join when it is closed", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)
		require.NoError(t, match.Close())

		// 2. Act
		err := match.CheckIsAvailableForJoin()

		// 3. Assert
		require.ErrorIsf(t, err, ErrRoomIsClosed, "shouldn't join to match while closed")
	})

	t.Run("match is not available for join when room is full", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 0,
			},
		}, loggerMock)

		// 2. Act
		err := match.CheckIsAvailableForJoin()

		// 3. Assert
		require.ErrorIsf(t, err, ErrRoomIsFull, "shouldn't join to match while room is full")
	})

	t.Run("match is not available for join when it is already started", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)
		match.isStarted = true

		// 2. Act
		err := match.CheckIsAvailableForJoin()

		// 3. Assert
		require.ErrorIsf(t, err, ErrAlreadyStarted, "shouldn't join to match while started")
	})

	t.Run("match is available for join", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)

		// 2. Act
		err := match.CheckIsAvailableForJoin()

		// 3. Assert
		require.NoError(t, err)
	})
}

func TestIsMatchReadyToStart(t *testing.T) {
	t.Run("match is not ready without players", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)

		// 2. Act
		got := match.IsReadyToStart()

		// 3. Assert
		require.False(t, got)
	})

	t.Run("match is not ready when closed", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)
		require.NoError(t, match.Close())

		// 2. Act
		got := match.IsReadyToStart()

		// 3. Assert
		require.False(t, got)
	})
}

func TestMatchClose(t *testing.T) {
	t.Run("idempotent close", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		match := NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, loggerMock)

		// 2. Act
		require.NoError(t, match.Close())
		require.NoError(t, match.Close())
		require.NoError(t, match.Close())
	})
}

func TestFireAtCell(t *testing.T) {
	for _, tt := range []struct {
		name         string
		board        domain.Board
		cellX        byte
		cellY        byte
		expectedType domain.CellType
	}{
		{
			name: "fire at empty cell, expecting miss",
			board: domain.Board{
				{domain.Empty},
			},
			cellX:        0,
			cellY:        0,
			expectedType: domain.Miss,
		},
		{
			name: "fire at non-initialized cell, also expecting miss",
			board: domain.Board{
				{},
			},
			cellX:        0,
			cellY:        0,
			expectedType: domain.Miss,
		},
		{
			name: "fire at ship cell, expecting dead",
			board: domain.Board{
				{domain.Ship},
			},
			cellX:        0,
			cellY:        0,
			expectedType: domain.Dead,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			match := Match{
				turningPlayer: domain.NewPlayerModel(tt.board, domain.ClientMetadata{ClientID: "player 1"}),
				targetPlayer:  domain.NewPlayerModel(tt.board, domain.ClientMetadata{ClientID: "player 2"}),
				visible:       make(map[string][]VisibleCell),
			}

			// 2. Act
			err := match.fireAtCell(tt.cellX, tt.cellY)

			// 3. Assert
			require.NoError(t, err)
			require.Equal(t, tt.expectedType, match.targetPlayer.Board.GetCellType(tt.cellX, tt.cellY))
		})
	}
}

func TestFireAtInvalidTarget(t *testing.T) {
	t.Run("fire at dead cell", func(t *testing.T) {
		// 1. Arrange
		board := domain.Board{
			{domain.Dead},
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		err := match.fireAtCell(0, 0)

		// 3. Assert
		require.ErrorIsf(t, err, ErrInvalidTarget, "expected error invalid target")
	})
}

func TestTurningPlayerVisionAfterFire(t *testing.T) {
	t.Run("hit an empty cell", func(t *testing.T) {
		// 1. Arrange
		board := domain.Board{
			{domain.Empty},
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		err := match.fireAtCell(0, 0)

		// 3. Assert
		require.NoError(t, err)
		require.EqualValuesf(t, []VisibleCell{
			{X: 0, Y: 0},
		}, match.visible["player 1"], "hit cell must be revealed for the turning player")
		require.Nil(t, match.visible["player 2"])
	})

	t.Run("hit some cells", func(t *testing.T) {
		// 1. Arrange
		board := domain.Board{
			{domain.Empty, domain.Ship, domain.Ship, domain.Miss},
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		require.NoError(t, match.fireAtCell(0, 0))
		require.NoError(t, match.fireAtCell(1, 0))
		require.NoError(t, match.fireAtCell(2, 0))

		// 3. Assert
		require.EqualValuesf(t, []VisibleCell{
			{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0},
		}, match.visible["player 1"], "hit cells must be revealed for the turning player")
		require.Nil(t, match.visible["player 2"])
	})
}

func TestPlayerIsDeadAfterAllShipCellsWereHit(t *testing.T) {
	t.Run("hit all ship cells 1", func(t *testing.T) {
		// 1. Arrange
		board := domain.Board{
			{domain.Empty, domain.Ship, domain.Ship, domain.Miss},
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		require.NoError(t, match.fireAtCell(1, 0))
		require.NoError(t, match.fireAtCell(2, 0))

		// 3. Assert
		require.Truef(t, match.targetPlayer.IsDead(), "target player must be dead")
	})

	t.Run("hit all ship cells 2", func(t *testing.T) {
		// 1. Arrange
		var board domain.Board
		for i := 0; i < board.Size(); i++ {
			for j := 0; j < board.Size(); j++ {
				board.SetCell(byte(j), byte(i), domain.Ship)
			}
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		for i := 0; i < board.Size(); i++ {
			for j := 0; j < board.Size(); j++ {
				require.NoError(t, match.fireAtCell(byte(j), byte(i)))
			}
		}

		// 3. Assert
		require.Truef(t, match.targetPlayer.IsDead(), "target player must be dead")
	})

	t.Run("player is not dead, when at least 1 ship cell is alive", func(t *testing.T) {
		// 1. Arrange
		board := domain.Board{
			{domain.Dead, domain.Ship, domain.Ship, domain.Dead},
		}

		match := Match{
			turningPlayer: domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 1"}),
			targetPlayer:  domain.NewPlayerModel(board, domain.ClientMetadata{ClientID: "player 2"}),
			visible:       make(map[string][]VisibleCell),
		}

		// 2. Act
		require.NoError(t, match.fireAtCell(1, 0))

		// 3. Assert
		require.Falsef(t, match.targetPlayer.IsDead(), "target player must not be dead")
	})
}

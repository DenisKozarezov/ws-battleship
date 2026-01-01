package domain

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
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

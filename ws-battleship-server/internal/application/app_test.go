package application

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/domain"
	"ws-battleship-shared/pkg/logger"

	"github.com/stretchr/testify/require"
)

func TestFindFreeRoom(t *testing.T) {
	t.Run("no rooms - cannot find", func(t *testing.T) {
		// 1. Arrange
		app := App{rooms: nil}

		// 2. Act
		got := app.findFreeRoom()

		// 3. Assert
		require.Nil(t, got)
	})

	t.Run("an empty room without players", func(t *testing.T) {
		// 1. Arrange
		logger, _ := logger.NewLogger(true, "")

		newRoom := domain.NewRoom(t.Context(), &config.AppConfig{
			KeepAlivePeriod: time.Second * 5,
			RoomCapacityMax: 5,
		}, logger)

		app := App{
			rooms: map[string]*domain.Room{
				newRoom.ID(): newRoom,
			},
		}

		// 2. Act
		got := app.findFreeRoom()

		// 3. Assert
		require.NotNil(t, got)
		err := got.Close()
		require.NoError(t, err)
		require.Equal(t, newRoom.ID(), got.ID())
	})
}

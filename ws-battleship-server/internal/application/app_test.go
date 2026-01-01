package application

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestFindFreeMatch(t *testing.T) {
	t.Run("cannot find a free match if there is not matches at all", func(t *testing.T) {
		// 1. Arrange
		var app App

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.Nil(t, got)
	})

	t.Run("find a free match when we have an empty match without players", func(t *testing.T) {
		// 1. Arrange
		newRoom := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, nil)

		app := App{
			matches: map[string]*domain.Match{
				newRoom.ID(): newRoom,
			},
		}

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.NotNil(t, got)
		require.Equal(t, newRoom.ID(), got.ID())
	})
}

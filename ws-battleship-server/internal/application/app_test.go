package application

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestFindFreeMatch(t *testing.T) {
	t.Run("cannot find a free match if there are no matches at all", func(t *testing.T) {
		// 1. Arrange
		var app App

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.Nil(t, got)
	})

	t.Run("find a free match when there is an empty match without players", func(t *testing.T) {
		// 1. Arrange
		newMatch := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 5,
			},
		}, nil)

		app := App{
			matches: map[string]*domain.Match{
				newMatch.ID(): newMatch,
			},
		}

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.NotNil(t, got)
		require.Equal(t, newMatch.ID(), got.ID())
	})

	t.Run("cannot find a free match when all rooms are full", func(t *testing.T) {
		// 1. Arrange
		newMatch1 := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 0,
			},
		}, nil)
		newMatch2 := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 0,
			},
		}, nil)

		app := App{
			matches: map[string]*domain.Match{
				newMatch1.ID(): newMatch1,
				newMatch2.ID(): newMatch2,
			},
		}

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.Nil(t, got)
	})

	t.Run("find a free match", func(t *testing.T) {
		// 1. Arrange
		newMatch1 := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 0,
			},
		}, nil)
		newMatch2 := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 0,
			},
		}, nil)
		newMatch3 := domain.NewMatch(t.Context(), &config.Config{
			App: config.AppConfig{
				KeepAlivePeriod: time.Second * 5,
				RoomCapacityMax: 1,
			},
		}, nil)

		app := App{
			matches: map[string]*domain.Match{
				newMatch1.ID(): newMatch1,
				newMatch2.ID(): newMatch2,
				newMatch3.ID(): newMatch3,
			},
		}

		// 2. Act
		got := app.findFreeMatch()

		// 3. Assert
		require.NotNil(t, got)
		require.Equal(t, newMatch3.ID(), got.ID())
	})
}

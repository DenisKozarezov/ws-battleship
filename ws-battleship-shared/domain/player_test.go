package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlayerIsDeadAfterAllShipCellsWereHit(t *testing.T) {
	t.Run("player is dead, when 1 only ship cell was decremented", func(t *testing.T) {
		// 1. Arrange
		board := Board{
			{Ship},
		}

		player := NewPlayerModel(board, ClientMetadata{})

		// 2. Act
		player.DecrementCell()

		// 3. Assert
		require.Zero(t, player.ShipCells)
		require.Truef(t, player.IsDead(), "player must be dead")
	})

	t.Run("decrement all ship cells", func(t *testing.T) {
		// 1. Arrange
		var board Board
		for i := 0; i < board.Size(); i++ {
			for j := 0; j < board.Size(); j++ {
				board.SetCell(byte(j), byte(i), Ship)
			}
		}

		player := NewPlayerModel(board, ClientMetadata{})

		// 2. Act
		for i := 0; i < board.Size(); i++ {
			for j := 0; j < board.Size(); j++ {
				player.DecrementCell()
			}
		}

		// 3. Assert
		require.Zero(t, player.ShipCells)
		require.Truef(t, player.IsDead(), "player must be dead")
	})

	t.Run("player is not dead, when at least 1 ship cell is alive", func(t *testing.T) {
		// 1. Arrange
		board := Board{
			{Ship}, {Ship},
		}

		player := NewPlayerModel(board, ClientMetadata{})

		// 2. Act
		player.DecrementCell()

		// 3. Assert
		require.Equalf(t, byte(1), player.ShipCells, "1 ship cell should remain")
		require.Falsef(t, player.IsDead(), "player must not be dead")
	})
}

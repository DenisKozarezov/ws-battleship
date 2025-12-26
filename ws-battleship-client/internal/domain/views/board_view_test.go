package views

import (
	"testing"
	"ws-battleship-client/internal/domain/models"

	"github.com/stretchr/testify/require"
)

func TestSelection(t *testing.T) {
	t.Run("selection up", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 5
		view.selectedColIdx = 255

		// 2. Act
		view.selectionUp()

		// 3. Assert
		require.Equal(t, 4, view.selectedRowIdx)
		require.Equal(t, 255, view.selectedColIdx)
	})

	t.Run("selection up - check bounds", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 0
		view.selectedColIdx = 255

		// 2. Act
		view.selectionUp()
		view.selectionUp()
		view.selectionUp()

		// 3. Assert
		require.Zero(t, view.selectedRowIdx)
		require.Equal(t, 255, view.selectedColIdx)
	})

	t.Run("selection down", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 5
		view.selectedColIdx = 255

		// 2. Act
		view.selectionDown()

		// 3. Assert
		require.Equal(t, 6, view.selectedRowIdx)
		require.Equal(t, 255, view.selectedColIdx)
	})

	t.Run("selection down - check bounds", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = view.boardSize - 1
		view.selectedColIdx = 255

		// 2. Act
		view.selectionDown()
		view.selectionDown()
		view.selectionDown()

		// 3. Assert
		require.Equal(t, view.boardSize-1, view.selectedRowIdx)
		require.Equal(t, 255, view.selectedColIdx)
	})

	t.Run("selection left", func(t *testing.T) {
		// 1. Arrange
		var view BoardView
		view.selectedRowIdx = 255
		view.selectedColIdx = 5

		// 2. Act
		view.selectionLeft() // -2

		// 3. Assert
		require.Equal(t, 255, view.selectedRowIdx)
		require.Equal(t, 3, view.selectedColIdx)
	})

	t.Run("selection left - check bounds", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 255
		view.selectedColIdx = 1

		// 2. Act
		view.selectionLeft()
		view.selectionLeft()
		view.selectionLeft()

		// 3. Assert
		require.Equal(t, 255, view.selectedRowIdx)
		require.Zero(t, view.selectedColIdx)
	})

	t.Run("selection right", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 255
		view.selectedColIdx = 5

		// 2. Act
		view.selectionRight() // +2

		// 3. Assert
		require.Equal(t, 255, view.selectedRowIdx)
		require.Equal(t, 7, view.selectedColIdx)
	})

	t.Run("selection right - check bounds", func(t *testing.T) {
		// 1. Arrange
		player := models.NewPlayer("")
		view := NewBoardView(player)
		view.selectedRowIdx = 255
		view.selectedColIdx = len(view.alphabet) - 1

		// 2. Act
		view.selectionRight()
		view.selectionRight()
		view.selectionRight()

		// 3. Assert
		require.Equal(t, 255, view.selectedRowIdx)
		require.Equal(t, len(view.alphabet)-1, view.selectedColIdx)
	})
}

package views

import (
	"testing"
	"ws-battleship-client/internal/domain/models"
	"ws-battleship-shared/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestSelection(t *testing.T) {
	t.Run("selection up", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionUp()

		// 3. Assert
		require.Equal(t, 4, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection up - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 0
		view.cellX = 5

		// 2. Act
		view.selectionUp()
		view.selectionUp()
		view.selectionUp()

		// 3. Assert
		require.Zero(t, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection down", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionDown()

		// 3. Assert
		require.Equal(t, 6, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection down - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = view.boardSize - 1
		view.cellX = 5

		// 2. Act
		view.selectionDown()
		view.selectionDown()
		view.selectionDown()

		// 3. Assert
		require.Equal(t, view.boardSize-1, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection left", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionLeft()

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 4, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection left - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = 1

		// 2. Act
		view.selectionLeft()
		view.selectionLeft()
		view.selectionLeft()

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Zero(t, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection right", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionRight()

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 6, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})

	t.Run("selection right - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.cellY = 5
		view.cellX = view.boardSize - 1

		// 2. Act
		view.selectionRight()
		view.selectionRight()
		view.selectionRight()

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, view.boardSize-1, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})
}

func TestSelectCell(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		cellX                  int
		cellY                  int
		expectedCellY          int
		expectedCellX          int
		expectedSelectedRowIdx int
		expectedSelectedColIdx int
	}{
		{
			name:  "(0; 0)",
			cellY: 0,
			cellX: 0,

			expectedCellY: 0,
			expectedCellX: 0,

			expectedSelectedRowIdx: 0,
			expectedSelectedColIdx: 0,
		},
		{
			name:  "(5; 5)",
			cellY: 5,
			cellX: 5,

			expectedCellY: 5,
			expectedCellX: 5,

			expectedSelectedRowIdx: 5,
			expectedSelectedColIdx: 10,
		},
		{
			name:  "out of bounds, clamping indices (255; 255) -> (9; 9)",
			cellX: 255,
			cellY: 255,

			expectedCellX: 9,
			expectedCellY: 9,

			expectedSelectedRowIdx: 9,
			expectedSelectedColIdx: 18,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			view := NewBoardView(models.NewPlayer(""))

			// 2. Act
			view.SelectCell(tt.cellY, tt.cellX)

			// 3. Assert
			require.Equal(t, tt.expectedCellX, view.cellX)
			require.Equal(t, tt.expectedCellY, view.cellY)
			require.Equal(t, tt.expectedSelectedRowIdx, view.selectedRowIdx)
			require.Equal(t, tt.expectedSelectedColIdx, view.selectedColIdx)
		})
	}
}

func TestSelectedCellHighlightStyle(t *testing.T) {
	for _, tt := range []struct {
		name     string
		cellX    int
		expected lipgloss.Style
	}{
		{
			name:     "missed cell is not empty, not allowed to strike",
			cellX:    0,
			expected: highlightForbiddenCell,
		},
		{
			name:     "alive cell is not empty, not allowed to strike",
			cellX:    1,
			expected: highlightForbiddenCell,
		},
		{
			name:     "dead cell is not empty, not allowed to strike",
			cellX:    2,
			expected: highlightForbiddenCell,
		},
		{
			name:     "empty cell, allowed to strike",
			cellX:    3,
			expected: highlightAllowedCell,
		},
		{
			name:     "not initialized cell, allowed to strike",
			cellX:    4,
			expected: highlightAllowedCell,
		},
		{
			name:     "out of bounds is clamped, allowed to strike",
			cellX:    255,
			expected: highlightAllowedCell,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			var board = domain.Board{
				{domain.Miss, domain.Alive, domain.Dead, domain.Empty, 0},
				// not free     not free      not free       free     free
				//    ∙             ■             □
			}
			view := NewBoardView(&models.Player{Board: board})
			view.SelectCell(0, tt.cellX)

			// 2. Act
			got := view.getSelectedCellHighlightStyle()

			// 3. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestSelectionWhenBoardIsNotSelectable(t *testing.T) {
	t.Run("selection up", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(&tea.KeyMsg{Type: tea.KeyUp})

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})
	t.Run("selection down", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(&tea.KeyMsg{Type: tea.KeyDown})

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})
	t.Run("selection left", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(&tea.KeyMsg{Type: tea.KeyLeft})

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})
	t.Run("selection right", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView(models.NewPlayer(""))
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(&tea.KeyMsg{Type: tea.KeyRight})

		// 3. Assert
		require.Equal(t, 5, view.cellY)
		require.Equal(t, 5, view.cellX)
		require.Equal(t, view.cellY, view.selectedRowIdx)
		require.Equal(t, view.cellX*2, view.selectedColIdx)
	})
}

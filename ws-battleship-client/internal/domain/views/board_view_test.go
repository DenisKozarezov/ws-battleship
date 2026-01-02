package views

import (
	"testing"
	"ws-battleship-shared/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestSelection(t *testing.T) {
	t.Run("selection up", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionUp()

		// 3. Assert
		assertSelectionCoordinates(t, view, 4, 5)
	})

	t.Run("selection up - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 0
		view.cellX = 5

		// 2. Act
		view.selectionUp()
		view.selectionUp()
		view.selectionUp()

		// 3. Assert
		assertSelectionCoordinates(t, view, 0, 5)
	})

	t.Run("selection down", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionDown()

		// 3. Assert
		assertSelectionCoordinates(t, view, 6, 5)
	})

	t.Run("selection down - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = view.board.Size() - 1
		view.cellX = 5

		// 2. Act
		view.selectionDown()
		view.selectionDown()
		view.selectionDown()

		// 3. Assert
		assertSelectionCoordinates(t, view, view.board.Size()-1, 5)
	})

	t.Run("selection left", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionLeft()

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 4)
	})

	t.Run("selection left - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = 1

		// 2. Act
		view.selectionLeft()
		view.selectionLeft()
		view.selectionLeft()

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 0)
	})

	t.Run("selection right", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = 5

		// 2. Act
		view.selectionRight()

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 6)
	})

	t.Run("selection right - check bounds", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.cellY = 5
		view.cellX = view.board.Size() - 1

		// 2. Act
		view.selectionRight()
		view.selectionRight()
		view.selectionRight()

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, view.board.Size()-1)
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
			view := NewBoardView()

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

func TestGetCellHighlightStyle(t *testing.T) {
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
			view := NewBoardView()
			view.board = board
			view.SelectCell(tt.cellX, 0)

			// 2. Act
			got := view.getCellHighlighStyle()

			// 3. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestSelectionWhenBoardIsNotSelectable(t *testing.T) {
	t.Run("selection up", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(tea.KeyMsg{Type: tea.KeyUp})

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 5)
	})
	t.Run("selection down", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(tea.KeyMsg{Type: tea.KeyDown})

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 5)
	})
	t.Run("selection left", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(tea.KeyMsg{Type: tea.KeyLeft})

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 5)
	})
	t.Run("selection right", func(t *testing.T) {
		// 1. Arrange
		view := NewBoardView()
		view.SelectCell(5, 5)

		// 2. Act
		view.SetSelectable(false)
		view.Update(tea.KeyMsg{Type: tea.KeyRight})

		// 3. Assert
		assertSelectionCoordinates(t, view, 5, 5)
	})
}

func assertSelectionCoordinates(t *testing.T, view *BoardView, expectedY, expectedX int) {
	require.Equal(t, expectedY, view.cellY)
	require.Equal(t, expectedX, view.cellX)
	require.Equal(t, view.cellY, view.selectedRowIdx)
	require.Equal(t, view.cellX*2, view.selectedColIdx)
}

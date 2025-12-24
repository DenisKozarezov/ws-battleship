package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderBoardRow(t *testing.T) {
	// 1. Arrange
	var b = Board{
		{Miss, Miss, Empty, Empty, Alive, Alive, Dead, Empty, Miss, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead},
	}

	tests := []struct {
		name     string
		rowIdx   int
		expected string
	}{
		{
			name:     "first row with some cells",
			rowIdx:   0,
			expected: "1│* *     O O X   *  │1",
		},
		{
			name:     "second row with empty cells",
			rowIdx:   1,
			expected: "2│                   │2",
		},
		{
			name:     "third row with filled cells",
			rowIdx:   2,
			expected: "3│X X X X X X X X X X│3",
		},
		{
			name:     "not initialized row",
			rowIdx:   b.size() - 1,
			expected: "10│                   │10",
		},
		{
			name:     "out of bounds",
			rowIdx:   255,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 2. Act
			got := b.renderRow(tt.rowIdx)

			// 3. Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBoardGetCellType(t *testing.T) {
	// 1. Arrange
	var b = Board{
		{Alive, 0, 0, 0, Alive, Alive, Dead, 0, 0, Dead},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
	}

	for _, tt := range []struct {
		name     string
		rowIdx   byte
		colIdx   byte
		expected CellType
	}{
		{
			name:     "out of bounds",
			rowIdx:   255,
			colIdx:   255,
			expected: 0,
		},
		{
			name:     "first cell of the first row",
			rowIdx:   0,
			colIdx:   0,
			expected: Alive,
		},
		{
			name:     "last cell of the first row",
			rowIdx:   0,
			colIdx:   byte(b.size() - 1),
			expected: Dead,
		},
		{
			name:     "first cell of the last row",
			rowIdx:   byte(b.size() - 1),
			colIdx:   0,
			expected: Empty,
		},
		{
			name:     "last cell of the last row",
			rowIdx:   byte(b.size() - 1),
			colIdx:   byte(b.size() - 1),
			expected: Empty,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 2. Act
			got := b.GetCellType(tt.rowIdx, tt.colIdx)

			// 3. Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expected: "∙ ∙     O O X   ∙  ",
		},
		{
			name:     "second row with empty cells",
			rowIdx:   1,
			expected: "                   ",
		},
		{
			name:     "third row with filled cells",
			rowIdx:   2,
			expected: "X X X X X X X X X X",
		},
		{
			name:     "not initialized row",
			rowIdx:   b.Size() - 1,
			expected: "                   ",
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
			colIdx:   byte(b.Size() - 1),
			expected: Dead,
		},
		{
			name:     "first cell of the last row",
			rowIdx:   byte(b.Size() - 1),
			colIdx:   0,
			expected: Empty,
		},
		{
			name:     "last cell of the last row",
			rowIdx:   byte(b.Size() - 1),
			colIdx:   byte(b.Size() - 1),
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

func TestBoardSize(t *testing.T) {
	t.Run("check board's size equals board's alphabet length", func(t *testing.T) {
		var b Board
		require.Equal(t, len(b.Alphabet()), b.Size())
	})
}

func TestAlphabet(t *testing.T) {
	t.Run("check board's alphabet equals const variable", func(t *testing.T) {
		var b Board
		require.Equal(t, boardAlphabet, string(b.Alphabet()))
	})
}

func TestBoardLines(t *testing.T) {
	t.Run("board's lines", func(t *testing.T) {
		// 1. Arrange
		var b = Board{
			{Miss, Miss, Miss, Miss, Miss, Miss, Miss, Miss, Miss, Miss},           // 1
			{Dead, Empty, Dead, Empty, Dead, Empty, Dead, Empty, Dead, Empty},      // 2
			{Alive, Dead, Alive, Dead, Alive, Dead, Alive, Dead, Alive, Dead},      // 3
			{Miss, Dead, Alive, Miss, Dead, Alive, Miss, Dead, Alive, Miss},        // 4
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty}, // 5
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty}, // 6
			{Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead},           // 7
			// 8 - Empty
			// 9 - Empty
			// 10 - Empty
		}

		// 2. Act
		got := b.Lines()

		// 3. Assert
		expected := []string{
			"∙ ∙ ∙ ∙ ∙ ∙ ∙ ∙ ∙ ∙", // 1
			"X   X   X   X   X  ", // 2
			"O X O X O X O X O X", // 3
			"∙ X O ∙ X O ∙ X O ∙", // 4
			"                   ", // 5
			"                   ", // 6
			"X X X X X X X X X X", // 7
			"                   ", // 8
			"                   ", // 9
			"                   ", // 10
		}
		require.EqualValues(t, expected, got)
	})
}

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoardGetCellType(t *testing.T) {
	// 1. Arrange
	var b = Board{
		{Ship, 0, 0, 0, Ship, Ship, Dead, 0, 0, Dead},
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
		cellX    byte
		cellY    byte
		expected CellType
	}{
		{
			name:     "out of bounds",
			cellX:    255,
			cellY:    255,
			expected: Null,
		},
		{
			name:     "first cell of the first row",
			cellX:    0,
			cellY:    0,
			expected: Ship,
		},
		{
			name:     "last cell of the first row",
			cellX:    byte(b.Size() - 1),
			cellY:    0,
			expected: Dead,
		},
		{
			name:     "first cell of the last row",
			cellX:    0,
			cellY:    byte(b.Size() - 1),
			expected: Empty,
		},
		{
			name:     "last cell of the last row",
			cellX:    byte(b.Size() - 1),
			cellY:    byte(b.Size() - 1),
			expected: Empty,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 2. Act
			got := b.GetCellType(tt.cellX, tt.cellY)

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
			{Ship, Dead, Ship, Dead, Ship, Dead, Ship, Dead, Ship, Dead},           // 3
			{Miss, Dead, Ship, Miss, Dead, Ship, Miss, Dead, Ship, Miss},           // 4
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
			"□   □   □   □   □  ", // 2
			"■ □ ■ □ ■ □ ■ □ ■ □", // 3
			"∙ □ ■ ∙ □ ■ ∙ □ ■ ∙", // 4
			"                   ", // 5
			"                   ", // 6
			"□ □ □ □ □ □ □ □ □ □", // 7
			"                   ", // 8
			"                   ", // 9
			"                   ", // 10
		}
		require.EqualValues(t, expected, got)
	})
}

func TestRenderBoardRow(t *testing.T) {
	// 1. Arrange
	var b = Board{
		{Miss, Miss, Empty, Empty, Ship, Ship, Dead, Empty, Miss, Empty},
		{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
		{Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead, Dead},
	}

	tests := []struct {
		name     string
		rowIdx   int
		expected string
	}{
		{
			name:     "first row with different cells",
			rowIdx:   0,
			expected: "∙ ∙     ■ ■ □   ∙  ",
		},
		{
			name:     "second row with empty cells",
			rowIdx:   1,
			expected: "                   ",
		},
		{
			name:     "filled third row with dead cells",
			rowIdx:   2,
			expected: "□ □ □ □ □ □ □ □ □ □",
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

func TestBoardMarshalBinary(t *testing.T) {
	t.Run("no error when marshal an empty board", func(t *testing.T) {
		// 1. Arrange
		var b Board

		// 2. Act
		got, err := b.MarshalBinary()

		// 3. Assert
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotZero(t, len(got))
	})
}

package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderBoardRow(t *testing.T) {
	// 1. Arrange
	var b = Board{
		{0, 0, 0, 0, alive, alive, dead, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{dead, dead, dead, dead, dead, dead, dead, dead, dead, dead},
	}

	tests := []struct {
		name     string
		rowIdx   int
		expected string
	}{
		{
			name:     "first row with some cells",
			rowIdx:   0,
			expected: "1│        O O X      │ 1",
		},
		{
			name:     "second row with empty cells",
			rowIdx:   1,
			expected: "2│                   │ 2",
		},
		{
			name:     "third row with filled cells",
			rowIdx:   2,
			expected: "3│X X X X X X X X X X│ 3",
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

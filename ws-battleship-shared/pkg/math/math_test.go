package math

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClamp(t *testing.T) {
	for _, tt := range []struct {
		name     string
		val      int
		minVal   int
		maxVal   int
		expected int
	}{
		{
			name:     "value is in range, no need to clamp",
			val:      10,
			minVal:   0,
			maxVal:   15,
			expected: 10,
		},
		{
			name:     "value is not in range, clamped to min",
			val:      -1,
			minVal:   0,
			maxVal:   15,
			expected: 0,
		},
		{
			name:     "value is not in range, clamped to max",
			val:      16,
			minVal:   0,
			maxVal:   15,
			expected: 15,
		},
		{
			name:     "same val",
			val:      15,
			minVal:   15,
			maxVal:   15,
			expected: 15,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Act
			got := Clamp(tt.val, tt.minVal, tt.maxVal)

			// 2. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

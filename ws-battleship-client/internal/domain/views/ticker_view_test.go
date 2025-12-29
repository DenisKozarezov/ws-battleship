package views

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeTicker(t *testing.T) {
	for _, tt := range []struct {
		elapsedTime string
		expected    string
	}{
		{
			elapsedTime: "0s",
			expected:    "00:00",
		},
		{
			elapsedTime: "1s",
			expected:    "00:01",
		},
		{
			elapsedTime: "30s",
			expected:    "00:30",
		},
		{
			elapsedTime: "59s",
			expected:    "00:59",
		},
		{
			elapsedTime: "60s",
			expected:    "01:00",
		},
		{
			elapsedTime: "1m30s",
			expected:    "01:30",
		},
		{
			elapsedTime: "1m59s",
			expected:    "01:59",
		},
		{
			elapsedTime: "2m0s",
			expected:    "02:00",
		},
		{
			elapsedTime: "10m0s",
			expected:    "10:00",
		},
		{
			elapsedTime: "59m0s",
			expected:    "59:00",
		},
		{
			elapsedTime: "59m59s",
			expected:    "59:59",
		},
		{
			elapsedTime: "1h0s",
			expected:    "60:00",
		},
		{
			elapsedTime: "2h0s",
			expected:    "120:00",
		},
	} {
		t.Run(tt.elapsedTime, func(t *testing.T) {
			// 1. Arrange
			elapsedTimeDuration, _ := time.ParseDuration(tt.elapsedTime)
			ticker := TickerView{
				elapsedTime: elapsedTimeDuration,
			}

			// 2. Act
			got := ticker.String()

			// 3. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

package domain

import (
	"testing"
	"ws-battleship-server/internal/config"

	"github.com/stretchr/testify/require"
)

func TestRoomCapacity(t *testing.T) {
	for _, tt := range []struct {
		name     string
		players  map[string]*Player
		expected int
	}{
		{
			name:     "no players - zero capacity",
			players:  map[string]*Player{},
			expected: 0,
		},
		{
			name:     "1 player = 1 capacity",
			players:  map[string]*Player{"1": nil},
			expected: 1,
		},
		{
			name: "3 players = 3 capacity",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
				"3": nil,
			},
			expected: 3,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			room := Room{players: tt.players}

			// 2. Act
			got := room.Capacity()

			// 3. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestRoomIsFull(t *testing.T) {
	for _, tt := range []struct {
		name        string
		players     map[string]*Player
		capacityMax int32
		expected    bool
	}{
		{
			name:        "no players - room is not full",
			players:     map[string]*Player{},
			capacityMax: 3,
			expected:    false,
		},
		{
			name: "2 players out of 3 - room is not full yet",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
			},
			capacityMax: 3,
			expected:    false,
		},
		{
			name: "3 players out of 3 - room is full",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
				"3": nil,
			},
			capacityMax: 3,
			expected:    true,
		},
		{
			name: "4 players out of 3 - room is full",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
				"3": nil,
				"4": nil,
			},
			capacityMax: 3,
			expected:    true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			room := Room{
				cfg:     &config.AppConfig{RoomCapacityMax: tt.capacityMax},
				players: tt.players,
			}

			// 2. Act
			got := room.IsFull()

			// 3. Assert
			require.Equal(t, tt.expected, got)
		})
	}
}

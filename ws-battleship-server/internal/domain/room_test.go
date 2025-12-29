package domain

import (
	"testing"
	"time"
	"ws-battleship-server/internal/config"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	"github.com/stretchr/testify/mock"
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
			name:        "no players, room is not full",
			players:     map[string]*Player{},
			capacityMax: 3,
			expected:    false,
		},
		{
			name: "2 players out of 3, room is not full yet",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
			},
			capacityMax: 3,
			expected:    false,
		},
		{
			name: "3 players out of 3, room is full",
			players: map[string]*Player{
				"1": nil,
				"2": nil,
				"3": nil,
			},
			capacityMax: 3,
			expected:    true,
		},
		{
			name: "4 players out of 3, room is full",
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

func TestCloseRoom(t *testing.T) {
	t.Run("close an empty room", func(t *testing.T) {
		// 1. Arrange
		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		room := NewRoom(t.Context(), &config.AppConfig{
			RoomCapacityMax: 5,
			KeepAlivePeriod: time.Second * 5,
		}, loggerMock)

		// 2. Act
		err := room.Close()

		// 3. Assert
		require.NoError(t, err)
		require.Zerof(t, room.Capacity(), "there should be no players after close")
		_, opened := <-room.readCh
		require.Falsef(t, opened, "readCh must be closed")
	})

	t.Run("close a room with some players", func(t *testing.T) {
		// 1. Arrange
		mockClient := new(MockClient)
		mockClient.On("Close").Return(nil)
		mockClient.On("ID").Return("123")
		mockClient.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient.On("WriteMessages", mock.Anything).Return()

		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		room := NewRoom(t.Context(), &config.AppConfig{
			RoomCapacityMax: 5,
			KeepAlivePeriod: time.Second * 5,
		}, loggerMock)

		room.RegisterNewPlayer(NewPlayer(mockClient, events.ClientMetadata{}))

		// 2. Act
		err := room.Close()

		// 3. Assert
		require.NoError(t, err)
		require.Zerof(t, room.Capacity(), "there should be no players after close")
		_, opened := <-room.readCh
		require.Falsef(t, opened, "readCh must be closed")
	})
}

func TestRegisterNewPlayer(t *testing.T) {
	t.Run("register a new player", func(t *testing.T) {
		// 1. Arrange
		mockClient := new(MockClient)
		mockClient.On("Close").Return(nil)
		mockClient.On("ID").Return("123")
		mockClient.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient.On("WriteMessages", mock.Anything).Return()

		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		room := NewRoom(t.Context(), &config.AppConfig{
			RoomCapacityMax: 5,
			KeepAlivePeriod: time.Second * 5,
		}, loggerMock)

		// 2. Act
		room.RegisterNewPlayer(NewPlayer(mockClient, events.ClientMetadata{}))

		// 3. Assert
		require.Equalf(t, 1, room.Capacity(), "there should be 1 player")
	})

	t.Run("register some new players", func(t *testing.T) {
		// 1. Arrange
		mockClient1 := new(MockClient)
		mockClient1.On("Close").Return(nil)
		mockClient1.On("ID").Return("123")
		mockClient1.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient1.On("WriteMessages", mock.Anything).Return()

		mockClient2 := new(MockClient)
		mockClient2.On("Close").Return(nil)
		mockClient2.On("ID").Return("456")
		mockClient2.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient2.On("WriteMessages", mock.Anything).Return()

		mockClient3 := new(MockClient)
		mockClient3.On("Close").Return(nil)
		mockClient3.On("ID").Return("567")
		mockClient3.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient3.On("WriteMessages", mock.Anything).Return()

		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		room := NewRoom(t.Context(), &config.AppConfig{
			RoomCapacityMax: 5,
			KeepAlivePeriod: time.Second * 5,
		}, loggerMock)

		// 2. Act
		room.RegisterNewPlayer(NewPlayer(mockClient1, events.ClientMetadata{}))
		room.RegisterNewPlayer(NewPlayer(mockClient2, events.ClientMetadata{}))
		room.RegisterNewPlayer(NewPlayer(mockClient3, events.ClientMetadata{}))

		// 3. Assert
		require.Equalf(t, 3, room.Capacity(), "there should be 3 players")
	})
}

func TestUnregisterNewPlayer(t *testing.T) {
	t.Run("unregister 1 player", func(t *testing.T) {
		// 1. Arrange
		mockClient := new(MockClient)
		mockClient.On("Close").Return(nil)
		mockClient.On("ID").Return("123")
		mockClient.On("ReadMessages", mock.Anything, mock.Anything).Return()
		mockClient.On("WriteMessages", mock.Anything).Return()

		loggerMock := new(logger.MockLogger)
		loggerMock.On("Infof", mock.Anything, mock.Anything)

		player := NewPlayer(mockClient, events.ClientMetadata{})
		room := Room{
			players: map[string]*Player{
				"123": player,
			},
			logger: loggerMock,
		}

		// 2. Act
		err := room.UnregisterPlayer(player)

		// 3. Assert
		require.NoError(t, err)
		require.Zerof(t, room.Capacity(), "there should be 0 players")
	})
}

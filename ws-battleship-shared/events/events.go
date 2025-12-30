package events

import (
	"encoding/json"
	"time"
	"ws-battleship-shared/domain"
)

type EventType = string

const (
	ReadBufferBytesMax  = 1024
	WriteBufferBytesMax = 1024
)

const (
	PlayerJoinedEventType = "join"
	GameStartEventType    = "game_start"
)

type Event struct {
	Timestamp string          `json:"timestamp"`
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
}

func NewEvent(eventType EventType, data any) (Event, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	return Event{
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      jsonData,
	}, nil
}

type GameStartEvent struct {
	GameModel domain.GameModel `json:"game_model"`
}

func NewGameStartEvent(gameModel domain.GameModel) (Event, error) {
	return NewEvent(GameStartEventType, GameStartEvent{
		GameModel: gameModel,
	})
}

type PlayerJoinedEvent struct {
	Player *domain.PlayerModel `json:"joined_player"`
}

func NewPlayerJoinedEvent(joinedPlayer *domain.PlayerModel) (Event, error) {
	return NewEvent(PlayerJoinedEventType, PlayerJoinedEvent{
		Player: joinedPlayer,
	})
}

type PlayerLeavedEvent struct {
	Player *domain.PlayerModel `json:"leave_player"`
}

func NewPlayerLeavedEvent(leavePlayer *domain.PlayerModel) (Event, error) {
	return NewEvent(PlayerJoinedEventType, PlayerLeavedEvent{
		Player: leavePlayer,
	})
}

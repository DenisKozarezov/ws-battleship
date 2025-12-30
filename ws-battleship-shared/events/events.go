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

const TimestampFormat = time.RFC3339

const (
	PlayerJoinedEventType = "join"
	PlayerLeavedEventType = "leave"
	GameStartEventType    = "game_start"
	SendMessageType       = "send_message"
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
		Timestamp: time.Now().Format(TimestampFormat),
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
	return NewEvent(PlayerLeavedEventType, PlayerLeavedEvent{
		Player: leavePlayer,
	})
}

type SendMessageEvent struct {
	Sender         string `json:"sender,omitzero"`
	Message        string `json:"message"`
	IsNotification bool   `json:"is_notify,omitzero"`
}

func NewSendMessageEvent(sender string, message string) (Event, error) {
	return NewEvent(SendMessageType, SendMessageEvent{
		Sender:  sender,
		Message: message,
	})
}

func NewChatNotificationEvent(message string) (Event, error) {
	return NewEvent(SendMessageType, SendMessageEvent{
		Message:        message,
		IsNotification: true,
	})
}

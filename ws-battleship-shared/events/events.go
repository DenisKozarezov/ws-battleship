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
	PlayerJoinedEventType      EventType = "player_join"
	PlayerLeftEventType        EventType = "player_leave"
	PlayerTurnEventType        EventType = "player_turn"
	PlayerUpdateStateEventType EventType = "player_update_state"
	GameStartEventType         EventType = "game_start"
	SendMessageType            EventType = "send_message"
)

type Event struct {
	Type      EventType       `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
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

type PlayerJoinedEvent struct {
	Player *domain.PlayerModel `json:"joined_player"`
}

func NewPlayerJoinedEvent(joinedPlayer *domain.PlayerModel) (Event, error) {
	return NewEvent(PlayerJoinedEventType, PlayerJoinedEvent{
		Player: joinedPlayer,
	})
}

type PlayerLeftEvent struct {
	Player *domain.PlayerModel `json:"left_player"`
}

func NewPlayerLeftEvent(leftPlayer *domain.PlayerModel) (Event, error) {
	return NewEvent(PlayerLeftEventType, PlayerLeftEvent{
		Player: leftPlayer,
	})
}

type PlayerTurnEvent struct {
	Player        *domain.PlayerModel `json:"current_player"`
	RemainingTime time.Duration       `json:"remaining_time"`
}

func NewPlayerTurnEvent(player *domain.PlayerModel, remainingTime time.Duration) (Event, error) {
	return NewEvent(PlayerTurnEventType, PlayerTurnEvent{
		Player:        player,
		RemainingTime: remainingTime,
	})
}

type GameStartEvent struct {
	GameModel *domain.GameModel `json:"game_model"`
}

func NewGameStartEvent(gameModel *domain.GameModel) (Event, error) {
	return NewEvent(GameStartEventType, GameStartEvent{
		GameModel: gameModel,
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

type PlayerUpdateStateEvent struct {
	GameModel *domain.GameModel
}

func NewPlayerUpdateStateEvent(gameModel *domain.GameModel) (Event, error) {
	return NewEvent(PlayerUpdateStateEventType, PlayerUpdateStateEvent{
		GameModel: gameModel,
	})
}

package events

import (
	"encoding/json"
	"fmt"
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
	PlayerFireEventType        EventType = "player_fire"
	PlayerUpdateStateEventType EventType = "player_update_state"
	GameStartEventType         EventType = "game_start"
	SendMessageType            EventType = "send_message"
)

type Event struct {
	Type      EventType       `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

func CastTo[T any](e Event) (result T, err error) {
	if err = json.Unmarshal(e.Data, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}
	return
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
	TurningPlayer *domain.PlayerModel `json:"turning_player"`
	TargetPlayer  *domain.PlayerModel `json:"target_player"`
	RemainingTime time.Duration       `json:"remaining_time"`
}

func NewPlayerTurnEvent(turningPlayer, targetPlayer *domain.PlayerModel, remainingTime time.Duration) (Event, error) {
	return NewEvent(PlayerTurnEventType, PlayerTurnEvent{
		TurningPlayer: turningPlayer,
		TargetPlayer:  targetPlayer,
		RemainingTime: remainingTime,
	})
}

type PlayerFireEvent struct {
	PlayerID       domain.ClientID `json:"player_id"`
	PlayerNickname string          `json:"nickname"`
	CellX          byte            `json:"cell_x"`
	CellY          byte            `json:"cell_y"`
}

func NewPlayerFireEvent(metadata domain.ClientMetadata, cellX, cellY byte) (Event, error) {
	return NewEvent(PlayerFireEventType, PlayerFireEvent{
		PlayerID:       metadata.ClientID,
		PlayerNickname: metadata.Nickname,
		CellX:          cellX,
		CellY:          cellY,
	})
}

type GameStartEvent struct {
}

func NewGameStartEvent() (Event, error) {
	return NewEvent(GameStartEventType, GameStartEvent{})
}

type ChatMessageType = string

const (
	MessageType          = "message"
	GameNotificationType = "game_notification"
	RoomNotificationType = "room_notification"
)

type SendMessageEvent struct {
	Sender  string          `json:"sender,omitzero"`
	Message string          `json:"message"`
	Type    ChatMessageType `json:"type"`
}

func NewSendMessageEvent(sender string, msg string) (Event, error) {
	return NewEvent(SendMessageType, SendMessageEvent{
		Sender:  sender,
		Message: msg,
		Type:    MessageType,
	})
}

func NewChatNotificationEvent(msg string, msgType ChatMessageType) (Event, error) {
	return NewEvent(SendMessageType, SendMessageEvent{
		Message: msg,
		Type:    msgType,
	})
}

type PlayerUpdateStateEvent struct {
	GameModel *domain.GameModel `json:"game_model"`
}

func NewPlayerUpdateStateEvent(gameModel *domain.GameModel) (Event, error) {
	return NewEvent(PlayerUpdateStateEventType, PlayerUpdateStateEvent{
		GameModel: gameModel,
	})
}

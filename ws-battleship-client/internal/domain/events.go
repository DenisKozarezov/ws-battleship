package domain

type EventType = string

const (
	PlayerJoinedEventType = "join"
)

type PlayerJoinedEvent struct {
	Nickname string `json:"nickname"`
}

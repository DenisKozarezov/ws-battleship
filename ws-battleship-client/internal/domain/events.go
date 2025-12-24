package domain

type EventType = string

const (
	ReadBufferBytesMax  = 1024
	WriteBufferBytesMax = 1024
)

const (
	PlayerJoinedEventType = "join"
)

type Event struct {
	Type      EventType `json:"type,omitempty"`
	Timestamp string    `json:"timestamp"`
	Data      []byte    `json:"data"`
}

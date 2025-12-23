package domain

import "net/http"

type EventType = string

const (
	PlayerJoinedEventType = "join"
)

type Event struct {
	Type      EventType `json:"type,omitempty"`
	Timestamp string    `json:"timestamp"`
	Data      []byte    `json:"data"`
}

type ClientMetadata struct {
	Nickname string
}

func ParseClientMetadata(r *http.Request) ClientMetadata {
	return ClientMetadata{
		Nickname: r.Header.Get("X-Nickname"),
	}
}

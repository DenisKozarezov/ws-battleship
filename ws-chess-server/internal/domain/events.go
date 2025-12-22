package domain

import "net/http"

type EventType = string

const (
	PlayerJoinedEventType = "join"
)

type ClientMetadata struct {
	Nickname string
}

func ParseClientMetadata(r *http.Request) ClientMetadata {
	return ClientMetadata{
		Nickname: r.Header.Get("X-Nickname"),
	}
}

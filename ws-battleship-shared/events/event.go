package events

import "net/http"

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

type ClientMetadata struct {
	Nickname string
}

func ParseClientMetadataToHeaders(metadata ClientMetadata) http.Header {
	headers := make(http.Header)
	headers.Set("X-Nickname", metadata.Nickname)
	return headers
}

func ParseClientMetadataFromHeaders(r *http.Request) ClientMetadata {
	return ClientMetadata{
		Nickname: r.Header.Get("X-Nickname"),
	}
}

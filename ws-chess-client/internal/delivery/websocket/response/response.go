package response

import "ws-chess-client/internal/domain"

type Event struct {
	Type      domain.EventType `json:"type,omitempty"`
	Timestamp string           `json:"timestamp"`
	Data      []byte           `json:"data"`
}

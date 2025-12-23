package response

import "ws-battleship-client/internal/domain"

type Event struct {
	Type      domain.EventType `json:"type,omitempty"`
	Timestamp string           `json:"timestamp"`
	Data      []byte           `json:"data"`
}

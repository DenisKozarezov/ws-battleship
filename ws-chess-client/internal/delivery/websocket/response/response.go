package response

import (
	"ws-chess-client/internal/domain"
)

type Response struct {
	Type      domain.EventType `json:"type,omitempty"`
	Timestamp string           `json:"timestamp"`
	Data      []byte           `json:"data"`
}

package websocket

import (
	"context"
	"net"
	"ws-battleship-shared/domain"
	serverEvents "ws-battleship-shared/events"
)

type Client interface {
	Metadata() domain.ClientMetadata
	Messages() <-chan serverEvents.Event
	Connect(ctx context.Context, ipv4 net.IP) error
	Shutdown() error
	SendMessage(e serverEvents.Event) error
}

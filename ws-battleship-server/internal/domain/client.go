package domain

import (
	"context"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
)

type Client interface {
	ID() domain.ClientID
	Ping() error
	Close()
	SendMessage(e events.Event) error
	ReadMessages(ctx context.Context, messagesCh chan<- events.Event)
	WriteMessages(ctx context.Context)
}

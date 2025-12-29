package domain

import (
	"context"
	"ws-battleship-shared/events"
)

type ClientID = string

type Client interface {
	ID() ClientID
	Ping() error
	Close()
	SendMessage(eventType events.EventType, obj any) error
	ReadMessages(ctx context.Context, messagesCh chan<- events.Event)
	WriteMessages(ctx context.Context)
}

package application

import (
	"context"
	"sync"
	"ws-battleship-shared/events"
)

type EventHandler = func(context.Context, events.Event) error

type EventBus struct {
	mu       sync.RWMutex
	handlers map[events.EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[events.EventType][]EventHandler),
	}
}

func (r *EventBus) Subscribe(eventType events.EventType, handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	r.mu.Lock()
	r.handlers[eventType] = append(r.handlers[eventType], handlers...)
	r.mu.Unlock()
}

func (r *EventBus) Unsubscribe(eventType events.EventType) {
	r.mu.RLock()
	_, found := r.handlers[eventType]
	r.mu.RUnlock()

	if !found {
		return
	}

	r.mu.Lock()
	delete(r.handlers, eventType)
	r.mu.Unlock()
}

func (r *EventBus) Invoke(ctx context.Context, e events.Event) {
	r.mu.RLock()
	handlers, found := r.handlers[e.Type]
	r.mu.RUnlock()

	if found {
		for i := range handlers {
			handlers[i](ctx, e)
		}
	}
}

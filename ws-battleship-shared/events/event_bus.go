package events

import (
	"context"
	"sync"
)

type EventHandler = func(context.Context, Event) error

type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (r *EventBus) Subscribe(eventType EventType, handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	r.mu.Lock()
	r.handlers[eventType] = append(r.handlers[eventType], handlers...)
	r.mu.Unlock()
}

func (r *EventBus) Unsubscribe(eventType EventType) {
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

func (r *EventBus) Invoke(ctx context.Context, e Event) error {
	r.mu.RLock()
	handlers, found := r.handlers[e.Type]
	r.mu.RUnlock()

	if found {
		for i := range handlers {
			if err := handlers[i](ctx, e); err != nil {
				return err
			}
		}
	}
	return nil
}

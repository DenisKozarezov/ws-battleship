package events

import (
	"reflect"
	"slices"
	"sync"
)

type EventHandler = func(Event) error

type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (b *EventBus) Subscribe(eventType EventType, handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	b.mu.Lock()
	b.handlers[eventType] = append(b.handlers[eventType], handlers...)
	b.mu.Unlock()
}

func (b *EventBus) UnsubscribeAll(eventType EventType) {
	b.mu.Lock()
	delete(b.handlers, eventType)
	b.mu.Unlock()
}

func (b *EventBus) Unsubscribe(eventType EventType, handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range handlers {
		b.handlers[eventType] = slices.DeleteFunc(b.handlers[eventType], func(item EventHandler) bool {
			lhs := reflect.ValueOf(handlers[i]).Pointer()
			rhs := reflect.ValueOf(item).Pointer()
			return rhs == lhs
		})
	}
}

func (b *EventBus) Invoke(event Event) error {
	b.mu.RLock()
	handlers, found := b.handlers[event.Type]
	b.mu.RUnlock()

	if found {
		for i := range handlers {
			if err := handlers[i](event); err != nil {
				return err
			}
		}
	}
	return nil
}

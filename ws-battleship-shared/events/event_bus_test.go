package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventBusSubscribe(t *testing.T) {
	t.Run("subscribe a single handler", func(t *testing.T) {
		// 1. Arrange
		eventBus := NewEventBus()

		// 2. Act
		callback := func(e Event) error { return nil }
		eventBus.Subscribe("some type", callback)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 1, "handler should be subscribed")
	})

	t.Run("subscribe 3 handlers", func(t *testing.T) {
		// 1. Arrange
		eventBus := NewEventBus()

		// 2. Act
		callback := func(e Event) error { return nil }
		eventBus.Subscribe("some type", callback)
		eventBus.Subscribe("some type", callback)
		eventBus.Subscribe("some type", callback)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 3, "all handlers should be subscribed")
	})
}

func TestEventBusInvoke(t *testing.T) {
	t.Run("invoke a single handler", func(t *testing.T) {
		// 1. Arrange
		var callbackInvoked bool
		callback := func(e Event) error {
			callbackInvoked = true
			return nil
		}

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback},
			},
		}

		// 2. Act
		err := eventBus.Invoke(Event{Type: "some type"})

		// 3. Assert
		require.NoError(t, err)
		require.Lenf(t, eventBus.handlers["some type"], 1, "handler should be subscribed")
		require.Truef(t, callbackInvoked, "callback should be invoked")
	})

	t.Run("invoke 3 handlers sequentially", func(t *testing.T) {
		// 1. Arrange
		var counter int
		callback := func(e Event) error {
			counter++
			return nil
		}

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback, callback, callback},
			},
		}

		// 2. Act
		err := eventBus.Invoke(Event{Type: "some type"})

		// 3. Assert
		require.NoError(t, err)
		require.Lenf(t, eventBus.handlers["some type"], 3, "all handlers should be subscribed")
		require.Equalf(t, 3, counter, "not all callbacks were invoked")
	})
}

func TestEventBusUnsubcribe(t *testing.T) {
	t.Run("unsubscribe a single handler in empty slice", func(t *testing.T) {
		// 1. Arrange
		callback := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback},
			},
		}

		// 2. Act
		eventBus.Unsubscribe("some type", callback)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 0, "handler should be unsubscribed")
	})

	t.Run("unsubscribe a single handler in filled-slice", func(t *testing.T) {
		// 1. Arrange
		callback1 := func(e Event) error { return nil }
		callback2 := func(e Event) error { return nil }
		callback3 := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback1, callback2, callback3},
			},
		}

		// 2. Act
		eventBus.Unsubscribe("some type", callback1)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 2, "handler should be unsubscribed")
		require.Equal(t, reflect.ValueOf(callback2).Pointer(), reflect.ValueOf(eventBus.handlers["some type"][0]).Pointer())
		require.Equal(t, reflect.ValueOf(callback3).Pointer(), reflect.ValueOf(eventBus.handlers["some type"][1]).Pointer())
	})

	t.Run("unsubscribe a single handler in filled-slice placed in-middle", func(t *testing.T) {
		// 1. Arrange
		callback1 := func(e Event) error { return nil }
		callback2 := func(e Event) error { return nil }
		callback3 := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback1, callback2, callback3},
			},
		}

		// 2. Act
		eventBus.Unsubscribe("some type", callback2)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 2, "handler should be unsubscribed")
		require.Equal(t, reflect.ValueOf(callback1).Pointer(), reflect.ValueOf(eventBus.handlers["some type"][0]).Pointer())
		require.Equal(t, reflect.ValueOf(callback3).Pointer(), reflect.ValueOf(eventBus.handlers["some type"][1]).Pointer())
	})

	t.Run("unsubscribe all handlers sequentially", func(t *testing.T) {
		// 1. Arrange
		callback1 := func(e Event) error { return nil }
		callback2 := func(e Event) error { return nil }
		callback3 := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback1, callback2, callback3},
			},
		}

		// 2. Act
		eventBus.Unsubscribe("some type", callback1, callback2, callback3)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 0, "all handlers should be unsubscribed")
	})

	t.Run("unsubscribe the whole topic", func(t *testing.T) {
		// 1. Arrange
		callback1 := func(e Event) error { return nil }
		callback2 := func(e Event) error { return nil }
		callback3 := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type": {callback1, callback2, callback3},
			},
		}

		// 2. Act
		eventBus.UnsubscribeAll("some type")

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type"], 0, "all handlers should be unsubscribed")
	})

	t.Run("unsubscribe a handler in empty topic", func(t *testing.T) {
		// 1. Arrange
		callback1 := func(e Event) error { return nil }
		callback2 := func(e Event) error { return nil }
		callback3 := func(e Event) error { return nil }

		eventBus := EventBus{
			handlers: map[EventType][]EventHandler{
				"some type 1": {callback1, callback2, callback3},
				"some type 2": {},
			},
		}

		// 2. Act
		callback := func(e Event) error { return nil }
		eventBus.Unsubscribe("some type 2", callback)

		// 3. Assert
		require.Lenf(t, eventBus.handlers["some type 1"], 3, "handlers in 'some type 1' should remain")
		require.Lenf(t, eventBus.handlers["some type 2"], 0, "handlers in 'some type 2' should remain empty")
	})
}

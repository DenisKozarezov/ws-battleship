package views

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimer(t *testing.T) {
	t.Run("when timer is created, it is stopped by default", func(t *testing.T) {
		// 1. Act
		view := NewTimerView()

		// 2. Assert
		require.True(t, view.isStopped)
	})

	t.Run("timer is immediately stopped after expiration and then invokes a callback", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()
		view.Reset(0.0)

		var callbackInvoked bool
		view.SetExpireCallback(func() {
			callbackInvoked = true
		})

		// 2. Act
		view.Start()
		view.FixedUpdate()

		// 3. Assert
		require.Truef(t, view.isStopped, "timer must be stopped")
		require.Truef(t, callbackInvoked, "callback must be invoked")
	})

	t.Run("reset a running timer", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()
		view.Start()

		// 2. Act
		view.Reset(15.0)
		view.FixedUpdate()

		// 3. Assert
		require.False(t, view.isStopped, "running timer must not be stopped when reset")
	})

	t.Run("reset a stopped timer", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()
		view.Start()

		// 2. Act
		view.Reset(15.0)
		view.FixedUpdate()
		view.Stop()

		// 3. Assert
		require.Truef(t, view.isStopped, "timer should remain stopped")
	})
}

package views

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/require"
)

func TestTimer(t *testing.T) {
	t.Run("when timer is created, it is not stopped", func(t *testing.T) {
		// 1. Act
		view := NewTimerView()

		// 2. Assert
		require.False(t, view.isStopped)
	})

	t.Run("timer is stopped immediatelty when resets to zero", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()

		var callbackInvoked bool
		view.SetExpireCallback(func() {
			callbackInvoked = true
		})

		// 2. Act
		view.Reset(0.0)
		view.Update(spinner.TickMsg{Time: time.Now()})

		// 3. Assert
		require.Truef(t, view.isStopped, "timer must be stopped")
		require.Truef(t, callbackInvoked, "callback must be invoked")
	})

	t.Run("reset a running timer", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()

		// 2. Act
		view.Reset(15.0)

		// 3. Assert
		require.Equal(t, float32(15.0), view.currentTime)
		require.False(t, view.isStopped)
	})

	t.Run("reset a stopped timer", func(t *testing.T) {
		// 1. Arrange
		view := NewTimerView()
		view.Stop()

		// 2. Act
		view.Reset(15.0)

		// 3. Assert
		require.Equal(t, float32(15.0), view.currentTime)
		require.False(t, view.isStopped)
	})
}

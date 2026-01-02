package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestChatClear(t *testing.T) {
	t.Run("clear an empty chat", func(t *testing.T) {
		// 1. Arrange
		chat := NewChatView()

		// 2. Act
		chat.Clear()

		// 3. Assert
		require.Nilf(t, chat.content, "content must be nil")
	})

	t.Run("clear a chat with some messages", func(t *testing.T) {
		// 1. Arrange
		chat := NewChatView()

		// 2. Act
		chat.AppendMessage(ChatMessage{
			Sender:  "sender",
			Message: "some message",
		})
		chat.Clear()

		// 3. Assert
		require.Nilf(t, chat.content, "content must be nil")
	})
}

func TestTypingMessage(t *testing.T) {
	t.Run("check callback is invoked after typing message", func(t *testing.T) {
		// 1. Arrange
		chat := NewChatView()
		chat.textarea.SetValue("some message")

		var callbackInvoked bool
		chat.SetMessageTypedHandler(func(s string) {
			callbackInvoked = true
		})

		// 2. Act
		chat.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// 3. Assert
		require.Truef(t, callbackInvoked, "callback must be invoked")
	})

	t.Run("check callback is not invoked when player didn't type anything but pressed Enter", func(t *testing.T) {
		// 1. Arrange
		chat := NewChatView()
		chat.textarea.SetValue("")

		var callbackInvoked bool
		chat.SetMessageTypedHandler(func(s string) {
			callbackInvoked = true
		})

		// 2. Act
		chat.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// 3. Assert
		require.Nilf(t, chat.content, "content must be nil")
		require.Falsef(t, callbackInvoked, "callback shouldn't be invoked")
	})
}

func TestFormatChatMessage(t *testing.T) {
	t.Run("append a non-notification message", func(t *testing.T) {
		// 1. Arrange
		now := time.Date(2025, 1, 1, 15, 0, 35, 0, time.UTC) // 2025-01-01 15:00:35 UTC+0

		// 2. Act
		got := formatChatMessage(ChatMessage{
			Sender:    "Nickname",
			Message:   "some message",
			Timestamp: now,
		})

		// 3. Assert
		require.Equal(t, "15:00:35 Nickname: some message", got)
	})

	t.Run("append a notification message", func(t *testing.T) {
		// 1. Arrange
		now := time.Date(2025, 1, 1, 15, 0, 35, 0, time.UTC) // 2025-01-01 15:00:35 UTC+0

		// 2. Act
		got := formatChatMessage(ChatMessage{
			Message:        "some message",
			Timestamp:      now,
			IsNotification: true,
		})

		// 3. Assert
		require.Equal(t, "15:00:35 some message", strings.TrimSpace(got))
	})
}

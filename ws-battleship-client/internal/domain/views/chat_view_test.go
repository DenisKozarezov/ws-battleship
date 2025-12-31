package views

import (
	"testing"

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

package views

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestClickStateIsResetAfterSomeTime(t *testing.T) {
	// 1. Arrange
	view := NewButtonView("")
	view.SetFocus(true)

	// 2. Act
	view.Click()
	time.Sleep(time.Millisecond * 100)
	view.FixedUpdate()

	// 3. Assert
	assertButtonIsNotClicked(t, view)
}

func TestClickOnceToAvoidUnintendedSpam(t *testing.T) {
	// 1. Arrange
	view := NewButtonView("")
	view.SetFocus(true)

	var counter int
	view.SetClickHandler(func() {
		counter++
	})

	// 2. Act
	view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// 3. Assert
	require.Equalf(t, 1, counter, "button should be clicked only once")
	assertButtonIsClicked(t, view)
}

func TestButtonClick(t *testing.T) {
	t.Run("click an unfocused button programmatically and check it's clicked", func(t *testing.T) {
		// 1. Arrange
		var callbackCalled bool
		view := NewButtonView("")
		view.SetFocus(false)
		view.SetClickHandler(func() {
			callbackCalled = true
		})

		// 2. Act
		view.Click()

		// 3. Assert
		require.Truef(t, callbackCalled, "callback should be invoked")
		assertButtonIsClicked(t, view)
	})

	t.Run("click a focused button programmatically and check it's clicked", func(t *testing.T) {
		// 1. Arrange
		var callbackCalled bool
		view := NewButtonView("")
		view.SetFocus(true)
		view.SetClickHandler(func() {
			callbackCalled = true
		})

		// 2. Act
		view.Click()

		// 3. Assert
		require.Truef(t, callbackCalled, "callback should be invoked")
		assertButtonIsClicked(t, view)
	})

	t.Run("press enter and check a focused button is clicked", func(t *testing.T) {
		// 1. Arrange
		var callbackCalled bool
		view := NewButtonView("")
		view.SetFocus(true)
		view.SetClickHandler(func() {
			callbackCalled = true
		})

		// 2. Act
		view.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// 3. Assert
		require.Truef(t, callbackCalled, "callback should be invoked")
		assertButtonIsClicked(t, view)
	})

	t.Run("press enter and check an unfocused button is not clicked", func(t *testing.T) {
		// 1. Arrange
		var callbackCalled bool
		view := NewButtonView("")
		view.SetFocus(false)
		view.SetClickHandler(func() {
			callbackCalled = true
		})

		// 2. Act
		view.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// 3. Assert
		require.Falsef(t, callbackCalled, "callback shouldn't be invoked")
		assertButtonIsNotClicked(t, view)
	})
}

func assertButtonIsClicked(t *testing.T, view *ButtonView) {
	require.Truef(t, view.isClicked, "button must be clicked")
	require.NotNilf(t, view.currentStyle, "current style must not be nil")
	require.Equalf(t, *view.currentStyle, view.styles.clickedStyle, "current style is not a click style")
}

func assertButtonIsNotClicked(t *testing.T, view *ButtonView) {
	require.Falsef(t, view.isClicked, "button shouldn't be clicked")
	require.NotNilf(t, view.currentStyle, "current style must not be nil")
	require.NotEqualf(t, *view.currentStyle, view.styles.clickedStyle, "current style shouldn't be a click style")
}

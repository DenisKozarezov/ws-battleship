package view

import (
	"strings"
	"ws-battleship-client/internal/domain/model"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	chatWidth  = 70
	chatHeight = 10
)

var (
	logsStyle = lipgloss.NewStyle().
			Width(chatWidth).Height(chatHeight). //Background(lipgloss.Color("#696969")).
			Border(lipgloss.RoundedBorder())

	senderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

type ChatView struct {
	game     *model.GameModel
	textarea textarea.Model
	viewport viewport.Model
}

func NewChatView(game *model.GameModel) *ChatView {
	ta := textarea.New()
	ta.Placeholder = "Press Enter to send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(chatWidth)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(chatWidth, chatHeight)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &ChatView{
		textarea: ta,
		viewport: vp,
		game:     game,
	}
}

func (m *ChatView) Init() tea.Cmd {
	return textarea.Blink
}

func (m *ChatView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.game.Messages = append(m.game.Messages, senderStyle.Render("You: ")+m.textarea.Value())
			m.SetContent(m.game.Messages)
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *ChatView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, logsStyle.Render(m.viewport.View()), m.textarea.View())
}

func (m *ChatView) SetContent(content []string) {
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(content, "\n")))
}

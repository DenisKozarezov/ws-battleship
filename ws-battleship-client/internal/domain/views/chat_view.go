package views

import (
	"strings"

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
	content  []string
	textarea textarea.Model
	viewport viewport.Model
}

func NewChatView() *ChatView {
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
	}
}

func (v *ChatView) Init() tea.Cmd {
	return textarea.Blink
}

func (v *ChatView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	v.textarea, tiCmd = v.textarea.Update(msg)
	v.viewport, vpCmd = v.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			v.content = append(v.content, senderStyle.Render("You: ")+v.textarea.Value())
			v.SetContent(v.content)
			v.textarea.Reset()
			v.viewport.GotoBottom()
		}
	}

	return v, tea.Batch(tiCmd, vpCmd)
}

func (v *ChatView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, logsStyle.Render(v.viewport.View()), v.textarea.View())
}

func (v *ChatView) SetContent(content []string) {
	v.viewport.SetContent(lipgloss.NewStyle().Width(v.viewport.Width).Render(strings.Join(content, "\n")))
}

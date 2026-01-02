package views

import (
	"strings"
	"time"

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
			Width(chatWidth).Height(chatHeight).
			Border(lipgloss.RoundedBorder())

	senderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	notificationStyle = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Background(lipgloss.Color("6")).
				Foreground(lipgloss.Color("#ffffff")).
				Bold(true)

	viewportStyle = lipgloss.NewStyle().Width(chatWidth)
)

type ChatView struct {
	content             []string
	textarea            textarea.Model
	viewport            viewport.Model
	messageTypedHandler func(msg string)
}

func NewChatView() *ChatView {
	ta := textarea.New()
	ta.Placeholder = "Press Enter to send a message..."

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(chatWidth)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	va := viewport.New(chatWidth, chatHeight)
	va.KeyMap = viewport.KeyMap{}
	va.SetContent("Welcome to the chat room!\nType a message and press Enter to send.")

	return &ChatView{
		textarea: ta,
		viewport: va,
	}
}

func (v *ChatView) Init() tea.Cmd {
	v.textarea.Focus()
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
			if len(v.textarea.Value()) == 0 {
				break
			}

			if v.messageTypedHandler != nil {
				v.messageTypedHandler(v.textarea.Value())
			}
			v.textarea.Reset()
		}
	}

	return v, tea.Batch(tiCmd, vpCmd)
}

func (v *ChatView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, logsStyle.Render(v.viewport.View()), v.textarea.View())
}

type ChatMessage struct {
	Sender         string
	Message        string
	IsNotification bool
	Timestamp      time.Time
}

func (v *ChatView) AppendMessage(msg ChatMessage) {
	v.setContent(append(v.content, formatChatMessage(msg)))
	v.viewport.GotoBottom()
}

func (v *ChatView) Clear() {
	v.setContent(nil)
}

func (v *ChatView) SetMessageTypedHandler(fn func(string)) {
	v.messageTypedHandler = fn
}

func formatChatMessage(msg ChatMessage) string {
	timestamp := msg.Timestamp.Format(time.TimeOnly)

	if msg.IsNotification {
		return lipgloss.PlaceHorizontal(chatWidth, lipgloss.Center, notificationStyle.Render(" "+timestamp+" "+msg.Message+" "))
	} else {
		return senderStyle.Render(timestamp+" "+msg.Sender+": ") + msg.Message
	}
}

func (v *ChatView) setContent(content []string) {
	v.content = content

	if len(content) == 0 {
		v.viewport.SetContent("")
		return
	}

	v.viewport.SetContent(viewportStyle.Render(strings.Join(content, "\n")))
}

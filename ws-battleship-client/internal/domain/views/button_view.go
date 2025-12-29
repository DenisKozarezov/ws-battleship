package views

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	defaultStyle = lipgloss.NewStyle().Bold(true).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#514A85"))

	focusedStyle = lipgloss.NewStyle().Bold(true).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7068BA"))

	clickedStyle = lipgloss.NewStyle().Bold(true).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#655AF2"))

	disabledStyle = lipgloss.NewStyle().Bold(true).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#55555E"))
)

const (
	clickTime = 0.05
)

type ButtonStyles struct {
	defaultStyle  lipgloss.Style
	focusedStyle  lipgloss.Style
	clickedStyle  lipgloss.Style
	disabledStyle lipgloss.Style
}

type ButtonView struct {
	styles ButtonStyles
	opts   []ButtonOption

	clickHandler func()

	text         string
	isEnabled    bool
	isFocused    bool
	isClicked    bool
	resetTime    time.Time
	currentStyle *lipgloss.Style
}

func NewButtonView(text string, opts ...ButtonOption) *ButtonView {
	return &ButtonView{
		styles: ButtonStyles{
			defaultStyle:  defaultStyle,
			focusedStyle:  focusedStyle,
			clickedStyle:  clickedStyle,
			disabledStyle: disabledStyle,
		},
		opts:         opts,
		text:         " " + text + " ",
		currentStyle: &defaultStyle,
	}
}

func (v *ButtonView) Init() tea.Cmd {
	for _, opt := range v.opts {
		opt(v)
	}

	v.SetFocus(false)
	v.SetEnabled(true)
	v.Reset()
	return nil
}

func (v *ButtonView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !v.isFocused {
		return v, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if !v.isClicked {
				v.Click()
			}
		}
	}

	return v, nil
}

func (v *ButtonView) FixedUpdate() {
	if v.isClicked && time.Since(v.resetTime).Seconds() >= clickTime {
		v.Reset()
	}
}

func (v *ButtonView) View() string {
	return v.currentStyle.Render(v.text)
}

func (v *ButtonView) Click() {
	v.isClicked = true
	v.currentStyle = &v.styles.clickedStyle
	v.resetTime = time.Now()

	if v.clickHandler != nil {
		v.clickHandler()
	}
}

func (v *ButtonView) Reset() {
	v.isClicked = false
	v.currentStyle = &v.styles.defaultStyle
}

func (v *ButtonView) SetFocus(isFocus bool) {
	v.isFocused = isFocus

	if isFocus {
		v.currentStyle = &v.styles.focusedStyle
	}
}

func (v *ButtonView) SetEnabled(isEnabled bool) {
	v.isEnabled = isEnabled

	if isEnabled {
		v.currentStyle = &v.styles.defaultStyle
	} else {
		v.currentStyle = &v.styles.disabledStyle
	}
}

func (v *ButtonView) SetClickHandler(fn func()) {
	v.clickHandler = fn
}

type ButtonOption = func(*ButtonView)

func WithWidth(width int) ButtonOption {
	return func(v *ButtonView) {
		v.styles.defaultStyle = v.styles.defaultStyle.Width(width)
		v.styles.focusedStyle = v.styles.focusedStyle.Width(width)
		v.styles.clickedStyle = v.styles.clickedStyle.Width(width)
		v.styles.disabledStyle = v.styles.disabledStyle.Width(width)
	}
}

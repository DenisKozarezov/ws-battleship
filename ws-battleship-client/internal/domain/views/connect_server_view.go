package views

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConnectServerView struct {
	spinner spinner.Model
}

func NewConnectServerView() *ConnectServerView {
	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &ConnectServerView{
		spinner: s,
	}
}

func (v *ConnectServerView) Init() tea.Cmd {
	return v.spinner.Tick
}

func (v *ConnectServerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	v.spinner, cmd = v.spinner.Update(msg)
	return v, cmd
}

func (v *ConnectServerView) FixedUpdate() {
}

func (v *ConnectServerView) View() string {
	return v.spinner.View() + " Connecting to server..."
}

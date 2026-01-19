package views

import (
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inputTextStyle = lipgloss.NewStyle().Align(lipgloss.Center).Border(lipgloss.ThickBorder())
)

type ConnectFunc func(ip net.IP)

type MainMenuView struct {
	ConnectFunc ConnectFunc

	IPv4Error     error
	ipv4InputView *IPv4InputView
	connectButton *ButtonView
}

func NewMainMenuView() *MainMenuView {
	return &MainMenuView{
		ipv4InputView: NewIPv4InputView(),
		connectButton: NewButtonView("Connect"),
	}
}

func (v *MainMenuView) Init() tea.Cmd {
	v.connectButton.SetClickHandler(v.onConnectHandler)
	return tea.Batch(v.ipv4InputView.Init(), v.connectButton.Init())
}

func (v *MainMenuView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return v, tea.Quit
		case tea.KeyEnter:
			v.connectButton.Click()
		}
	}

	var cmds []tea.Cmd
	_, cmd = v.ipv4InputView.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = v.connectButton.Update(msg)
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

func (v *MainMenuView) FixedUpdate() {
	v.connectButton.FixedUpdate()
}

func (v *MainMenuView) View() string {
	ipv4Input := lipgloss.JoinVertical(lipgloss.Center, inputTextStyle.Render(v.ipv4InputView.View()), v.connectButton.View())
	if v.IPv4Error != nil {
		err := lipgloss.PlaceVertical(3, lipgloss.Center, v.IPv4Error.Error())
		ipv4Input = lipgloss.JoinHorizontal(lipgloss.Top, ipv4Input, err)
	}

	return ipv4Input
}

func (v *MainMenuView) onConnectHandler() {
	var ipv4 net.IP
	ipv4, v.IPv4Error = v.ipv4InputView.IPAddress()
	if v.IPv4Error != nil {
		return
	}

	if v.ConnectFunc != nil {
		v.ConnectFunc(ipv4)
	}
}

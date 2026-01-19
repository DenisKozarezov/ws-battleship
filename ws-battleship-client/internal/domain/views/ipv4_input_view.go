package views

import (
	"errors"
	"net"
	"regexp"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type IPv4InputView struct {
	textInput textinput.Model
}

func NewIPv4InputView() *IPv4InputView {
	textInput := textinput.New()
	textInput.Placeholder = "Enter Server IP..."
	textInput.Focus()
	textInput.CharLimit = 15
	textInput.Width = 20

	return &IPv4InputView{textInput: textInput}
}

func (v *IPv4InputView) Init() tea.Cmd {
	return v.textInput.Cursor.BlinkCmd()
}

func (v *IPv4InputView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	v.textInput, cmd = v.textInput.Update(msg)
	return v, cmd
}

func (v *IPv4InputView) View() string {
	return v.textInput.View()
}

func (v *IPv4InputView) IPAddress() (net.IP, error) {
	ipv4Addr := v.textInput.Value()
	return net.ParseIP(ipv4Addr), validateIPInputText(ipv4Addr)
}

var (
	ErrInvalidAddress = errors.New("invalid IPv4 format: expected format 255.255.255.255")
)

func validateIPInputText(input string) error {
	const ipv4Regex = `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, err := regexp.MatchString(ipv4Regex, input)
	if !matched {
		return errors.Join(ErrInvalidAddress, err)
	}
	return nil
}

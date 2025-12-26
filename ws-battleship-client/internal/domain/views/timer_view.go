package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type TimerView struct {
	spinner        spinner.Model
	expireTime     time.Time
	currentTime    float32
	isStopped      bool
	expireCallback func()
}

func NewTimerView() *TimerView {
	return &TimerView{
		spinner: spinner.New(spinner.WithSpinner(spinner.Points)),
	}
}

func (m *TimerView) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *TimerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isStopped {
		return m, nil
	}

	switch val := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.currentTime = max(0, float32(m.expireTime.Sub(val.Time).Seconds()))

		if m.currentTime <= 0 {
			m.Stop()
			if m.expireCallback != nil {
				m.expireCallback()
			}
		}

		return m, cmd
	}
	return m, nil
}

func (m *TimerView) View() string {
	return fmt.Sprintf("%s %.0f %s", m.spinner.View(), m.currentTime, "sec")
}

func (m *TimerView) Reset(startTime float32) {
	m.expireTime = time.Now().Add(time.Second * time.Duration(startTime))
	m.currentTime = startTime
	m.isStopped = false
}

func (m *TimerView) Stop() {
	m.isStopped = true
}

func (m *TimerView) SetExpireCallback(fn func()) {
	m.expireCallback = fn
}

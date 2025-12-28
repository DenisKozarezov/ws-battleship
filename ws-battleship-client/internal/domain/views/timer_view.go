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
	currentTime    time.Duration
	isStopped      bool
	expireCallback func()
}

func NewTimerView() *TimerView {
	return &TimerView{
		isStopped: true,
		spinner:   spinner.New(spinner.WithSpinner(spinner.Points)),
	}
}

func (m *TimerView) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *TimerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *TimerView) FixedUpdate() {
	if m.isStopped {
		return
	}

	m.currentTime = time.Until(m.expireTime)

	if m.currentTime.Seconds() <= 0.0 {
		m.Stop()
		if m.expireCallback != nil {
			m.expireCallback()
		}
	}
}

func (m *TimerView) View() string {
	return fmt.Sprintf("%s %.0f %s", m.spinner.View(), m.currentTime.Abs().Seconds(), "sec")
}

func (m *TimerView) Reset(elapsedTime float32) {
	if elapsedTime < 0.0 {
		elapsedTime = 0.0
	}

	m.expireTime = time.Now().Add(time.Second * time.Duration(elapsedTime))
}

func (m *TimerView) Start() {
	m.isStopped = false
}

func (m *TimerView) Stop() {
	m.isStopped = true
}

func (m *TimerView) SetExpireCallback(fn func()) {
	m.expireCallback = fn
}

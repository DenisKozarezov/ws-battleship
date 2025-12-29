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

func (v *TimerView) Init() tea.Cmd {
	return v.spinner.Tick
}

func (v *TimerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *TimerView) FixedUpdate() {
	if v.isStopped {
		return
	}

	v.currentTime = time.Until(v.expireTime)

	if v.currentTime.Seconds() <= 0.0 {
		v.Stop()
		if v.expireCallback != nil {
			v.expireCallback()
		}
	}
}

func (v *TimerView) View() string {
	return fmt.Sprintf("%s %.0f %s", v.spinner.View(), v.currentTime.Abs().Seconds(), "sec")
}

func (v *TimerView) Reset(timeInSeconds int) {
	if timeInSeconds < 0 {
		timeInSeconds = 0
	}

	v.expireTime = time.Now().Add(time.Second * time.Duration(timeInSeconds))
}

func (v *TimerView) Start() {
	v.isStopped = false
}

func (v *TimerView) Stop() {
	v.isStopped = true
}

func (v *TimerView) SetExpireCallback(fn func()) {
	v.expireCallback = fn
}

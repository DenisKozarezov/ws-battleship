package views

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickerView struct {
	isStopped   bool
	startTime   time.Time
	elapsedTime time.Duration
}

func NewTickerView() *TickerView {
	return &TickerView{
		isStopped: true,
	}
}

func (m *TickerView) Init() tea.Cmd {
	m.Reset()
	return nil
}

func (m *TickerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TickerView) FixedUpdate() {
	if m.isStopped {
		return
	}

	m.elapsedTime = time.Since(m.startTime)
}

func (m *TickerView) View() string {
	return m.elapsedTimeString()
}

func (m *TickerView) Start() {
	m.isStopped = false
	m.Reset()
}

func (m *TickerView) Stop() {
	m.isStopped = true
}

func (m *TickerView) Reset() {
	m.startTime = time.Now()
}

func (m *TickerView) ElapsedTime() time.Duration {
	return m.elapsedTime
}

func (m *TickerView) elapsedTimeString() string {
	elapsedSeconds := int(m.elapsedTime.Seconds())
	minutes := elapsedSeconds / 60
	elapsedSeconds %= 60

	var builder strings.Builder
	builder.Grow(5)
	if minutes < 10 {
		builder.WriteRune('0')
	}
	builder.WriteString(strconv.FormatInt(int64(minutes), 10))
	builder.WriteRune(':')

	if elapsedSeconds < 10 {
		builder.WriteRune('0')
	}
	builder.WriteString(strconv.FormatInt(int64(elapsedSeconds), 10))

	return builder.String()
}

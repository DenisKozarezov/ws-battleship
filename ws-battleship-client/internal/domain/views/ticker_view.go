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

func (v *TickerView) Init() tea.Cmd {
	v.Reset()
	return nil
}

func (v *TickerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

func (v *TickerView) FixedUpdate() {
	if v.isStopped {
		return
	}

	v.elapsedTime = time.Since(v.startTime)
}

func (v *TickerView) View() string {
	return v.String()
}

func (v *TickerView) Start() {
	v.isStopped = false
	v.Reset()
}

func (v *TickerView) Stop() {
	v.isStopped = true
}

func (v *TickerView) Reset() {
	v.startTime = time.Now()
}

func (v *TickerView) ElapsedTime() time.Duration {
	return v.elapsedTime
}

func (v *TickerView) String() string {
	elapsedSeconds := int(v.elapsedTime.Seconds())
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

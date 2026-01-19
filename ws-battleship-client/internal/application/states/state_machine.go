package states

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type State interface {
	OnEnter()
	OnExit()
	FixedUpdate()
	View() tea.Model
}

type StateMachine interface {
	Context() context.Context
	SwitchState(newState State)
	FixedUpdate()
}

type stateMachine struct {
	currentState         State
	stateSwitchedHandler func(view tea.Model)
}

func NewStateMachine() *stateMachine {
	return &stateMachine{}
}

func (m *stateMachine) SwitchState(newState State) {
	if m.currentState != nil {
		m.currentState.OnExit()
	}

	if newState == nil {
		return
	}

	m.currentState = newState
	m.currentState.OnEnter()

	if m.stateSwitchedHandler != nil {
		m.stateSwitchedHandler(m.currentState.View())
	}
}

func (m *stateMachine) FixedUpdate() {
	if m.currentState != nil {
		m.currentState.FixedUpdate()
	}
}

func (m *stateMachine) CurrentState() State {
	return m.currentState
}

func (m *stateMachine) Context() context.Context {
	return context.Background()
}

func (m *stateMachine) SetStateSwitchedHandler(fn func(model tea.Model)) {
	m.stateSwitchedHandler = fn
}

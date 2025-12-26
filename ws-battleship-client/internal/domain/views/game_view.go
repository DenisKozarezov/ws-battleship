package views

import (
	"ws-battleship-client/internal/domain/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	defaultText = lipgloss.NewStyle().Background(lipgloss.NoColor{})

	boardStyle = lipgloss.NewStyle().Padding(1).Align(lipgloss.Center).Border(lipgloss.NormalBorder())

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type GameView struct {
	game       *models.GameModel
	chatView   *ChatView
	leftBoard  *BoardView
	rightBoard *BoardView
	timerView  *TimerView

	currentBoard *BoardView
}

func NewGameView(game *models.GameModel) *GameView {
	return &GameView{
		game:       game,
		leftBoard:  NewBoardView(game.Player1),
		rightBoard: NewBoardView(game.Player2),
		chatView:   NewChatView(game),
		timerView:  NewTimerView(),
	}
}

func (m *GameView) Init() tea.Cmd {
	m.timerView.Reset(30.0)

	m.GiveTurnToPlayer(m.rightBoard)

	return tea.Batch(m.leftBoard.Init(),
		m.rightBoard.Init(),
		m.chatView.Init(),
		m.timerView.Init(),
		tea.SetWindowTitle("Battleship"))
}

func (m *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd
	_, cmd := m.leftBoard.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = m.rightBoard.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = m.chatView.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = m.timerView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *GameView) View() string {
	boards := boardStyle.Render(m.renderPlayersBoards())
	return lipgloss.JoinHorizontal(lipgloss.Top, boards, " ", m.chatView.View())
}

func (m *GameView) GiveTurnToPlayer(board *BoardView) {
	m.currentBoard = board
	m.currentBoard.SetSelectable(true)
}

func (m *GameView) renderPlayersBoards() string {
	help := helpStyle.Align(lipgloss.Center).Render("Press ↑ ↓ → ← to Navigate\nPress Enter to Fire")

	margin := lipgloss.Place(30, 0, lipgloss.Center, lipgloss.Top, "YOUR TURN!\n"+m.timerView.View()+"\n\n"+help+"\n\n")

	return lipgloss.JoinHorizontal(lipgloss.Center, m.leftBoard.View(), margin, m.rightBoard.View())
}

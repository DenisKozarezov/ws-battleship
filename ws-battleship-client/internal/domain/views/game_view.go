package views

import (
	"ws-battleship-client/internal/domain/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	defaultText = lipgloss.NewStyle().Background(lipgloss.NoColor{})

	boardStyle = lipgloss.NewStyle().
			Padding(1).Align(lipgloss.Center).Border(lipgloss.NormalBorder())

	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#4a4a8a")).
			Foreground(lipgloss.Color("#ffffff"))

	highlightCell = lipgloss.NewStyle().
			Background(lipgloss.Color("#37DB76")).
			Foreground(lipgloss.Color("#ffffff")).Bold(true)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type GameView struct {
	game       *model.GameModel
	chatView   *ChatView
	leftBoard  *BoardView
	rightBoard *BoardView
}

func NewGameView(game *model.GameModel) *GameView {
	return &GameView{
		game:       game,
		chatView:   NewChatView(game),
		leftBoard:  NewBoardView(game.Player1),
		rightBoard: NewBoardView(game.Player2),
	}
}

func (m *GameView) Init() tea.Cmd {
	return tea.SetWindowTitle("Battleship")
}

func (m *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.leftBoard.Update(msg)
	m.rightBoard.Update(msg)
	m.chatView.Update(msg)
	return m, nil
}

func (m *GameView) View() string {
	boards := boardStyle.Render(m.renderPlayersBoards())
	return lipgloss.JoinHorizontal(lipgloss.Top, boards, " ", m.chatView.View())
}

func (m *GameView) renderPlayersBoards() string {
	help := helpStyle.Align(lipgloss.Center).Render("Press Arrows to Navigate\nPress Enter to Fire")

	margin := lipgloss.Place(30, 0, lipgloss.Center, lipgloss.Top, "YOUR TURN!\n15 sec\n\n"+help+"\n\n")

	return lipgloss.JoinHorizontal(lipgloss.Center, m.leftBoard.View(), margin, m.rightBoard.View())
}

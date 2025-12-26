package view

import (
	"fmt"
	"strings"
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
	game           *model.GameModel
	chatView       *ChatView
	selectedRowIdx int
	selectedColIdx int
	boardSize      int
	alphabet       string
}

func NewGameView(game *model.GameModel) *GameView {
	var alphabet []rune
	for i, r := range game.Player1.Board.Alphabet() {
		alphabet = append(alphabet, r)

		if i < len(game.Player1.Board.Alphabet())-1 {
			alphabet = append(alphabet, ' ')
		}
	}

	return &GameView{
		chatView:  NewChatView(game),
		game:      game,
		boardSize: game.Player1.Board.Size(),
		alphabet:  string(alphabet),
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
		case tea.KeyUp:
			m.selectionUp()
		case tea.KeyDown:
			m.selectionDown()
		case tea.KeyLeft:
			m.selectionLeft()
		case tea.KeyRight:
			m.selectionRight()
		}
	}

	m.chatView.Update(msg)
	return m, nil
}

func (m *GameView) View() string {
	boards := boardStyle.Render(m.renderPlayersBoards())
	return lipgloss.JoinHorizontal(lipgloss.Top, boards, " ", m.chatView.View())
}

func (m *GameView) selectionUp() {
	m.selectedRowIdx = max(0, m.selectedRowIdx-1)
}

func (m *GameView) selectionDown() {
	m.selectedRowIdx = min(m.boardSize-1, m.selectedRowIdx+1)
}

func (m *GameView) selectionLeft() {
	m.selectedColIdx = max(0, m.selectedColIdx-2)
}

func (m *GameView) selectionRight() {
	m.selectedColIdx = min(len(m.alphabet)-1, m.selectedColIdx+2)
}

func (m *GameView) renderPlayersBoards() string {
	b1Lines := m.game.Player1.Board.Lines()
	b2Lines := m.game.Player2.Board.Lines()

	var builder1 strings.Builder
	var builder2 strings.Builder
	var numbers string = "\n"
	for i := range b1Lines {
		builder1.WriteString(m.renderBorderRow(b1Lines[i], i))
		builder2.WriteString(m.renderBorderRow(b2Lines[i], i))
		numbers += fmt.Sprintf("%d\n", i+1)

		if i < len(b1Lines)-1 {
			builder1.WriteRune('\n')
			builder2.WriteRune('\n')
		}
	}

	leftBoard := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(builder1.String())
	rightBoard := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(builder2.String())

	leftBoard = lipgloss.JoinVertical(lipgloss.Center, m.alphabet, leftBoard)
	rightBoard = lipgloss.JoinVertical(lipgloss.Center, m.alphabet, rightBoard)

	leftBoard = lipgloss.JoinHorizontal(lipgloss.Center, numbers, leftBoard)
	rightBoard = lipgloss.JoinHorizontal(lipgloss.Center, numbers, rightBoard)

	leftBoard = lipgloss.JoinVertical(lipgloss.Center, leftBoard, m.game.Player1.Nickname)
	rightBoard = lipgloss.JoinVertical(lipgloss.Center, rightBoard, m.game.Player2.Nickname)

	help := helpStyle.Align(lipgloss.Center).Render("Press Arrows to Navigate\nPress Enter to Fire")

	margin := lipgloss.Place(30, 0, lipgloss.Center, lipgloss.Top, "YOUR TURN!\n15 sec\n\n"+help+"\n\n")

	return lipgloss.JoinHorizontal(lipgloss.Center, leftBoard, margin, rightBoard)
}

func (m *GameView) renderBorderRow(str string, currentRowIdx int) string {
	if m.selectedRowIdx == currentRowIdx {
		return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, highlightCell, highlightStyle)
	}
	return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, highlightStyle, defaultText)
}

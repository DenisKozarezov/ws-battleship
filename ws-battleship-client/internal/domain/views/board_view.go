package views

import (
	"strconv"
	"strings"
	"ws-battleship-client/internal/domain/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#4a4a8a")).
			Foreground(lipgloss.Color("#ffffff"))

	highlightCell = lipgloss.NewStyle().
			Background(lipgloss.Color("#37DB76")).
			Foreground(lipgloss.Color("#ffffff")).Bold(true)
)

type BoardView struct {
	player         *models.Player
	selectedRowIdx int
	selectedColIdx int
	boardSize      int
	alphabet       string
	isSelectable   bool
}

func NewBoardView(player *models.Player) *BoardView {
	var alphabet []rune
	for i, r := range player.Board.Alphabet() {
		alphabet = append(alphabet, r)

		if i < len(player.Board.Alphabet())-1 {
			alphabet = append(alphabet, ' ')
		}
	}
	return &BoardView{
		player:       player,
		boardSize:    player.Board.Size(),
		alphabet:     string(alphabet),
		isSelectable: false,
	}
}

func (m *BoardView) Init() tea.Cmd {
	return nil
}

func (m *BoardView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.isSelectable {
				m.selectionUp()
			}
		case tea.KeyDown:
			if m.isSelectable {
				m.selectionDown()
			}
		case tea.KeyLeft:
			if m.isSelectable {
				m.selectionLeft()
			}
		case tea.KeyRight:
			if m.isSelectable {
				m.selectionRight()
			}
		}
	}

	return m, nil
}

func (m *BoardView) View() string {
	b1Lines := m.player.Board.Lines()

	var board strings.Builder
	var numbers strings.Builder
	numbers.WriteRune('\n')

	for i := range b1Lines {
		board.WriteString(m.renderBorderRow(b1Lines[i], i))
		numbers.WriteString(strconv.FormatInt(int64(i+1), 10))

		if i < len(b1Lines)-1 {
			board.WriteRune('\n')
			numbers.WriteRune('\n')
		}
	}

	boardView := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(board.String())
	boardView = lipgloss.JoinVertical(lipgloss.Center, m.alphabet, boardView)
	boardView = lipgloss.JoinHorizontal(lipgloss.Center, numbers.String(), boardView)
	boardView = lipgloss.JoinVertical(lipgloss.Center, boardView, m.player.Nickname)

	return boardView
}

func (m *BoardView) SetSelectable(isSelectable bool) {
	m.isSelectable = isSelectable
}

func (m *BoardView) renderBorderRow(str string, currentRowIdx int) string {
	if !m.isSelectable {
		return str
	}

	if m.selectedRowIdx == currentRowIdx {
		return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, highlightCell, highlightStyle)
	}
	return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, highlightStyle, defaultText)
}

func (m *BoardView) selectionUp() {
	m.selectedRowIdx = max(0, m.selectedRowIdx-1)
}

func (m *BoardView) selectionDown() {
	m.selectedRowIdx = min(m.boardSize-1, m.selectedRowIdx+1)
}

func (m *BoardView) selectionLeft() {
	m.selectedColIdx = max(0, m.selectedColIdx-2)
}

func (m *BoardView) selectionRight() {
	m.selectedColIdx = min(len(m.alphabet)-1, m.selectedColIdx+2)
}

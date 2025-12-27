package views

import (
	"strconv"
	"strings"
	"ws-battleship-client/internal/domain/models"
	"ws-battleship-client/pkg/math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#4a4a8a")).
			Foreground(lipgloss.Color("#ffffff"))

	highlightAllowedCell = lipgloss.NewStyle().
				Background(lipgloss.Color("#37DB76")).
				Foreground(lipgloss.Color("#ffffff")).Bold(true)

	highlightForbiddenCell = lipgloss.NewStyle().
				Background(lipgloss.Color("#B83921")).
				Foreground(lipgloss.Color("#ffffff")).Bold(true)
)

type BoardView struct {
	player         *models.Player
	selectedRowIdx int
	selectedColIdx int
	cellX, cellY   int
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
	m.SelectCell(0, 0)
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
	numbers.Grow(m.boardSize)
	numbers.WriteRune('\n')

	for i := range b1Lines {
		board.WriteString(m.renderBoardRow(b1Lines[i], i))
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

func (m *BoardView) SelectCell(rowIdx, colIdx int) {
	m.cellY = math.Clamp(rowIdx, 0, m.boardSize-1)
	m.cellX = math.Clamp(colIdx, 0, m.boardSize-1)

	m.selectedRowIdx = math.Clamp(rowIdx, 0, m.boardSize-1)

	// String: a   b   c   d   e    f     g     h     j     k
	// Index:  0   1   2   3   4    5     6     7     8     9
	// Column: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18
	m.selectedColIdx = math.Clamp(colIdx*2, 0, len(m.alphabet)-1)
}

func (m *BoardView) renderBoardRow(str string, currentRowIdx int) string {
	if !m.isSelectable {
		return str
	}

	if m.selectedRowIdx == currentRowIdx {
		return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, m.getSelectedCellHighlightStyle(), highlightStyle)
	}
	return lipgloss.StyleRunes(str, []int{m.selectedColIdx}, highlightStyle, defaultText)
}

func (m *BoardView) getSelectedCellHighlightStyle() lipgloss.Style {
	if m.player.Board.IsCellEmpty(byte(m.cellY), byte(m.cellX)) {
		return highlightAllowedCell
	} else {
		return highlightForbiddenCell
	}
}

func (m *BoardView) selectionUp() {
	m.SelectCell(m.cellY-1, m.cellX)
}

func (m *BoardView) selectionDown() {
	m.SelectCell(m.cellY+1, m.cellX)
}

func (m *BoardView) selectionLeft() {
	m.SelectCell(m.cellY, m.cellX-1)
}

func (m *BoardView) selectionRight() {
	m.SelectCell(m.cellY, m.cellX+1)
}

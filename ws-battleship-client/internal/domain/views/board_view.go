package views

import (
	"strconv"
	"strings"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/pkg/math"

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
	nickname string

	selectedRowIdx int
	selectedColIdx int
	cellX, cellY   int
	board          domain.Board
	alphabet       string
	isSelectable   bool
}

func NewBoardView() *BoardView {
	var emptyBoard domain.Board

	alphabet := make([]rune, 0, emptyBoard.Size()*2)
	for i, r := range emptyBoard.Alphabet() {
		alphabet = append(alphabet, r)

		if i < len(emptyBoard.Alphabet())-1 {
			alphabet = append(alphabet, ' ')
		}
	}
	return &BoardView{
		nickname: "Unknown",
		board:    emptyBoard,
		alphabet: string(alphabet),
	}
}

func (v *BoardView) Init() tea.Cmd {
	v.SelectCell(0, 0)
	return nil
}

func (v *BoardView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if v.isSelectable {
				v.selectionUp()
			}
		case tea.KeyDown:
			if v.isSelectable {
				v.selectionDown()
			}
		case tea.KeyLeft:
			if v.isSelectable {
				v.selectionLeft()
			}
		case tea.KeyRight:
			if v.isSelectable {
				v.selectionRight()
			}
		}
	}

	return v, nil
}

func (v *BoardView) View() string {
	boardLines := v.board.Lines()

	var board strings.Builder
	var numbers strings.Builder
	numbers.Grow(v.board.Size())
	numbers.WriteRune('\n')

	for i := range boardLines {
		board.WriteString(v.renderBoardRow(boardLines[i], i))
		numbers.WriteString(strconv.FormatInt(int64(i+1), 10))

		if i < len(boardLines)-1 {
			board.WriteRune('\n')
			numbers.WriteRune('\n')
		}
	}

	boardView := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(board.String())
	boardView = lipgloss.JoinVertical(lipgloss.Center, v.alphabet, boardView)
	boardView = lipgloss.JoinHorizontal(lipgloss.Center, numbers.String(), boardView)
	boardView = lipgloss.JoinVertical(lipgloss.Center, boardView, v.nickname)

	return boardView
}

func (v *BoardView) SetPlayer(player *domain.PlayerModel) {
	v.board = player.Board
	v.nickname = player.Nickname
}

func (v *BoardView) SetSelectable(isSelectable bool) {
	v.isSelectable = isSelectable
}

func (v *BoardView) SelectCell(cellX, cellY int) {
	v.cellX = math.Clamp(cellX, 0, v.board.Size()-1)
	v.cellY = math.Clamp(cellY, 0, v.board.Size()-1)

	v.selectedRowIdx = v.cellY

	// String: a   b   c   d   e    f     g     h     j     k
	// Index:  0   1   2   3   4    5     6     7     8     9
	// Column: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18
	v.selectedColIdx = math.Clamp(cellX*2, 0, len(v.alphabet)-1)
}

func (v *BoardView) IsAllowedToFire() bool {
	return v.board.IsCellEmpty(byte(v.cellX), byte(v.cellY)) || v.board.GetCellType(byte(v.cellX), byte(v.cellY)) == domain.Alive
}

func (v *BoardView) renderBoardRow(str string, currentRowIdx int) string {
	if !v.isSelectable {
		return str
	}

	if v.selectedRowIdx == currentRowIdx {
		return lipgloss.StyleRunes(str, []int{v.selectedColIdx}, v.getCellHighlighStyle(), highlightStyle)
	}
	return lipgloss.StyleRunes(str, []int{v.selectedColIdx}, highlightStyle, defaultText)
}

func (v *BoardView) getCellHighlighStyle() lipgloss.Style {
	if v.IsAllowedToFire() {
		return highlightAllowedCell
	} else {
		return highlightForbiddenCell
	}
}

func (v *BoardView) selectionUp() {
	v.SelectCell(v.cellX, v.cellY-1)
}

func (v *BoardView) selectionDown() {
	v.SelectCell(v.cellX, v.cellY+1)
}

func (v *BoardView) selectionLeft() {
	v.SelectCell(v.cellX-1, v.cellY)
}

func (v *BoardView) selectionRight() {
	v.SelectCell(v.cellX+1, v.cellY)
}

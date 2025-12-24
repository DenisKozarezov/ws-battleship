package view

import (
	"strings"
	"ws-battleship-client/internal/domain/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GameView struct {
	game *model.GameModel
}

func NewGameView(game *model.GameModel) *GameView {
	return &GameView{game: game}
}

func (m *GameView) Init() tea.Cmd {
	return nil
}

func (m *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *GameView) View() string {
	boardStyle := lipgloss.NewStyle().
		Padding(1).Align(lipgloss.Center).
		Border(lipgloss.DoubleBorder())

	logsStyle := lipgloss.NewStyle().
		Width(70).Height(10).
		Border(lipgloss.DoubleBorder())

	left := boardStyle.Render(m.renderPlayersBoards())
	right := logsStyle.Render(m.renderGameLogs())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m *GameView) renderPlayersBoards() string {
	b1Lines := m.game.Player1.Board.Lines()
	b2Lines := m.game.Player2.Board.Lines()

	const margin = "        "

	var builder strings.Builder
	for i := range b1Lines {
		builder.WriteString(b1Lines[i] + margin + b2Lines[i])

		if i < len(b1Lines)-1 {
			builder.WriteRune('\n')
		}
	}
	return builder.String()
}

func (m *GameView) renderGameLogs() string {
	colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("21")).SetString("System").String()

	var builder strings.Builder
	builder.WriteString(colorStyle + ": " + "Player 1 joined the match. \n")
	builder.WriteString(colorStyle + ": " + "Player 1 is striking D5! And missed...\n")
	builder.WriteString(colorStyle + ": " + "KABOOM! Player 1 destroyed the enemy ship! +50 score!")
	return builder.String()
}

package views

import (
	"ws-battleship-client/internal/config"
	"ws-battleship-shared/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	defaultText = lipgloss.NewStyle().Background(lipgloss.NoColor{})

	boardStyle = lipgloss.NewStyle().Padding(1).Align(lipgloss.Center).Border(lipgloss.NormalBorder())

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	FixedUpdate()
	View() string
}

type GameView struct {
	cfg            *config.GameConfig
	chatView       *ChatView
	leftBoard      *BoardView
	rightBoard     *BoardView
	turnTimerView  *TimerView
	gameTickerView *TickerView

	currentBoard *BoardView
}

func NewGameView(cfg *config.GameConfig) *GameView {
	return &GameView{
		cfg:            cfg,
		leftBoard:      NewBoardView(),
		rightBoard:     NewBoardView(),
		chatView:       NewChatView(),
		turnTimerView:  NewTimerView(),
		gameTickerView: NewTickerView(),
	}
}

func (v *GameView) Init() tea.Cmd {
	v.turnTimerView.SetExpireCallback(func() {
		v.currentBoard.SetSelectable(false)
	})

	return tea.Batch(v.leftBoard.Init(),
		v.rightBoard.Init(),
		v.chatView.Init(),
		v.turnTimerView.Init(),
		tea.SetWindowTitle("Battleship"))
}

func (v *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return v, tea.Quit
		}
	}

	var cmds []tea.Cmd
	_, cmd := v.leftBoard.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = v.rightBoard.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = v.chatView.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = v.turnTimerView.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = v.gameTickerView.Update(msg)
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

func (v *GameView) FixedUpdate() {
	v.gameTickerView.FixedUpdate()
	v.turnTimerView.FixedUpdate()
}

func (v *GameView) View() string {
	gameTime := "GAME TIME: " + v.gameTickerView.View()

	boards := boardStyle.Render(lipgloss.JoinVertical(lipgloss.Center, gameTime, v.renderPlayersBoards()))
	gameView := lipgloss.JoinHorizontal(lipgloss.Top, boards, " ", v.chatView.View())
	return gameView
}

func (v *GameView) StartGame(gameModel domain.GameModel) {
	players := make([]*domain.PlayerModel, 0, len(gameModel.Players))
	for _, player := range gameModel.Players {
		players = append(players, player)
	}

	v.leftBoard.SetPlayer(players[0])
	v.rightBoard.SetPlayer(players[1])

	v.gameTickerView.Start()
	v.GiveTurnToPlayer(v.leftBoard)
}

func (v *GameView) GiveTurnToPlayer(board *BoardView) {
	if v.currentBoard != nil {
		v.currentBoard.SetSelectable(false)
	}
	v.currentBoard = board
	v.currentBoard.SetSelectable(true)

	v.turnTimerView.Reset(int(v.cfg.TurnTime.Seconds()))
	v.turnTimerView.Start()
}

func (v *GameView) AppendMessageInChat(msg ChatMessage) {
	v.chatView.AppendMessage(msg)
}

func (v *GameView) renderPlayersBoards() string {
	return lipgloss.JoinHorizontal(lipgloss.Center, v.leftBoard.View(), v.renderGameTurn(), v.rightBoard.View())
}

func (v *GameView) renderGameTurn() string {
	var turn string
	if v.isLocalPlayerTurn() {
		turn = highlightAllowedCell.Render(" YOUR TURN ")
	} else {
		turn = highlightForbiddenCell.Render(" ENEMY TURN ")
	}
	turn = lipgloss.JoinVertical(lipgloss.Center, turn, v.turnTimerView.View())

	if v.isLocalPlayerTurn() {
		help := helpStyle.Align(lipgloss.Center).Render("Press ↑ ↓ → ← to Navigate\nPress Enter to Fire")
		return lipgloss.PlaceHorizontal(30, lipgloss.Center, turn+"\n\n"+help)
	} else {
		return lipgloss.PlaceHorizontal(30, lipgloss.Center, turn)
	}
}

func (v *GameView) isLocalPlayerTurn() bool {
	return v.currentBoard == v.leftBoard
}

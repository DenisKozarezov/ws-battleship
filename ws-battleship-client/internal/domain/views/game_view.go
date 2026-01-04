package views

import (
	"fmt"
	"time"
	clientEvents "ws-battleship-client/internal/domain/events"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"

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
	isLocalPlayerTurn bool

	boards       map[string]*BoardView
	leftBoard    *BoardView
	rightBoard   *BoardView
	turningBoard *BoardView

	turnTimerView  *TimerView
	gameTickerView *TickerView
	chatView       *ChatView

	playerFiredHandler func(cellX, cellY byte)
}

func NewGameView(eventBus *events.EventBus, metadata domain.ClientMetadata) *GameView {
	chatView := NewChatView()
	chatView.SetMessageTypedHandler(func(msg string) {
		// ONLY FOR CLIENT USAGE! To avoid circual appending in chat we have to use other event type instead of
		// server's [SendMessageType].
		// That's why we need something only for internal usage, that won't be sended to server.
		// Consider this as internal events for local machine.
		event, _ := clientEvents.NewPlayerTypedMessageEvent(metadata.Nickname, msg)
		_ = eventBus.Invoke(event)
	})

	return &GameView{
		boards:         make(map[string]*BoardView),
		leftBoard:      NewBoardView(),
		rightBoard:     NewBoardView(),
		turnTimerView:  NewTimerView(),
		gameTickerView: NewTickerView(),
		chatView:       chatView,
	}
}

func (v *GameView) Init() tea.Cmd {
	v.turnTimerView.SetExpireCallback(func() {
		v.turningBoard.SetSelectable(false)
	})

	return tea.Batch(v.leftBoard.Init(),
		v.rightBoard.Init(),
		v.turnTimerView.Init(),
		v.gameTickerView.Init(),
		v.chatView.Init(),
		tea.SetWindowTitle("Battleship"))
}

func (v *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return v, tea.Quit
		case tea.KeyEnter:
			if v.isLocalPlayerTurn {
				v.onPlayerFiredHandler()
			}
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

func (v *GameView) StartGame() {
	v.gameTickerView.Start()
}

func (v *GameView) SetGameModel(gameModel *domain.GameModel) {
	clear(v.boards)

	if gameModel.LeftPlayer != nil {
		v.leftBoard.SetPlayer(gameModel.LeftPlayer)
		v.boards[gameModel.LeftPlayer.ID] = v.leftBoard
	}

	if gameModel.RightPlayer != nil {
		v.rightBoard.SetPlayer(gameModel.RightPlayer)
		v.boards[gameModel.RightPlayer.ID] = v.rightBoard
	}
}

func (v *GameView) GiveTurnToPlayer(turningPlayer *domain.PlayerModel, remainingTime time.Duration, isLocalPlayer bool) error {
	v.isLocalPlayerTurn = isLocalPlayer

	if v.turningBoard != nil {
		v.turningBoard.SetSelectable(false)
	}

	var found bool
	if v.turningBoard, found = v.boards[turningPlayer.ID]; !found {
		return fmt.Errorf("player not found")
	}

	if isLocalPlayer {
		v.turningBoard.SetSelectable(true)
	}

	v.turnTimerView.Reset(int(remainingTime.Seconds()))
	v.turnTimerView.Start()
	return nil
}

func (v *GameView) AppendMessageInChat(msg ChatMessage) error {
	v.chatView.AppendMessage(msg)
	return nil
}

func (v *GameView) SetPlayerFiredHandler(fn func(cellY, cellX byte)) {
	v.playerFiredHandler = fn
}

func (v *GameView) renderPlayersBoards() string {
	return lipgloss.JoinHorizontal(lipgloss.Center, v.leftBoard.View(), v.renderGameTurn(), v.rightBoard.View())
}

func (v *GameView) renderGameTurn() string {
	var turn string
	if v.isLocalPlayerTurn {
		turn = highlightAllowedCell.Render(" YOUR TURN ")
	} else {
		turn = highlightForbiddenCell.Render(" ENEMY TURN ")
	}
	turn = lipgloss.JoinVertical(lipgloss.Center, turn, v.turnTimerView.View())

	if v.isLocalPlayerTurn {
		help := helpStyle.Align(lipgloss.Center).Render("Press ↑ ↓ → ← to Navigate\nPress Enter to Fire")
		return lipgloss.PlaceHorizontal(30, lipgloss.Center, turn+"\n\n"+help)
	} else {
		return lipgloss.PlaceHorizontal(30, lipgloss.Center, turn)
	}
}

func (v *GameView) onPlayerFiredHandler() {
	v.isLocalPlayerTurn = false

	if v.turningBoard == nil {
		return
	}
	v.turningBoard.SetSelectable(false)

	if v.playerFiredHandler != nil {
		v.playerFiredHandler(byte(v.turningBoard.cellX), byte(v.turningBoard.cellY))
	}
}

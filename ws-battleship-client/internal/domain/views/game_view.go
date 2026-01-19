package views

import (
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

type GameView struct {
	isLocalPlayerTurn bool
	localPlayerID     string

	boards     map[string]*BoardView
	yourBoard  *BoardView
	enemyBoard *BoardView

	turnTimerView  *TimerView
	gameTickerView *TickerView
	chatView       *ChatView

	playerFiredHandler func(targetPlayerID string, cellX, cellY byte)
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
		localPlayerID:  metadata.ClientID,
		boards:         make(map[string]*BoardView),
		yourBoard:      NewBoardView(),
		enemyBoard:     NewBoardView(),
		turnTimerView:  NewTimerView(),
		gameTickerView: NewTickerView(),
		chatView:       chatView,
	}
}

func (v *GameView) Init() tea.Cmd {
	v.turnTimerView.SetExpireCallback(func() {
		v.enemyBoard.SetSelectable(false)
	})

	return tea.Batch(v.yourBoard.Init(),
		v.enemyBoard.Init(),
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
	_, cmd := v.enemyBoard.Update(msg)
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

func (v *GameView) EndGame() {
	v.gameTickerView.Stop()
	v.turnTimerView.Stop()

	if v.yourBoard != nil {
		v.yourBoard.SetSelectable(false)
	}

	if v.enemyBoard != nil {
		v.enemyBoard.SetSelectable(false)
	}
}

func (v *GameView) SetGameModel(gameModel *domain.GameModel) {
	clear(v.boards)

	for playerID, player := range gameModel.Players {
		if playerID == v.localPlayerID {
			v.yourBoard.SetPlayer(gameModel.Players[v.localPlayerID])
		} else {
			v.enemyBoard.SetPlayer(player)
		}
		v.boards[playerID] = v.enemyBoard
	}
}

func (v *GameView) GiveTurnToPlayer(event events.PlayerTurnEvent, isLocalPlayer bool) error {
	v.isLocalPlayerTurn = isLocalPlayer
	v.enemyBoard.SetSelectable(isLocalPlayer)
	v.turnTimerView.Reset(int(event.RemainingTime.Seconds()))
	v.turnTimerView.Start()
	return nil
}

func (v *GameView) AppendMessageInChat(msg ChatMessage) error {
	v.chatView.AppendMessage(msg)
	return nil
}

func (v *GameView) SetPlayerFiredHandler(fn func(targetPlayerID string, cellY, cellX byte)) {
	v.playerFiredHandler = fn
}

func (v *GameView) renderPlayersBoards() string {
	return lipgloss.JoinHorizontal(lipgloss.Center, v.yourBoard.View(), v.renderGameTurn(), v.enemyBoard.View())
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
	if v.enemyBoard == nil || !v.enemyBoard.IsAllowedToFire() {
		return
	}

	v.isLocalPlayerTurn = false
	v.enemyBoard.SetSelectable(false)

	if v.playerFiredHandler != nil {
		v.playerFiredHandler(v.enemyBoard.playerID, byte(v.enemyBoard.cellX), byte(v.enemyBoard.cellY))
	}
}

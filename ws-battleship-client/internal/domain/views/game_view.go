package views

import (
	"encoding/json"
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
	localNickname string

	boards       map[string]*BoardView
	leftBoard    *BoardView
	rightBoard   *BoardView
	turningBoard *BoardView

	turnTimerView  *TimerView
	gameTickerView *TickerView
	chatView       *ChatView
}

func NewGameView(eventBus *events.EventBus, metadata domain.ClientMetadata) *GameView {
	chatView := NewChatView()

	v := &GameView{
		localNickname:  metadata.Nickname,
		boards:         make(map[string]*BoardView),
		leftBoard:      NewBoardView(),
		rightBoard:     NewBoardView(),
		turnTimerView:  NewTimerView(),
		gameTickerView: NewTickerView(),
		chatView:       chatView,
	}

	eventBus.Subscribe(events.SendMessageType, v.onMessageReceivedHandler)
	eventBus.Subscribe(events.GameStartEventType, v.onGameStartedHandler)
	eventBus.Subscribe(events.PlayerUpdateStateEventType, v.onPlayerUpdateState)
	eventBus.Subscribe(events.PlayerTurnEventType, v.onPlayerTurnHandler)
	chatView.SetMessageTypedHandler(func(msg string) {
		// ONLY FOR CLIENT USAGE! To avoid circual broadcasts we have to use other event type instead of
		// server's [SendMessageType].
		// That's why we need something only for internal usage, that won't be sended to server.
		// Consider this as internal events.
		event, _ := clientEvents.NewPlayerTypedMessageEvent(metadata.Nickname, msg)
		eventBus.Invoke(event)
	})

	return v
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

func (v *GameView) StartGame(gameModel *domain.GameModel) {
	v.gameTickerView.Start()
}

func (v *GameView) GiveTurnToPlayer(turningPlayer *domain.PlayerModel, remainingTime time.Duration) {
	for _, board := range v.boards {
		board.SetSelectable(false)
	}

	v.turningBoard = v.boards[turningPlayer.ID]
	if v.isLocalPlayerTurn() {
		v.turningBoard.SetSelectable(true)
	}

	v.turnTimerView.Reset(int(remainingTime.Seconds()))
	v.turnTimerView.Start()
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
	if v.turningBoard == nil {
		return false
	}

	return v.turningBoard.nickname == v.localNickname
}

func (v *GameView) onGameStartedHandler(e events.Event) error {
	var gameStartEvent events.GameStartEvent
	if err := json.Unmarshal(e.Data, &gameStartEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	v.StartGame(gameStartEvent.GameModel)
	return nil
}

func (v *GameView) onPlayerUpdateState(e events.Event) error {
	var playerUpdateEvent events.PlayerUpdateStateEvent
	if err := json.Unmarshal(e.Data, &playerUpdateEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	if playerUpdateEvent.GameModel.LeftPlayer != nil {
		v.leftBoard.SetPlayer(playerUpdateEvent.GameModel.LeftPlayer)
		v.boards[playerUpdateEvent.GameModel.LeftPlayer.ID] = v.leftBoard
	}

	if playerUpdateEvent.GameModel.RightPlayer != nil {
		v.rightBoard.SetPlayer(playerUpdateEvent.GameModel.RightPlayer)
		v.boards[playerUpdateEvent.GameModel.RightPlayer.ID] = v.rightBoard
	}
	return nil
}

func (v *GameView) onPlayerTurnHandler(e events.Event) error {
	var playerTurnEvent events.PlayerTurnEvent
	if err := json.Unmarshal(e.Data, &playerTurnEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	v.GiveTurnToPlayer(playerTurnEvent.Player, playerTurnEvent.RemainingTime)
	return nil
}

func (v *GameView) onMessageReceivedHandler(e events.Event) error {
	var sendMessageEvent events.SendMessageEvent
	if err := json.Unmarshal(e.Data, &sendMessageEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	timestamp, err := time.Parse(events.TimestampFormat, e.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	v.chatView.AppendMessage(ChatMessage{
		Sender:         sendMessageEvent.Sender,
		Message:        sendMessageEvent.Message,
		IsNotification: sendMessageEvent.IsNotification,
		Timestamp:      timestamp.Format(time.TimeOnly),
	})
	return nil
}

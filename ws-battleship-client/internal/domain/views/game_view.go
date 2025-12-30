package views

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"ws-battleship-client/internal/config"
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
	cfg            *config.GameConfig
	leftBoard      *BoardView
	rightBoard     *BoardView
	turnTimerView  *TimerView
	gameTickerView *TickerView
	chatView       *ChatView

	eventBus     *events.EventBus
	currentBoard *BoardView
}

func NewGameView(cfg *config.GameConfig, eventBus *events.EventBus, metadata domain.ClientMetadata) *GameView {
	chatView := NewChatView()

	v := &GameView{
		cfg:            cfg,
		eventBus:       eventBus,
		leftBoard:      NewBoardView(),
		rightBoard:     NewBoardView(),
		turnTimerView:  NewTimerView(),
		gameTickerView: NewTickerView(),
		chatView:       chatView,
	}

	eventBus.Subscribe(events.SendMessageType, v.onMessageReceivedHandler)
	eventBus.Subscribe(events.GameStartEventType, v.onGameStartedHandler)
	chatView.SetMessageTypedHandler(func(msg string) {
		// ONLY FOR CLIENT USAGE! To avoid circual broadcasts we have to use other event type instead of
		// server's [SendMessageType].
		// That's why we need something only for internal usage, that won't be sended to server.
		// Consider this as internal events.
		event, _ := clientEvents.NewPlayerTypedMessageEvent(metadata.Nickname, msg)
		eventBus.Invoke(context.Background(), event)
	})

	return v
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

func (v *GameView) onGameStartedHandler(_ context.Context, e events.Event) error {
	var gameStartEvent events.GameStartEvent
	if err := json.Unmarshal(e.Data, &gameStartEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	v.StartGame(gameStartEvent.GameModel)
	return nil
}

func (v *GameView) onMessageReceivedHandler(ctx context.Context, e events.Event) error {
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

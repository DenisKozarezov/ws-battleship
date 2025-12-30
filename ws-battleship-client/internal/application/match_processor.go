package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"
)

type MatchProcessor struct {
	logger   logger.Logger
	gameView *views.GameView
}

func NewMatchProcessor(logger logger.Logger, gameView *views.GameView) *MatchProcessor {
	return &MatchProcessor{
		logger:   logger,
		gameView: gameView,
	}
}

func (p *MatchProcessor) OnGameStartHandler(ctx context.Context, e events.Event) error {
	var gameStartEvent events.GameStartEvent
	if err := json.Unmarshal(e.Data, &gameStartEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	p.gameView.StartGame(gameStartEvent.GameModel)
	return nil
}

func (p *MatchProcessor) OnPlayerJoinedHandler(ctx context.Context, e events.Event) error {
	var playerJoinedEvent events.PlayerJoinedEvent
	if err := json.Unmarshal(e.Data, &playerJoinedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	//p.logger.Infof("player %s joined the match", playerJoinedEvent.Player.Nickname)
	return nil
}

func (p *MatchProcessor) OnPlayerLeavedHandler(ctx context.Context, e events.Event) error {
	var playerLeavedEvent events.PlayerLeavedEvent
	if err := json.Unmarshal(e.Data, &playerLeavedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	//p.logger.Infof("player %s leaved the match", playerLeavedEvent.Player.Nickname)
	return nil
}

func (p *MatchProcessor) OnSendMessageHandler(ctx context.Context, e events.Event) error {
	var sendMessageEvent events.SendMessageEvent
	if err := json.Unmarshal(e.Data, &sendMessageEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	timestamp, err := time.Parse(events.TimestampFormat, e.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	p.gameView.AppendMessageInChat(views.ChatMessage{
		Sender:         sendMessageEvent.Sender,
		Message:        sendMessageEvent.Message,
		IsNotification: sendMessageEvent.IsNotification,
		Timestamp:      timestamp.Format(time.TimeOnly),
	})
	return nil
}

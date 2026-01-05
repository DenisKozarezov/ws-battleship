package application

import (
	"fmt"
	"time"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/events"
)

func (a *App) onGameStartedHandler(e events.Event) error {
	a.gameView.StartGame()
	return nil
}

func (a *App) onGameEndHandler(e events.Event) error {
	if _, err := events.CastTo[events.GameEndEvent](e); err != nil {
		return err
	}
	a.gameView.EndGame()
	return nil
}

func (a *App) onPlayerUpdateState(e events.Event) error {
	playerUpdateEvent, err := events.CastTo[events.PlayerUpdateStateEvent](e)
	if err != nil {
		return err
	}

	a.gameView.SetGameModel(playerUpdateEvent.GameModel)
	return nil
}

func (a *App) onPlayerTurnHandler(e events.Event) error {
	playerTurnEvent, err := events.CastTo[events.PlayerTurnEvent](e)
	if err != nil {
		return err
	}

	isLocalPlayer := playerTurnEvent.TurningPlayer != nil &&
		a.metadata.ClientID == playerTurnEvent.TurningPlayer.ID

	return a.gameView.GiveTurnToPlayer(playerTurnEvent, isLocalPlayer)
}

func (a *App) onPlayerSendMessageHandler(e events.Event) error {
	sendMessageEvent, err := events.CastTo[events.SendMessageEvent](e)
	if err != nil {
		return err
	}

	timestamp, err := time.Parse(events.TimestampFormat, e.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return a.gameView.AppendMessageInChat(views.ChatMessage{
		Sender:    sendMessageEvent.Sender,
		Message:   sendMessageEvent.Message,
		Type:      sendMessageEvent.Type,
		Timestamp: timestamp,
	})
}

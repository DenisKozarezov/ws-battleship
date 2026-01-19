package states

import (
	"fmt"
	"time"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/events"
)

func (s *GameState) onGameStartedHandler(e events.Event) error {
	s.gameView.StartGame()
	return nil
}

func (s *GameState) onGameEndHandler(e events.Event) error {
	if _, err := events.CastTo[events.GameEndEvent](e); err != nil {
		return err
	}
	s.gameView.EndGame()
	return nil
}

func (s *GameState) onPlayerUpdateState(e events.Event) error {
	playerUpdateEvent, err := events.CastTo[events.PlayerUpdateStateEvent](e)
	if err != nil {
		return err
	}

	s.gameView.SetGameModel(playerUpdateEvent.GameModel)
	return nil
}

func (a *GameState) onPlayerTurnHandler(e events.Event) error {
	playerTurnEvent, err := events.CastTo[events.PlayerTurnEvent](e)
	if err != nil {
		return err
	}

	isLocalPlayer := a.metadata.ClientID == playerTurnEvent.TurningPlayerID
	return a.gameView.GiveTurnToPlayer(playerTurnEvent, isLocalPlayer)
}

func (s *GameState) onPlayerSendMessageHandler(e events.Event) error {
	sendMessageEvent, err := events.CastTo[events.SendMessageEvent](e)
	if err != nil {
		return err
	}

	timestamp, err := time.Parse(events.TimestampFormat, e.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return s.gameView.AppendMessageInChat(views.ChatMessage{
		Sender:    sendMessageEvent.Sender,
		Message:   sendMessageEvent.Message,
		Type:      sendMessageEvent.Type,
		Timestamp: timestamp,
	})
}

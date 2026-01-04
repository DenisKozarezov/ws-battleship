package application

import "ws-battleship-shared/events"

func (a *App) onPlayerTypedMessage(e events.Event) error {
	e.Type = events.SendMessageType
	return a.client.SendMessage(e)
}

func (a *App) onPlayerPressedFireHandler(cellX, cellY byte) {
	event, _ := events.NewPlayerFireEvent(a.metadata, cellX, cellY)
	if err := a.client.SendMessage(event); err != nil {
		a.logger.Errorf("failed to send a message: %s", err)
	}
}

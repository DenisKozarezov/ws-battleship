package application

import "ws-battleship-shared/events"

func (a *App) onPlayerTypedMessage(e events.Event) error {
	e.Type = events.SendMessageType
	return a.client.SendMessage(e)
}

func (a *App) onPlayerPressedFireHandler(targetPlayerID string, cellX, cellY byte) {
	args := events.FireCommandArgs{
		FiringPlayerID: a.metadata.ClientID,
		TargetPlayerID: targetPlayerID,
		CellX:          cellX,
		CellY:          cellY,
	}

	event, _ := events.NewPlayerFireEvent(args)

	if err := a.client.SendMessage(event); err != nil {
		a.logger.Errorf("failed to send a message: %s", err)
	}
}

package states

import "ws-battleship-shared/events"

func (s *GameState) onPlayerTypedMessage(e events.Event) error {
	e.Type = events.SendMessageType
	return s.client.SendMessage(e)
}

func (s *GameState) onPlayerPressedFireHandler(targetPlayerID string, cellX, cellY byte) {
	args := events.FireCommandArgs{
		FiringPlayerID: s.metadata.ClientID,
		TargetPlayerID: targetPlayerID,
		CellX:          cellX,
		CellY:          cellY,
	}

	event, _ := events.NewPlayerFireEvent(args)

	if err := s.client.SendMessage(event); err != nil {
		s.logger.Errorf("failed to send a message: %s", err)
	}
}

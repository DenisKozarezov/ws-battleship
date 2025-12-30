package events

import (
	"ws-battleship-shared/events"
	serverEvents "ws-battleship-shared/events"
)

const (
	// Local client events. ONLY FOR INTERNAL USAGE! We don't need to send them to server.
	PlayerTypedMessageType serverEvents.EventType = "player_typed_message"
)

func NewPlayerTypedMessageEvent(sender string, message string) (events.Event, error) {
	event, err := events.NewSendMessageEvent(sender, message)
	if err != nil {
		return event, err
	}
	event.Type = PlayerTypedMessageType
	return event, nil
}

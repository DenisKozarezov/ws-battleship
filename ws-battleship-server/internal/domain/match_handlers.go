package domain

import (
	"fmt"
	"ws-battleship-server/internal/delivery/websocket"
	"ws-battleship-shared/events"
)

func (m *Match) onPlayerFiredHandler(e events.Event) error {
	playerFiredEvent, err := events.CastTo[events.PlayerFireEvent](e)
	if err != nil {
		return err
	}

	args := events.FireCommandArgs{
		FiringPlayerID: playerFiredEvent.FiringPlayerID,
		TargetPlayerID: playerFiredEvent.TargetPlayerID,
		CellX:          playerFiredEvent.CellX,
		CellY:          playerFiredEvent.CellY,
	}

	m.Dispatch(NewFireCommand(args))
	return nil
}

func (m *Match) onPlayerJoinedHandler(joinedClient websocket.Client) {
	player := m.players[joinedClient.ID()]

	if err := m.allPlayersUpdate(); err != nil {
		m.logger.Errorf("failed to update players: %s", err)
	}

	event, err := events.NewPlayerJoinedEvent(player.Model)
	if err != nil {
		m.logger.Error(err)
		return
	}

	if err = m.room.Broadcast(event); err != nil {
		m.logger.Error(err)
		return
	}

	m.room.logger.Infof("player %s joined the match id=%s [players: %d]", player, m.ID(), m.room.Capacity())
	if err := m.SendNotification(fmt.Sprintf("Player '%s' joined the game.", player.Nickname()), events.RoomNotificationType); err != nil {
		m.logger.Error(err)
	}

	if m.IsReadyToStart() {
		m.Dispatch(NewGameStartCommand(m.logger))
	}
}

func (m *Match) onPlayerLeftHandler(leftClient websocket.Client) {
	defer delete(m.players, leftClient.ID())

	player := m.players[leftClient.ID()]

	event, err := events.NewPlayerLeftEvent(player.Model)
	if err != nil {
		m.logger.Error(err)
		return
	}

	if err = m.room.Broadcast(event); err != nil {
		m.logger.Error(err)
		return
	}

	m.room.logger.Infof("player %s left the match id=%s [players: %d]", player, m.ID(), m.room.Capacity())
	if err := m.SendNotification(fmt.Sprintf("Player '%s' left the game.", player.Nickname()), events.RoomNotificationType); err != nil {
		m.logger.Error(err)
	}
}

func (m *Match) onPlayerSentMessageHandler(e events.Event) error {
	return m.room.Broadcast(e)
}

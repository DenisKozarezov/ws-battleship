package domain

import "ws-battleship-shared/events"

type Command interface {
	Execute(CommandExecutor) error
}

type CommandExecutor interface {
	ID() string
	Fire(args events.FireCommandArgs) error
	GiveTurnToNextPlayer() error
	JoinNewPlayer(joinedPlayer *Player) error
	StartMatch() error
	EndMatch(winningPlayer *Player) error
}

package domain

type GameTurnCommand struct{}

func NewGameTurnCommand() *GameTurnCommand {
	return &GameTurnCommand{}
}

func (c *GameTurnCommand) Execute(executor CommandExecutor) error {
	return executor.GiveTurnToNextPlayer()
}

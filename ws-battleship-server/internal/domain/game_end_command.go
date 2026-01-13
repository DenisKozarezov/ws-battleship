package domain

import "ws-battleship-shared/pkg/logger"

type GameEndCommand struct {
	logger        logger.Logger
	winningPlayer *Player
}

func NewGameEndCommand(logger logger.Logger, winningPlayer *Player) *GameEndCommand {
	return &GameEndCommand{logger: logger, winningPlayer: winningPlayer}
}

func (c *GameEndCommand) Execute(executor CommandExecutor) error {
	c.logger.Infof("match id=%s is ended; player id=%s has won!", executor.ID(), c.winningPlayer.ID)
	return executor.EndMatch(c.winningPlayer)
}

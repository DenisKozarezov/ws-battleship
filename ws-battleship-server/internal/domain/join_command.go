package domain

import "ws-battleship-shared/pkg/logger"

type JoinCommand struct {
	logger logger.Logger
	player *Player
}

func NewJoinCommand(logger logger.Logger, player *Player) *JoinCommand {
	return &JoinCommand{logger: logger, player: player}
}

func (c *JoinCommand) Execute(executor CommandExecutor) error {
	c.logger.Infof("player '%s' is joining...", c.player.String())
	return executor.JoinNewPlayer(c.player)
}

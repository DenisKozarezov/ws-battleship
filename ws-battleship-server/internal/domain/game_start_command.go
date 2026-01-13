package domain

import "ws-battleship-shared/pkg/logger"

type GameStartCommand struct {
	logger logger.Logger
}

func NewGameStartCommand(logger logger.Logger) *GameStartCommand {
	return &GameStartCommand{logger: logger}
}

func (c *GameStartCommand) Execute(executor CommandExecutor) error {
	c.logger.Infof("match is starting in room id=%s", executor.ID())
	return executor.StartMatch()
}

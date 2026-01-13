package domain

import "ws-battleship-shared/events"

type FireCommand struct {
	events.FireCommandArgs
}

func NewFireCommand(args events.FireCommandArgs) *FireCommand {
	return &FireCommand{FireCommandArgs: args}
}

func (c *FireCommand) Execute(executor CommandExecutor) error {
	return executor.Fire(c.FireCommandArgs)
}

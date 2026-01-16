package domain

type CloseMatchCommand struct{}

func NewCloseMatchCommand() *CloseMatchCommand {
	return &CloseMatchCommand{}
}

func (c *CloseMatchCommand) Execute(executor CommandExecutor) error {
	return executor.Close()
}

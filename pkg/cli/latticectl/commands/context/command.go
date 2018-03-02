package context

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	Subcommands []command.Command2
}

func (c *Command) BaseCommand() (*command.BaseCommand2, error) {
	cmd := &latticectl.BaseCommand{
		Name:        "context",
		Subcommands: c.Subcommands,
	}

	return cmd.BaseCommand()
}

package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.BaseCommand{
		Name:        "context",
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

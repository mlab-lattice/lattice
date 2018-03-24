package local

import (
	"github.com/mlab-lattice/system/pkg/latticectl"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.BaseCommand{
		Name:        "local",
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

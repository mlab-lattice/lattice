package bootstrap

import (
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.BaseCommand{
		Name:        "bootstrap",
		Subcommands: c.Subcommands,
	}

	return cmd, nil
}

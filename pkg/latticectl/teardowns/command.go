package teardowns

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	teardownFlagName = "teardown"
)

// TeardownCommandContext contains the information available to any TeardownCommand.
type TeardownCommandContext struct {
	*command.SystemCommandContext
	Teardown v1.TeardownID
}

// TeardownCommand is a Command that acts on a specific teardown in a specific system.
// More practically, it is a valid SystemCommand and also validates that a teardown was specified.
type TeardownCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *TeardownCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the TeardownCommand.
func (c *TeardownCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var teardown string
	c.Flags[teardownFlagName] = &flags.String{
		Required: true,
		Target:   &teardown,
	}

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			teardownCtx := &TeardownCommandContext{
				SystemCommandContext: ctx,
				Teardown:             v1.TeardownID(teardown),
			}
			return c.Run(teardownCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}

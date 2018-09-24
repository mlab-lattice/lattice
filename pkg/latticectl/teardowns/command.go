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

type Command struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *TeardownCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

type TeardownCommandContext struct {
	*command.SystemCommandContext
	Teardown v1.TeardownID
}

func (c *Command) Command() *cli.Command {
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

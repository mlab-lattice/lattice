package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

type SystemCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *SystemCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

type SystemCommandContext struct {
	*LatticeCommandContext
	System v1.SystemID
}

func (c *SystemCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var system string
	c.Flags[SystemFlagName] = SystemFlag(&system)

	cmd := &LatticeCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *LatticeCommandContext, args []string, f cli.Flags) error {
			system := v1.SystemID(system)

			// Try to retrieve the lattice from the context if there is one
			if system == "" {
				system = ctx.Context.System
			}

			if system == "" {
				return flags.NewFlagsNotSetError([]string{SystemFlagName})
			}

			systemCtx := &SystemCommandContext{
				LatticeCommandContext: ctx,
				System:                system,
			}
			return c.Run(systemCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}

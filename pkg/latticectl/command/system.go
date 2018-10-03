package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

// SystemCommandContext contains the information available to any LatticeCommand.
type SystemCommandContext struct {
	*LatticeCommandContext
	System v1.SystemID
}

// SystemCommand is a Command that acts on a specific system in a specific lattice.
// More practically, it is a valid LatticeCommand and also validates that a system was specified.
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

// Command returns a *cli.Command for the SystemCommand.
func (c *SystemCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	// allow system to be overridden via flag
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

			// if no system was explicitly set, check the context
			if !f[SystemFlagName].Set() {
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

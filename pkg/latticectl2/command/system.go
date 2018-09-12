package command

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

type SystemCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx *SystemCommandContext, args []string, flags cli.Flags)
	Subcommands map[string]*cli.Command
}

type SystemCommandContext struct {
	*LatticeCommandContext
	System v1.SystemID
}

func (c *SystemCommand) Command() *cli.Command {
	c.Flags["system"] = &flags.String{
		Required: false,
	}

	cmd := &LatticeCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		Run: func(ctx *LatticeCommandContext, args []string, flags cli.Flags) {
			system := v1.SystemID(c.Flags["system"].Value().(string))
			// Try to retrieve the lattice from the context if there is one
			if system == "" {
				system = ctx.Context.System
			}

			if system == "" {
				log.Fatal("required flag system must be set")
			}

			systemCtx := &SystemCommandContext{
				LatticeCommandContext: ctx,
				System:                system,
			}
			c.Run(systemCtx, args, flags)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}

package latticectl

import (
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx SystemCommandContext, args []string)
	Subcommands []Command
}

type SystemCommandContext interface {
	LatticeCommandContext
	SystemID() types.SystemID
}

type systemCommandContext struct {
	LatticeCommandContext
	systemID types.SystemID
}

func (c *systemCommandContext) SystemID() types.SystemID {
	return c.systemID
}

func (c *SystemCommand) Base() (*BaseCommand, error) {
	var system string
	systemNameFlag := &command.StringFlag{
		Name:     "system",
		Required: false,
		Target:   &system,
	}
	flags := append(c.Flags, systemNameFlag)

	cmd := &LatticeCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(lctx LatticeCommandContext, args []string) {
			// Try to retrieve the lattice from the context if there is one
			systemID := types.SystemID(system)
			if systemID == "" && lctx.Latticectl().Context != nil {
				ctx, err := lctx.Latticectl().Context.Get()
				if err != nil {
					log.Fatal(err)
				}

				systemID = ctx.System()
			}

			if systemID == "" {
				log.Fatal("required flag system must be set")
			}

			ctx := &systemCommandContext{
				LatticeCommandContext: lctx,
				systemID:              systemID,
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

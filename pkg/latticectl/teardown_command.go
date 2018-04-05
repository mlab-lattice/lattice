package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type TeardownCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx TeardownCommandContext, args []string)
	Subcommands []Command
}

type TeardownCommandContext interface {
	SystemCommandContext
	TeardownID() v1.TeardownID
}

type teardownCommandContext struct {
	SystemCommandContext
	teardownID v1.TeardownID
}

func (c *teardownCommandContext) TeardownID() v1.TeardownID {
	return c.teardownID
}

func (c *TeardownCommand) Base() (*BaseCommand, error) {
	var teardownID string
	teardownIDFlag := &cli.StringFlag{
		Name:     "teardown",
		Required: true,
		Target:   &teardownID,
	}
	flags := append(c.Flags, teardownIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &teardownCommandContext{
				SystemCommandContext: sctx,
				teardownID:           v1.TeardownID(teardownID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

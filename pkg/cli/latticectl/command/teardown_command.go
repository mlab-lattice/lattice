package command

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type TeardownCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx TeardownCommandContext, args []string)
	Subcommands []latticectl.Command
}

type TeardownCommandContext interface {
	SystemCommandContext
	TeardownID() types.SystemTeardownID
}

type teardownCommandContext struct {
	SystemCommandContext
	teardownID types.SystemTeardownID
}

func (c *teardownCommandContext) TeardownID() types.SystemTeardownID {
	return c.teardownID
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	var teardownID string
	teardownIDFlag := &command.StringFlag{
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
				teardownID:           types.SystemTeardownID(teardownID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

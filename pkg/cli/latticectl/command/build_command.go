package command

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type BuildCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx BuildCommandContext, args []string)
	Subcommands []latticectl.Command
}

type BuildCommandContext interface {
	SystemCommandContext
	BuildID() types.BuildID
}

type buildCommandContext struct {
	SystemCommandContext
	buildID types.BuildID
}

func (c *buildCommandContext) BuildID() types.BuildID {
	return c.buildID
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	var buildID string
	buildIDFlag := &command.StringFlag{
		Name:     "build",
		Required: true,
		Target:   &buildID,
	}
	flags := append(c.Flags, buildIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &buildCommandContext{
				SystemCommandContext: sctx,
				buildID:              types.BuildID(buildID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

package command

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/latticectl"
)

type BuildCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx BuildCommandContext, args []string)
	Subcommands []latticectl.Command
}

type BuildCommandContext interface {
	SystemCommandContext
	BuildID() v1.BuildID
}

type buildCommandContext struct {
	SystemCommandContext
	buildID v1.BuildID
}

func (c *buildCommandContext) BuildID() v1.BuildID {
	return c.buildID
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	var buildID string
	buildIDFlag := &cli.StringFlag{
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
				buildID:              v1.BuildID(buildID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

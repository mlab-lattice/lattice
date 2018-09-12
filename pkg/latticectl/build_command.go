package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type BuildCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx BuildCommandContext, args []string)
	Subcommands []Command
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

func (c *BuildCommand) Base() (*BaseCommand, error) {
	var buildID string
	buildIDFlag := &flags.String{
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

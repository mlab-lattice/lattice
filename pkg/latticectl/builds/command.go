package builds

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	buildFlagName = "build"
)

type Command struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *BuildCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

type BuildCommandContext struct {
	*command.SystemCommandContext
	Build v1.BuildID
}

func (c *Command) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var build string
	c.Flags[buildFlagName] = &flags.String{
		Required: true,
		Target:   &build,
	}

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			buildCtx := &BuildCommandContext{
				SystemCommandContext: ctx,
				Build:                v1.BuildID(build),
			}
			return c.Run(buildCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}

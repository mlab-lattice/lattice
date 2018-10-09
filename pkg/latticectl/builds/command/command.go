package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	buildFlagName = "build"
)

// BuildCommandContext contains the information available to any BuildCommand.
type BuildCommandContext struct {
	*command.SystemCommandContext
	Build v1.BuildID
}

// BuildCommand is a Command that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a build was specified.
type BuildCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *BuildCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the BuildCommand.
func (c *BuildCommand) Command() *cli.Command {
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

package deploys

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	deployFlagName = "deploy"
)

// DeployCommandContext contains the information available to any LatticeCommand.
type DeployCommandContext struct {
	*command.SystemCommandContext
	Deploy v1.DeployID
}

// DeployCommand is a Command that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a build was specified.
type DeployCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *DeployCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the DeployCommand.
func (c *DeployCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var deploy string
	c.Flags[deployFlagName] = &flags.String{
		Required: true,
		Target:   &deploy,
	}

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			deployCtx := &DeployCommandContext{
				SystemCommandContext: ctx,
				Deploy:               v1.DeployID(deploy),
			}
			return c.Run(deployCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}

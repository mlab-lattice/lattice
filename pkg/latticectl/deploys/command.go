package deploys

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	deployFlagName = "deploy"
)

type Command struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *DeployCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

type DeployCommandContext struct {
	*command.SystemCommandContext
	Deploy v1.DeployID
}

func (c *Command) Command() *cli.Command {
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

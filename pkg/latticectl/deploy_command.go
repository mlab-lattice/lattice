package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type DeployCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx DeployCommandContext, args []string)
	Subcommands []Command
}

type DeployCommandContext interface {
	SystemCommandContext
	DeployID() v1.DeployID
}

type deployCommandContext struct {
	SystemCommandContext
	deployID v1.DeployID
}

func (c *deployCommandContext) DeployID() v1.DeployID {
	return c.deployID
}

func (c *DeployCommand) Base() (*BaseCommand, error) {
	var deployID string
	deployIDFlag := &flags.String{
		Name:     "deploy",
		Required: true,
		Target:   &deployID,
	}
	flags := append(c.Flags, deployIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &deployCommandContext{
				SystemCommandContext: sctx,
				deployID:             v1.DeployID(deployID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

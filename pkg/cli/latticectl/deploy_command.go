package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeployCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx DeployCommandContext)
	Subcommands []LatticeCommand
	*SystemCommand
}

func (c *DeployCommand) deployBase() *DeployCommand {
	return c
}

type DeployCommandContext interface {
	SystemCommandContext
	DeployID() types.SystemRolloutID
}

type deployCommandContext struct {
	SystemCommandContext
	deployID types.SystemRolloutID
}

func (c *deployCommandContext) DeployID() types.SystemRolloutID {
	return c.deployID
}

func (c *DeployCommand) Init() error {
	var deployID string
	deployIDFlag := &command.StringFlag{
		Name:     "deploy",
		Required: true,
		Target:   &deployID,
	}
	flags := append(c.Flags, deployIDFlag)

	c.SystemCommand = &SystemCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string, sctx SystemCommandContext) {
			ctx := &deployCommandContext{
				SystemCommandContext: sctx,
				deployID:             types.SystemRolloutID(deployID),
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.SystemCommand.Init()
}

package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemCommand struct {
	Name           string
	Short          string
	Args           command.Args
	Flags          command.Flags
	PreRun         func()
	Run            func(args []string, ctx SystemCommandContext)
	ContextCreator func(ctx LatticeCommandContext, systemID types.SystemID) SystemCommandContext
	Subcommands    []command.Command
	*LatticeCommand
}

type SystemCommandContext interface {
	LatticeCommandContext
	SystemID() types.SystemID
}

type systemCommandContext struct {
	LatticeCommandContext
	systemID types.SystemID
}

func (c *systemCommandContext) SystemID() types.SystemID {
	return c.systemID
}

func DefaultSystemContextCreator(ctx LatticeCommandContext, systemID types.SystemID) SystemCommandContext {
	return &systemCommandContext{
		LatticeCommandContext: ctx,
		systemID:              systemID,
	}
}

func (c *SystemCommand) Init() error {
	var systemID string
	systemNameFlag := &command.StringFlag{
		Name:     "system",
		Required: true,
		Target:   &systemID,
	}
	flags := append(c.Flags, systemNameFlag)

	c.LatticeCommand = &LatticeCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string, lctx LatticeCommandContext) {
			ctx := c.ContextCreator(lctx, types.SystemID(systemID))
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.LatticeCommand.Init()
}

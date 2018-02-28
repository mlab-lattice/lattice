package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx SystemCommandContext)
	Subcommands []LatticeCommand
	*BaseLatticeCommand
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

func (c *SystemCommand) Init() error {
	var systemID string
	systemNameFlag := &command.StringFlag{
		Name:     "system",
		Required: true,
		Target:   &systemID,
	}
	flags := append(c.Flags, systemNameFlag)

	c.BaseLatticeCommand = &BaseLatticeCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string, lctx LatticeCommandContext) {
			ctx := &systemCommandContext{
				LatticeCommandContext: lctx,
				systemID:              types.SystemID(systemID),
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.BaseLatticeCommand.Init()
}

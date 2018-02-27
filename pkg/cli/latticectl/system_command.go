package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx SystemCommandContext)
	Subcommands []command.Command
	*LatticeCommand
}

type SystemCommandContext interface {
	LatticeCommandContext
	SystemID() types.SystemID
	Systems() client.SystemClient
}

type systemCommandContext struct {
	LatticeCommandContext
	systemID     types.SystemID
	systemClient client.SystemClient
}

func (c *systemCommandContext) SystemID() types.SystemID {
	return c.systemID
}

func (c *systemCommandContext) Systems() client.SystemClient {
	return c.LatticeCommandContext.Lattice().Systems()
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
			ctx := &systemCommandContext{
				LatticeCommandContext: lctx,
				systemID:              types.SystemID(systemID),
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.LatticeCommand.Init()
}

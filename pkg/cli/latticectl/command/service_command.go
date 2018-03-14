package command

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type ServiceCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx ServiceCommandContext, args []string)
	Subcommands []latticectl.Command
}

type ServiceCommandContext interface {
	SystemCommandContext
	ServiceID() types.ServiceID
}

type serviceCommandContext struct {
	SystemCommandContext
	serviceID types.ServiceID
}

func (c *serviceCommandContext) ServiceID() types.ServiceID {
	return c.serviceID
}

func (c *ServiceCommand) Base() (*latticectl.BaseCommand, error) {
	var serviceID string
	serviceIDFlag := &command.StringFlag{
		Name:     "service",
		Required: true,
		Target:   &serviceID,
	}
	flags := append(c.Flags, serviceIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &serviceCommandContext{
				SystemCommandContext: sctx,
				serviceID:            types.ServiceID(serviceID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

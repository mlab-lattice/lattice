package latticectl

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type ServiceCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx ServiceCommandContext, args []string)
	Subcommands []Command
}

type ServiceCommandContext interface {
	SystemCommandContext
	ServiceID() v1.ServiceID
}

type serviceCommandContext struct {
	SystemCommandContext
	serviceID v1.ServiceID
}

func (c *serviceCommandContext) ServiceID() v1.ServiceID {
	return c.serviceID
}

func (c *ServiceCommand) Base() (*BaseCommand, error) {
	var serviceStr string
	serviceIDFlag := &cli.StringFlag{
		Name:     "service",
		Required: true,
		Target:   &serviceStr,
	}

	flags := append(c.Flags, serviceIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			var serviceID v1.ServiceID
			// resolve service id

			nodePath, err := tree.NewNodePath(serviceStr)
			if err == nil {
				c := sctx.Client().Systems().Services(sctx.SystemID())
				service, err := c.GetByServicePath(nodePath)

				if err != nil {
					log.Fatalf("error looking up service by path: %v", err)
				}

				serviceID = service.ID
			} else {
				//TODO validate that serviceStr is a valid service id
				serviceID = v1.ServiceID(serviceStr)
			}

			ctx := &serviceCommandContext{
				SystemCommandContext: sctx,
				serviceID:            serviceID,
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

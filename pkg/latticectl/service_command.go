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
	ServiceId() v1.ServiceID
}

type serviceCommandContext struct {
	SystemCommandContext
	serviceID v1.ServiceID
}

func (c *serviceCommandContext) ServiceId() v1.ServiceID {
	return c.serviceID
}

func (c *ServiceCommand) Base() (*BaseCommand, error) {
	var serviceIDStr string
	var servicePathStr string
	serviceIDFlag := &cli.StringFlag{
		Name:     "service",
		Required: false,
		Target:   &serviceIDStr,
	}

	servicePathFlag := &cli.StringFlag{
		Name:     "service-path",
		Required: false,
		Target:   &servicePathStr,
	}

	flags := append(c.Flags, serviceIDFlag)
	flags = append(flags, servicePathFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			var serviceID v1.ServiceID
			// resolve service id
			if serviceIDStr == "" && servicePathStr == "" {
				log.Fatal("Need to specify service or servicePath")
			} else if serviceIDStr != "" {
				serviceID = v1.ServiceID(serviceIDStr)
			} else if servicePathStr != "" {
				// lookup service by node path
				nodePath, err := tree.NewNodePath(servicePathStr)
				if err != nil {
					log.Fatal("invalid service path: " + servicePathStr)
				}

				c := sctx.Client().Systems().Services(sctx.SystemID())
				service, err := c.GetByServicePath(nodePath)

				if err != nil {
					log.Fatalf("error looking up service by path: %v", err)
				}

				serviceID = service.ID
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

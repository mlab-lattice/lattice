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
	ServicePath() tree.NodePath
	ServiceId() v1.ServiceID
}

type serviceCommandContext struct {
	SystemCommandContext
	servicePath tree.NodePath
	serviceID   v1.ServiceID
}

func (c *serviceCommandContext) ServicePath() tree.NodePath {
	return c.servicePath
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

			var servicePath tree.NodePath

			if serviceIDStr == "" && servicePathStr == "" {
				log.Fatal("Need to specify service or servicePath")
			} else if servicePathStr != "" {
				nodePath, err := tree.NewNodePath(servicePathStr)
				if err != nil {
					log.Fatal("invalid service path: " + servicePathStr)
				}
				servicePath = nodePath
			}

			ctx := &serviceCommandContext{
				SystemCommandContext: sctx,
				serviceID:            v1.ServiceID(serviceIDStr),
				servicePath:          servicePath,
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

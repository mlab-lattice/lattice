package latticectl

import (
	"log"

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
}

type serviceCommandContext struct {
	SystemCommandContext
	servicePath tree.NodePath
}

func (c *serviceCommandContext) ServicePath() tree.NodePath {
	return c.servicePath
}

func (c *ServiceCommand) Base() (*BaseCommand, error) {
	var servicePathStr string
	serviceIDFlag := &cli.StringFlag{
		Name:     "service",
		Required: true,
		Target:   &servicePathStr,
	}
	flags := append(c.Flags, serviceIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			servicePath, err := tree.NewNodePath(servicePathStr)
			if err != nil {
				log.Fatal("invalid service path: " + servicePathStr)
			}

			ctx := &serviceCommandContext{
				SystemCommandContext: sctx,
				servicePath:          servicePath,
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

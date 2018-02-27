package command

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

type LatticeCommand struct {
	Name        string
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string, ctx LatticeCommandContext)
	Subcommands []Command
	*BaseCommand
}

type LatticeCommandContext interface {
	URL() string
	Lattice() client.Interface
}

type latticeCommandContext struct {
	url           string
	latticeClient client.Interface
}

func (c *latticeCommandContext) URL() string {
	return c.url
}

func (c *latticeCommandContext) Lattice() client.Interface {
	if c.latticeClient == nil {
		c.latticeClient = rest.NewClient(c.url)
	}

	return c.latticeClient
}

func (c *LatticeCommand) Init() error {
	var latticeURL string
	latticeURLFlag := &StringFlag{
		Name:     "lattice-url",
		Required: true,
		Target:   &latticeURL,
	}
	flags := append(c.Flags, latticeURLFlag)

	c.BaseCommand = &BaseCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string) {
			ctx := &latticeCommandContext{
				url: latticeURL,
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.BaseCommand.Init()
}

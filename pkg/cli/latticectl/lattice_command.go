package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

type LatticeCommand struct {
	Name           string
	Short          string
	Args           command.Args
	Flags          command.Flags
	PreRun         func()
	Run            func(args []string, ctx LatticeCommandContext)
	ContextCreator func(lattice string) LatticeCommandContext
	Subcommands    []command.Command
	*command.BaseCommand
}

type LatticeCommandContext interface {
	Lattice() string
	Client() client.Interface
}

type latticeCommandContext struct {
	lattice       string
	latticeClient client.Interface
}

func (c *latticeCommandContext) Lattice() string {
	return c.lattice
}

func (c *latticeCommandContext) Client() client.Interface {
	return c.latticeClient
}

func DefaultLatticeContextCreator(lattice string) LatticeCommandContext {
	return &latticeCommandContext{
		lattice:       lattice,
		latticeClient: rest.NewClient(lattice),
	}
}

func (c *LatticeCommand) Init() error {
	var lattice string
	latticeURLFlag := &command.StringFlag{
		Name:     "lattice",
		Required: true,
		Target:   &lattice,
	}
	flags := append(c.Flags, latticeURLFlag)

	c.BaseCommand = &command.BaseCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string) {
			ctx := c.ContextCreator(lattice)
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.BaseCommand.Init()
}

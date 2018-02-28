package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

type LatticeClientGenerator func(lattice string) client.Interface

type LatticeCommand interface {
	command.Command
	setClient(generator LatticeClientGenerator, init bool) error
}

type BaseLatticeCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx LatticeCommandContext)
	Client      LatticeClientGenerator
	Subcommands []LatticeCommand
	*command.BaseCommand
}

func (c *BaseLatticeCommand) client() LatticeClientGenerator {
	return c.Client
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

func DefaultLatticeClient(lattice string) client.Interface {
	return rest.NewClient(lattice)
}

func (c *BaseLatticeCommand) Init() error {
	var lattice string
	latticeURLFlag := &command.StringFlag{
		Name:     "lattice",
		Required: true,
		Target:   &lattice,
	}
	flags := append(c.Flags, latticeURLFlag)

	var subcommands []command.Command
	for _, subcommand := range c.Subcommands {
		subcommands = append(subcommands, subcommand)
	}

	c.BaseCommand = &command.BaseCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string) {
			ctx := &latticeCommandContext{
				lattice:       lattice,
				latticeClient: c.Client(lattice),
			}
			c.Run(args, ctx)
		},
		Subcommands: subcommands,
	}

	if err := c.BaseCommand.Init(); err != nil {
		return err
	}

	return c.initClient()
}

func (c *BaseLatticeCommand) initClient() error {
	if c.Client == nil {
		return nil
	}

	return c.setClient(c.Client, true)
}

func (c *BaseLatticeCommand) setClient(clientFunc LatticeClientGenerator, init bool) error {
	// if my client func has already been set, and this isn't an init pass,
	// then I've also already set all my subcommands' client funcs. nothing else to do
	if c.Client != nil && !init {
		return nil
	}

	c.Client = clientFunc

	// otherwise, offer the client up to my subcommands.
	// if they have their own clients set already then the'll just
	// decline the client via the above guard. if they don't, they'll
	// accept it pass it down to their children.
	// this should result in all of the subcommands inheriting the
	// client func closest to them in the tree
	for _, subcommand := range c.Subcommands {
		if err := subcommand.setClient(clientFunc, false); err != nil {
			return err
		}
	}

	return nil
}

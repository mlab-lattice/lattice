package latticectl

import (
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type LatticeCommandContext interface {
	Lattice() string
	Client() client.Interface
	Latticectl() *Latticectl
}

type latticeCommandContext struct {
	lattice       string
	latticeClient client.Interface
	latticectl    *Latticectl
}

func (c *latticeCommandContext) Lattice() string {
	return c.lattice
}

func (c *latticeCommandContext) Client() client.Interface {
	return c.latticeClient
}

func (c *latticeCommandContext) Latticectl() *Latticectl {
	return c.latticectl
}

type LatticeCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx LatticeCommandContext, args []string)
	Subcommands []Command
}

func (c *LatticeCommand) Base() (*BaseCommand, error) {
	var lattice string
	latticeURLFlag := &command.StringFlag{
		Name:     "lattice",
		Required: false,
		Target:   &lattice,
	}
	flags := append(c.Flags, latticeURLFlag)

	cmd := &BaseCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(latticectl *Latticectl, args []string) {
			// Try to retrieve the lattice from the context if there is one
			if lattice == "" && latticectl.Context != nil {
				ctx, err := latticectl.Context.Get()
				if err != nil {
					log.Fatal(err)
				}

				lattice = ctx.Lattice()
			}

			if latticectl.Client == nil {
				log.Fatal("client must be set")
			}

			if lattice == "" {
				log.Fatal("required flag lattice must be set")
			}

			ctx := &latticeCommandContext{
				lattice:       lattice,
				latticeClient: latticectl.Client(lattice),
				latticectl:    latticectl,
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

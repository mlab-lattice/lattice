package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	PreRun      func()
	Client      latticectl.LatticeClientGenerator
	Subcommands []latticectl.LatticeCommand
	*latticectl.BaseLatticeCommand
}

func (c *Command) Init() error {
	c.BaseLatticeCommand = &latticectl.BaseLatticeCommand{
		Name:   "systems",
		PreRun: c.PreRun,
		Run: func(args []string, ctx latticectl.LatticeCommandContext) {
			c.run(ctx)
		},
		Client:      c.Client,
		Subcommands: c.Subcommands,
	}

	return c.BaseLatticeCommand.Init()
}

func (c *Command) run(ctx latticectl.LatticeCommandContext) {
	systems, err := ctx.Client().Systems().List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", systems)
}

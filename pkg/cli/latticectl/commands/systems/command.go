package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	PreRun         func()
	ContextCreator func(lattice string) latticectl.LatticeCommandContext
	Subcommands    []command.Command
	*latticectl.LatticeCommand
}

func (c *Command) Init() error {
	if c.ContextCreator == nil {
		c.ContextCreator = latticectl.DefaultLatticeContextCreator
	}

	c.LatticeCommand = &latticectl.LatticeCommand{
		Name:   "systems",
		PreRun: c.PreRun,
		Run: func(args []string, ctx latticectl.LatticeCommandContext) {
			c.run(ctx)
		},
		ContextCreator: c.ContextCreator,
		Subcommands:    c.Subcommands,
	}

	return c.LatticeCommand.Init()
}

func (c *Command) run(ctx latticectl.LatticeCommandContext) {
	systems, err := ctx.Client().Systems().List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", systems)
}

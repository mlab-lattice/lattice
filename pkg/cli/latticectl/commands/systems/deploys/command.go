package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	PreRun      func()
	Subcommands []latticectl.LatticeCommand
	*latticectl.SystemCommand
}

func (c *Command) Init() error {
	c.SystemCommand = &latticectl.SystemCommand{
		Name:   "deploys",
		PreRun: c.PreRun,
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx)
		},
		Subcommands: c.Subcommands,
	}

	return c.SystemCommand.Init()
}

func (c *Command) run(ctx latticectl.SystemCommandContext) {
	deploys, err := ctx.Client().Systems().Rollouts(ctx.SystemID()).List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploys)
}

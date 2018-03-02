package context

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) BaseCommand() (*command.BaseCommand2, error) {
	cmd := &latticectl.BaseCommand{
		Name: "get",
		Run: func(args []string, ctxm latticectl.ContextManager, client latticectl.LatticeClientGenerator) {
			c.run(ctxm)
		},
	}

	return cmd.BaseCommand()
}

func (c *GetCommand) run(ctxm latticectl.ContextManager) {
	var lattice string
	var system types.SystemID

	if ctxm != nil {
		ctx, err := ctxm.Get()
		if err != nil {
			log.Fatal(err)
		}

		lattice = ctx.Lattice()
		system = ctx.System()
	}

	fmt.Printf("lattice: %v\nsystem: %v\n", lattice, system)
}

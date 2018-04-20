package context

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.BaseCommand{
		Name: "context",
		Run: func(lctl *latticectl.Latticectl, args []string) {
			GetContext(lctl.Context)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func GetContext(ctxm latticectl.ContextManager) {
	var lattice string
	var system v1.SystemID

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

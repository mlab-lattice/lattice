package context

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.BaseCommand{
		Name: "get",
		Run: func(lctl *latticectl.Latticectl, args []string) {
			GetContext(lctl.Context)
		},
	}

	return cmd.Base()
}

func GetContext(ctxm latticectl.ContextManager) {
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

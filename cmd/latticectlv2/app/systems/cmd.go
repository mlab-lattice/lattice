package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

var Cmd = &latticectl.LatticeCommand{
	Name: "systems",
	Run: func(args []string, ctx latticectl.LatticeCommandContext) {
		systems, err := ctx.Lattice().Systems().List()
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("%v\n", systems)
	},
	Subcommands: []command.Command{
		create,
		get,
		delete,
	},
}

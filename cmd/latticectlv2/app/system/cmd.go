package system

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
)

var Cmd = &command.LatticeCommand{
	Name: "systems",
	Run: func(args []string, ctx command.LatticeCommandContext) {
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

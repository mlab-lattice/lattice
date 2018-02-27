package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

var get = &latticectl.SystemCommand{
	Name: "get",
	Run: func(args []string, ctx latticectl.SystemCommandContext) {
		system, err := ctx.Systems().Get(ctx.SystemID())
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("%v\n", system)
	},
}

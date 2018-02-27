package system

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
)

var get = &command.SystemCommand{
	Name: "get",
	Run: func(args []string, ctx command.SystemCommandContext) {
		system, err := ctx.Systems().Get(ctx.SystemID())
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("%v\n", system)
	},
}

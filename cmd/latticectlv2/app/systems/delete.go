package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

var delete = &latticectl.SystemCommand{
	Name: "delete",
	Run: func(args []string, ctx latticectl.SystemCommandContext) {
		if err := ctx.Systems().Delete(ctx.SystemID()); err != nil {
			log.Panic(err)
		}

		fmt.Println("succesful")
	},
}

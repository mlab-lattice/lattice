package system

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
)

var delete = &command.SystemCommand{
	Name: "delete",
	Run: func(args []string, ctx command.SystemCommandContext) {
		if err := ctx.Systems().Delete(ctx.SystemID()); err != nil {
			log.Panic(err)
		}

		fmt.Println("succesful")
	},
}

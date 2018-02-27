package system

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

var (
	definitionURL string
	systemName    string
)

var create = &command.LatticeCommand{
	Name: "create",
	Flags: []command.Flag{
		&command.StringFlag{
			Name:     "definition",
			Required: true,
			Target:   &definitionURL,
		},
		&command.StringFlag{
			Name:     "name",
			Required: true,
			Target:   &systemName,
		},
	},
	Run: func(args []string, ctx command.LatticeCommandContext) {
		system, err := ctx.Lattice().Systems().Create(types.SystemID(systemName), definitionURL)
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("%v\n", system)
	},
}

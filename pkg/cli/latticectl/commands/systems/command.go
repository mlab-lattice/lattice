package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.LatticeCommand{
		Name: "systems",
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {
			ListSystems(ctx.Client().Systems())
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSystems(client client.SystemClient) {
	systems, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", systems)
}

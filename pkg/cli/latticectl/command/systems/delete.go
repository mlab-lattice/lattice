package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeleteCommand struct {
}

func (c *DeleteCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.SystemCommand{
		Name: "delete",
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			DeleteSystem(ctx.Client().Systems(), ctx.SystemID())
		},
	}

	return cmd.Base()
}

func DeleteSystem(client client.SystemClient, name types.SystemID) {
	err := client.Delete(name)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Deleted: %v\n", name)
}

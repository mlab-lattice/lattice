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
			err := DeleteSystem(ctx.Client().Systems(), ctx.SystemID())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func DeleteSystem(client client.SystemClient, name types.SystemID) error {
	err := client.Delete(name)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted: %v\n", name)
	return nil
}

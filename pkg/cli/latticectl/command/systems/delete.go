package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeleteCommand struct {
}

func (c *DeleteCommand) Base() (*latticectl.BaseCommand, error) {
	var system string

	cmd := &lctlcommand.SystemCommand{
		Name: "delete",
		Flags: command.Flags{
			&command.StringFlag{
				Name:     "system",
				Required: true,
				Target:   &system,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			err := DeleteSystem(ctx.Client().Systems(), types.SystemID(system))
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

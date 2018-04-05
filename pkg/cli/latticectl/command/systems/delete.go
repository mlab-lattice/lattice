package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/color"
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

	cmd := &lctlcommand.LatticeCommand{
		Name: "delete",
		Flags: command.Flags{
			&command.StringFlag{
				Name:     "system",
				Required: true,
				Target:   &system,
			},
		},
		Run: func(ctx lctlcommand.LatticeCommandContext, args []string) {
			err := DeleteSystem(ctx.Client().Systems(), types.SystemID(system), os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func DeleteSystem(client client.SystemClient, name types.SystemID, writer io.Writer) error {
	err := client.Delete(name)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "System %s deleted.\n", color.ID(string(name)))
	return nil
}

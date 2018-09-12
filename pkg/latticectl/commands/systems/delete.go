package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type DeleteCommand struct {
}

func (c *DeleteCommand) Base() (*latticectl.BaseCommand, error) {
	var system string

	cmd := &latticectl.LatticeCommand{
		Name: "delete",
		Flags: cli.Flags{
			&flags.String{
				Name:     "system",
				Required: true,
				Target:   &system,
			},
		},
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {
			err := DeleteSystem(ctx.Client().Systems(), v1.SystemID(system), os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func DeleteSystem(client v1client.SystemClient, name v1.SystemID, writer io.Writer) error {
	err := client.Delete(name)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "System %s deleted.\n", color.ID(string(name)))
	return nil
}

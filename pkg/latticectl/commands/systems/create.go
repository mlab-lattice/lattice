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
)

type CreateCommand struct {
}

func (c *CreateCommand) Base() (*latticectl.BaseCommand, error) {
	var definitionURL string
	var systemName string
	cmd := &latticectl.LatticeCommand{
		Name: "create",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "definition",
				Required: true,
				Target:   &definitionURL,
			},
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &systemName,
			},
		},
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {

			err := CreateSystem(ctx.Client().Systems(), v1.SystemID(systemName), definitionURL, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func CreateSystem(client v1client.SystemClient, name v1.SystemID, definitionURL string, writer io.Writer) error {
	system, err := client.Create(name, definitionURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "System %s created. To rollout a version of this system run:\n\n", color.ID(string(system.ID)))
	fmt.Fprintf(writer, "    lattice systems:deploy --system %s --version <tag>\n", system.ID)
	return nil
}

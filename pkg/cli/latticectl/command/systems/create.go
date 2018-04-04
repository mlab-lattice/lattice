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

type CreateCommand struct {
}

func (c *CreateCommand) Base() (*latticectl.BaseCommand, error) {
	var definitionURL string
	var systemName string
	cmd := &lctlcommand.LatticeCommand{
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
		Run: func(ctx lctlcommand.LatticeCommandContext, args []string) {

			err := CreateSystem(ctx.Client().Systems(), types.SystemID(systemName), definitionURL, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func CreateSystem(client client.SystemClient, name types.SystemID, definitionURL string, writer io.Writer) error {
	system, err := client.Create(name, definitionURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "System %s created. To rollout a version of this system run:\n\n", color.ID(string(system.ID)))
	fmt.Fprintf(writer, "    lattice systems:deploy --system %s --version <tag>\n", system.ID)
	return nil
}

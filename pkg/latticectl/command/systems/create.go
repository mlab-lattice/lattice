package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type CreateCommand struct {
}

func (c *CreateCommand) Base() (*latticectl.BaseCommand, error) {
	var definitionURL string
	var systemName string
	cmd := &command.LatticeCommand{
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
		Run: func(ctx command.LatticeCommandContext, args []string) {
			CreateSystem(ctx.Client().Systems(), v1.SystemID(systemName), definitionURL)
		},
	}

	return cmd.Base()
}

func CreateSystem(client clientv1.SystemClient, name v1.SystemID, definitionURL string) {
	system, err := client.Create(name, definitionURL)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}

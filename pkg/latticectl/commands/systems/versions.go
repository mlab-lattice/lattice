package systems

import (
	"fmt"
	"log"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type ListVersionsCommand struct {
}

func (c *ListVersionsCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.SystemCommand{
		Name:  "versions",
		Flags: []cli.Flag{},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			err := ListVersions(ctx.Client().Systems(), ctx.SystemID())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func ListVersions(client v1client.SystemClient, id v1.SystemID) error {
	versions, err := client.Versions(id)
	if err != nil {
		return err
	}

	for _, version := range versions {
		fmt.Printf("%v\n", version)
	}
	return nil
}

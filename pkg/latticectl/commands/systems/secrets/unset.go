package secrets

import (
	"log"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type UnsetCommand struct {
}

func (c *UnsetCommand) Base() (*latticectl.BaseCommand, error) {
	var path string

	cmd := &latticectl.SystemCommand{
		Name: "unset",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "path",
				Required: true,
				Target:   &path,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			secretPath, err := tree.NewPathSubcomponent(path)
			if err != nil {
				log.Fatal("invalid secret path format")
			}

			err = UnsetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), secretPath)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func UnsetSecret(client v1client.SystemSecretClient, path tree.PathSubcomponent) error {
	err := client.Unset(path)
	if err != nil {
		return err
	}

	return nil
}

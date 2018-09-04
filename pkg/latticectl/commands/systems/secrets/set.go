package secrets

import (
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"log"
)

type SetCommand struct {
}

func (c *SetCommand) Base() (*latticectl.BaseCommand, error) {
	var path string
	var value string

	cmd := &latticectl.SystemCommand{
		Name: "set",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "path",
				Required: true,
				Target:   &path,
			},
			&cli.StringFlag{
				Name:     "value",
				Required: true,
				Target:   &value,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			secretPath, err := tree.NewPathSubcomponent(path)
			if err != nil {
				log.Fatal("invalid secret path format")
			}

			err = SetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), secretPath, value)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func SetSecret(client v1client.SystemSecretClient, path tree.PathSubcomponent, value string) error {
	err := client.Set(path, value)
	if err != nil {
		return err
	}

	return nil
}

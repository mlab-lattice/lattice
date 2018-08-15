package secrets

import (
	"log"
	"strings"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type SetCommand struct {
}

func (c *SetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var value string

	cmd := &latticectl.SystemCommand{
		Name: "set",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &name,
			},
			&cli.StringFlag{
				Name:     "value",
				Required: true,
				Target:   &value,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			splitName := strings.Split(name, ":")
			if len(splitName) != 2 {
				log.Fatal("invalid secret name format")
			}

			path := tree.Path(splitName[0])
			name = splitName[1]

			err := SetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name, value)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func SetSecret(client v1client.SecretClient, path tree.Path, name, value string) error {
	err := client.Set(path, name, value)
	if err != nil {
		return err
	}

	return nil
}

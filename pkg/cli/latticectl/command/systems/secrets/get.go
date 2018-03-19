package secrets

import (
	"fmt"
	"log"
	"strings"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string

	cmd := &lctlcommand.SystemCommand{
		Name: "get",
		Flags: command.Flags{
			&command.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &name,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			splitName := strings.Split(name, ":")
			if len(splitName) != 2 {
				log.Fatal("invalid secret name format")
			}

			path := tree.NodePath(splitName[0])
			name = splitName[1]

			err := GetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetSecret(client client.SystemSecretClient, path tree.NodePath, name string) error {
	secret, err := client.Get(path, name)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", secret)
	return nil
}

package secrets

import (
	"log"
	"strings"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type UnsetCommand struct {
}

func (c *UnsetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var value string

	cmd := &lctlcommand.SystemCommand{
		Name: "unset",
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

			err := UnsetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name, value)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func UnsetSecret(client client.SystemSecretClient, path tree.NodePath, name, value string) error {
	err := client.Unset(path, name, value)
	if err != nil {
		return err
	}

	return nil
}

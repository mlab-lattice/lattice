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

type SetCommand struct {
}

func (c *SetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var value string

	cmd := &lctlcommand.SystemCommand{
		Name: "set",
		Flags: command.Flags{
			&command.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &name,
			},
			&command.StringFlag{
				Name:     "value",
				Required: true,
				Target:   &value,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			splitName := strings.Split(name, ":")
			if len(splitName) != 2 {
				log.Fatal("invalid secret name format")
			}

			path := tree.NodePath(splitName[0])
			name = splitName[1]

			SetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name, value)
		},
	}

	return cmd.Base()
}

func SetSecret(client client.SecretClient, path tree.NodePath, name, value string) {
	err := client.Set(path, name, value)
	if err != nil {
		log.Fatal(err)
	}
}

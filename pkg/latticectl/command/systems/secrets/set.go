package secrets

import (
	"log"
	"strings"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"
)

type SetCommand struct {
}

func (c *SetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var value string

	cmd := &command.SystemCommand{
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
		Run: func(ctx command.SystemCommandContext, args []string) {
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

func SetSecret(client clientv1.SecretClient, path tree.NodePath, name, value string) {
	err := client.Set(path, name, value)
	if err != nil {
		log.Fatal(err)
	}
}

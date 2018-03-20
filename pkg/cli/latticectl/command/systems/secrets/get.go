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

			GetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name)
		},
	}

	return cmd.Base()
}

func GetSecret(client client.SecretClient, path tree.NodePath, name string) {
	secret, err := client.Get(path, name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", secret)
}

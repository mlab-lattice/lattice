package secrets

import (
	"fmt"
	"log"
	"strings"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string

	cmd := &command.SystemCommand{
		Name: "get",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &name,
			},
		},
		Run: func(ctx command.SystemCommandContext, args []string) {
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

func GetSecret(client clientv1.SecretClient, path tree.NodePath, name string) {
	secret, err := client.Get(path, name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", secret)
}

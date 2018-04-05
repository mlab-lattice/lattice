package secrets

import (
	"fmt"
	"log"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
)

type ListSecretsCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListSecretsCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.SystemCommand{
		Name: "secrets",
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			err := ListSecrets(ctx.Client().Systems().Secrets(ctx.SystemID()))
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSecrets(client v1client.SecretClient) error {
	secrets, err := client.List()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", secrets)
	return nil
}

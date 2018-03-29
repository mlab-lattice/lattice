package secrets

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type ListSecretsCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListSecretsCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.SystemCommand{
		Name: "secrets",
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			err := ListSecrets(ctx.Client().Systems().Secrets(ctx.SystemID()))
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSecrets(client client.SystemSecretClient) error {
	secrets, err := client.List()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", secrets)
	return nil
}

package secrets

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.SystemCommand{
		Name: "secrets",
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			ListSecrets(ctx.Client().Systems().Secrets(ctx.SystemID()))
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSecrets(client client.SystemSecretClient) {
	secrets, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", secrets)
}

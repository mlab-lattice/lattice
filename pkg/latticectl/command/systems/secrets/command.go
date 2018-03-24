package secrets

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &command.SystemCommand{
		Name: "secrets",
		Run: func(ctx command.SystemCommandContext, args []string) {
			ListSecrets(ctx.Client().Systems().Secrets(ctx.SystemID()))
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSecrets(client clientv1.SecretClient) {
	secrets, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", secrets)
}

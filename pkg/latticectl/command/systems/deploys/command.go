package deploys

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
		Name: "deploys",
		Run: func(ctx command.SystemCommandContext, args []string) {
			ListDeploys(ctx.Client().Systems().Deploys(ctx.SystemID()))
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListDeploys(client clientv1.DeployClient) {
	deploys, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploys)
}

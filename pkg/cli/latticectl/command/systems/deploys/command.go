package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
)

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.SystemCommand{
		Name: "deploys",
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			ListDeploys(ctx.Client().Systems().Deploys(ctx.SystemID()))
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListDeploys(client client.DeployClient) {
	deploys, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploys)
}

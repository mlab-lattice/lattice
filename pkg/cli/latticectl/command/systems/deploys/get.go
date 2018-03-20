package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.DeployCommand{
		Name: "get",
		Run: func(ctx lctlcommand.DeployCommandContext, args []string) {
			GetDeploy(ctx.Client().Systems().Deploys(ctx.SystemID()), ctx.DeployID())
		},
	}

	return cmd.Base()
}

func GetDeploy(client client.DeployClient, deployID types.DeployID) {
	deploy, err := client.Get(deployID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}

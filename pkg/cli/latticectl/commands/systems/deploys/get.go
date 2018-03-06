package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.DeployCommand{
		Name: "get",
		Run: func(ctx latticectl.DeployCommandContext, args []string) {
			GetDeploy(ctx.Client().Systems().Rollouts(ctx.SystemID()), ctx.DeployID())
		},
	}

	return cmd.Base()
}

func GetDeploy(client client.RolloutClient, deployID types.SystemRolloutID) {
	deploy, err := client.Get(deployID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}

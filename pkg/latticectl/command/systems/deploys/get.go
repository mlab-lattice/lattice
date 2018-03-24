package deploys

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &command.DeployCommand{
		Name: "get",
		Run: func(ctx command.DeployCommandContext, args []string) {
			GetDeploy(ctx.Client().Systems().Deploys(ctx.SystemID()), ctx.DeployID())
		},
	}

	return cmd.Base()
}

func GetDeploy(client clientv1.DeployClient, deployID v1.DeployID) {
	deploy, err := client.Get(deployID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}

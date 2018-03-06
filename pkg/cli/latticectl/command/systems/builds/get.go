package builds

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
	cmd := &lctlcommand.BuildCommand{
		Name: "get",
		Run: func(ctx lctlcommand.BuildCommandContext, args []string) {
			GetDeploy(ctx.Client().Systems().SystemBuilds(ctx.SystemID()), ctx.BuildID())
		},
	}

	return cmd.Base()
}

func GetDeploy(client client.SystemBuildClient, buildID types.SystemBuildID) {
	deploy, err := client.Get(buildID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}

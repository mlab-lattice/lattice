package services

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
			GetService(ctx.Client().Systems().Services(ctx.SystemID()), types.ServiceID(0))
		},
	}

	return cmd.Base()
}

func GetService(client client.ServiceClient, service types.ServiceID) {
	deploy, err := client.Get(service)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}

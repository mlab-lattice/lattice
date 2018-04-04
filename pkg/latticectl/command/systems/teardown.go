package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &command.SystemCommand{
		Name: "teardown",
		Run: func(ctx command.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			TearDownSystem(ctx.Client().Systems().Teardowns(systemID))
		},
	}

	return cmd.Base()
}

func TearDownSystem(client clientv1.TeardownClient) {
	teardownID, err := client.Create()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", teardownID)
}

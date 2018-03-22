package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.SystemCommand{
		Name: "teardown",
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			TearDownSystem(ctx.Client().Systems().Teardowns(systemID))
		},
	}

	return cmd.Base()
}

func TearDownSystem(client client.TeardownClient) {
	teardownID, err := client.Create()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", teardownID)
}

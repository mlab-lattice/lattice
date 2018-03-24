package systems

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
	cmd := &command.SystemCommand{
		Name: "get",
		Run: func(ctx command.SystemCommandContext, args []string) {
			GetSystem(ctx.Client().Systems(), ctx.SystemID())
		},
	}

	return cmd.Base()
}

func GetSystem(client clientv1.SystemClient, name v1.SystemID) {
	system, err := client.Get(name)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}

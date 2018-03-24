package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"
)

type DeleteCommand struct {
}

func (c *DeleteCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &command.SystemCommand{
		Name: "delete",
		Run: func(ctx command.SystemCommandContext, args []string) {
			DeleteSystem(ctx.Client().Systems(), ctx.SystemID())
		},
	}

	return cmd.Base()
}

func DeleteSystem(client clientv1.SystemClient, name v1.SystemID) {
	system, err := client.Get(name)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}

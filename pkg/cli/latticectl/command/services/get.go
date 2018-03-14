package services

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.ServiceCommand{
		Name: "status",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.ServiceCommandContext, args []string) {
			//format, err := output.Value()
			//if err != nil {
			//	log.Fatal(err)
			//}

			c := ctx.Client().Systems().Services(ctx.SystemID())

			GetService(c, ctx.ServiceID())
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

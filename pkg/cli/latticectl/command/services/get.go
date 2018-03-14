package services

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
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
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Services(ctx.SystemID())

			if watch {
				WatchService(c, ctx.ServiceID(), format, os.Stdout)
			}

			GetService(c, ctx.ServiceID(), format, os.Stdout)
		},
	}

	return cmd.Base()
}

func GetService(client client.ServiceClient, serviceID types.ServiceID, format printer.Format, writer io.Writer) {
	service, err := client.Get(serviceID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", service)
}

func WatchService(client client.ServiceClient, serviceID types.ServiceID, format printer.Format, writer io.Writer) {
	service, err := client.Get(serviceID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", service)
}

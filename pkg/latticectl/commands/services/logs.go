package services

import (
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var follow bool
	var component string
	var instance string

	cmd := &latticectl.ServiceCommand{
		Name: "logs",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "component",
				Short:    "c",
				Required: true,
				Target:   &component,
			},
			&cli.StringFlag{
				Name:     "instance",
				Short:    "i",
				Required: true,
				Target:   &instance,
			},
			&cli.BoolFlag{
				Name:    "follow",
				Short:   "f",
				Default: false,
				Target:  &follow,
			},
		},
		Run: func(ctx latticectl.ServiceCommandContext, args []string) {
			c := ctx.Client().Systems().Services(ctx.SystemID())
			err := GetServiceLogs(c, ctx.ServiceId(), component, instance, follow, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceLogs(client v1client.ServiceClient, serviceID v1.ServiceID, component string, instance string,
	follow bool, w io.Writer) error {

	logs, err := client.Logs(serviceID, component, instance, follow)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

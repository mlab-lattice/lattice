package services

import (
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var follow bool
	var component string

	cmd := &latticectl.ServiceCommand{
		Name: "logs",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "component",
				Short:    "c",
				Required: true,
				Target:   &component,
			},
			&cli.BoolFlag{
				Name:    "follow",
				Short:   "f",
				Default: false,
				Target:  &follow,
			},
		},
		Run: func(ctx latticectl.ServiceCommandContext, args []string) {
			err := GetServiceLogs(ctx, component, follow, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceLogs(ctx latticectl.ServiceCommandContext, component string, follow bool, w io.Writer) error {
	c := ctx.Client().Systems().Services(ctx.SystemID())
	service, err := lookupService(ctx)

	if err != nil {
		return err
	}

	logs, err := c.Logs(service.ID, component, follow)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

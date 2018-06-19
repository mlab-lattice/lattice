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
	var component string
	var instance string
	var follow bool
	var previous bool
	var timestamps bool
	var sinceTime string
	var since string
	var tail int

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
				Required: false,
				Target:   &instance,
			},
			&cli.BoolFlag{
				Name:    "follow",
				Short:   "f",
				Default: false,
				Target:  &follow,
			},
			&cli.BoolFlag{
				Name:    "previous",
				Default: false,
				Target:  &previous,
			},
			&cli.BoolFlag{
				Name:    "timestamps",
				Default: false,
				Target:  &timestamps,
			},
			&cli.StringFlag{
				Name:     "since-time",
				Required: false,
				Target:   &sinceTime,
			},
			&cli.StringFlag{
				Name:     "since",
				Required: false,
				Target:   &since,
			},
			&cli.IntFlag{
				Name:     "tail",
				Required: false,
				Short:    "t",
				Target:   &tail,
			},
		},
		Run: func(ctx latticectl.ServiceCommandContext, args []string) {
			c := ctx.Client().Systems().Services(ctx.SystemID())
			logOptions := v1.NewContainerLogOptions()
			logOptions.Follow = follow
			logOptions.Previous = previous
			logOptions.Timestamps = timestamps
			logOptions.SinceTime = sinceTime
			logOptions.Since = since

			if tail != 0 {
				tl := int64(tail)
				logOptions.Tail = &tl
			}

			err := GetServiceLogs(c, ctx.ServiceID(), component, instance, logOptions, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceLogs(client v1client.ServiceClient, serviceID v1.ServiceID, component string, instance string,
	logOptions *v1.ContainerLogOptions, w io.Writer) error {

	logs, err := client.Logs(serviceID, component, instance, logOptions)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

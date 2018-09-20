package services

import (
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	instanceFlag = "instance"
)

func Logs() *cli.Command {
	var (
		follow     bool
		instance   string
		previous   bool
		sidecar    string
		timestamps bool
		since      string
		tail       int
	)

	cmd := Command{
		Flags: map[string]cli.Flag{
			"follow":                &flags.Bool{Target: &follow},
			instanceFlag:            &flags.String{Target: &instance},
			"previous":              &flags.Bool{Target: &previous},
			command.SidecarFlagName: command.SidecarFlag(&sidecar),
			"timestamps":            &flags.Bool{Target: &timestamps},
			"since":                 &flags.String{Target: &since},
			"tail":                  &flags.Int{Target: &tail},
		},
		Run: func(ctx *ServiceCommandContext, args []string, flags cli.Flags) error {
			var instancePtr *string
			if flags[instanceFlag].Set() {
				instancePtr = &instance
			}

			var sidecarPtr *string
			if flags[command.SidecarFlagName].Set() {
				sidecarPtr = &sidecar
			}

			return ServiceLogs(
				ctx.Client,
				ctx.System,
				ctx.Service,
				instancePtr,
				sidecarPtr,
				follow,
				previous,
				timestamps,
				since,
				int64(tail),
				os.Stdout,
			)
		},
	}

	return cmd.Command()
}

func ServiceLogs(
	client client.Interface,
	system v1.SystemID,
	id v1.ServiceID,
	instance *string,
	sidecar *string,
	follow bool,
	previous bool,
	timestamps bool,
	since string,
	tail int64,
	w io.Writer,
) error {
	options := &v1.ContainerLogOptions{
		Follow:     follow,
		Previous:   previous,
		Timestamps: timestamps,
		Since:      since,
		Tail:       &tail,
	}
	logs, err := client.V1().Systems().Services(system).Logs(id, instance, sidecar, options)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

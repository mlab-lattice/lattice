package jobs

import (
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"time"
)

func Logs() *cli.Command {
	var (
		follow     bool
		previous   bool
		sidecar    string
		timestamps bool
		since      string
		tail       int
	)

	cmd := Command{
		Flags: map[string]cli.Flag{
			"follow":                &flags.Bool{Target: &follow},
			"previous":              &flags.Bool{Target: &previous},
			command.SidecarFlagName: command.SidecarFlag(&sidecar),
			"timestamps":            &flags.Bool{Target: &timestamps},
			"since":                 &flags.String{Target: &since},
			"tail":                  &flags.Int{Target: &tail},
		},
		Run: func(ctx *JobCommandContext, args []string, flags cli.Flags) error {
			var sidecarPtr *string
			if flags[command.SidecarFlagName].Set() {
				sidecarPtr = &sidecar
			}

			return JobLogs(
				ctx.Client,
				ctx.System,
				ctx.Job,
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

func JobLogs(
	client client.Interface,
	system v1.SystemID,
	id v1.JobID,
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
	for {
		logs, err := client.V1().Systems().Jobs(system).Logs(id, sidecar, options)
		if err != nil {
			v1err, ok := err.(*v1.Error)
			if !ok {
				return err
			}

			// INVALID_INSTANCE indicates the job has not run yet, so if that is
			// the error we should wait
			// TODO(kevindrosendahl): ^ is somewhat a lie.
			//                        currently in the k8s backend, it means either there are 0 or multiple
			//                        pods for the job. should refine this in both latticectl and the api
			if v1err.Code != v1.ErrorCodeInvalidInstance {
				return err
			}

			time.Sleep(time.Second)
			continue
		}

		defer logs.Close()
		_, err = io.Copy(w, logs)
		return err
	}
}

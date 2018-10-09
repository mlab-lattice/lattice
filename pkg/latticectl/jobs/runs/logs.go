package runs

import (
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	runcommand "github.com/mlab-lattice/lattice/pkg/latticectl/jobs/runs/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

// Logs returns a *cli.Command to retrieve the logs of a job.
func Logs() *cli.Command {
	var (
		follow     bool
		previous   bool
		sidecar    string
		timestamps bool
		since      string
		tail       int
	)

	cmd := runcommand.JobRunCommand{
		Flags: map[string]cli.Flag{
			"follow":                &flags.Bool{Target: &follow},
			"previous":              &flags.Bool{Target: &previous},
			command.SidecarFlagName: command.SidecarFlag(&sidecar),
			"timestamps":            &flags.Bool{Target: &timestamps},
			"since":                 &flags.String{Target: &since},
			"tail":                  &flags.Int{Target: &tail},
		},
		Run: func(ctx *runcommand.JobRunCommandContext, args []string, flags cli.Flags) error {
			var sidecarPtr *string
			if flags[command.SidecarFlagName].Set() {
				sidecarPtr = &sidecar
			}

			return RunLogs(
				ctx.Client,
				ctx.System,
				ctx.Job,
				ctx.JobRun,
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

// RunLogs prints the logs for the specified job to the supplied writer.
func RunLogs(
	client client.Interface,
	system v1.SystemID,
	job v1.JobID,
	id v1.JobRunID,
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
		logs, err := client.V1().Systems().Jobs(system).Runs(job).Logs(id, sidecar, options)
		if err != nil {
			v1err, ok := err.(*v1.Error)
			if !ok {
				return err
			}

			if v1err.Code != v1.ErrorCodeLogsUnavailable {
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

package jobs

import (
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var sidecarStr string
	var follow bool
	var timestamps bool
	var sinceTime string
	var since string
	var tail int

	cmd := &latticectl.JobCommand{
		Name: "logs",
		Flags: cli.Flags{
			&flags.String{
				Name:   "sidecar",
				Short:  "s",
				Target: &sidecarStr,
			},
			&flags.Bool{
				Name:    "follow",
				Short:   "f",
				Default: false,
				Target:  &follow,
			},
			&flags.Bool{
				Name:    "timestamps",
				Default: false,
				Target:  &timestamps,
			},
			&flags.String{
				Name:     "since-time",
				Required: false,
				Target:   &sinceTime,
			},
			&flags.String{
				Name:     "since",
				Required: false,
				Target:   &since,
			},
			&flags.Int{
				Name:     "tail",
				Required: false,
				Short:    "t",
				Target:   &tail,
			},
		},
		Run: func(ctx latticectl.JobCommandContext, args []string) {
			c := ctx.Client().Systems().Jobs(ctx.SystemID())

			logOptions := &v1.ContainerLogOptions{
				Follow:     follow,
				Timestamps: timestamps,
				SinceTime:  sinceTime,
				Since:      since,
			}

			if tail != 0 {
				tl := int64(tail)
				logOptions.Tail = &tl
			}

			var sidecar *string
			if sidecarStr != "" {
				sidecar = &sidecarStr
			}

			err := GetJobLogs(c, ctx.JobID(), sidecar, logOptions, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetJobLogs(
	client v1client.SystemJobClient,
	jobID v1.JobID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
	w io.Writer,
) error {
	logs, err := client.Logs(jobID, sidecar, logOptions)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

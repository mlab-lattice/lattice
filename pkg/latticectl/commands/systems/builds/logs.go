package builds

import (
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var pathStr string
	var sidecarStr string
	var follow bool
	var previous bool
	var timestamps bool
	var sinceTime string
	var since string
	var tail int

	cmd := &latticectl.BuildCommand{
		Name: "logs",
		Flags: cli.Flags{
			&flags.String{
				Name:     "path",
				Short:    "p",
				Required: true,
				Target:   &pathStr,
			},
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
				Name:    "previous",
				Default: false,
				Target:  &previous,
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
		Run: func(ctx latticectl.BuildCommandContext, args []string) {
			path, err := tree.NewPath(pathStr)
			if err != nil {
				log.Fatal("invalid node path: " + pathStr)
			}

			var sidecar *string
			if sidecarStr != "" {
				sidecar = &sidecarStr
			}

			logOptions := &v1.ContainerLogOptions{
				Follow:     follow,
				Previous:   previous,
				Timestamps: timestamps,
				SinceTime:  sinceTime,
				Since:      since,
			}

			if tail != 0 {
				tl := int64(tail)
				logOptions.Tail = &tl
			}

			c := ctx.Client().Systems().Builds(ctx.SystemID())
			err = GetBuildLogs(c, ctx.BuildID(), path, sidecar, logOptions, os.Stdout)

			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetBuildLogs(
	client v1client.SystemBuildClient,
	buildID v1.BuildID,
	path tree.Path,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
	w io.Writer,
) error {
	logs, err := client.Logs(buildID, path, sidecar, logOptions)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

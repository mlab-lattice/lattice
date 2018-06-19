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
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var pathStr string
	var component string
	var follow bool
	var previous bool
	var timestamps bool
	var sinceTime string
	var since string
	var tail int

	cmd := &latticectl.BuildCommand{
		Name: "logs",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "path",
				Short:    "p",
				Required: true,
				Target:   &pathStr,
			},
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
		Run: func(ctx latticectl.BuildCommandContext, args []string) {

			path, err := tree.NewNodePath(pathStr)
			if err != nil {
				log.Fatal("invalid node path: " + pathStr)
			}

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

			c := ctx.Client().Systems().Builds(ctx.SystemID())
			err = GetBuildLogs(c, ctx.BuildID(), path, component, logOptions, os.Stdout)

			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetBuildLogs(client v1client.BuildClient, buildID v1.BuildID, path tree.NodePath,
	component string, logOptions *v1.ContainerLogOptions, w io.Writer) error {
	logs, err := client.Logs(buildID, path, component, logOptions)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

package builds

import (
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"io"
	"log"
	"os"
)

type LogsCommand struct {
}

func (c *LogsCommand) Base() (*latticectl.BaseCommand, error) {
	var follow bool
	var component string

	cmd := &latticectl.BuildCommand{
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
		Run: func(ctx latticectl.BuildCommandContext, args []string) {
			c := ctx.Client().Systems().Builds(ctx.SystemID())
			err := GetBuildLogs(c, ctx.BuildID(), component, follow, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetBuildLogs(client v1client.BuildClient, buildID v1.BuildID, component string, follow bool, w io.Writer) error {
	logs, err := client.Logs(buildID, component, follow)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}

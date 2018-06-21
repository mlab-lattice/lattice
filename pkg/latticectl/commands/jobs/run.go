package jobs

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type BuildCommand struct {
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListJobsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}
	var pathStr string

	cmd := &latticectl.SystemCommand{
		Name: "run",
		Flags: []cli.Flag{
			output.Flag(),
			&cli.StringFlag{
				Name:     "path",
				Required: true,
				Target:   &pathStr,
			},
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			path, err := tree.NewNodePath(pathStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid path: %v", err)
				os.Exit(1)
			}

			c := ctx.Client().Systems().Jobs(ctx.SystemID())

			err = RunJob(c, path, format, os.Stdout, watch)
			if err != nil {
				//log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func RunJob(
	client v1client.JobClient,
	path tree.NodePath,
	format printer.Format,
	writer io.Writer,
	watch bool,
) error {
	job, err := client.Create(path)
	if err != nil {
		return err
	}

	//if watch {
	//	if format == printer.FormatTable {
	//		fmt.Fprintf(writer, "\nBuild ID: %s\n", color.ID(string(job.ID)))
	//	}
	//	return builds.WatchBuild(client, job.ID, format, os.Stdout, builds.PrintBuildStateDuringWatchBuild)
	//}

	fmt.Fprintf(writer, "Running job %s, ID: %s\n\n", path.String(), color.ID(string(job.ID)))
	fmt.Fprintf(writer, "To view the status of the job, run:\n\n    latticectl jobs:status --job %s [--watch]\n", color.ID(string(job.ID)))
	return nil
}

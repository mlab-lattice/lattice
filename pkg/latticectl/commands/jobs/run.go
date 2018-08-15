package jobs

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"strings"
	"time"
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
	var follow bool
	var command string
	var jobArgs []string
	var envStrings []string

	cmd := &latticectl.SystemCommand{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "path",
				Required: true,
				Target:   &pathStr,
			},
			&cli.BoolFlag{
				Name:   "follow",
				Short:  "f",
				Target: &follow,
			},
			&cli.StringFlag{
				Name:   "command",
				Short:  "c",
				Target: &command,
			},
			&cli.StringSliceFlag{
				Name:   "arg",
				Short:  "a",
				Target: &jobArgs,
			},
			&cli.StringSliceFlag{
				Name:   "env",
				Short:  "e",
				Target: &envStrings,
			},
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			path, err := tree.NewPath(pathStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid path: %v", err)
				os.Exit(1)
			}

			fullCommand := []string{command}
			fullCommand = append(fullCommand, jobArgs...)

			env := definitionv1.ContainerEnvironment{}
			for _, kv := range envStrings {
				// FIXME: support comma escaping
				parts := strings.Split(kv, "=")
				if len(parts) != 2 {
					fmt.Fprintf(os.Stderr, "invalid environment variable %v", kv)
					os.Exit(1)
				}

				// FIXME: support secrets
				env[parts[0]] = definitionv1.ValueOrSecret{
					Value: &parts[1],
				}
			}

			c := ctx.Client().Systems().Jobs(ctx.SystemID())

			err = RunJob(c, path, fullCommand, env, format, os.Stdout, watch, follow)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func RunJob(
	client v1client.JobClient,
	path tree.Path,
	command []string,
	environment definitionv1.ContainerEnvironment,
	format printer.Format,
	writer io.Writer,
	watch bool,
	follow bool,
) error {
	job, err := client.Create(path, command, environment)
	if err != nil {
		return err
	}

	if follow {
		// need to wait until job is at least queued
		// FIXME: do this better
		time.Sleep(2 * time.Second)

		logOptions := &v1.ContainerLogOptions{Follow: true}
		return GetJobLogs(client, job.ID, nil, logOptions, writer)
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

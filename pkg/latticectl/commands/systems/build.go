package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/builds"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type BuildCommand struct {
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}
	var version string

	cmd := &latticectl.SystemCommand{
		Name: "build",
		Flags: []cli.Flag{
			output.Flag(),
			&cli.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Builds(ctx.SystemID())

			err = BuildSystem(c, v1.SystemVersion(version), format, os.Stdout, watch)
			if err != nil {
				//log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func BuildSystem(
	client v1client.SystemBuildClient,
	version v1.SystemVersion,
	format printer.Format,
	writer io.Writer,
	watch bool,
) error {
	build, err := client.Create(version)
	if err != nil {
		return err
	}

	if watch {
		if format == printer.FormatTable {
			fmt.Fprintf(writer, "\nBuild ID: %s\n", color.ID(string(build.ID)))
		}
		return builds.WatchBuild(client, build.ID, format, os.Stdout, builds.PrintBuildStateDuringWatchBuild)
	}

	fmt.Fprintf(writer, "Building version %s, Build ID: %s\n\n", version, color.ID(string(build.ID)))
	fmt.Fprintf(writer, "To view the status of the build, run:\n\n    latticectl system:builds:status --build %s [--watch]\n", color.ID(string(build.ID)))
	return nil
}

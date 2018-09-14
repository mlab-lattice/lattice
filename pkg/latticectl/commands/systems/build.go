package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/builds"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
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
	var path tree.Path

	cmd := &latticectl.SystemCommand{
		Name: "build",
		Flags: []cli.Flag{
			output.Flag(),
			&flags.Path{
				Name:    "path",
				Default: tree.RootPath(),
				Target:  &path,
			},
			&flags.String{
				Name:   "version",
				Target: &version,
			},
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Builds(ctx.SystemID())

			version := v1.Version(version)
			var v *v1.Version
			if version != "" {
				v = &version
			}

			err = BuildSystem(c, v, &path, format, os.Stdout, watch)
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
	version *v1.Version,
	path *tree.Path,
	format printer.Format,
	writer io.Writer,
	watch bool,
) error {
	var build *v1.Build
	var err error
	if version != nil {
		build, err = client.CreateFromVersion(*version)
	} else {
		build, err = client.CreateFromPath(*path)
	}
	if err != nil {
		return err
	}

	if watch {
		if format == printer.FormatTable {
			fmt.Fprintf(writer, "\nBuild ID: %s\n", color.ID(string(build.ID)))
		}
		return builds.WatchBuild(client, build.ID, format, os.Stdout, builds.PrintBuildStateDuringWatchBuild)
	}

	fmt.Fprintf(writer, "Build ID: %s\n\n", color.ID(string(build.ID)))
	fmt.Fprintf(writer, "To view the status of the build, run:\n\n    latticectl system:builds:status --build %s [--watch]\n", color.ID(string(build.ID)))
	return nil
}

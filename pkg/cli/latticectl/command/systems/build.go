package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/builds"
)

type BuildCommand struct {
}

//type PrintBuildState func(io.Writer, *spinner.Spinner, *types.SystemBuild, string)

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	var version string
	
	cmd := &lctlcommand.SystemCommand{
		Name: "build",
		Flags: []command.Flag{
			output.Flag(),
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}
			
			c := ctx.Client().Systems().SystemBuilds(ctx.SystemID())
			
			err = BuildSystem(c, version, format, os.Stdout, watch)
			if err != nil {
				//log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func BuildSystem(
	client client.SystemBuildClient,
	version string,
	format printer.Format,
	writer io.Writer,
	watch bool,
	) error {
	buildID, err := client.Create(version)
	if err != nil {
		return err
	}
	
	if watch {
		if format == printer.FormatDefault || format == printer.FormatTable {
			fmt.Fprintf(writer, "\nBuild ID: %s\n", color.ID(string(buildID)))
		}
		return builds.WatchBuild(client, buildID, format, os.Stdout, builds.PrintBuildStateDuringWatchBuild)
	} else {
		fmt.Fprintf(writer, "Building version %s, Build ID: %s\n\n", version, color.ID(string(buildID)))
		fmt.Fprintf(writer, "To view the status of the build, run:\n\n    latticectl system:builds:status --build %s [--watch]\n", color.ID(string(buildID)))
	}
	return nil
}

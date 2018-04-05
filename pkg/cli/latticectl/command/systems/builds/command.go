package builds

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	tw "github.com/tfogo/tablewriter"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ListBuildsSupportedFormats is the list of printer.Formats supported
// by the ListBuilds function.
var ListBuildsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

// ListBuildsCommand is a type that implements the latticectl.Command interface
// for listing the Builds in a System.
type ListBuildsCommand struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *ListBuildsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListBuildsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "builds",
		Flags: command.Flags{
			output.Flag(),
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

			if watch {
				WatchBuilds(c, format, os.Stdout)
				return
			}

			err = ListBuilds(c, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

// ListBuilds writes the current Builds to the supplied io.Writer in the given printer.Format.
func ListBuilds(client client.SystemBuildClient, format printer.Format, writer io.Writer) error {
	builds, err := client.List()
	if err != nil {
		return err
	}

	p := buildsPrinter(builds, format)
	p.Print(writer)
	return nil
}

// WatchBuilds polls the API for the current Builds, and writes out the Builds to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchBuilds(client client.SystemBuildClient, format printer.Format, writer io.Writer) {
	// Poll the API for the builds and send it to the channel
	buildLists := make(chan []types.SystemBuild)
	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			buildList, err := client.List()
			if err != nil {
				return false, err
			}

			buildLists <- buildList
			return false, nil
		},
	)

	for buildList := range buildLists {
		p := buildsPrinter(buildList, format)
		lastHeight = p.Overwrite(b, lastHeight)

		// Note: Watching builds is never exitable.
		// There is no fail state for an entire list of builds.
	}
}

func buildsPrinter(builds []types.SystemBuild, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"ID", "Version", "State"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, build := range builds {
			var stateColor color.Color
			switch build.State {
			case types.SystemBuildStateSucceeded:
				stateColor = color.Success
			case types.SystemBuildStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(build.ID),
				string(build.Version),
				stateColor(string(build.State)),
			})
		}

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: builds,
		}
	}

	return p
}

package teardowns

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

// ListTeardownsSupportedFormats is the list of printer.Formats supported
// by the ListTeardowns function.
var ListTeardownsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

// ListTeardownsCommand is a type that implements the latticectl.Command interface
// for listing the Teardowns in a System.
type ListTeardownsCommand struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *ListTeardownsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListTeardownsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "teardowns",
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

			c := ctx.Client().Systems().Teardowns(ctx.SystemID())

			if watch {
				WatchTeardowns(c, format, os.Stdout)
				return
			}

			err = ListTeardowns(c, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

// ListTeardowns writes the current Teardowns to the supplied io.Writer in the given printer.Format.
func ListTeardowns(client client.TeardownClient, format printer.Format, writer io.Writer) error {
	teardowns, err := client.List()
	if err != nil {
		return err
	}

	p := teardownsPrinter(teardowns, format)
	p.Print(writer)
	return nil
}

// WatchTeardowns polls the API for the current Teardowns, and writes out the Teardowns to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchTeardowns(client client.TeardownClient, format printer.Format, writer io.Writer) {
	// Poll the API for the teardowns and send it to the channel
	teardownLists := make(chan []types.SystemTeardown)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			teardownList, err := client.List()
			if err != nil {
				return false, err
			}

			teardownLists <- teardownList
			return false, nil
		},
	)

	for teardownList := range teardownLists {
		p := teardownsPrinter(teardownList, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}

func teardownsPrinter(teardowns []types.SystemTeardown, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"ID", "State"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, teardown := range teardowns {
			var stateColor color.Color
			switch teardown.State {
			case types.SystemTeardownStateSucceeded:
				stateColor = color.Success
			case types.SystemTeardownStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(teardown.ID),
				stateColor(string(teardown.State)),
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
			Value: teardowns,
		}
	}

	return p
}

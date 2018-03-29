package teardowns

import (
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
	printerChan := make(chan printer.Interface)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			teardowns, err := client.List()
			if err != nil {
				return false, err
			}

			p := teardownsPrinter(teardowns, format)
			printerChan <- p
			return false, nil
		},
	)

	// If displaying a table, use the overwritting terminal watcher, if JSON
	// use the scrolling watcher
	var w printer.Watcher
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		w = &printer.OverwrittingTerminalWatcher{}

	case printer.FormatJSON:
		w = &printer.ScrollingWatcher{}
	}

	w.Watch(printerChan, writer)
}

func teardownsPrinter(teardowns []types.SystemTeardown, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"ID", "State"}

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
			Headers: headers,
			Rows:    rows,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: teardowns,
		}
	}

	return p
}

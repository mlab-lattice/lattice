package systems

import (
	"io"
	"log"
	"os"
	"time"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"

	"k8s.io/apimachinery/pkg/util/wait"
)

// ListSystemsSupportedFormats is the list of printer.Formats supported
// by the ListSystems function.
var ListSystemsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

// ListSystemsCommand is a type that implements the latticectl.Command interface
// for listing the Systems in a Lattice.
type ListSystemsCommand struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *ListSystemsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &command.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool

	cmd := &command.LatticeCommand{
		Name: "systems",
		Flags: cli.Flags{
			output.Flag(),
			&cli.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx command.LatticeCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				WatchSystems(c, format, os.Stdout)
				return
			}

			ListSystems(c, format, os.Stdout)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

// ListSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func ListSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) {
	systems, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	p := systemsPrinter(systems, format)
	p.Print(writer)
}

// WatchSystems polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) {
	// Poll the API for the systems and send it to the channel
	printerChan := make(chan printer.Interface)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			systems, err := client.List()
			if err != nil {
				return false, err
			}

			p := systemsPrinter(systems, format)
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

func systemsPrinter(systems []v1.System, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Name", "Definition", "Status"}

		var rows [][]string
		for _, system := range systems {
			var stateColor color.Color
			switch system.State {
			case v1.SystemStateStable:
				stateColor = color.Success
			case v1.SystemStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(system.ID),
				system.DefinitionURL,
				stateColor(string(system.State)),
			})
		}

		p = &printer.Table{
			Headers: headers,
			Rows:    rows,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: systems,
		}
	}

	return p
}

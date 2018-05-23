package systems

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

// ListSystemsSupportedFormats is the list of printer.Formats supported
// by the ListSystems function.
var ListSystemsSupportedFormats = []printer.Format{
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
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.LatticeCommand{
		Name: "systems",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				err = WatchSystems(c, format, os.Stdout)
			} else {
				err = ListSystems(c, format, os.Stdout)
			}
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

// ListSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func ListSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) error {
	systems, err := client.List()
	if err != nil {
		return err
	}

	p := systemsPrinter(systems, format)
	if err := p.Print(writer); err != nil {
		return err
	}

	return nil
}

// WatchSystems polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) error {
	// Poll the API for the systems and send it to the channel
	systemLists := make(chan []v1.System)
	lastHeight := 0
	var b bytes.Buffer
	var err error

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			systemList, err := client.List()
			if err != nil {
				return false, err
			}

			systemLists <- systemList
			return false, nil
		},
	)

	for systemList := range systemLists {
		p := systemsPrinter(systemList, format)
		err, lastHeight = p.Overwrite(b, lastHeight)
		if err != nil {
			return err
		}
		// Note: Watching systems is never exitable.
		// There is no fail state for an entire lattice of systems.
	}

	return nil
}

func systemsPrinter(systems []v1.System, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		var rows [][]string
		headers := []string{"Name", "Definition", "Status"}

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
				color.ID(string(system.ID)),
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

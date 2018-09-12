package systems

import (
	"fmt"
	"io"
	"os"
	"time"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

// SystemsSupportedFormats is the list of printer.Formats supported
// by the ListSystems function.
var SystemsSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

func Command() *cli.Command {
	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			"output": command.OutputFlag(SystemsSupportedFormats, printer.FormatTable),
			"watch":  command.WatchFlag(),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) {
			format := printer.Format(flags["watch"].Value().(string))

			if flags["watch"].Value().(bool) {
				WatchSystems(ctx.Client.V1().Systems(), format, os.Stdout)
				return
			}

			err := ListSystems(ctx.Client.V1().Systems(), format, os.Stdout)
			if err != nil {
				panic(err)
			}
		},
		Subcommands: map[string]*cli.Command{
			"status": Status(),
		},
	}

	return cmd.Command()
}

// WatchSystems polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) {
	// Poll the API for the systems and send it to the channel
	systems := make(chan []v1.System)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			s, err := client.List()
			if err != nil {
				return false, err
			}

			systems <- s
			return false, nil
		},
	)

	p := systemsPrinter(s, format)

	for s := range systems {
		p.Stream(writer)

		// Note: Watching systems is never exitable.
		// There is no fail state for an entire lattice of systems.
	}
}

// ListSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func ListSystems(client clientv1.SystemClient, format printer.Format, writer io.Writer) error {
	systems, err := client.List()
	if err != nil {
		return err
	}

	p := systemsPrinter(systems, format)
	p.Print(writer)

	return nil
}

func systemsPrinter(systems []v1.System, format printer.Format) printer.Interface {
	switch format {
	case printer.FormatTable:
		t := &printer.Table{
			Columns: []printer.TableColumn{
				{
					Header:    "Name",
					Color:     color.ID,
					Alignment: printer.TableAlignLeft,
				},
				{
					Header:    "Definition",
					Alignment: printer.TableAlignLeft,
				},
				{
					Header:    "Status",
					Alignment: printer.TableAlignLeft,
				},
			},
		}

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

			t.AppendRow([]string{
				string(system.ID),
				system.DefinitionURL,
				stateColor(string(system.State)),
			})
		}

		return t

	case printer.FormatJSON:
		return &printer.JSON{
			Value: systems,
		}
	}

	panic(fmt.Sprintf("unexpected format %v", format))
}

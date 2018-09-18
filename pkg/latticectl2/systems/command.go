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
	"sort"
)

// ListSupportedFormats is the list of printer.Formats supported
// by the ListSystems function.
var ListSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

func Command() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(&output, ListSupportedFormats, printer.FormatTable),
			command.WatchFlagName:  command.WatchFlag(&watch),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchSystems(ctx.Client.V1().Systems(), format, os.Stdout)
				return nil
			}

			return ListSystems(ctx.Client.V1().Systems(), format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"status": Status(),
		},
	}

	return cmd.Command()
}

// ListSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func ListSystems(client clientv1.SystemClient, format printer.Format, w io.Writer) error {
	systems, err := client.List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := systemsTable(w)
		r := systemsTableRows(systems)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(systems)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchSystems polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchSystems(client clientv1.SystemClient, format printer.Format, w io.Writer) {
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

	var handle func([]v1.System)
	switch format {
	case printer.FormatTable:
		t := systemsTable(w)
		handle = func(systems []v1.System) {
			r := systemsTableRows(systems)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(systems []v1.System) {
			j.Print(systems)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", format))
	}

	for s := range systems {
		handle(s)
	}
}

func systemsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []printer.TableColumn{
		{
			Header:    "name",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "definition",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "status",
			Alignment: printer.TableAlignLeft,
		},
	})
}

func systemsTableRows(systems []v1.System) []printer.TableRow {
	var rows []printer.TableRow
	for _, system := range systems {
		var stateColor color.Formatter
		switch system.State {
		case v1.SystemStateStable:
			stateColor = color.SuccessString
		case v1.SystemStateFailed:
			stateColor = color.FailureString
		default:
			stateColor = color.WarningString
		}

		rows = append(rows, []string{
			color.IDString(string(system.ID)),
			system.DefinitionURL,
			stateColor(string(system.State)),
		})
	}

	// sort the rows by system ID
	sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })

	return rows
}

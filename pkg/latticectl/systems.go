package latticectl

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"github.com/mlab-lattice/lattice/pkg/latticectl/systems"
	"k8s.io/apimachinery/pkg/util/wait"
	"sort"
)

func Systems() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchSystems(ctx.Client, format, os.Stdout)
				return nil
			}

			return PrintSystems(ctx.Client, format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"create":   systems.Create(),
			"delete":   systems.Delete(),
			"status":   systems.Status(),
			"versions": systems.Versions(),
		},
	}

	return cmd.Command()
}

// PrintSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func PrintSystems(client client.Interface, format printer.Format, w io.Writer) error {
	systems, err := client.V1().Systems().List()
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
func WatchSystems(client client.Interface, format printer.Format, w io.Writer) {
	// Poll the API for the systems and send it to the channel
	systems := make(chan []v1.System)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			s, err := client.V1().Systems().List()
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
			Header:    "version",
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
		switch system.Status.State {
		case v1.SystemStateStable:
			stateColor = color.SuccessString
		case v1.SystemStateFailed:
			stateColor = color.FailureString
		default:
			stateColor = color.WarningString
		}

		version := "-"
		if system.Status.Version != nil {
			version = string(*system.Status.Version)
		}

		rows = append(rows, []string{
			color.IDString(string(system.ID)),
			system.DefinitionURL,
			version,
			stateColor(string(system.Status.State)),
		})
	}

	// sort the rows by id
	idIdx := 0
	sort.Slice(rows, func(i, j int) bool { return rows[i][idIdx] < rows[j][idIdx] })

	return rows
}

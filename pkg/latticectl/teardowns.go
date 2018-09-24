package latticectl

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/teardowns"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

func Teardowns() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := command.SystemCommand{
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
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchTeardowns(ctx.Client, ctx.System, format, os.Stdout)
				return nil
			}

			return PrintTeardowns(ctx.Client, ctx.System, format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"status": teardowns.Status(),
		},
	}

	return cmd.Command()
}

// PrintTeardowns writes the current Systems to the supplied io.Writer in the given printer.Format.
func PrintTeardowns(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) error {
	teardowns, err := client.V1().Systems().Teardowns(system).List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := teardownsTable(w)
		r := teardownsTableRows(teardowns)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(teardowns)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchTeardowns polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchTeardowns(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) {
	// Poll the API for the systems and send it to the channel
	teardowns := make(chan []v1.Teardown)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			d, err := client.V1().Systems().Teardowns(system).List()
			if err != nil {
				return false, err
			}

			teardowns <- d
			return false, nil
		},
	)

	var handle func([]v1.Teardown)
	switch format {
	case printer.FormatTable:
		t := teardownsTable(w)
		handle = func(teardowns []v1.Teardown) {
			r := teardownsTableRows(teardowns)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(teardowns []v1.Teardown) {
			j.Print(teardowns)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", format))
	}

	for d := range teardowns {
		handle(d)
	}
}

func teardownsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []printer.TableColumn{
		{
			Header:    "id",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "state",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "message",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "started",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "completed",
			Alignment: printer.TableAlignLeft,
		},
	})
}

func teardownsTableRows(teardowns []v1.Teardown) []printer.TableRow {
	var rows []printer.TableRow
	for _, teardown := range teardowns {
		stateColor := color.WarningString
		switch teardown.Status.State {
		case v1.TeardownStateSucceeded:
			stateColor = color.SuccessString

		case v1.TeardownStateFailed:
			stateColor = color.FailureString
		}

		message := "-"
		if teardown.Status.Message != "" {
			message = teardown.Status.Message
		}

		started := "-"
		if teardown.Status.StartTimestamp != nil {
			started = teardown.Status.StartTimestamp.Format(time.RFC1123)
		}

		completed := "-"
		if teardown.Status.StartTimestamp != nil {
			completed = teardown.Status.StartTimestamp.Format(time.RFC1123)
		}

		rows = append(rows, []string{
			color.IDString(string(teardown.ID)),
			stateColor(string(teardown.Status.State)),
			message,
			started,
			completed,
		})
	}

	// sort the rows by start timestamp
	startedIdx := 3
	sort.Slice(
		rows,
		func(i, j int) bool {
			ts1, ts2 := rows[i][startedIdx], rows[j][startedIdx]
			if ts1 == "-" {
				return true
			}

			if ts2 == "-" {
				return false
			}

			t1, err := time.Parse(time.RFC1123, ts1)
			if err != nil {
				panic(err)
			}

			t2, err := time.Parse(time.RFC1123, ts2)
			if err != nil {
				panic(err)
			}
			return t1.After(t2)
		},
	)

	return rows
}

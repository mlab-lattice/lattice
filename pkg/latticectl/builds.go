package latticectl

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/builds"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

func Builds() *cli.Command {
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
				WatchBuilds(ctx.Client, ctx.System, format, os.Stdout)
				return nil
			}

			return PrintBuilds(ctx.Client, ctx.System, format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"logs":   builds.Logs(),
			"status": builds.Status(),
		},
	}

	return cmd.Command()
}

// PrintBuilds writes the current Systems to the supplied io.Writer in the given printer.Format.
func PrintBuilds(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) error {
	builds, err := client.V1().Systems().Builds(system).List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := buildsTable(w)
		r := buildsTableRows(builds)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(builds)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchBuilds polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchBuilds(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) {
	// Poll the API for the systems and send it to the channel
	builds := make(chan []v1.Build)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			d, err := client.V1().Systems().Builds(system).List()
			if err != nil {
				return false, err
			}

			builds <- d
			return false, nil
		},
	)

	var handle func([]v1.Build)
	switch format {
	case printer.FormatTable:
		t := buildsTable(w)
		handle = func(builds []v1.Build) {
			r := buildsTableRows(builds)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(builds []v1.Build) {
			j.Print(builds)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", format))
	}

	for d := range builds {
		handle(d)
	}
}

func buildsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []printer.TableColumn{
		{
			Header:    "id",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "target",
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

func buildsTableRows(builds []v1.Build) []printer.TableRow {
	var rows []printer.TableRow
	for _, build := range builds {
		stateColor := color.WarningString
		switch build.Status.State {
		case v1.BuildStateSucceeded:
			stateColor = color.SuccessString

		case v1.BuildStateFailed:
			stateColor = color.FailureString
		}

		target := "-"
		switch {
		case build.Path != nil:
			target = fmt.Sprintf("path %v", build.Path.String())

		case build.Version != nil:
			target = fmt.Sprintf("version %v", *build.Version)
		}

		message := "-"
		if build.Status.Message != "" {
			message = build.Status.Message
		}

		started := "-"
		if build.Status.StartTimestamp != nil {
			started = build.Status.StartTimestamp.Format(time.RFC1123)
		}

		completed := "-"
		if build.Status.StartTimestamp != nil {
			completed = build.Status.StartTimestamp.Format(time.RFC1123)
		}

		rows = append(rows, []string{
			color.IDString(string(build.ID)),
			target,
			stateColor(string(build.Status.State)),
			message,
			started,
			completed,
		})
	}

	// sort the rows by start timestamp
	startedIdx := 4
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

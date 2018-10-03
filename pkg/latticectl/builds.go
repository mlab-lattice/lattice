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

// Builds returns a *cli.Command to list a system's builds with subcommands to interact
// with individual builds.
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
				return WatchBuilds(ctx.Client, ctx.System, format)
			}

			return PrintBuilds(ctx.Client, ctx.System, os.Stdout, format)
		},
		Subcommands: map[string]*cli.Command{
			"logs":   builds.Logs(),
			"status": builds.Status(),
		},
	}

	return cmd.Command()
}

// PrintBuilds prints the system's builds to the supplied writer.
func PrintBuilds(client client.Interface, system v1.SystemID, w io.Writer, format printer.Format) error {
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

// WatchBuilds watches the system's builds, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchBuilds(client client.Interface, system v1.SystemID, format printer.Format) error {
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
		t := buildsTable(os.Stdout)
		handle = func(builds []v1.Build) {
			r := buildsTableRows(builds)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(builds []v1.Build) {
			j.Print(builds)
		}

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	for d := range builds {
		handle(d)
	}
	return nil
}

func buildsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []string{"ID", "TARGET", "STATE", "STARTED", "COMPLETED"})
}

func buildsTableRows(builds []v1.Build) [][]string {
	var rows [][]string
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

		started := "-"
		if build.Status.StartTimestamp != nil {
			started = build.Status.StartTimestamp.Local().Format(time.RFC1123)
		}

		completed := "-"
		if build.Status.CompletionTimestamp != nil {
			completed = build.Status.CompletionTimestamp.Local().Format(time.RFC1123)
		}

		rows = append(rows, []string{
			color.IDString(string(build.ID)),
			target,
			stateColor(string(build.Status.State)),
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

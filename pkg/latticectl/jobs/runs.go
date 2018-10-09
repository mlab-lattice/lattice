package jobs

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	jobcommand "github.com/mlab-lattice/lattice/pkg/latticectl/jobs/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/jobs/runs"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Runs returns a *cli.Command to list a job's runs with subcommands to interact
// with individual runs.
func Runs() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := jobcommand.JobCommand{
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
		Run: func(ctx *jobcommand.JobCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				return WatchRuns(ctx.Client, ctx.System, ctx.Job, format)
			}

			return PrintRuns(ctx.Client, ctx.System, ctx.Job, os.Stdout, format)
		},
		Subcommands: map[string]*cli.Command{
			"logs":   runs.Logs(),
			"status": runs.Status(),
		},
	}

	return cmd.Command()
}

// PrintRuns prints the job's runs to the supplied writer.
func PrintRuns(client client.Interface, system v1.SystemID, job v1.JobID, w io.Writer, format printer.Format) error {
	runs, err := client.V1().Systems().Jobs(system).Runs(job).List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := jobRunsTable(w)
		r := jobRunsTableRows(runs)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(runs)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchRuns watches the job's runs, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchRuns(client client.Interface, system v1.SystemID, job v1.JobID, format printer.Format) error {
	// Poll the API for the systems and send it to the channel
	runs := make(chan []v1.JobRun)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			d, err := client.V1().Systems().Jobs(system).Runs(job).List()
			if err != nil {
				return false, err
			}

			runs <- d
			return false, nil
		},
	)

	var handle func([]v1.JobRun)
	switch format {
	case printer.FormatTable:
		t := jobRunsTable(os.Stdout)
		handle = func(runs []v1.JobRun) {
			r := jobRunsTableRows(runs)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(runs []v1.JobRun) {
			j.Print(runs)
		}

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	for d := range runs {
		handle(d)
	}

	return nil
}

func jobRunsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []string{"ID", "STATE", "STARTED", "COMPLETED"})
}

func jobRunsTableRows(runs []v1.JobRun) [][]string {
	var rows [][]string
	for _, job := range runs {
		stateColor := color.WarningString
		switch job.Status.State {
		case v1.JobRunStateSucceeded:
			stateColor = color.SuccessString

		case v1.JobRunStateFailed:
			stateColor = color.FailureString
		}

		started := "-"
		if job.Status.StartTimestamp != nil {
			started = job.Status.StartTimestamp.Local().Format(time.RFC1123)
		}

		completed := "-"
		if job.Status.CompletionTimestamp != nil {
			completed = job.Status.CompletionTimestamp.Local().Format(time.RFC1123)
		}

		rows = append(rows, []string{
			color.IDString(string(job.ID)),
			stateColor(string(job.Status.State)),
			started,
			completed,
		})
	}

	// sort the rows by start timestamp
	startedIdx := 2
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

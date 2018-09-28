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
	"github.com/mlab-lattice/lattice/pkg/latticectl/jobs"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

func Jobs() *cli.Command {
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
				WatchJobs(ctx.Client, ctx.System, format, os.Stdout)
				return nil
			}

			return PrintJobs(ctx.Client, ctx.System, format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"logs":   jobs.Logs(),
			"run":    jobs.Run(),
			"status": jobs.Status(),
		},
	}

	return cmd.Command()
}

// PrintJobs writes the current Systems to the supplied io.Writer in the given printer.Format.
func PrintJobs(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) error {
	jobs, err := client.V1().Systems().Jobs(system).List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := jobsTable(w)
		r := jobsTableRows(jobs)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(jobs)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchJobs polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchJobs(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) {
	// Poll the API for the systems and send it to the channel
	jobs := make(chan []v1.Job)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			d, err := client.V1().Systems().Jobs(system).List()
			if err != nil {
				return false, err
			}

			jobs <- d
			return false, nil
		},
	)

	var handle func([]v1.Job)
	switch format {
	case printer.FormatTable:
		t := jobsTable(w)
		handle = func(jobs []v1.Job) {
			r := jobsTableRows(jobs)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(jobs []v1.Job) {
			j.Print(jobs)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", format))
	}

	for d := range jobs {
		handle(d)
	}
}

func jobsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []string{"ID", "PATH", "STATE", "STARTED", "COMPLETED"})
}

func jobsTableRows(jobs []v1.Job) [][]string {
	var rows [][]string
	for _, job := range jobs {
		stateColor := color.WarningString
		switch job.Status.State {
		case v1.JobStateSucceeded:
			stateColor = color.SuccessString

		case v1.JobStateFailed:
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
			job.Path.String(),
			stateColor(string(job.Status.State)),
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

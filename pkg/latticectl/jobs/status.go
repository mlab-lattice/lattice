package jobs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	jobcommand "github.com/mlab-lattice/lattice/pkg/latticectl/jobs/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

// Status returns a *cli.Command to retrieve the status of a job.
func Status() *cli.Command {
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
				return WatchJobStatus(ctx.Client, ctx.System, ctx.Job, format)
			}

			return PrintJobStatus(ctx.Client, ctx.System, ctx.Job, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// PrintJobStatus prints the specified job's status to the supplied writer.
func PrintJobStatus(client client.Interface, system v1.SystemID, id v1.JobID, w io.Writer, f printer.Format) error {
	job, err := client.V1().Systems().Jobs(system).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := jobWriter(w)
		s := jobString(job)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		j.Print(job)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

// WatchJobStatus watches the specified job, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchJobStatus(client client.Interface, system v1.SystemID, id v1.JobID, f printer.Format) error {
	var handle func(*v1.Job) bool
	switch f {
	case printer.FormatTable:
		dw := jobWriter(os.Stdout)

		handle = func(job *v1.Job) bool {
			s := jobString(job)
			dw.Overwrite(s)

			switch job.Status.State {
			case v1.JobStateFailed:
				fmt.Print(color.BoldHiSuccessString("✘ job failed\n"))
				return true

			case v1.JobStateSucceeded:
				fmt.Print(color.BoldHiSuccessString("✓ job succeeded\n"))
				return true
			}

			return false
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(job *v1.Job) bool {
			j.Print(job)
			return false
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		job, err := client.V1().Systems().Jobs(system).Get(id)
		if err != nil {
			return err
		}

		done := handle(job)
		if done {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func jobWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func jobString(job *v1.Job) string {
	stateColor := color.BoldString
	switch job.Status.State {
	case v1.JobStatePending, v1.JobStateQueued, v1.JobStateRunning:
		stateColor = color.BoldHiWarningString

	case v1.JobStateSucceeded:
		stateColor = color.BoldHiSuccessString

	case v1.JobStateFailed, v1.JobStateDeleting:
		stateColor = color.BoldHiFailureString
	}

	return fmt.Sprintf(`job %s (%s)
  state: %s
`,
		color.IDString(string(job.ID)),
		job.Path.String(),
		stateColor(string(job.Status.State)),
	)
}

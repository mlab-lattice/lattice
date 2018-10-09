package runs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	runcommand "github.com/mlab-lattice/lattice/pkg/latticectl/jobs/runs/command"
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

	cmd := runcommand.JobRunCommand{
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
		Run: func(ctx *runcommand.JobRunCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				return WatchRunStatus(ctx.Client, ctx.System, ctx.Job, ctx.JobRun, format)
			}

			return PrintRunStatus(ctx.Client, ctx.System, ctx.Job, ctx.JobRun, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// PrintRunStatus prints the specified job's status to the supplied writer.
func PrintRunStatus(
	client client.Interface,
	system v1.SystemID,
	job v1.JobID,
	id v1.JobRunID,
	w io.Writer,
	f printer.Format,
) error {
	run, err := client.V1().Systems().Jobs(system).Runs(job).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := runWriter(w)
		s := runString(run)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(run)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

// WatchRunStatus watches the specified job, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchRunStatus(client client.Interface, system v1.SystemID, job v1.JobID, id v1.JobRunID, f printer.Format) error {
	var handle func(*v1.JobRun) bool
	switch f {
	case printer.FormatTable:
		dw := runWriter(os.Stdout)

		handle = func(run *v1.JobRun) bool {
			s := runString(run)
			dw.Overwrite(s)

			switch run.Status.State {
			case v1.JobRunStateFailed:
				fmt.Print(color.BoldHiSuccessString("✘ run failed\n"))
				return true

			case v1.JobRunStateSucceeded:
				fmt.Print(color.BoldHiSuccessString("✓ run succeeded\n"))
				return true
			}

			return false
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(run *v1.JobRun) bool {
			j.Print(run)
			return false
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		run, err := client.V1().Systems().Jobs(system).Runs(job).Get(id)
		if err != nil {
			return err
		}

		done := handle(run)
		if done {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func runWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func runString(run *v1.JobRun) string {
	stateColor := color.BoldString
	switch run.Status.State {
	case v1.JobRunStatePending, v1.JobRunStateRunning, v1.JobRunStateUnknown:
		stateColor = color.BoldHiWarningString

	case v1.JobRunStateSucceeded:
		stateColor = color.BoldHiSuccessString

	case v1.JobRunStateFailed:
		stateColor = color.BoldHiFailureString
	}

	exitCode := ""
	if run.Status.ExitCode != nil {
		exitCode = fmt.Sprintf(`
  exit: %v`,
			*run.Status.ExitCode,
		)
	}

	return fmt.Sprintf(`job %s
  state: %s%s
`,
		color.IDString(string(run.ID)),
		stateColor(string(run.Status.State)),
		exitCode,
	)
}

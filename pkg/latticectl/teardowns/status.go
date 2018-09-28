package teardowns

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
)

func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := Command{
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
		Run: func(ctx *TeardownCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchTeardown(ctx.Client, ctx.System, ctx.Teardown, os.Stdout, format)
				return nil
			}

			return PrintTeardown(ctx.Client, ctx.System, ctx.Teardown, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func PrintTeardown(client client.Interface, system v1.SystemID, id v1.TeardownID, w io.Writer, f printer.Format) error {
	teardown, err := client.V1().Systems().Teardowns(system).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := teardownWriter(w)
		s := teardownString(teardown)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(teardown)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

func WatchTeardown(client client.Interface, system v1.SystemID, id v1.TeardownID, w io.Writer, f printer.Format) error {
	var handle func(*v1.Teardown) bool
	switch f {
	case printer.FormatTable:
		dw := teardownWriter(w)

		handle = func(teardown *v1.Teardown) bool {
			s := teardownString(teardown)
			dw.Overwrite(s)

			switch teardown.Status.State {
			case v1.TeardownStateFailed:
				fmt.Fprint(w, color.BoldHiFailureString("✘ teardown failed\n"))
				return true

			case v1.TeardownStateSucceeded:
				fmt.Fprint(w, color.BoldHiSuccessString("✓ teardown succeeded\n"))
				return true
			}

			return false
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(teardown *v1.Teardown) bool {
			j.Print(teardown)
			return false
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		teardown, err := client.V1().Systems().Teardowns(system).Get(id)
		if err != nil {
			return err
		}

		done := handle(teardown)
		if done {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func teardownWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func teardownString(teardown *v1.Teardown) string {
	stateColor := color.BoldString
	switch teardown.Status.State {
	case v1.TeardownStatePending, v1.TeardownStateInProgress:
		stateColor = color.BoldHiWarningString

	case v1.TeardownStateSucceeded:
		stateColor = color.BoldHiSuccessString

	case v1.TeardownStateFailed:
		stateColor = color.BoldHiFailureString
	}

	additional := ""
	if teardown.Status.Message != "" {
		additional += fmt.Sprintf(`
  message: %v`,
			teardown.Status.Message,
		)
	}

	if teardown.Status.StartTimestamp != nil {
		additional += fmt.Sprintf(`
  started: %v`,
			teardown.Status.StartTimestamp.Local().String(),
		)
	}

	if teardown.Status.CompletionTimestamp != nil {
		additional += fmt.Sprintf(`
  completed: %v`,
			teardown.Status.CompletionTimestamp.Local().String(),
		)
	}

	return fmt.Sprintf(`teardown %v
  state: %v%v
`,
		color.IDString(string(teardown.ID)),
		stateColor(string(teardown.Status.State)),
		additional,
	)
}

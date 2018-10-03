package systems

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

// Status returns a *cli.Command to retrieve the status of a system.
func Status() *cli.Command {
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
				return WatchSystemStatus(ctx.Client, ctx.System, format)
			}

			return PrintSystemStatus(ctx.Client, ctx.System, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// PrintSystemStatus prints the specified system's status to the supplied writer.
func PrintSystemStatus(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format) error {
	system, err := client.V1().Systems().Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := systemWriter(w)
		s := systemString(system)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(system)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

// WatchSystemStatus watches the specified status, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchSystemStatus(client client.Interface, id v1.SystemID, f printer.Format) error {
	var handle func(*v1.System)
	switch f {
	case printer.FormatTable:
		dw := systemWriter(os.Stdout)

		handle = func(system *v1.System) {
			s := systemString(system)
			dw.Overwrite(s)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(system *v1.System) {
			j.Print(system)
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		system, err := client.V1().Systems().Get(id)
		if err != nil {
			return err
		}

		handle(system)

		time.Sleep(5 * time.Second)
	}
}

func systemWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func systemString(system *v1.System) string {
	stateColor := color.BoldString
	switch system.Status.State {
	case v1.SystemStatePending, v1.SystemStateScaling, v1.SystemStateUpdating:
		stateColor = color.BoldHiWarningString

	case v1.SystemStateStable:
		stateColor = color.BoldHiSuccessString

	case v1.SystemStateDegraded, v1.SystemStateDeleting:
		stateColor = color.BoldHiFailureString
	}

	return fmt.Sprintf(`system %s
  state: %s
`,
		color.IDString(string(system.ID)),
		stateColor(string(system.Status.State)),
	)
}

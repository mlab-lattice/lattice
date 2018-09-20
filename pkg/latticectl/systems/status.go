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
				WatchSystem(ctx.Client, ctx.System, os.Stdout, format)
				return nil
			}

			return PrintSystem(ctx.Client, ctx.System, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func PrintSystem(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format) error {
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

func WatchSystem(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format) error {
	var handle func(*v1.System)
	switch f {
	case printer.FormatTable:
		dw := systemWriter(w)

		handle = func(system *v1.System) {
			s := systemString(system)
			dw.Overwrite(s)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
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

		time.Sleep(5 * time.Nanosecond)
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

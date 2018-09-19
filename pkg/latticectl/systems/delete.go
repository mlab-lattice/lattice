package systems

import (
	"fmt"
	"io"
	"os"
	//"sort"
	//"strings"
	//"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	//"k8s.io/apimachinery/pkg/util/wait"
	//"github.com/briandowns/spinner"
	"github.com/briandowns/spinner"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"time"
)

func Delete() *cli.Command {
	var (
		output string
		system string
		watch  bool
	)

	// teardown explicitly makes the user set the system flag rather than
	// using the set context
	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			command.SystemFlagName: &flags.String{
				Required: true,
				Target:   &system,
			},
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			return DeleteSystem(ctx.Client, v1.SystemID(system), os.Stdout, format, watch)
		},
	}

	return cmd.Command()
}

func DeleteSystem(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format, watch bool) error {
	err := client.V1().Systems().Delete(id)
	if err != nil {
		return err
	}

	if watch {
		return WatchSystemDelete(client, id, w, f)
	}

	fmt.Fprintf(
		w, `system %v deleting

to watch progress, run:
  latticectl systems status --system %v -w
`,
		color.IDString(string(id)),
		string(id),
	)
	return nil
}

func WatchSystemDelete(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()
	s.Suffix = " system is deleting"

	for {
		system, err := client.V1().Systems().Get(id)
		if err != nil {
			v1err, ok := err.(*v1.Error)
			if !ok {
				return err
			}

			switch v1err.Code {
			case v1.ErrorCodeInvalidSystemID:
				s.Stop()
				fmt.Fprintf(w, color.BoldHiSuccessString("✓ system has been deleted\n"))
				return nil

			default:
				s.Stop()
				return err
			}
		}

		if system.Status.State != v1.SystemStateDeleting {
			s.Stop()
			return fmt.Errorf("system %v is not deleting", id)
		}

		time.Sleep(5 * time.Second)
	}
}

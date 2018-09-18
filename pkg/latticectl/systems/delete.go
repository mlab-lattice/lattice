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
)

func Delete() *cli.Command {
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
			return DeleteSystem(ctx.Client, ctx.System, os.Stdout, format, watch)
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
	return nil
}

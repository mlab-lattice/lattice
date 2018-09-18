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
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

func Create() *cli.Command {
	var (
		definition string
		name       string
		output     string
		watch      bool
	)

	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			"definition": &flags.String{
				Required: true,
				Target:   &definition,
			},
			"name": &flags.String{
				Required: true,
				Target:   &name,
			},
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
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			return CreateSystem(ctx.Client, v1.SystemID(name), definition, os.Stdout, format, watch)
		},
	}

	return cmd.Command()
}

func CreateSystem(client client.Interface, id v1.SystemID, definition string, w io.Writer, f printer.Format, watch bool) error {
	_, err := client.V1().Systems().Create(id, definition)
	if err != nil {
		return err
	}

	if watch {
		return WatchSystemCreate(client, id, w, f)
	}

	fmt.Fprintf(
		w, `system %v initializing

to watch progress, run:
  latticectl systems status --system %v -w
`,
		color.IDString(string(id)),
		string(id),
	)
	return nil
}

func WatchSystemCreate(client client.Interface, id v1.SystemID, w io.Writer, f printer.Format) error {
	return nil
}

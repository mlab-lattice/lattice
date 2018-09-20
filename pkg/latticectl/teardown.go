package latticectl

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/teardowns"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

func Teardown() *cli.Command {
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
			return TeardownSystem(ctx.Client, v1.SystemID(system), os.Stdout, format, watch)
		},
	}

	return cmd.Command()
}

func TeardownSystem(
	client client.Interface,
	system v1.SystemID,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	teardown, err := client.V1().Systems().Teardowns(system).Create()
	if err != nil {
		return err
	}

	return displayTeardown(client, system, teardown, w, f, watch)
}

func displayTeardown(
	client client.Interface,
	system v1.SystemID,
	teardown *v1.Teardown,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	if watch {
		return teardowns.WatchTeardown(client, system, teardown.ID, w, f)
	}

	fmt.Fprintf(
		w,
		`
tearing down system %s. teardown ID: %s

to watch teardown, run:
    latticectl teardowns status --system %s --teardown %s -w
`,
		color.IDString(string(system)),
		color.IDString(string(teardown.ID)),
		system,
		teardown.ID,
	)
	return nil
}

package latticectl

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/teardowns"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

// Teardown returns a *cli.Command to tear down a system.
func Teardown() *cli.Command {
	var (
		output string
		system string
		watch  bool
	)

	// explicitly make the user set the system flag rather than
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
			return TeardownSystem(ctx.Client, v1.SystemID(system), format, watch)
		},
	}

	return cmd.Command()
}

// TeardownSystem tears down a system.
func TeardownSystem(
	client client.Interface,
	system v1.SystemID,
	f printer.Format,
	watch bool,
) error {
	teardown, err := client.V1().Systems().Teardowns(system).Create()
	if err != nil {
		return err
	}

	return displayTeardown(client, system, teardown, f, watch)
}

func displayTeardown(
	client client.Interface,
	system v1.SystemID,
	teardown *v1.Teardown,
	f printer.Format,
	watch bool,
) error {
	if watch {
		return teardowns.WatchTeardownStatus(client, system, teardown.ID, f)
	}

	fmt.Printf(
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

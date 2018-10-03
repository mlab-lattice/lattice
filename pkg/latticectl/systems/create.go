package systems

import (
	"fmt"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"

	"github.com/briandowns/spinner"
)

// Create returns a *cli.Command to create a system.
func Create() *cli.Command {
	var (
		definition string
		name       string
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
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			return CreateSystem(ctx.Client, v1.SystemID(name), definition, watch)
		},
	}

	return cmd.Command()
}

// CreateSystem creates the system with the specified options.
func CreateSystem(client client.Interface, id v1.SystemID, definition string, watch bool) error {
	_, err := client.V1().Systems().Create(id, definition)
	if err != nil {
		return err
	}

	if watch {
		return WatchSystemCreate(client, id)
	}

	fmt.Printf(
		`system %s initializing

to watch progress, run:
  latticectl systems status --system %s -w
`,
		color.IDString(string(id)),
		id,
	)
	return nil
}

// WatchSystemCreate spins until the system has successfully been created.
func WatchSystemCreate(client client.Interface, id v1.SystemID) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()
	s.Suffix = " system is creating"

	for {
		system, err := client.V1().Systems().Get(id)
		if err != nil {
			return err
		}

		if system.Status.State == v1.SystemStateStable {
			s.Stop()
			fmt.Printf(color.BoldHiSuccessString("✓ system has been created\n"))
		}

		time.Sleep(5 * time.Second)
	}
}

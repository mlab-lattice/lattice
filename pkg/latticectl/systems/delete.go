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

// Delete returns a *cli.Command to delete a system.
func Delete() *cli.Command {
	var (
		system string
		watch  bool
	)

	// explicitly make the user set the system flag rather than
	// using the set context
	cmd := command.LatticeCommand{
		Flags: map[string]cli.Flag{
			command.SystemFlagName: &flags.String{
				Required: true,
				Target:   &system,
			},
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *command.LatticeCommandContext, args []string, flags cli.Flags) error {
			return DeleteSystem(ctx.Client, v1.SystemID(system), watch)
		},
	}

	return cmd.Command()
}

// DeleteSystem deletes the system.
func DeleteSystem(client client.Interface, id v1.SystemID, watch bool) error {
	err := client.V1().Systems().Delete(id)
	if err != nil {
		return err
	}

	if watch {
		return WatchSystemDelete(client, id)
	}

	fmt.Printf(
		`system %s deleting

to watch progress, run:
  latticectl systems status --system %s -w
`,
		color.IDString(string(id)),
		string(id),
	)
	return nil
}

// WatchSystemDelete spins until the system has successfully been deleted.
func WatchSystemDelete(client client.Interface, id v1.SystemID) error {
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
				fmt.Printf(color.BoldHiSuccessString("✓ system has been deleted\n"))
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

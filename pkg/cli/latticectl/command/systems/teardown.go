package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/teardowns"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/briandowns/spinner"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: teardowns.ListTeardownsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "teardown",
		Flags: []command.Flag{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			systemID := ctx.SystemID()

			err = TeardownSystem(ctx.Client().Systems(), systemID, format, os.Stdout, watch)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func TeardownSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer, watch bool) error {
	// TODO :: Add watch of this. Same with deploy / build - link to behavior of teardowns/get.go etc
	teardownID, err := client.Teardowns(systemID).Create()
	if err != nil {
		log.Panic(err)
	}

	if watch {
		if format == printer.FormatDefault || format == printer.FormatTable {
			fmt.Fprintf(writer, "\nTearing down system %s. Teardown ID: %s\n\n", color.ID(string(systemID)), color.ID(string(teardownID)))
		}
		err = WatchSystem(client, systemID, format, writer, printSystemStateDuringTeardown, true)
		if err != nil {
			log.Panic(err)
		}
	} else {
		fmt.Fprintf(writer, "\nTearing down system %s. Teardown ID: %s\n\n", color.ID(string(systemID)), color.ID(string(teardownID)))
		fmt.Fprint(writer, "To watch teardown, run:\n\n")
		fmt.Fprintf(writer, "    lattice system:teardowns:status -w --teardown %s\n", string(teardownID))
	}

	return nil
}

//TODO: Need to get the flavour text the correct context for tearing down
func printSystemStateDuringTeardown(writer io.Writer, s *spinner.Spinner, system *types.System) {
	switch system.State {
	case types.SystemStateScaling:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is scaling...", color.ID(string(system.ID)))
	case types.SystemStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is updating...", color.ID(string(system.ID)))
	case types.SystemStateDeleting:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is terminating...", color.ID(string(system.ID)))
	case types.SystemStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("System %s is stable.", string(system.ID)))
	case types.SystemStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("System %s has failed.", string(system.ID)))

		var serviceErrors [][]string

		for serviceName, service := range system.Services {
			if service.State == types.ServiceStateFailed {
				serviceErrors = append(serviceErrors, []string{
					fmt.Sprintf("%s", serviceName),
					string(*service.FailureMessage),
				})
			}
		}

		printSystemFailure(writer, system.ID, serviceErrors)
	}
}

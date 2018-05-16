package systems

import (
	"fmt"
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/teardowns"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"github.com/briandowns/spinner"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: teardowns.ListTeardownsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.SystemCommand{
		Name: "teardown",
		Flags: []cli.Flag{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
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

func TeardownSystem(client v1client.SystemClient, systemID v1.SystemID, format printer.Format, writer io.Writer, watch bool) error {
	teardown, err := client.Teardowns(systemID).Create()
	if err != nil {
		log.Panic(err)
	}

	if watch {
		if format == printer.FormatTable {
			fmt.Fprintf(writer, "\nTearing down system %s. Teardown ID: %s\n\n", color.ID(string(systemID)), color.ID(string(teardown.ID)))
		}
		err = WatchSystem(client, systemID, format, writer, printSystemStateDuringTeardown, true)
		if err != nil {
			log.Panic(err)
		}
	} else {
		fmt.Fprintf(writer, "\nTearing down system %s. Teardown ID: %s\n\n", color.ID(string(systemID)), color.ID(string(teardown.ID)))
		fmt.Fprint(writer, "To watch teardown, run:\n\n")
		fmt.Fprintf(writer, "    lattice system:teardowns:status -w --teardown %s\n", string(teardown.ID))
	}

	return nil
}

//TODO: Need to get the flavour text the correct context for tearing down
func printSystemStateDuringTeardown(writer io.Writer, s *spinner.Spinner, system *v1.System, services []v1.Service) {
	switch system.State {
	case v1.SystemStateScaling:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is scaling...", color.ID(string(system.ID)))
	case v1.SystemStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is updating...", color.ID(string(system.ID)))
	case v1.SystemStateDeleting:
		s.Start()
		s.Suffix = fmt.Sprintf(" System %s is terminating...", color.ID(string(system.ID)))
	case v1.SystemStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("System %s is stable.", string(system.ID)))
	case v1.SystemStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("System %s has failed.", string(system.ID)))

		var serviceErrors [][]string

		for _, service := range services {
			if service.State == v1.ServiceStateFailed {
				message := "unknown"
				if service.FailureInfo != nil {
					message = service.FailureInfo.Message
				}

				serviceErrors = append(serviceErrors, []string{
					service.Path.String(),
					message,
				})
			}
		}

		printSystemFailure(writer, system.ID, serviceErrors)
	}
}

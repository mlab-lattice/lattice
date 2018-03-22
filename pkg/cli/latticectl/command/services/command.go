package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

// ListServicesSupportedFormats is the list of printer.Formats supported
// by the ListDeploys function.
var ListServicesSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type ListServicesCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListServicesCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "services",
		Flags: command.Flags{
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

			if watch {
				WatchServices(ctx.Client().Systems().Services(ctx.SystemID()), format, os.Stdout)
			} else {
				err := ListServices(ctx.Client().Systems().Services(ctx.SystemID()), format, os.Stdout)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListServices(client client.ServiceClient, format printer.Format, writer io.Writer) error {
	deploys, err := client.List()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", deploys)
	return nil
}

func WatchServices(client client.ServiceClient, format printer.Format, writer io.Writer) {
	// Poll the API for the builds and send it to the channel
	printerChan := make(chan printer.Interface)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			services, err := client.List()
			if err != nil {
				return false, err
			}

			p := servicesPrinter(services, format)
			printerChan <- p
			return false, nil
		},
	)

	// If displaying a table, use the overwriting terminal watcher, if JSON
	// use the scrolling watcher
	var w printer.Watcher
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		w = &printer.OverwrittingTerminalWatcher{}

	case printer.FormatJSON:
		w = &printer.ScrollingWatcher{}
	}

	w.Watch(printerChan, writer)
}

func servicesPrinter(services []types.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"ID", "Path", "State"}

		var rows [][]string
		for _, service := range services {
			var stateColor color.Color
			switch service.State {
			case types.ServiceStateFailed:
				stateColor = color.Failure
			case types.ServiceStateStable:
				stateColor = color.Success
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(service.ID),
				string(service.Path),
				stateColor(string(service.State)),
			})
		}

		p = &printer.Table{
			Headers: headers,
			Rows:    rows,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: services,
		}
	}

	return p
}

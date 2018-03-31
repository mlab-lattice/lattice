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

	tw "github.com/tfogo/tablewriter"
)

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.ServiceCommand{
		Name: "status",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.ServiceCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Services(ctx.SystemID())

			if watch {
				WatchService(c, ctx.ServiceID(), format, os.Stdout)
			}

			err = GetService(c, ctx.ServiceID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetService(client client.ServiceClient, serviceID types.ServiceID, format printer.Format, writer io.Writer) error {
	service, err := client.Get(serviceID)
	if err != nil {
		return err
	}

	printer := servicePrinter(service, format)
	printer.Print(writer)
	return nil
}

func WatchService(client client.ServiceClient, serviceID types.ServiceID, format printer.Format, writer io.Writer) {
	// Poll the API for the builds and send it to the channel
	printerChan := make(chan printer.Interface)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			service, err := client.Get(serviceID)
			if err != nil {
				return false, err
			}

			p := servicePrinter(service, format)
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

func servicePrinter(service *types.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Path", "State", "Instances", "Info"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
			{},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_RIGHT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string

		var stateColor color.Color
		switch service.State {
		case types.ServiceStateFailed:
			stateColor = color.Failure
		case types.ServiceStateStable:
			stateColor = color.Success
		default:
			stateColor = color.Warning
		}

		var info string
		if service.FailureMessage == nil {
			info = ""
		} else {
			info = *service.FailureMessage
		}

		rows = append(rows, []string{
			string(service.Path),
			stateColor(string(service.State)),
			fmt.Sprintf("%v", service.UpdatedInstances),
			string(info),
		})

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: service,
		}
	}

	return p
}

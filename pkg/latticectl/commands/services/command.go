package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"

	tw "github.com/tfogo/tablewriter"
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
	output := &latticectl.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool

	cmd := &latticectl.SystemCommand{
		Name: "services",
		Flags: cli.Flags{
			output.Flag(),
			&cli.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
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

func ListServices(client v1client.ServiceClient, format printer.Format, writer io.Writer) error {
	deploys, err := client.List()
	if err != nil {
		return err
	}

	printer := servicesPrinter(deploys, format)
	printer.Print(writer)
	//fmt.Printf("%v\n", deploys)
	return nil
}

func WatchServices(client v1client.ServiceClient, format printer.Format, writer io.Writer) {
	serviceLists := make(chan []v1.Service)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			serviceList, err := client.List()
			if err != nil {
				return false, err
			}

			serviceLists <- serviceList
			return false, nil
		},
	)

	for serviceList := range serviceLists {
		p := servicesPrinter(serviceList, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}

func servicesPrinter(services []v1.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Service", "State", "Available", "Updated", "Stale", "Terminating", "Addresses", "Info"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
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
			{},
			{},
			{},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, service := range services {
			var stateColor color.Color
			switch service.State {
			case v1.ServiceStateFailed:
				stateColor = color.Failure
			case v1.ServiceStateStable:
				stateColor = color.Success
			default:
				stateColor = color.Warning
			}

			var info string
			if service.Message != nil {
				info = *service.Message
			}
			if service.FailureInfo != nil {
				info = service.FailureInfo.Message
			}

			var addresses []string
			for port, address := range service.Ports {
				addresses = append(addresses, fmt.Sprintf("%v: %v", port, address))
			}

			rows = append(rows, []string{
				string(service.Path),
				stateColor(string(service.State)),
				fmt.Sprintf("%d", service.AvailableInstances),
				fmt.Sprintf("%d", service.UpdatedInstances),
				fmt.Sprintf("%d", service.StaleInstances),
				fmt.Sprintf("%d", service.TerminatingInstances),
				strings.Join(addresses, ","),
				string(info),
			})
		}

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: services,
		}
	}

	return p
}

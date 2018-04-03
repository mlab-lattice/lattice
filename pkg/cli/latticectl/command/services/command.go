package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

	printer := servicesPrinter(deploys, format)
	printer.Print(writer)
	//fmt.Printf("%v\n", deploys)
	return nil
}

func WatchServices(client client.ServiceClient, format printer.Format, writer io.Writer) {
	serviceLists := make(chan []types.Service)

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

func servicesPrinter(services []types.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Service", "State", "Updated", "Stale", "Addresses", "Info"}

		headerColors := []tw.Colors{
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
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_RIGHT,
			tw.ALIGN_RIGHT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

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

			var info string
			if service.FailureMessage == nil {
				info = ""
			} else {
				info = *service.FailureMessage
			}

			var addresses []string
			for port, address := range service.PublicPorts {
				addresses = append(addresses, fmt.Sprintf("%v: %v", port, address.Address))
			}

			rows = append(rows, []string{
				string(service.Path),
				stateColor(string(service.State)),
				fmt.Sprintf("%d", service.UpdatedInstances),
				fmt.Sprintf("%d", service.StaleInstances),
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

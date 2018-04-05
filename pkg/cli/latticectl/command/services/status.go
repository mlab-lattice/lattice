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

	"github.com/briandowns/spinner"
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
	services := make(chan *types.Service)

	lastHeight := 0
	var b bytes.Buffer
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			service, err := client.Get(serviceID)
			if err != nil {
				return false, err
			}

			services <- service
			return false, nil
		},
	)

	for service := range services {
		p := servicePrinter(service, format)
		lastHeight = p.Overwrite(b, lastHeight)

		if format == printer.FormatDefault || format == printer.FormatTable {
			printServiceState(writer, s, service)
		}
	}
}

func printServiceState(writer io.Writer, s *spinner.Spinner, service *types.Service) {
	switch service.State {
	case types.ServiceStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is pending...", color.ID(string(service.Path)))
	case types.ServiceStateScalingDown:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is scaling down...", color.ID(string(service.Path)))
	case types.ServiceStateScalingUp:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is scaling up...", color.ID(string(service.Path)))
	case types.ServiceStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is updating...", color.ID(string(service.Path)))
	case types.ServiceStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("Service %s is stable.", string(service.Path)))
	case types.ServiceStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("Service %s has failed. Error: %s", string(service.Path), service.FailureMessage))
	}
}

func servicePrinter(service *types.Service, format printer.Format) printer.Interface {
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

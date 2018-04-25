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
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/briandowns/spinner"
	tw "github.com/tfogo/tablewriter"
)

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.ServiceCommand{
		Name: "status",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.ServiceCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Services(ctx.SystemID())

			if watch {
				WatchService(c, ctx.ServicePath(), format, os.Stdout)
			}

			err = GetService(c, ctx.ServicePath(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetService(client v1client.ServiceClient, servicePath tree.NodePath, format printer.Format, writer io.Writer) error {
	service, err := client.Get(servicePath)
	if err != nil {
		return err
	}

	printer := servicePrinter(service, format)
	printer.Print(writer)
	return nil
}

func WatchService(client v1client.ServiceClient, servicePath tree.NodePath, format printer.Format, writer io.Writer) {
	services := make(chan *v1.Service)

	lastHeight := 0
	var b bytes.Buffer
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			service, err := client.Get(servicePath)
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

func printServiceState(writer io.Writer, s *spinner.Spinner, service *v1.Service) {
	switch service.State {
	case v1.ServiceStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is pending...", color.ID(string(service.Path)))
	case v1.ServiceStateScaling:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is scaling...", color.ID(string(service.Path)))
	case v1.ServiceStateUpdating:
		s.Start()
		s.Suffix = fmt.Sprintf(" Service %s is updating...", color.ID(string(service.Path)))
	case v1.ServiceStateStable:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiSuccess("Service %s is stable.", string(service.Path)))
	case v1.ServiceStateFailed:
		s.Stop()
		fmt.Fprint(writer, color.BoldHiFailure("Service %s has failed. Error: %s", string(service.Path), service.FailureMessage))
	}
}

func servicePrinter(service *v1.Service, format printer.Format) printer.Interface {
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
		case v1.ServiceStateFailed:
			stateColor = color.Failure
		case v1.ServiceStateStable:
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

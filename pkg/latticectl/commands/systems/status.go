package systems

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
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

type PrintSystemState func(io.Writer, *spinner.Spinner, *v1.System)

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target:  &watch,
	}

	cmd := &latticectl.SystemCommand{
		Name: "status",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				err = WatchSystem(c, ctx.SystemID(), format, os.Stdout, PrintSystemStateDuringStatus, false)
				if err != nil {
					os.Exit(1)
				}
			}

			err = GetSystem(c, ctx.SystemID(), format, os.Stdout)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	return cmd.Base()
}

func GetSystem(client v1client.SystemClient, systemID v1.SystemID, format printer.Format, writer io.Writer) error {
	system, err := client.Get(systemID)
	if err != nil {
		return err
	}

	p := SystemPrinter(system, format)
	p.Print(writer)
	return nil
}

func WatchSystem(client v1client.SystemClient, systemID v1.SystemID, format printer.Format, writer io.Writer, PrintSystemStateDuringStatus PrintSystemState, exitable bool) error {
	systems := make(chan *v1.System)

	lastHeight := 0
	var returnError error
	var exit bool
	var b bytes.Buffer
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	// Poll the API for the builds and send it to the channel
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			system, err := client.Get(systemID)
			if err != nil {
				return false, err
			}

			systems <- system
			return false, nil
		},
	)

	for system := range systems {
		p := SystemPrinter(system, format)
		lastHeight = p.Overwrite(b, lastHeight)

		if format == printer.FormatDefault || format == printer.FormatTable {
			PrintSystemStateDuringStatus(writer, s, system)
		}

		exit, returnError = systemInExitableState(system)

		if exitable && exit {
			return returnError
		}
	}

	return nil
}

func systemInExitableState(system *v1.System) (bool, error) {
	switch system.State {
	case v1.SystemStateStable:
		return true, nil
	case v1.SystemStateFailed:
		return true, errors.New("System Failed")
	default:
		return false, nil
	}
}

func PrintSystemStateDuringStatus(writer io.Writer, s *spinner.Spinner, system *v1.System) {
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

		for serviceName, service := range system.Services {
			if service.State == v1.ServiceStateFailed {
				serviceErrors = append(serviceErrors, []string{
					fmt.Sprintf("%s", serviceName),
					string(*service.FailureMessage),
				})
			}
		}

		printSystemFailure(writer, system.ID, serviceErrors)
	}
}

func printSystemFailure(writer io.Writer, systemID v1.SystemID, serviceErrors [][]string) {
	fmt.Fprint(writer, color.BoldHiFailure("✘ Error encountered in system "))
	fmt.Fprint(writer, color.BoldHiFailure(string(systemID)))
	fmt.Fprint(writer, color.BoldHiFailure(":\n\n"))
	for _, serviceError := range serviceErrors {
		fmt.Fprintf(writer, color.Failure("Error in service %s, Error message:\n\n    %s\n"), serviceError[0], serviceError[1])
	}
}

// Currently just prints systems. In the future, could print more details (e.g. jobs, node pools)
func SystemPrinter(system *v1.System, format printer.Format) printer.Interface {
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
		// fmt.Fprintln(os.Stdout, system)
		for serviceName, service := range system.Services {
			// fmt.Fprintln(os.Stdout, service)

			// fmt.Fprintln(os.Stdout, component)
			// fmt.Fprint(os.Stdout, "COMPONENT STATE", component.State, "    ")
			var infoMessage string

			if service.FailureMessage == nil {
				infoMessage = ""
			} else {
				infoMessage = string(*service.FailureMessage)
			}

			var stateColor color.Color
			switch service.State {
			case v1.ServiceStateStable:
				stateColor = color.Success
			case v1.ServiceStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			var addresses []string
			for port, address := range service.PublicPorts {
				addresses = append(addresses, fmt.Sprintf("%v: %v", port, address.Address))
			}

			rows = append(rows, []string{
				string(serviceName),
				stateColor(string(service.State)),
				fmt.Sprintf("%d", service.UpdatedInstances),
				fmt.Sprintf("%d", service.StaleInstances),
				strings.Join(addresses, ","),
				string(infoMessage),
			})

			sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
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
			Value: system,
		}
	}

	return p
}

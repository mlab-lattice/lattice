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

type GetCommand struct {
}

type PrintSystemState func(io.Writer, *spinner.Spinner, *types.System)

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
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
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				err = WatchSystem(c, ctx.SystemID(), format, os.Stdout, printSystemState, false)
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

func GetSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) error {
	system, err := client.Get(systemID)
	if err != nil {
		return err
	}

	p := SystemPrinter(system, format)
	p.Print(writer)
	return nil
}

func WatchSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer, printSystemState PrintSystemState, exitable bool) error {
	systems := make(chan *types.System)

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
			printSystemState(writer, s, system)
		}

		exit, returnError = systemInExitableState(system)

		if exitable && exit {
			return returnError
		}
	}

	return nil
}

func systemInExitableState(system *types.System) (bool, error) {
	switch system.State {
	case types.SystemStateStable:
		return true, nil
	case types.SystemStateFailed:
		return true, errors.New("System Failed")
	default:
		return false, nil
	}
}

func printSystemState(writer io.Writer, s *spinner.Spinner, system *types.System) {
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

func printSystemFailure(writer io.Writer, systemID types.SystemID, serviceErrors [][]string) {
	fmt.Fprint(writer, color.BoldHiFailure("✘ Error encountered in system "))
	fmt.Fprint(writer, color.BoldHiFailure(string(systemID)))
	fmt.Fprint(writer, color.BoldHiFailure(":\n\n"))
	for _, serviceError := range serviceErrors {
		fmt.Fprintf(writer, color.Failure("Error in service %s, Error message:\n\n    %s\n"), serviceError[0], serviceError[1])
	}
}

func SystemPrinter(system *types.System, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Service", "State", "Info", "Updated", "Stale", "Addresses"}

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
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_LEFT,
			tw.ALIGN_RIGHT,
			tw.ALIGN_RIGHT,
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
			case types.ServiceStateStable:
				stateColor = color.Success
			case types.ServiceStateFailed:
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
				string(infoMessage),
				fmt.Sprintf("%d", service.UpdatedInstances),
				fmt.Sprintf("%d", service.StaleInstances),
				strings.Join(addresses, ","),
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

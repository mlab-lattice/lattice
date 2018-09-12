package systems

import (
	"bytes"
	"fmt"
	"github.com/briandowns/spinner"
	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"io"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"sort"
	"strings"
	"time"
)

func Status() *cli.Command {
	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			"output": command.OutputFlag(SystemsSupportedFormats, printer.FormatTable),
			"watch":  command.WatchFlag(),
		},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) {
			format := printer.Format(flags["watch"].Value().(string))

			if flags["watch"].Value().(bool) {
				WatchSystem(ctx.Client.V1().Systems(), ctx.System, format, os.Stdout)
				return
			}

			err := GetSystem(ctx.Client.V1().Systems(), ctx.System, format, os.Stdout)
			if err != nil {
				panic(err)
			}
		},
	}

	return cmd.Command()
}

func GetSystem(client clientv1.SystemClient, id v1.SystemID, format printer.Format, writer io.Writer) error {
	// TODO: Make requests in parallel
	system, err := client.Get(id)
	if err != nil {
		return err
	}

	serviceList, err := client.Services(id).List()
	if err != nil {
		return err
	}

	p := systemPrinter(system, serviceList, format)
	p.Print(writer)
	return nil
}

func WatchSystem(client clientv1.SystemClient, id v1.SystemID, format printer.Format, writer io.Writer) {
	type status struct {
		system   *v1.System
		services []v1.Service
	}

	statuses := make(chan status)

	// Poll the API for the builds and send it to the channel
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			// TODO: Make requests in parallel
			system, err := client.Get(id)
			if err != nil {
				return false, err
			}

			services, err := client.Services(id).List()
			if err != nil {
				return false, err
			}

			statuses <- status{system, services}

			return false, nil
		},
	)
	lastHeight := 0
	var b bytes.Buffer
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	for s := range statuses {
		p := systemPrinter(s.system, s.services, format)
		lastHeight = p.Overwrite(b, lastHeight)

		if format == printer.FormatTable {
			printState(writer, spin, s.system, s.services)
		}

		switch s.system.State {
		case v1.SystemStateStable, v1.SystemStateFailed:
			return
		}
	}
}

func printState(writer io.Writer, s *spinner.Spinner, system *v1.System, services []v1.Service) {
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

		fmt.Fprint(writer, color.BoldHiFailure("✘ Error encountered in system "))
		fmt.Fprint(writer, color.BoldHiFailure(string(system.ID)))
		fmt.Fprint(writer, color.BoldHiFailure(":\n\n"))
		for _, serviceError := range serviceErrors {
			fmt.Fprintf(writer, color.Failure("Error in service %s, Error message:\n\n    %s\n"), serviceError[0], serviceError[1])
		}
	}
}

// Currently just prints systems. In the future, could print more details (e.g. jobs, node pools)
func systemPrinter(system *v1.System, services []v1.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		t := &printer.Table{
			Columns: []printer.TableColumn{
				{
					Header:    "Service",
					Color:     color.ID,
					Alignment: printer.TableAlignLeft,
				},
				{
					Header:    "State",
					Alignment: printer.TableAlignLeft,
				},
				{
					Header:    "Available",
					Alignment: printer.TableAlignRight,
				},
				{
					Header:    "Updated",
					Alignment: printer.TableAlignRight,
				},
				{
					Header:    "Stale",
					Alignment: printer.TableAlignRight,
				},
				{
					Header:    "Terminating",
					Alignment: printer.TableAlignRight,
				},
				{
					Header:    "Ports",
					Alignment: printer.TableAlignLeft,
				},
				{
					Header:    "Info",
					Alignment: printer.TableAlignLeft,
				},
			},
		}

		var rows [][]string
		for _, service := range services {
			var message string
			if service.Message != nil {
				message = *service.Message
			}
			if service.FailureInfo != nil {
				message = service.FailureInfo.Message
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
			for port, address := range service.Ports {
				addresses = append(addresses, fmt.Sprintf("%v: %v", port, address))
			}

			rows = append(rows, []string{
				service.Path.String(),
				stateColor(string(service.State)),
				fmt.Sprintf("%d", service.AvailableInstances),
				fmt.Sprintf("%d", service.UpdatedInstances),
				fmt.Sprintf("%d", service.StaleInstances),
				fmt.Sprintf("%d", service.TerminatingInstances),
				strings.Join(addresses, ","),
				string(message),
			})

		}

		// sort the rows by service ID
		sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
		t.AppendRows(rows)

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: system,
		}
	}

	return p
}

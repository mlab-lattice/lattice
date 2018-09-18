package systems

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/briandowns/spinner"
)

func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(&output, ListSupportedFormats, printer.FormatTable),
			command.WatchFlagName:  command.WatchFlag(&watch),
		},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchSystem(ctx.Client.V1().Systems(), ctx.System, format, os.Stdout)
				return nil
			}

			return GetSystem(ctx.Client.V1().Systems(), ctx.System, format, os.Stdout)
		},
	}

	return cmd.Command()
}

func GetSystem(client clientv1.SystemClient, id v1.SystemID, format printer.Format, w io.Writer) error {
	// TODO: Make requests in parallel
	system, err := client.Get(id)
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		services, err := client.Services(id).List()
		if err != nil {
			return err
		}

		t := systemTable(w)
		r := systemTableRows(services)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(system)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

func WatchSystem(client clientv1.SystemClient, id v1.SystemID, f printer.Format, w io.Writer) {
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
				if e, ok := err.(*v1.Error); ok && e.Code == v1.ErrorCodeInvalidSystemID {
					close(statuses)
					return true, nil
				}

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

	var handle func(s status)
	switch f {
	case printer.FormatTable:
		t := systemTable(w)
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		handle = func(status status) {
			r := systemTableRows(status.services)
			t.Overwrite(r)

			switch status.system.State {
			case v1.SystemStatePending:
				s.Suffix = " pending..."

			case v1.SystemStateDeleting:
				s.Suffix = " deleting..."

			case v1.SystemStateFailed:
				s.Stop()
				fmt.Fprintln(w, color.BoldHiFailureString(" ✘ system has failed"))

			case v1.SystemStateScaling:
				s.Suffix = " scaling..."

			case v1.SystemStateUpdating:
				s.Suffix = " updating..."

			case v1.SystemStateStable:
				s.Suffix = color.BoldHiSuccessString(" ✓ system is stable")

			case v1.SystemStateDegraded:
				s.Suffix = " system has become degraded"
			}
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(status status) {
			j.Print(status.system)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", f))
	}

	for s := range statuses {
		handle(s)
	}
}

func systemTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []printer.TableColumn{
		{
			Header:    "service",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "state",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "available",
			Alignment: printer.TableAlignRight,
		},
		{
			Header:    "updated",
			Alignment: printer.TableAlignRight,
		},
		{
			Header:    "stale",
			Alignment: printer.TableAlignRight,
		},
		{
			Header:    "terminating",
			Alignment: printer.TableAlignRight,
		},
		{
			Header:    "ports",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "info",
			Alignment: printer.TableAlignLeft,
		},
	})
}

func systemTableRows(services []v1.Service) []printer.TableRow {
	var rows []printer.TableRow
	for _, service := range services {
		var message string
		if service.Message != nil {
			message = *service.Message
		}
		if service.FailureInfo != nil {
			message = service.FailureInfo.Message
		}

		var stateColor color.Formatter
		switch service.State {
		case v1.ServiceStateStable:
			stateColor = color.SuccessString
		case v1.ServiceStateFailed:
			stateColor = color.FailureString
		default:
			stateColor = color.WarningString
		}

		var addresses []string
		for port, address := range service.Ports {
			addresses = append(addresses, fmt.Sprintf("%v: %v", port, address))
		}

		rows = append(rows, []string{
			color.IDString(service.Path.String()),
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

	return rows
}

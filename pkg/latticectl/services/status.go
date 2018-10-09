package services

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	servicecommand "github.com/mlab-lattice/lattice/pkg/latticectl/services/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

// Status returns a *cli.Command to retrieve the status of a service.
func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := servicecommand.ServiceCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *servicecommand.ServiceCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				return WatchServiceStatus(ctx.Client, ctx.System, ctx.Service, format)
			}

			return PrintServiceStatus(ctx.Client, ctx.System, ctx.Service, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// PrintServiceStatus prints the specified service's status to the supplied writer.
func PrintServiceStatus(client client.Interface, system v1.SystemID, id v1.ServiceID, w io.Writer, f printer.Format) error {
	service, err := client.V1().Systems().Services(system).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := serviceWriter(w)
		s := serviceString(service)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(service)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

// WatchServiceStatus watches the specified service, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchServiceStatus(client client.Interface, system v1.SystemID, id v1.ServiceID, f printer.Format) error {
	var handle func(*v1.Service)
	switch f {
	case printer.FormatTable:
		dw := serviceWriter(os.Stdout)

		handle = func(service *v1.Service) {
			s := serviceString(service)
			dw.Overwrite(s)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(os.Stdout)
		handle = func(service *v1.Service) {
			j.Print(service)
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		service, err := client.V1().Systems().Services(system).Get(id)
		if err != nil {
			return err
		}

		handle(service)

		time.Sleep(5 * time.Second)
	}
}

func serviceWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func serviceString(service *v1.Service) string {
	stateColor := color.BoldString
	switch service.Status.State {
	case v1.ServiceStatePending, v1.ServiceStateScaling, v1.ServiceStateUpdating:
		stateColor = color.BoldHiWarningString

	case v1.ServiceStateStable:
		stateColor = color.BoldHiSuccessString

	case v1.ServiceStateFailed, v1.ServiceStateDeleting:
		stateColor = color.BoldHiFailureString
	}

	updatedInstancesColor := color.NormalString
	if service.Status.UpdatedInstances != 0 {
		updatedInstancesColor = color.SuccessString
	}

	staleInstancesColor := color.NormalString
	if service.Status.StaleInstances != 0 {
		staleInstancesColor = color.WarningString
	}

	terminatingInstancesColor := color.NormalString
	if service.Status.StaleInstances != 0 {
		terminatingInstancesColor = color.FailureString
	}

	message := ""
	if service.Status.Message != nil {
		message += fmt.Sprintf(`
  message: %s`,
			*service.Status.Message,
		)
	}

	ports := ""
	if len(service.Status.Ports) != 0 {
		ports = `
  ports:`
	}
	for port, address := range service.Status.Ports {
		ports += fmt.Sprintf(`
    %d: %s`,
			port,
			address,
		)
	}

	instances := ""
	if len(service.Status.Ports) != 0 {
		instances = `
  instances:`
	}
	for _, instance := range service.Status.Instances {
		instances += fmt.Sprintf(`
    %s`,
			instance,
		)
	}

	return fmt.Sprintf(`service %s (%s)
  state: %s
  available instances: %s
  updated instances: %s
  stale instances: %s
  terminating instances: %s%s%s%s
`,
		color.IDString(string(service.ID)),
		service.Path.String(),
		stateColor(string(service.Status.State)),
		strconv.Itoa(int(service.Status.AvailableInstances)),
		updatedInstancesColor(strconv.Itoa(int(service.Status.UpdatedInstances))),
		staleInstancesColor(strconv.Itoa(int(service.Status.StaleInstances))),
		terminatingInstancesColor(strconv.Itoa(int(service.Status.TerminatingInstances))),
		message,
		ports,
		instances,
	)
}

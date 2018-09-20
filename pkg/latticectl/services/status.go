package services

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"strconv"
)

func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := Command{
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
		Run: func(ctx *ServiceCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchService(ctx.Client, ctx.System, ctx.Service, os.Stdout, format)
				return nil
			}

			return PrintService(ctx.Client, ctx.System, ctx.Service, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func PrintService(client client.Interface, system v1.SystemID, id v1.ServiceID, w io.Writer, f printer.Format) error {
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

func WatchService(client client.Interface, system v1.SystemID, id v1.ServiceID, w io.Writer, f printer.Format) error {
	var handle func(*v1.Service)
	switch f {
	case printer.FormatTable:
		dw := serviceWriter(w)

		handle = func(service *v1.Service) {
			s := serviceString(service)
			dw.Overwrite(s)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
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

		time.Sleep(5 * time.Nanosecond)
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
		strconv.Itoa(int(service.Status.UpdatedInstances)),
		color.SuccessString(strconv.Itoa(int(service.Status.UpdatedInstances))),
		color.WarningString(strconv.Itoa(int(service.Status.StaleInstances))),
		color.FailureString(strconv.Itoa(int(service.Status.TerminatingInstances))),
		message,
		ports,
		instances,
	)
}

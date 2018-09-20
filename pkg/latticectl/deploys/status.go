package deploys

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"time"
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
		Run: func(ctx *DeployCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchDeploy(ctx.Client, ctx.System, ctx.Deploy, os.Stdout, format)
				return nil
			}

			return PrintDeploy(ctx.Client, ctx.System, ctx.Deploy, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func PrintDeploy(client client.Interface, system v1.SystemID, id v1.DeployID, w io.Writer, f printer.Format) error {
	deploy, err := client.V1().Systems().Deploys(system).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := deployWriter(w)
		s := deployString(deploy)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(system)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

func WatchDeploy(client client.Interface, system v1.SystemID, id v1.DeployID, w io.Writer, f printer.Format) error {
	var handle func(*v1.Deploy) bool
	switch f {
	case printer.FormatTable:
		dw := deployWriter(w)
		handle = func(deploy *v1.Deploy) bool {
			s := deployString(deploy)
			dw.Overwrite(s)

			switch deploy.Status.State {
			case v1.DeployStateFailed:
				fmt.Fprint(w, color.BoldHiFailureString("✘ deploy failed\n"))
				return true

			case v1.DeployStateSucceeded:
				fmt.Fprint(w, color.BoldHiSuccessString("✓ deploy succeeded\n"))
				return true
			}

			return false
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(deploy *v1.Deploy) bool {
			j.Print(system)
			return false
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		deploy, err := client.V1().Systems().Deploys(system).Get(id)
		if err != nil {
			return err
		}

		done := handle(deploy)
		if done {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func deployWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func deployString(deploy *v1.Deploy) string {
	var spec string
	switch {
	case deploy.Build != nil:
		spec = fmt.Sprintf("build %v", *deploy.Build)

	case deploy.Path != nil:
		spec = fmt.Sprintf("path %v", deploy.Path.String())

	case deploy.Version != nil:
		spec = fmt.Sprintf("version %v", *deploy.Version)
	}

	stateColor := color.BoldString
	switch deploy.Status.State {
	case v1.DeployStatePending, v1.DeployStateAccepted, v1.DeployStateInProgress:
		stateColor = color.BoldHiWarningString

	case v1.DeployStateSucceeded:
		stateColor = color.BoldHiSuccessString

	case v1.DeployStateFailed:
		stateColor = color.BoldHiFailureString
	}

	additional := ""
	if deploy.Status.Message != "" {
		additional += fmt.Sprintf(`
  message: %v`,
			deploy.Status.Message,
		)
	}

	if deploy.Status.StartTimestamp != nil {
		additional += fmt.Sprintf(`
  started: %v`,
			deploy.Status.StartTimestamp.String(),
		)
	}

	if deploy.Status.CompletionTimestamp != nil {
		additional += fmt.Sprintf(`
  completed: %v`,
			deploy.Status.CompletionTimestamp.String(),
		)
	}

	return fmt.Sprintf(`deploy %v (%v)
  state: %v%v
`,
		color.IDString(string(deploy.ID)),
		spec,
		stateColor(string(deploy.Status.State)),
		additional,
	)
}

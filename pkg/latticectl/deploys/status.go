package deploys

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	deploycommand "github.com/mlab-lattice/lattice/pkg/latticectl/deploys/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

// Status returns a *cli.Command to retrieve the status of a deploy.
func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := deploycommand.DeployCommand{
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
		Run: func(ctx *deploycommand.DeployCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				return WatchDeployStatus(ctx.Client, ctx.System, ctx.Deploy, os.Stdout, format)
			}

			return PrintDeployStatus(ctx.Client, ctx.System, ctx.Deploy, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// PrintDeployStatus prints the specified deploy's status to the supplied writer.
func PrintDeployStatus(client client.Interface, system v1.SystemID, id v1.DeployID, w io.Writer, f printer.Format) error {
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
		j.Print(deploy)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

// WatchDeployStatus watches the specified build, updating output based on changes.
// When passed in printer.Table as f, the table uses some ANSI escapes to overwrite some of the terminal buffer,
// so it always writes to stdout and does not accept an io.Writer.
func WatchDeployStatus(client client.Interface, system v1.SystemID, id v1.DeployID, w io.Writer, f printer.Format) error {
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
			j.Print(deploy)
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

	if deploy.Status.Build != nil {
		additional += fmt.Sprintf(`
  build: %s`,
			string(*deploy.Status.Build),
		)
	}

	if deploy.Status.Path != nil {
		additional += fmt.Sprintf(`
  path: %s`,
			deploy.Status.Path.String(),
		)
	}

	if deploy.Status.Version != nil {
		additional += fmt.Sprintf(`
  version: %s`,
			string(*deploy.Status.Version),
		)
	}

	if deploy.Status.StartTimestamp != nil {
		additional += fmt.Sprintf(`
  started: %v`,
			deploy.Status.StartTimestamp.Local().String(),
		)
	}

	if deploy.Status.CompletionTimestamp != nil {
		additional += fmt.Sprintf(`
  completed: %v`,
			deploy.Status.CompletionTimestamp.Local().String(),
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

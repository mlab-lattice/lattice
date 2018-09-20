package latticectl

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/deploys"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"
)

func Deploys() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := command.SystemCommand{
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
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchDeploys(ctx.Client, ctx.System, format, os.Stdout)
				return nil
			}

			return PrintDeploys(ctx.Client, ctx.System, format, os.Stdout)
		},
		Subcommands: map[string]*cli.Command{
			"status": deploys.Status(),
		},
	}

	return cmd.Command()
}

// PrintDeploys writes the current Systems to the supplied io.Writer in the given printer.Format.
func PrintDeploys(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) error {
	deploys, err := client.V1().Systems().Deploys(system).List()
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatTable:
		t := deploysTable(w)
		r := deploysTableRows(deploys)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(deploys)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

// WatchDeploys polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchDeploys(client client.Interface, system v1.SystemID, format printer.Format, w io.Writer) {
	// Poll the API for the systems and send it to the channel
	deploys := make(chan []v1.Deploy)

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			d, err := client.V1().Systems().Deploys(system).List()
			if err != nil {
				return false, err
			}

			deploys <- d
			return false, nil
		},
	)

	var handle func([]v1.Deploy)
	switch format {
	case printer.FormatTable:
		t := deploysTable(w)
		handle = func(deploys []v1.Deploy) {
			r := deploysTableRows(deploys)
			t.Overwrite(r)
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(deploys []v1.Deploy) {
			j.Print(deploys)
		}

	default:
		panic(fmt.Sprintf("unexpected format %v", format))
	}

	for d := range deploys {
		handle(d)
	}
}

func deploysTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []printer.TableColumn{
		{
			Header:    "id",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "target",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "state",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "build",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "message",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "started",
			Alignment: printer.TableAlignLeft,
		},
		{
			Header:    "completed",
			Alignment: printer.TableAlignLeft,
		},
	})
}

func deploysTableRows(deploys []v1.Deploy) []printer.TableRow {
	var rows []printer.TableRow
	for _, deploy := range deploys {
		stateColor := color.WarningString
		switch deploy.Status.State {
		case v1.DeployStateSucceeded:
			stateColor = color.SuccessString

		case v1.DeployStateFailed:
			stateColor = color.FailureString
		}

		target := "-"
		switch {
		case deploy.Build != nil:
			target = fmt.Sprintf("build %v", *deploy.Build)

		case deploy.Path != nil:
			target = fmt.Sprintf("path %v", deploy.Path.String())

		case deploy.Version != nil:
			target = fmt.Sprintf("version %v", *deploy.Version)
		}

		build := "-"
		if deploy.Status.Build != nil {
			build = string(*deploy.Status.Build)
		}

		message := "-"
		if deploy.Status.Message != "" {
			message = deploy.Status.Message
		}

		started := "-"
		if deploy.Status.StartTimestamp != nil {
			started = deploy.Status.StartTimestamp.Format(time.RFC1123)
		}

		completed := "-"
		if deploy.Status.StartTimestamp != nil {
			completed = deploy.Status.StartTimestamp.Format(time.RFC1123)
		}

		rows = append(rows, []string{
			color.IDString(string(deploy.ID)),
			target,
			stateColor(string(deploy.Status.State)),
			build,
			message,
			started,
			completed,
		})
	}

	// sort the rows by start timestamp
	startedIdx := 5
	sort.Slice(
		rows,
		func(i, j int) bool {
			ts1, ts2 := rows[i][startedIdx], rows[j][startedIdx]
			if ts1 == "-" {
				return true
			}

			if ts2 == "-" {
				return false
			}

			t1, err := time.Parse(time.RFC1123, ts1)
			if err != nil {
				panic(err)
			}

			t2, err := time.Parse(time.RFC1123, ts2)
			if err != nil {
				panic(err)
			}
			return t1.After(t2)
		},
	)

	return rows
}

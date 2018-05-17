package deploys

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	tw "github.com/tfogo/tablewriter"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ListDeploysSupportedFormats is the list of printer.Formats supported
// by the ListDeploys function.
var ListDeploysSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

type ListDeploysCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListDeploysCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListDeploysSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.SystemCommand{
		Name: "deploys",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			if watch {
				WatchDeploys(ctx.Client().Systems().Deploys(ctx.SystemID()), format, os.Stdout)
			}

			err = ListDeploys(ctx.Client().Systems().Deploys(ctx.SystemID()), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListDeploys(client v1client.DeployClient, format printer.Format, writer io.Writer) error {
	deploys, err := client.List()
	if err != nil {
		return err
	}

	p := deploysPrinter(deploys, format)
	p.Print(writer)
	return nil
}

func WatchDeploys(client v1client.DeployClient, format printer.Format, writer io.Writer) {
	deployLists := make(chan []v1.Deploy)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			deployList, err := client.List()
			if err != nil {
				return false, err
			}

			deployLists <- deployList
			return false, nil
		},
	)

	for deployList := range deployLists {
		p := deploysPrinter(deployList, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}

func deploysPrinter(deploys []v1.Deploy, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		headers := []string{"ID", "Build ID", "State"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{tw.FgHiCyanColor},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, deploy := range deploys {
			var stateColor color.Color
			switch deploy.State {
			case v1.DeployStateSucceeded:
				stateColor = color.Success
			case v1.DeployStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(deploy.ID),
				string(deploy.BuildID),
				stateColor(string(deploy.State)),
			})
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
			Value: deploys,
		}
	}

	return p
}

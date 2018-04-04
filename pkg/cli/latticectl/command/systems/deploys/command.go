package deploys

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	tw "github.com/tfogo/tablewriter"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ListDeploysSupportedFormats is the list of printer.Formats supported
// by the ListDeploys function.
var ListDeploysSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type ListDeploysCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListDeploysCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListDeploysSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "deploys",
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

			if watch {
				WatchDeploys(ctx.Client().Systems().Rollouts(ctx.SystemID()), format, os.Stdout)
			}

			err = ListDeploys(ctx.Client().Systems().Rollouts(ctx.SystemID()), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListDeploys(client client.RolloutClient, format printer.Format, writer io.Writer) error {
	deploys, err := client.List()
	if err != nil {
		return err
	}

	p := deploysPrinter(deploys, format)
	p.Print(writer)
	return nil
}

func WatchDeploys(client client.RolloutClient, format printer.Format, writer io.Writer) {
	deployLists := make(chan []types.SystemRollout)

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

func deploysPrinter(deploys []types.SystemRollout, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
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
			case types.SystemRolloutStateSucceeded:
				stateColor = color.Success
			case types.SystemRolloutStateFailed:
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

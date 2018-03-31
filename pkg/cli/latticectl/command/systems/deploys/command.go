package deploys

import (
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

	"k8s.io/apimachinery/pkg/util/wait"
	tw "github.com/tfogo/tablewriter"
)

// ListDeploysSupportedFormats is the list of printer.Formats supported
// by the ListDeploys function.
var ListDeploysSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
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
	// Poll the API for the builds and send it to the channel
	printerChan := make(chan printer.Interface)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			deploys, err := client.List()
			if err != nil {
				return false, err
			}
	
			p := deploysPrinter(deploys, format)
			printerChan <- p
			return false, nil
		},
	)
	
	// If displaying a table, use the overwritting terminal watcher, if JSON
	// use the scrolling watcher
	var w printer.Watcher
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		w = &printer.OverwrittingTerminalWatcher{}
	
	case printer.FormatJSON:
		w = &printer.ScrollingWatcher{}
	}
	
	w.Watch(printerChan, writer)
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
			Headers: 			headers,
			Rows:    			rows,
			HeaderColors: headerColors,
			ColumnColors: columnColors,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: deploys,
		}
	}

	return p
}

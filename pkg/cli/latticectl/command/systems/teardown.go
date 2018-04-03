package systems

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
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/teardowns"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"

	tw "github.com/tfogo/tablewriter"
)

type TeardownCommand struct {
}

func (c *TeardownCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: teardowns.ListTeardownsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "teardown",
		Flags: []command.Flag{
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

			systemID := ctx.SystemID()

			err = TeardownSystem(ctx.Client().Systems().Teardowns(systemID), format, os.Stdout, watch)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func TeardownSystem(client client.TeardownClient, format printer.Format, writer io.Writer, watch bool) error {
	// TODO :: Add watch of this. Same with deploy / build - link to behavior of teardowns/get.go etc
	teardownID, err := client.Create()
	if err != nil {
		log.Panic(err)
	}

	if watch {
		WatchTeardown(client, teardownID, format, writer)
	} else {
		teardown, err := client.Get(teardownID)
		if err != nil {
			return err
		}

		p := teardownPrinter([]types.SystemTeardown{*teardown}, format)
		p.Print(writer)
	}

	return nil
}

func WatchTeardown(client client.TeardownClient, teardownID types.SystemTeardownID, format printer.Format, writer io.Writer) {
	teardowns := make(chan *types.SystemTeardown)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			teardown, err := client.Get(teardownID)
			if err != nil {
				return false, err
			}

			teardowns <- teardown
			return false, nil
		},
	)

	for teardown := range teardowns {
		p := teardownPrinter([]types.SystemTeardown{*teardown}, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}

func teardownPrinter(teardowns []types.SystemTeardown, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"ID", "State"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, teardown := range teardowns {
			var stateColor color.Color
			switch teardown.State {
			case types.SystemTeardownStateFailed:
				stateColor = color.Failure
			case types.SystemTeardownStateSucceeded:
				stateColor = color.Success
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(teardown.ID),
				stateColor(string(teardown.State)),
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
			Value: teardowns,
		}
	}

	return p
}

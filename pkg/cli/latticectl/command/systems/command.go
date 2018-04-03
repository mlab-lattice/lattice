package systems

import (
	"io"
	"log"
	"os"
	"time"
	"bytes"

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

// ListSystemsSupportedFormats is the list of printer.Formats supported
// by the ListSystems function.
var ListSystemsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

// ListSystemsCommand is a type that implements the latticectl.Command interface
// for listing the Systems in a Lattice.
type ListSystemsCommand struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *ListSystemsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.LatticeCommand{
		Name: "systems",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.LatticeCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				WatchSystems(c, format, os.Stdout)
				return
			}

			err = ListSystems(c, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

// ListSystems writes the current Systems to the supplied io.Writer in the given printer.Format.
func ListSystems(client client.SystemClient, format printer.Format, writer io.Writer) error {
	systems, err := client.List()
	if err != nil {
		return err
	}

	p := systemsPrinter(systems, format)
	p.Print(writer)

	return nil
}

// WatchSystems polls the API for the current Systems, and writes out the Systems to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchSystems(client client.SystemClient, format printer.Format, writer io.Writer) {
	// Poll the API for the systems and send it to the channel
	systemLists := make(chan []types.System)
	lastHeight := 0
	var b bytes.Buffer
	
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			systemList, err := client.List()
			if err != nil {
				return false, err
			}

			systemLists <- systemList
			return false, nil
		},
	)

	for systemList := range systemLists {
		p := systemsPrinter(systemList, format)
		lastHeight = p.Overwrite(b, lastHeight)
		
		// Note: Watching systems is never exitable.
		// There is no fail state for an entire lattice of systems.
	}
}

func systemsPrinter(systems []types.System, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Name", "Definition", "Status"}
		
		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}
		
		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
			{},
		}
		
		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, system := range systems {
			var stateColor color.Color
			switch system.State {
			case types.SystemStateStable:
				stateColor = color.Success
			case types.SystemStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}

			rows = append(rows, []string{
				string(system.ID),
				system.DefinitionURL,
				stateColor(string(system.State)),
			})
		}

		p = &printer.Table{
			Headers: headers,
			Rows:    rows,
			HeaderColors: headerColors,
			ColumnColors: columnColors,
			ColumnAlignment: 	columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: systems,
		}
	}

	return p
}

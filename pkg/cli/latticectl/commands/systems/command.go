package systems

import (
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type Command struct {
	Subcommands []latticectl.Command
}

var ListSystemSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSystemSupportedFormats,
	}

	cmd := &latticectl.LatticeCommand{
		Name:  "systems",
		Flags: command.Flags{output.Flag()},
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}
			ListSystems(ctx.Client().Systems(), format, os.Stdout)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSystems(client client.SystemClient, format printer.Format, writer io.Writer) {
	systems, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Name", "Definition", "Status"}
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
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: systems,
		}
	}

	p.Print(writer)
}

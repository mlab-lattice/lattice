package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"os"
)

var ListSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

func Command() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(),
			command.OutputFlagName: command.OutputFlag(ListSupportedFormats, printer.FormatTable),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			contextName, err := configFile.CurrentContext()
			if err != nil {
				return err
			}

			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			format := printer.Format(flags[command.OutputFlagName].Value().(string))
			return PrintContext(context, format)
		},
		Subcommands: map[string]*cli.Command{
			"create": Create(),
		},
	}
}

func PrintContext(ctx *command.Context, format printer.Format) error {
	switch format {
	case printer.FormatTable, printer.FormatJSON:
		j := printer.NewJSONIndented(os.Stdout, 4)
		j.Print(ctx)
	}

	return nil
}

package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"os"
)

var ListSupportedFormats = []printer.Format{
	printer.FormatJSON,
}

func List() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(),
			command.OutputFlagName: command.OutputFlag(GetSupportedFormats, printer.FormatJSON),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			contexts, err := configFile.Contexts()
			if err != nil {
				return err
			}

			format := printer.Format(flags[command.OutputFlagName].Value().(string))
			return PrintContexts(contexts, format)
		},
	}
}

func PrintContexts(contexts map[string]command.Context, format printer.Format) error {
	// FIXME: probably want to support a more natural table-like format
	switch format {
	case printer.FormatJSON:
		j := printer.NewJSONIndented(os.Stdout, 4)
		j.Print(contexts)
	}

	return nil
}

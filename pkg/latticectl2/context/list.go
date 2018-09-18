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
	var (
		configPath string
		output     string
	)

	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			command.OutputFlagName: command.OutputFlag(&output, GetSupportedFormats, printer.FormatJSON),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			contexts, err := configFile.Contexts()
			if err != nil {
				return err
			}

			format := printer.Format(output)
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

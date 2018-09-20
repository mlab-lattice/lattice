package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"io"
	"os"
)

func List() *cli.Command {
	var (
		configPath string
		output     string
	)

	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			command.OutputFlagName: command.OutputFlag(&output, []printer.Format{printer.FormatJSON}, printer.FormatJSON),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			contexts, err := configFile.Contexts()
			if err != nil {
				return err
			}

			format := printer.Format(output)
			return PrintContexts(contexts, os.Stdout, format)
		},
	}
}

func PrintContexts(contexts map[string]command.Context, w io.Writer, f printer.Format) error {
	// FIXME: probably want to support a more natural table-like format
	switch f {
	case printer.FormatJSON:
		j := printer.NewJSONIndented(w, 4)
		j.Print(contexts)
	}

	return nil
}

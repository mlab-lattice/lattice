package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/context"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"os"
)

func Context() *cli.Command {
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

			contextName, err := configFile.CurrentContext()
			if err != nil {
				return err
			}

			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			format := printer.Format(output)
			return PrintContext(context, format)
		},
		Subcommands: map[string]*cli.Command{
			"create": context.Create(),
			"delete": context.Delete(),
			"list":   context.List(),
			"switch": context.Switch(),
			"update": context.Update(),
		},
	}
}

func PrintContext(ctx *command.Context, format printer.Format) error {
	// FIXME: probably want to support a more natural table-like format
	switch format {
	case printer.FormatJSON:
		j := printer.NewJSONIndented(os.Stdout, 4)
		j.Print(ctx)
	}

	return nil
}

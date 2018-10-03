package latticectl

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/context"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"io"
	"os"
)

// Context returns a *cli.Command to print the current context, with subcommands
// to interact with contexts.
func Context() *cli.Command {
	var (
		configPath string
		output     string
	)

	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			name, err := configFile.CurrentContext()
			if err != nil {
				return err
			}

			context, err := configFile.Context(name)
			if err != nil {
				return err
			}

			format := printer.Format(output)
			return PrintContext(name, context, os.Stdout, format)
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

func PrintContext(name string, context *command.Context, w io.Writer, f printer.Format) error {
	switch f {
	case printer.FormatTable:
		dw := contextWriter(w)
		s := contextString(name, context)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSONIndented(os.Stdout, 4)
		j.Print(context)
	}

	return nil
}

func contextWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func contextString(name string, context *command.Context) string {
	authType := color.FailureString("unauthenticated")
	if context.Auth != nil {
		switch {
		case context.Auth.BearerToken != nil:
			authType = "bearer token"
		}
	}

	defaultSystem := ""
	if context.System != "" {
		defaultSystem = fmt.Sprintf(`
  system: %v`,
			context.System,
		)
	}

	return fmt.Sprintf(`context %v
  url: %v
  auth: %v%v
`,
		color.IDString(name),
		context.URL,
		authType,
		defaultSystem,
	)
}

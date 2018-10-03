package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"io"
	"os"
	"sort"
)

// Create returns a *cli.Command to delete a context.
func List() *cli.Command {
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
				}, printer.FormatTable,
			),
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

// PrintContexts prints the contexts to the supplied writer.
func PrintContexts(contexts map[string]command.Context, w io.Writer, f printer.Format) error {
	switch f {
	case printer.FormatTable:
		t := contextsTable(w)
		r := contextsTableRows(contexts)
		t.AppendRows(r)
		t.Print()

	case printer.FormatJSON:
		j := printer.NewJSONIndented(w, 4)
		j.Print(contexts)
	}

	return nil
}

func contextsTable(w io.Writer) *printer.Table {
	return printer.NewTable(w, []string{"NAME", "URL", "AUTH TYPE", "DEFAULT SYSTEM"})
}

func contextsTableRows(contexts map[string]command.Context) [][]string {
	var rows [][]string
	for name, context := range contexts {
		authType := color.FailureString("unauthenticated")
		if context.Auth != nil {
			switch {
			case context.Auth.BearerToken != nil:
				authType = "bearer token"
			}
		}

		rows = append(rows, []string{
			color.IDString(name),
			context.URL,
			authType,
			string(context.System),
		})
	}

	// sort the rows by name
	sort.Slice(
		rows,
		func(i, j int) bool {
			return rows[i][0] < rows[j][0]
		},
	)

	return rows
}

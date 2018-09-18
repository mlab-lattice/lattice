package systems

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

func Versions() *cli.Command {
	var (
		output string
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(&output, []printer.Format{printer.FormatJSON}, printer.FormatJSON),
		},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			return ListVersions(ctx.Client, ctx.System, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func ListVersions(client client.Interface, id v1.SystemID, w io.Writer, format printer.Format) error {
	versions, err := client.V1().Systems().Versions(id)
	if err != nil {
		return err
	}

	switch format {
	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(versions)

	default:
		return fmt.Errorf("unexpected format %v", format)
	}

	return nil
}

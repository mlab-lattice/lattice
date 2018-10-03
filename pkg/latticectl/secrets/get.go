package secrets

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

// Get returns a *cli.Command to retrieve the value of a secret.
func Get() *cli.Command {
	var (
		output string
	)

	cmd := SecretCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
		},
		Run: func(ctx *SecretCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			return GetSecret(ctx.Client, ctx.System, ctx.Secret, os.Stdout, format)
		},
	}

	return cmd.Command()
}

// GetSecret retrieves the value of the secret and prints it to the supplied writer.
func GetSecret(
	client client.Interface,
	system v1.SystemID,
	secret tree.PathSubcomponent,
	w io.Writer,
	f printer.Format,
) error {
	result, err := client.V1().Systems().Secrets(system).Get(secret)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := secretWriter(w)
		s := secretString(result)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(result)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}
func secretWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func secretString(secret *v1.Secret) string {
	return fmt.Sprintf(`secret %v
  value: %v
`,
		color.IDString(secret.Path.String()),
		secret.Value,
	)
}

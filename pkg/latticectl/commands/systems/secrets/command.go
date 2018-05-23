package secrets

import (
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type ListSecretsCommand struct {
	Subcommands []latticectl.Command
}

var ListSecretsSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

func (c *ListSecretsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListSecretsSupportedFormats,
	}

	cmd := &latticectl.SystemCommand{
		Name: "secrets",
		Flags: cli.Flags{
			output.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			err = ListSecrets(ctx.Client().Systems().Secrets(ctx.SystemID()), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListSecrets(client v1client.SecretClient, format printer.Format, writer io.Writer) error {
	secrets, err := client.List()
	if err != nil {
		return err
	}

	p := secretsPrinter(secrets, format)
	if err := p.Print(writer); err != nil {
		return err
	}
	return nil
}

func secretsPrinter(secrets []v1.Secret, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		var rows [][]string
		headers := []string{"Path", "Name", "Value"}

		for _, secret := range secrets {

			rows = append(rows, []string{
				color.ID(string(secret.Path)),
				color.ID(string(secret.Name)),
				string(secret.Value),
			})
		}

		p = &printer.Table{
			Headers: headers,
			Rows:    rows,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: secrets,
		}
	}

	return p
}

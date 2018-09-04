package secrets

import (
	"io"
	"log"
	"os"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	tw "github.com/tfogo/tablewriter"
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

func ListSecrets(client v1client.SystemSecretClient, format printer.Format, writer io.Writer) error {
	secrets, err := client.List()
	if err != nil {
		return err
	}

	p := secretsPrinter(secrets, format)
	p.Print(writer)
	return nil
}

func secretsPrinter(secrets []v1.Secret, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		headers := []string{"Path", "Value"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, secret := range secrets {

			rows = append(rows, []string{
				string(secret.Path),
				string(secret.Value),
			})
		}

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: secrets,
		}
	}

	return p
}

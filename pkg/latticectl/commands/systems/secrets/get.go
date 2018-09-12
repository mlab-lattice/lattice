package secrets

import (
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
	"io"
	"log"
	"os"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	var path string

	output := &latticectl.OutputFlag{
		SupportedFormats: ListSecretsSupportedFormats,
	}

	cmd := &latticectl.SystemCommand{
		Name: "get",
		Flags: cli.Flags{
			output.Flag(),
			&flags.String{
				Name:     "path",
				Required: true,
				Target:   &path,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			secretPath, err := tree.NewPathSubcomponent(path)
			if err != nil {
				log.Fatal("invalid secret path format")
			}

			err = GetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), secretPath, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetSecret(client v1client.SystemSecretClient, path tree.PathSubcomponent, format printer.Format, writer io.Writer) error {
	secret, err := client.Get(path)
	if err != nil {
		return err
	}

	p := secretsPrinter([]v1.Secret{*secret}, format)
	p.Print(writer)
	return nil
}

package secrets

import (
	"io"
	"log"
	"os"
	"strings"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	var name string

	output := &latticectl.OutputFlag{
		SupportedFormats: ListSecretsSupportedFormats,
	}

	cmd := &latticectl.SystemCommand{
		Name: "get",
		Flags: cli.Flags{
			output.Flag(),
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &name,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			splitName := strings.Split(name, ":")
			if len(splitName) != 2 {
				log.Fatal("invalid secret name format")
			}

			path := tree.Path(splitName[0])
			name = splitName[1]

			err = GetSecret(ctx.Client().Systems().Secrets(ctx.SystemID()), path, name, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetSecret(client v1client.SecretClient, path tree.Path, name string, format printer.Format, writer io.Writer) error {
	secret, err := client.Get(path, name)
	if err != nil {
		return err
	}

	p := secretsPrinter([]v1.Secret{*secret}, format)
	p.Print(writer)
	return nil
}

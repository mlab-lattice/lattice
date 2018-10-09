package secrets

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	secretcommand "github.com/mlab-lattice/lattice/pkg/latticectl/secrets/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
)

// Unset returns a *cli.Command to unset the value of a secret.
func Unset() *cli.Command {
	cmd := secretcommand.SecretCommand{
		Run: func(ctx *secretcommand.SecretCommandContext, args []string, flags cli.Flags) error {
			return UnsetSecret(ctx.Client, ctx.System, ctx.Secret, os.Stdout)
		},
	}

	return cmd.Command()
}

// UnsetSecret sets the value of the secret and prints a success message to the supplied writer if it succeeded.
func UnsetSecret(
	client client.Interface,
	system v1.SystemID,
	secret tree.PathSubcomponent,
	w io.Writer,
) error {
	err := client.V1().Systems().Secrets(system).Unset(secret)
	if err != nil {
		return err
	}

	fmt.Fprint(w, color.BoldHiSuccessString(fmt.Sprintf("✓ succesfully unset %v\n", secret.String())))
	return nil
}

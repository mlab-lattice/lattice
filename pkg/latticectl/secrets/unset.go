package secrets

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
)

func Unset() *cli.Command {
	cmd := Command{
		Run: func(ctx *SecretCommandContext, args []string, flags cli.Flags) error {
			return UnsetSecret(ctx.Client, ctx.System, ctx.Secret, os.Stdout)
		},
	}

	return cmd.Command()
}

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

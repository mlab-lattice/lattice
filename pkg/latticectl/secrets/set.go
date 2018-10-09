package secrets

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	secretcommand "github.com/mlab-lattice/lattice/pkg/latticectl/secrets/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	setFileFlag  = "file"
	setValueFlag = "value"
)

var setContentFlags = []string{setFileFlag, setValueFlag}

// Set returns a *cli.Command to set the value of a secret.
func Set() *cli.Command {
	var (
		file  string
		value string
	)

	cmd := secretcommand.SecretCommand{
		Flags: map[string]cli.Flag{
			setFileFlag:  &flags.String{Target: &file},
			setValueFlag: &flags.String{Target: &value},
		},
		MutuallyExclusiveFlags: [][]string{setContentFlags},
		RequiredFlagSet:        [][]string{setContentFlags},
		Run: func(ctx *secretcommand.SecretCommandContext, args []string, flags cli.Flags) error {
			if flags[setFileFlag].Set() {
				var data []byte
				var err error
				if file == "-" {
					data, err = ioutil.ReadAll(os.Stdin)
				} else {
					data, err = ioutil.ReadFile(file)
				}
				if err != nil {
					return err
				}

				value = string(data)
			}

			return SetSecret(ctx.Client, ctx.System, ctx.Secret, value, os.Stdout)
		},
	}

	return cmd.Command()
}

// SetSecret sets the value of the secret and prints a success message to the supplied writer if it succeeded.
func SetSecret(
	client client.Interface,
	system v1.SystemID,
	secret tree.PathSubcomponent,
	value string,
	w io.Writer,
) error {
	err := client.V1().Systems().Secrets(system).Set(secret, value)
	if err != nil {
		return err
	}

	fmt.Fprint(w, color.BoldHiSuccessString(fmt.Sprintf("✓ succesfully set %v\n", secret.String())))
	return nil
}

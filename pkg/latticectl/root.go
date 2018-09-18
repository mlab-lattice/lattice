package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
)

var Command = cli.RootCommand{
	Name: "latticectl",
	Command: &cli.Command{
		Short: "utility for interacting with lattices",
		Subcommands: map[string]*cli.Command{
			"build":     Build(),
			"builds":    Builds(),
			"context":   Context(),
			"deploy":    Deploy(),
			"deploys":   Deploys(),
			"secrets":   Secrets(),
			"systems":   Systems(),
			"teardown":  Teardown(),
			"teardowns": Teardowns(),
		},
	},
}

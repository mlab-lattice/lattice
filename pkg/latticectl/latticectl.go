package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

// Latticectl is the root command for latticectl.
var Latticectl = cli.RootCommand{
	Name: "latticectl",
	Command: &cli.Command{
		Short: "utility for interacting with lattices",
		Subcommands: map[string]*cli.Command{
			"build":     Build(),
			"builds":    Builds(),
			"context":   Context(),
			"deploy":    Deploy(),
			"deploys":   Deploys(),
			"jobs":      Jobs(),
			"secrets":   Secrets(),
			"services":  Services(),
			"systems":   Systems(),
			"teardown":  Teardown(),
			"teardowns": Teardowns(),
		},
	},
}

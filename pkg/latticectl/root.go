package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

var Command = cli.RootCommand{
	Name: "latticectl",
	Command: &cli.Command{
		Short: "utility for interacting with lattices",
		Subcommands: map[string]*cli.Command{
			"build":      Build(),
			"builds":     Builds(),
			"context":    Context(),
			"deploy":     Deploy(),
			"deploys":    Deploys(),
			"kubernetes": Kubernetes(),
			"local":      Local(),
			"secrets":    Secrets(),
			"services":   Services(),
			"systems":    Systems(),
			"teardown":   Teardown(),
			"teardowns":  Teardowns(),
		},
	},
}

package latticectl2

import (
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
)

var Command = cli.RootCommand{
	Name: "latticectl",
	Command: &cli.Command{
		Short: "utility for interacting with lattices",
		Subcommands: map[string]*cli.Command{
			"context": Context(),
			"deploy":  Deploy(),
			"systems": Systems(),
		},
	},
}

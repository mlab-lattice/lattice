package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/local"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

func Local() *cli.Command {
	return &cli.Command{
		Subcommands: map[string]*cli.Command{
			"down": local.Down(),
			"up":   local.Up(),
		},
	}
}

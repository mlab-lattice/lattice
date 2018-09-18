package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/secrets"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
)

func Secrets() *cli.Command {
	return &cli.Command{
		Subcommands: map[string]*cli.Command{
			"get": secrets.Get(),
			"set": secrets.Set(),
		},
	}
}

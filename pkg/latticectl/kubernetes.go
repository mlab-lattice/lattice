package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
)

func Kubernetes() *cli.Command {
	return &cli.Command{
		Subcommands: map[string]*cli.Command{
			"bootstrap": kubernetes.Bootstrap(),
		},
	}
}

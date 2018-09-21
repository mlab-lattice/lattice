package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

func Kubernetes() *cli.Command {
	return &cli.Command{
		Subcommands: map[string]*cli.Command{
			"bootstrap": kubernetes.Bootstrap(),
		},
	}
}

package servicemesh

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type LatticeBootstrapperOptions struct {
	Envoy *envoy.LatticeBootstrapperOptions
}

func NewLatticeBootstrapper(namespacePrefix string, options *LatticeBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewLatticeBootstrapper(namespacePrefix, options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}

func LatticeBoostrapperFlag(serviceMesh *string) (cli.Flag, *LatticeBootstrapperOptions) {
	envoyFlags, envoyOptions := envoy.LatticeBootstrapperFlags()
	options := &LatticeBootstrapperOptions{}

	flag := &cli.DelayedEmbeddedFlag{
		Name:     "service-mesh-var",
		Required: true,
		Usage:    "configuration for the service mesh cluster bootstrapper",
		Flags: map[string]cli.Flags{
			Envoy: envoyFlags,
		},
		FlagChooser: func() (*string, error) {
			if serviceMesh == nil {
				return nil, fmt.Errorf("serviceMesh cannot be nil")
			}

			switch *serviceMesh {
			case Envoy:
				options.Envoy = envoyOptions
			default:
				return nil, fmt.Errorf("unsupported service mesh %v", *serviceMesh)
			}

			return serviceMesh, nil
		},
	}

	return flag, options
}

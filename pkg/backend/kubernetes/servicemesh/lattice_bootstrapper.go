package servicemesh

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type ClusterBootstrapperOptions struct {
	Envoy *envoy.LatticeBootstrapperOptions
}

func NewLatticeBootstrapper(namespacePrefix string, options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewLatticeBootstrapper(namespacePrefix, options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}

func LatticeBoostrapperFlag(serviceMesh *string) (cli.Flag, *ClusterBootstrapperOptions) {
	envoyFlags, envoyOptions := envoy.LatticeBootstrapperFlags()
	options := &ClusterBootstrapperOptions{}

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

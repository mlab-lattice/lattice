package servicemesh

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

type ClusterBootstrapperOptions struct {
	Envoy *envoy.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewClusterBootstrapper(options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}

func ClusterBoostrapperFlag(serviceMesh *string) (command.Flag, *ClusterBootstrapperOptions) {
	envoyFlags, envoyOptions := envoy.ClusterBootstrapperFlags()
	options := &ClusterBootstrapperOptions{}

	flag := &command.DelayedEmbeddedFlag{
		Name:     "service-mesh-var",
		Required: true,
		Usage:    "configuration for the service mesh cluster bootstrapper",
		Flags: map[string]command.Flags{
			Envoy: envoyFlags,
		},
		FlagChooser: func() (string, error) {
			if serviceMesh == nil {
				return "", fmt.Errorf("serviceMesh cannot be nil")
			}

			switch *serviceMesh {
			case Envoy:
				options.Envoy = envoyOptions
			default:
				return "", fmt.Errorf("unsupported service mesh %v", *serviceMesh)
			}

			return *serviceMesh, nil
		},
	}

	return flag, options
}

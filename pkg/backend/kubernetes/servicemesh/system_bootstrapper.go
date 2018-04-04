package servicemesh

import (
	"fmt"

	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
)

type SystemBootstrapperOptions struct {
	Envoy *envoy.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewSystemBootstrapper(options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}

func SystemBootstrapperFromFlags(serviceMesh string, serviceMeshVars []string) (systembootstrapper.Interface, error) {
	options, err := ParseSystemBootstrapperFlags(serviceMesh, serviceMeshVars)
	if err != nil {
		return nil, err
	}

	return NewSystemBootstrapper(options)
}

func ParseSystemBootstrapperFlags(serviceMesh string, serviceMeshVars []string) (*SystemBootstrapperOptions, error) {
	var options *SystemBootstrapperOptions

	switch serviceMesh {
	case Envoy:
		envoyOptions := envoy.ParseSystemBootstrapperFlags(serviceMeshVars)

		options = &SystemBootstrapperOptions{
			Envoy: envoyOptions,
		}

	default:
		return nil, fmt.Errorf("unsupported service mesh: %v", serviceMesh)
	}

	return options, nil
}

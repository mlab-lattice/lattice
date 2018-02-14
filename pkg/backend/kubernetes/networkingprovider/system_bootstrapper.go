package networkingprovider

import (
	"fmt"

	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
)

type SystemBootstrapperOptions struct {
	Flannel *flannel.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.Flannel != nil {
		return flannel.NewSystemBootstrapper(options.Flannel), nil
	}

	return nil, fmt.Errorf("must provide networking provider options")
}

func SystemBootstrapperFromFlags(networkingProvider string, networkingProviderVars []string) (systembootstrapper.Interface, error) {
	options, err := ParseSystemBootstrapperFlags(networkingProvider, networkingProviderVars)
	if err != nil {
		return nil, err
	}

	return NewSystemBootstrapper(options)
}

func ParseSystemBootstrapperFlags(serviceMesh string, serviceMeshVars []string) (*SystemBootstrapperOptions, error) {
	var options *SystemBootstrapperOptions

	switch serviceMesh {
	case Flannel:
		envoyOptions, err := flannel.ParseSystemBootstrapperFlags(serviceMeshVars)
		if err != nil {
			return nil, err
		}

		options = &SystemBootstrapperOptions{
			Flannel: envoyOptions,
		}

	case "":
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported networking provider: %v", serviceMesh)
	}

	return options, nil
}

package networkingprovider

import (
	"fmt"

	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/none"
)

type SystemBootstrapperOptions struct {
	Flannel *flannel.SystemBootstrapperOptions
	None    *none.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.Flannel != nil {
		return flannel.NewSystemBootstrapper(options.Flannel), nil
	}

	if options.None != nil {
		return none.NewSystemBootstrapper(options.None), nil
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
		flannelOptions, err := flannel.ParseSystemBootstrapperFlags(serviceMeshVars)
		if err != nil {
			return nil, err
		}

		options = &SystemBootstrapperOptions{
			Flannel: flannelOptions,
		}

	case None:
		noneOptions, err := none.ParseSystemBootstrapperFlags(serviceMeshVars)
		if err != nil {
			return nil, err
		}

		options = &SystemBootstrapperOptions{
			None: noneOptions,
		}

	default:
		return nil, fmt.Errorf("unsupported networking provider: %v", serviceMesh)
	}

	return options, nil
}

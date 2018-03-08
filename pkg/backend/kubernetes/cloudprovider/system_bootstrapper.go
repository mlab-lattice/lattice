package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
)

type SystemBootstrapperOptions struct {
	AWS   *aws.SystemBootstrapperOptions
	Local *local.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewSystemBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewSystemBootstrapper(options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

func SystemBootstrapperFromFlags(cloudProvider string, cloudProviderVars []string) (systembootstrapper.Interface, error) {
	options, err := ParseSystemBootstrapperFlags(cloudProvider, cloudProviderVars)
	if err != nil {
		return nil, err
	}

	return NewSystemBootstrapper(options)
}

func ParseSystemBootstrapperFlags(cloudProvider string, cloudProviderVars []string) (*SystemBootstrapperOptions, error) {
	var options *SystemBootstrapperOptions

	switch cloudProvider {
	case Local:
		options = &SystemBootstrapperOptions{
			Local: local.ParseSystemBootstrapperFlags(cloudProviderVars),
		}

	case AWS:
		options = &SystemBootstrapperOptions{
			AWS: aws.ParseSystemBootstrapperFlags(cloudProviderVars),
		}

	default:
		return nil, fmt.Errorf("unsupported cloud provider: %v", cloudProvider)
	}

	return options, nil
}

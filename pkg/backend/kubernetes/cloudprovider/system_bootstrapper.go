package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
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

func SystemBoostrapperFlag(cloudProvider *string) (cli.Flag, *SystemBootstrapperOptions) {
	awsFlags, awsOptions := aws.SystemBootstrapperFlags()
	localFlags, localOptions := local.SystemBootstrapperFlags()
	options := &SystemBootstrapperOptions{
		AWS:   awsOptions,
		Local: localOptions,
	}

	flag := &cli.DelayedEmbeddedFlag{
		Name:     "cloud-provider-system-bootstrapper-var",
		Required: true,
		Flags: map[string]cli.Flags{
			AWS:   awsFlags,
			Local: localFlags,
		},
		FlagChooser: func() (string, error) {
			if cloudProvider == nil {
				return "", fmt.Errorf("cloud provider cannot be nil")
			}

			switch *cloudProvider {
			case Local, AWS:
				return *cloudProvider, nil
			default:
				return "", fmt.Errorf("unsupported cloud provider %v", *cloudProvider)
			}
		},
	}

	return flag, options
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

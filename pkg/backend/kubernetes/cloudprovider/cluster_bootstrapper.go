package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type ClusterBootstrapperOptions struct {
	AWS   *aws.ClusterBootstrapperOptions
	Local *local.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(clusterID types.LatticeID, options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewClusterBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewClusterBootstrapper(clusterID, options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

func ClusterBoostrapperFlag(cloudProvider *string) (command.Flag, *ClusterBootstrapperOptions) {
	awsFlags, awsOptions := aws.ClusterBootstrapperFlags()
	localFlags, localOptions := local.ClusterBootstrapperFlags()
	options := &ClusterBootstrapperOptions{
		AWS:   awsOptions,
		Local: localOptions,
	}

	flag := &command.DelayedEmbeddedFlag{
		Name:     "cloud-provider-cluster-bootstrapper-var",
		Required: true,
		Usage:    "configuration for the cloud provider cluster bootstrapper",
		Flags: map[string]command.Flags{
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

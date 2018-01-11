package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

type ClusterBootstrapperOptions struct {
	AWS   *aws.ClusterBootstrapperOptions
	Local *local.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewClusterBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewClusterBootstrapper(options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

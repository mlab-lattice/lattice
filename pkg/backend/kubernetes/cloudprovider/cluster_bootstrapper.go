package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/types"
)

type ClusterBootstrapperOptions struct {
	AWS   *aws.ClusterBootstrapperOptions
	Local *local.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(ClusterID types.ClusterID, options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewClusterBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewClusterBootstrapper(ClusterID, options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

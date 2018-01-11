package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"
)

type ClusterProvisionerOptions struct {
	AWS   *aws.ClusterProvisionerOptions
	Local *local.ClusterProvisionerOptions
}

func NewClusterProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir string, options *ClusterProvisionerOptions) (provisioner.Interface, error) {
	if options.AWS != nil {
		return aws.NewClusterProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir, options.AWS), nil
	}

	if options.Local != nil {
		return local.NewClusterProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir, options.Local)
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

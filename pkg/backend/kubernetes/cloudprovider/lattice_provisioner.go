package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/lifecycle/lattice/provisioner"
)

type LatticeProvisionerOptions struct {
	AWS   *aws.LatticeProvisionerOptions
	Local *local.LatticeProvisionerOptions
}

func NewLatticeProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir string, options *LatticeProvisionerOptions) (provisioner.Interface, error) {
	if options.AWS != nil {
		return aws.NewLatticeProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir, options.AWS), nil
	}

	if options.Local != nil {
		return local.NewLatticeProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir, options.Local)
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

package noop

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

func NewBootstrapper() *DefaultBootstrapper {
	return &DefaultBootstrapper{}
}

type DefaultBootstrapper struct {
}

func (b *DefaultBootstrapper) BootstrapClusterResources(resources *bootstrapper.ClusterResources) {
}

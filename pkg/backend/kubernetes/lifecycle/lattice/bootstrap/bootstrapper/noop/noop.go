package noop

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
)

func NewBootstrapper() *DefaultBootstrapper {
	return &DefaultBootstrapper{}
}

type DefaultBootstrapper struct {
}

func (b *DefaultBootstrapper) BootstrapClusterResources(resources *bootstrapper.Resources) {
}

package noop

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
)

func NewBootstrapper() *DefaultBootstrapper {
	return &DefaultBootstrapper{}
}

type DefaultBootstrapper struct {
}

func (b *DefaultBootstrapper) BootstrapSystemResources(resources *bootstrapper.SystemResources) {
}

package aws

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *noop.DefaultBootstrapper {
	return noop.NewBootstrapper()
}
